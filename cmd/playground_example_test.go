/*
Experimenting with a runnable "whole file" example.

Go test runs this locally without issue. This _should_
also run in playground, but when spinning up pkgsite locally
I'm getting a weird import error when hitting "Run".
*/

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/benorgil/exectester/configs"
)

// Calling the root Cobra command func wrapper directly
func Example_playground() {
	oLogLine := configs.OutputFormat{}
	eLogLine := configs.OutputFormat{}
	o := bytes.NewBufferString("")
	e := bytes.NewBufferString("")

	cmd := RootCmd(configs.FallbackLogger)
	cmd.SetOut(o)
	cmd.SetErr(e)
	cmd.SetArgs([]string{"--stdout=output_to_stdout", "--stderr=output_to_stderr"})

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal([]byte(o.String()), &oLogLine)
	json.Unmarshal([]byte(e.String()), &eLogLine)
	fmt.Println(oLogLine.Msg)
	fmt.Println(eLogLine.Msg)
	// Output:
	// output_to_stdout
	// output_to_stderr
}
