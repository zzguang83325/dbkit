package dbkit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// 结构体反射缓存的存储库名称（内部使用，不对外暴露）
const structCacheRepository = "__dbkit_struct_cache__"

// structFieldInfo 存储单个字段的缓存信息
type structFieldInfo struct {
	fieldIndex int          // 字段索引
	columnName string       // 列名（从 tag 解析）
	fieldType  reflect.Type // 字段类型
	fieldKind  reflect.Kind // 字段种类
	canSet     bool         // 是否可设置（可导出）
}

// structCacheInfo 存储整个结构体的缓存信息
type structCacheInfo struct {
	fields []structFieldInfo // 字段信息列表
}

// getStructCacheInfo 获取或创建结构体的缓存信息
// 使用 localCache 缓存结构体的反射信息，避免重复解析
func getStructCacheInfo(structType reflect.Type) *structCacheInfo {
	// 使用 Type 的字符串表示作为缓存键
	// 这样可以自动处理多数据库同名表的问题（不同包的同名结构体有不同的 Type）
	cacheKey := structType.String()

	// 尝试从本地缓存获取
	if cached, ok := LocalCacheGet(structCacheRepository, cacheKey); ok {
		return cached.(*structCacheInfo)
	}

	// 缓存未命中，解析结构体字段信息
	info := &structCacheInfo{
		fields: make([]structFieldInfo, 0, structType.NumField()),
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// 解析列名（只解析一次，后续从缓存读取）
		colName := field.Tag.Get("column")
		if colName == "" {
			colName = field.Tag.Get("db")
		}
		if colName == "" {
			colName = field.Tag.Get("json")
		}
		if colName == "-" {
			continue
		}
		if colName == "" {
			colName = strings.ToLower(field.Name)
		}

		// 处理逗号分隔的 tag（如 json:"id,omitempty"）
		if idx := strings.Index(colName, ","); idx != -1 {
			colName = colName[:idx]
		}

		if colName == "-" {
			continue
		}

		// 存储字段信息
		info.fields = append(info.fields, structFieldInfo{
			fieldIndex: i,
			columnName: colName,
			fieldType:  field.Type,
			fieldKind:  field.Type.Kind(),
			canSet:     field.IsExported(), // Go 1.17+ 使用 IsExported 判断是否可导出
		})
	}

	// 存入本地缓存（永不过期，因为结构体定义在运行时不会改变）
	LocalCacheSet(structCacheRepository, cacheKey, info, 0)

	return info
}

// ToStruct converts a single Record to a struct.
// dest must be a pointer to a struct.
func ToStruct(r *Record, dest interface{}) error {
	if r == nil {
		return fmt.Errorf("dbkit: record is nil")
	}
	if dest == nil {
		return fmt.Errorf("dbkit: dest is nil")
	}

	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || (val.Elem().Kind() != reflect.Slice && val.Elem().Kind() != reflect.Struct) {
		return fmt.Errorf("dbkit: dest must be a pointer to a struct")
	}

	return setStructFromRecord(val.Elem(), r)
}

// FromStruct populates a Record from a struct.
func FromStruct(src interface{}, r *Record) error {
	if src == nil {
		return fmt.Errorf("dbkit: src is nil")
	}
	if r == nil {
		return fmt.Errorf("dbkit: record is nil")
	}

	val := reflect.ValueOf(src)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("dbkit: src must be a struct or pointer to struct")
	}

	return setRecordFromStruct(r, val)
}

// ToRecord converts a struct to a new Record.
func ToRecord(src interface{}) *Record {
	r := NewRecord()
	_ = FromStruct(src, r)
	return r
}

func setStructFromRecord(structVal reflect.Value, r *Record) error {
	structType := structVal.Type()

	// 获取缓存的结构体信息（首次调用会解析并缓存，后续直接使用缓存）
	cacheInfo := getStructCacheInfo(structType)

	// 使用缓存的字段信息，避免重复反射解析
	for _, fieldInfo := range cacheInfo.fields {
		fieldVal := structVal.Field(fieldInfo.fieldIndex)

		// 使用缓存的 canSet 信息
		if !fieldInfo.canSet {
			continue
		}

		val := r.Get(fieldInfo.columnName)
		if val == nil {
			continue
		}

		if err := setFieldValue(fieldVal, val); err != nil {
			// 获取字段名用于错误信息
			fieldName := structType.Field(fieldInfo.fieldIndex).Name
			return fmt.Errorf("field %s: %v", fieldName, err)
		}
	}
	return nil
}

