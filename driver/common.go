package driver

import (
	"fmt"
	"reflect"
	"strconv"
)

type DiskStatus string

func (s DiskStatus) Equals(status string) bool {
	return string(s) == status
}

func applyOptions(opts map[string]string, target interface{}) error {
	ptr := reflect.ValueOf(target)
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("apply should receive a pointer")
	}

	str := ptr.Elem()
	if str.Kind() != reflect.Struct {
		return fmt.Errorf("apply should receive a pointer to a struct")
	}

	for name, value := range opts {
		f := str.FieldByName(name)
		if !f.IsValid() {
			return fmt.Errorf("unknown property %q at %q", name, str.Type())
		}

		if err := applyString(value, &f); err != nil {
			return fmt.Errorf("invalid value for property %q at %q: %s",
				name, str.Type(), err.Error(),
			)
		}
	}

	return nil
}

func applyString(value string, field *reflect.Value) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		field.SetInt(value)
	}

	return nil
}
