/*
Copyright © 2023 Ben Orgil

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

/*
This Cobra flag is just a string

The "interpolatorEnum" defined here behaves like an enum. If the user enters a
value for the flag not defined in the enum they immediately get back a good error.
*/
type interpolatorEnum string

// An enum of allowed values for this flag
const (
	interpolatorEnumIntCounter interpolatorEnum = "int_counter"
	interpolatorEnumString     interpolatorEnum = "string"
)

// Defining flags error message and redefining allowed values as slice
// to be able to loop over them dynamically
var (
	interpolatorEnumValues        = []string{"int_counter", "string"}
	interpolatorEnumValuesStr     = strings.Join(interpolatorEnumValues, ", ")
	interpolatorEnumValuesInfoMsg = fmt.Sprintf(
		"The interpolator to use on the interpolate_key. Allowed: '%v'", interpolatorEnumValuesStr)
	interpolatorEnumValuesErrMsg = fmt.Sprintf(
		"must be one of: '%v'", interpolatorEnumValuesStr)
)

// Used by FlagSet.VarP() method
// It's used both by fmt.Print and by Cobra in help text
func (e *interpolatorEnum) String() string {
	return string(*e)
}

// Used by FlagSet.VarP() method
// Needs to have pointer receiver so it doesn't change the value of a copy
func (e *interpolatorEnum) Set(v string) error {
	if slices.Contains(interpolatorEnumValues, v) {
		*e = interpolatorEnum(v)
		return nil
	} else {
		return fmt.Errorf(interpolatorEnumValuesErrMsg)
	}
}

// Used by FlagSet.VarP() method
// Only used in help text
func (e *interpolatorEnum) Type() string {
	return "interpolatorEnum"
}

// This func is called to process the custom type flags value before returning it
func interpolatorEnumHookFunc(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if _, ok := data.(string); ok {
		return data, nil
	} else {
		return data, fmt.Errorf("interpolator value of '%v' is not a string", data)
	}
}
