package data

import (
	"reflect"
	"strconv"
)

// Tries to read and convert an interface{} value (eg. the ones in log.Properties) to an int value - returns an int (or 0 on failure) and a bool (ok)
func ToInt(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int8, int16, int32, int64:
		return int(reflect.ValueOf(v).Int()), true
	case uint, uint8, uint16, uint32, uint64:
		return int(reflect.ValueOf(v).Uint()), true
	case float32, float64:
		return int(reflect.ValueOf(v).Float()), true
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, true
		}
	}
	return 0, false
}
