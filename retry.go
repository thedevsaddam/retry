// Copyright @2018 Saddam Hossain.  All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package retry is a simple and easy retry mechanism package for Go
package retry

import (
	"errors"
	"reflect"
	"time"
)

// DoFunc try to execute the function, it only expect that the function will return an error only
func DoFunc(attempt int, sleep time.Duration, fn func() error) error {

	if err := fn(); err != nil {
		attempt--
		if attempt > 0 {
			time.Sleep(sleep)
			return DoFunc(attempt, sleep, fn)
		}
		return err
	}

	return nil
}

// Do try to execute the function by its value, function can take variadic arguments and return multiple return.
// You must put error as the last return value so that DoFunc can take decision that the call failed or not
func Do(attempt int, sleep time.Duration, fn interface{}, args ...interface{}) ([]reflect.Value, error) {

	// if the fn is not a function then return error
	vfn := reflect.ValueOf(fn)
	if vfn.Type().Kind() != reflect.Func {
		return nil, errors.New("retry: fn is not a function")
	}

	// if the functions in not variadic then return the argument missmatch error
	if !vfn.Type().IsVariadic() {
		if vfn.Type().NumIn() != len(args) {
			return nil, errors.New("retry: fn argument mismatch")
		}
	}

	// if the function does not return anything, we can't catch if an error occur or not
	if vfn.Type().NumOut() <= 0 {
		return nil, errors.New("retry: fn return's can not empty, at least an error")
	}

	// build args for reflect value Call
	in := make([]reflect.Value, len(args))
	for k, a := range args {
		in[k] = reflect.ValueOf(a)
	}

	// call the fn with arguments
	out := vfn.Call(in)

	// if the last return value is type of error and it is not nil then attempt to retry
	if err, ok := out[len(out)-1:][0].Interface().(error); ok {
		if err != nil {
			attempt--
			if attempt > 0 {
				time.Sleep(sleep)
				return Do(attempt, sleep, fn, args...)
			}
			return out, err
		}
	}

	return out, nil
}
