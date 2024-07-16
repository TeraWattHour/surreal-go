package surreal

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// autoScan parses the raw data into the destination, raw data must always represent a JSON array.
func autoScan(raw []byte, destination any) error {
	if reflect.TypeOf(destination).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer to destination")
	}

	// query returned an array
	if len(raw) > 0 && raw[0] == '[' {
		switch reflect.Indirect(reflect.ValueOf(destination)).Kind() {
		case reflect.Slice, reflect.Array:
			if err := json.Unmarshal(raw, destination); err != nil {
				return fmt.Errorf("failed to decode result: %s", err)
			}
		default:
			sliceType := reflect.SliceOf(reflect.Indirect(reflect.ValueOf(destination)).Type())
			value := reflect.New(reflect.MakeSlice(sliceType, 1, 1).Type()).Elem()

			if err := json.Unmarshal(raw, value.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to decode result: %s", err)
			}

			if value.Len() >= 1 {
				reflect.Indirect(reflect.ValueOf(destination)).Set(value.Index(0))
			}
		}
	} else {
		switch reflect.Indirect(reflect.ValueOf(destination)).Kind() {
		case reflect.Slice, reflect.Array:
			sliceType := reflect.Indirect(reflect.ValueOf(destination)).Type().Elem()
			instance := reflect.New(sliceType).Elem()

			if err := json.Unmarshal(raw, instance.Addr().Interface()); err != nil {
				return fmt.Errorf("failed to decode result: %s", err)
			}

			reflect.Indirect(reflect.ValueOf(destination)).Set(reflect.Append(reflect.Indirect(reflect.ValueOf(destination)), instance))
		default:
			if err := json.Unmarshal(raw, destination); err != nil {
				return fmt.Errorf("failed to decode result: %s", err)
			}
		}
	}

	return nil
}
