/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package prepare

import (
	"fmt"
	"reflect"
	"strings"
)

// tagName is the name of the tag used to specify the JSON field name.
const tagName = "json"

// Preparer is used for recursive validation of configuration structures.
type Preparer interface {
	// Prepare validates something.
	Prepare() error
}

// Prepare calls Prepare() method on the object and its fields recursively.
func Prepare(src any) error {
	if src == nil {
		return nil
	}

	return traverse(reflect.ValueOf(src), false)
}

// traverse recursively validates the configuration structure.
func traverse(v reflect.Value, preparedByParent bool) error {
	if !v.IsValid() {
		return nil
	}

	//nolint:exhaustive // reflect.Kind handling is intentionally grouped, default covers all other kinds.
	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		return traversePointerOrInterface(v)
	case reflect.Struct:
		return traverseStruct(v, preparedByParent)
	default:
		return tryPrepareValue(v)
	}
}

// traversePointerOrInterface traverses a pointer or an interface.
func traversePointerOrInterface(v reflect.Value) error {
	if v.IsNil() {
		return nil
	}

	err := tryPrepareValue(v)
	if err != nil {
		return err
	}

	return traverse(v.Elem(), true)
}

// traverseStruct traverses a struct.
func traverseStruct(v reflect.Value, preparedByParent bool) error {
	if !preparedByParent {
		err := tryPrepareValue(v)
		if err != nil {
			return err
		}
	}

	structType := v.Type()

	numFields := v.NumField()
	for j := range numFields {
		err := traverse(v.Field(j), false)
		if err != nil {
			field := structType.Field(j)

			return fmt.Errorf("invalid section '%s': %w", fieldTagOrName(&field), err)
		}
	}

	return nil
}

// fieldTagOrName returns the tag value or the field name.
func fieldTagOrName(field *reflect.StructField) string {
	tagValue := field.Tag.Get(tagName)
	if idx := strings.Index(tagValue, ","); idx >= 0 {
		tagValue = tagValue[:idx]
	}

	if tagValue == "" {
		return field.Name
	}

	return tagValue
}

// tryPrepareValue attempts to prepare the value by checking
// if it implements the Preparer interface and calling its Prepare method.
func tryPrepareValue(v reflect.Value) error {
	if !v.IsValid() {
		return nil
	}

	if v.CanInterface() {
		if preparer, ok := v.Interface().(Preparer); ok {
			return preparer.Prepare()
		}
	}

	if v.CanAddr() && v.Addr().CanInterface() {
		if preparer, ok := v.Addr().Interface().(Preparer); ok {
			return preparer.Prepare()
		}
	}

	return nil
}
