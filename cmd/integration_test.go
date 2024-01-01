/*
Using "example" snippets to double as an integration test suite.

Calling the compiled bin embedded in this repo with exec and capturing output
to test actual cli functionality
*/

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/benorgil/exectester/configs"
)

var exe_err = "This test depends on the compiled exe at the" +
	" location defined in the env var: " +
	configs.TestArgExePath

// Send to stdout stream
// (Not runnable in playground)
func Example_sendStdout() {
	// Find the exe
	testArgExePath, present := os.LookupEnv(configs.TestArgExePath)
	if !present {
		println(exe_err)
		os.Exit(1)
	}
	out, err := exec.Command(testArgExePath, "--stdout=sent_to_stdout").CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	l := configs.OutputFormat{}
	json.Unmarshal([]byte(string(out)), &l)
	fmt.Println(l.Msg)
	// Output:
	// sent_to_stdout
}

// Send to stderr stream
// (Not runnable in playground)
func Example_sendStderr() {
	// Find the exe
	testArgExePath, present := os.LookupEnv(configs.TestArgExePath)
	if !present {
		println(exe_err)
		os.Exit(1)
	}

	out, err := exec.Command(testArgExePath, "--stderr=sent_to_stderr").CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	l := configs.OutputFormat{}
	json.Unmarshal([]byte(string(out)), &l)
	fmt.Println(l.Msg)
	// Output:
	// sent_to_stderr
}
