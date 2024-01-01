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
This Cobra flag returns an [cmd.OutputFormatter] instance which is used by the app for
sending output to the console.

Slog already provided structured output. Using it to format the cli's
console output.

The "outputFormatterEnum" defined here behaves like an enum. If the user enters a
value for the flag not defined in the enum they immediately get back a good error.
*/
type outputFormatterEnum string

// An enum of allowed values for this flag
const (
	outputFormatterEnumHuman      outputFormatterEnum = "human_readable"
	outputFormatterEnumStructured outputFormatterEnum = "structured"
)

// Used by FlagSet.VarP() method
// It's used both by fmt.Print and by Cobra in help text
func (e *outputFormatterEnum) String() string {
	return string(*e)
}

// Used by FlagSet.VarP() method
// Needs to have pointer receiver so it doesn't change the value of a copy
func (e *outputFormatterEnum) Set(v string) error {
	if slices.Contains(outputFormatterEnumValues, v) {
		*e = outputFormatterEnum(v)
		return nil
	} else {
		return fmt.Errorf(outputFormatterEnumValuesErrMsg)
	}
}

// Used by FlagSet.VarP() method
// Only used in help text
func (e *outputFormatterEnum) Type() string {
	return "outputFormatterEnum"
}

// Defining flags error message and redefining allowed values as slice
// to be able to loop over them dynamically
var (
	outputFormatterEnumValues        = []string{"human_readable", "structured"}
	outputFormatterEnumValuesStr     = strings.Join(outputFormatterEnumValues, ", ")
	outputFormatterEnumValuesInfoMsg = fmt.Sprintf(
		"The output format to console. Allowed: '%v'", outputFormatterEnumValuesStr)
	outputFormatterEnumValuesErrMsg = fmt.Sprintf(
		"must be one of: '%v'", outputFormatterEnumValuesStr)
)

// This func if called to process the custom type flags value before returning it
// This flag returns a custom type that includes the loggers to use for console output
func outputFormatterEnumHookFunc(f reflect.Type, t reflect.Type, flagValue interface{}) (interface{}, error) {
	if _, ok := flagValue.(string); ok {
		of := new(OutputFormatter)
		return of.getLogger(flagValue.(string)), nil
	} else {
		return flagValue, fmt.Errorf("format_output value of '%v' is not a string", flagValue)
	}
}
