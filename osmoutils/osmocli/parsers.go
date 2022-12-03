package osmocli

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func ParseField(v reflect.Value, t reflect.Type, fieldIndex int, arg string) error {
	fVal := v.Field(fieldIndex)
	fType := t.Field(fieldIndex)

	// fmt.Printf("Field %d: %s %s %s\n", fieldIndex, fType.Name, fType.Type, fType.Type.Kind())
	switch fType.Type.Kind() {
	case reflect.Uint64:
		u, err := ParseUint(arg, fType.Name)
		if err != nil {
			return err
		}
		fVal.SetUint(u)
		return nil
	case reflect.Uint:
		u, err := ParseUint(arg, fType.Name)
		if err != nil {
			return err
		}
		fVal.SetUint(u)
		return nil
	case reflect.Int:
		// Handle int type
		// ...
	case reflect.String:
		// Handle string type
		// ...
	case reflect.Struct:
		// Handle struct type
		// ...
	}
	return fmt.Errorf("field type not recognized. Got type %v", fType)
}

func ParseUint(arg string, fieldName string) (uint64, error) {
	v, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %s as uint for field %s: %w", arg, fieldName, err)
	}
	return v, nil
}

func ParseUnixTime(arg string, fieldName string) (time.Time, error) {
	timeUnix, err := strconv.ParseInt(arg, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse %s as unix time for field %s: %w", arg, fieldName, err)
	}
	startTime := time.Unix(timeUnix, 0)
	return startTime, nil
}