func setRecordFromStruct(r *Record, structVal reflect.Value) error {
	structType := structVal.Type()

	// 获取缓存的结构体信息（首次调用会解析并缓存，后续直接使用缓存）
	cacheInfo := getStructCacheInfo(structType)

	// 使用缓存的字段信息，避免重复反射解析
	for _, fieldInfo := range cacheInfo.fields {
		fieldVal := structVal.Field(fieldInfo.fieldIndex)

		if !fieldVal.CanInterface() {
			continue
		}

		// 跳过 nil 指针字段，这样它们就不会被包含在 Record 中
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			continue
		}

		r.Set(fieldInfo.columnName, fieldVal.Interface())
	}
	return nil
}

func setFieldValue(field reflect.Value, value interface{}) error {
	v := reflect.ValueOf(value)

	// Handle pointer target
	if field.Kind() == reflect.Ptr {
		if v.Kind() == reflect.Ptr && v.IsNil() {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		// Create new instance of pointer type if needed
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldValue(field.Elem(), value)
	}

	// Unpack pointer value
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		v = v.Elem()
		value = v.Interface()
	}

	// Try direct set
	if v.Type().AssignableTo(field.Type()) {
		field.Set(v)
		return nil
	}

	// Advanced conversions
	switch field.Kind() {
	case reflect.String:
		field.SetString(fmt.Sprint(value))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := toInt64(value)
		if err != nil {
			return err
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := toUint64(value)
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := toFloat64(value)
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Bool:
		val, err := toBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)
	default:
		// Special handling for time.Time
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t, err := toTime(value)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(t))
			return nil
		}
		return fmt.Errorf("cannot convert %T to %s", value, field.Type())
	}

	return nil
}

func toInt64(v interface{}) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	case []byte:
		return strconv.ParseInt(string(val), 10, 64)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("cannot convert %T to int64", v)
}

func toUint64(v interface{}) (uint64, error) {
	switch val := v.(type) {
	case uint:
		return uint64(val), nil
	case uint8:
		return uint64(val), nil
	case uint16:
		return uint64(val), nil
	case uint32:
		return uint64(val), nil
	case uint64:
		return val, nil
	case int:
		return uint64(val), nil
	case int8:
		return uint64(val), nil
	case int16:
		return uint64(val), nil
	case int32:
		return uint64(val), nil
	case int64:
		return uint64(val), nil
	case float32:
		return uint64(val), nil
	case float64:
		return uint64(val), nil
	case string:
		return strconv.ParseUint(val, 10, 64)
	case []byte:
		return strconv.ParseUint(string(val), 10, 64)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("cannot convert %T to uint64", v)
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	case []byte:
		return strconv.ParseFloat(string(val), 64)
	}
	return 0, fmt.Errorf("cannot convert %T to float64", v)
}

func toBool(v interface{}) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case int, int8, int16, int32, int64:
		i, _ := toInt64(v)
		return i != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		i, _ := toUint64(v)
		return i != 0, nil
	case string:
		return strconv.ParseBool(val)
	case []byte:
		return strconv.ParseBool(string(val))
	}
	return false, fmt.Errorf("cannot convert %T to bool", v)
}

func toTime(v interface{}) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case string:
		// Try common formats
		formats := []string{
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, f := range formats {
			if t, err := time.Parse(f, val); err == nil {
				return t, nil
			}
		}
	case int64:
		return time.Unix(val, 0), nil
	}
	return time.Time{}, fmt.Errorf("cannot convert %T to time.Time", v)
}

// ToStructs converts a slice of Records to a slice of structs or struct pointers.
// dest must be a pointer to a slice.
func ToStructs(records []Record, dest interface{}) error {
	if dest == nil {
		return fmt.Errorf("dbkit: dest cannot be nil")
	}

	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dbkit: dest must be a pointer to a slice")
	}

	sliceVal := val.Elem()
	// Clear the slice before filling it
	sliceVal.Set(reflect.MakeSlice(sliceVal.Type(), 0, len(records)))
	elemType := sliceVal.Type().Elem()

	// Handle both slice of structs and slice of struct pointers
	isPtr := elemType.Kind() == reflect.Ptr
	var baseType reflect.Type
	if isPtr {
		baseType = elemType.Elem()
	} else {
		baseType = elemType
	}

	for i := range records {
		newElem := reflect.New(baseType)
		if err := ToStruct(&records[i], newElem.Interface()); err != nil {
			return err
		}

		if isPtr {
			sliceVal.Set(reflect.Append(sliceVal, newElem))
		} else {
			sliceVal.Set(reflect.Append(sliceVal, newElem.Elem()))
		}
	}

	return nil
}
