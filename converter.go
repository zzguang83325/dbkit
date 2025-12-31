package dbkit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Slice && val.Elem().Kind() != reflect.Struct {
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
	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Get column name from tags
		colName := field.Tag.Get("column")
		if colName == "" {
			colName = field.Tag.Get("db")
		}
		if colName == "" {
			colName = field.Tag.Get("json")
		}
		if colName == "" || colName == "-" {
			colName = strings.ToLower(field.Name)
		}

		// Handle comma in tags (like json:"id,omitempty")
		if idx := strings.Index(colName, ","); idx != -1 {
			colName = colName[:idx]
		}

		val := r.Get(colName)
		if val == nil {
			continue
		}

		if err := setFieldValue(fieldVal, val); err != nil {
			return fmt.Errorf("field %s: %v", field.Name, err)
		}
	}
	return nil
}

func setRecordFromStruct(r *Record, structVal reflect.Value) error {
	structType := structVal.Type()
	for i := 0; i < structVal.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)
		if !fieldVal.CanInterface() {
			continue
		}

		// Get column name from tags
		colName := field.Tag.Get("column")
		if colName == "" {
			colName = field.Tag.Get("db")
		}
		if colName == "" {
			colName = field.Tag.Get("json")
		}
		if colName == "" || colName == "-" {
			colName = strings.ToLower(field.Name)
		}

		if idx := strings.Index(colName, ","); idx != -1 {
			colName = colName[:idx]
		}

		r.Set(colName, fieldVal.Interface())
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
