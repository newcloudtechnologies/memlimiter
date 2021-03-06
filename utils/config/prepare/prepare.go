/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package prepare

import (
	"reflect"

	"github.com/pkg/errors"
)

const (
	tagName        = "json"
	prepareTagName = "prepare"
	optValue       = "optional"
)

// Preparer is used for recursive validation of configuration structures.
type Preparer interface {
	// Prepare validates something.
	Prepare() error
}

// Prepare calls Prepare() method on the object and its fields recursively.
func Prepare(src interface{}) error {
	if src == nil {
		return nil
	}

	v := reflect.ValueOf(src)

	pr, ok := src.(Preparer)
	if ok {
		err := pr.Prepare()
		if err != nil {
			return errors.Wrap(err, "prepare error")
		}
	}

	return traverse(v, true)
}

//nolint:gocognit,gocyclo,exhaustive,cyclop
func traverse(v reflect.Value, parentTraversed bool) (err error) {
	switch v.Kind() {
	case reflect.Interface, reflect.Ptr:
		if !v.IsNil() && v.CanInterface() {
			if err := tryPrepareInterface(v.Interface()); err != nil {
				return err
			}

			if err := traverse(v.Elem(), true); err != nil {
				return err
			}
		}
	case reflect.Struct:
		if !parentTraversed && v.CanInterface() {
			if err := tryPrepareInterface(v.Interface()); err != nil {
				return err
			}
		}

		for j := 0; j < v.NumField(); j++ {
			optTag := v.Type().Field(j).Tag.Get(prepareTagName)
			if optTag == optValue && v.Field(j).IsNil() {
				continue
			}

			err := traverse(v.Field(j), false)
			if err != nil {
				tagValue := v.Type().Field(j).Tag.Get(tagName)

				return errors.Errorf("invalid section '%s': %v", tagValue, err)
			}

			// call Prepare() on children.
			child := v.Field(j)
			if child.CanAddr() {
				if child.Addr().MethodByName("Prepare").Kind() != reflect.Invalid {
					child.Addr().MethodByName("Prepare").Call([]reflect.Value{})
				}
			}
		}
	default:
		if v.CanInterface() {
			return tryPrepareInterface(v.Interface())
		}
	}

	return nil
}

func tryPrepareInterface(v interface{}) (err error) {
	pr, ok := v.(Preparer)
	if ok {
		err = pr.Prepare()
	}

	return
}
