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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/benorgil/exectester/configs"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/suite"
)

/*
Grouping tests in a testify suite

There is a huge limitation though, because testify
doesn't seem to support any parallelism :(((
https://github.com/stretchr/testify/issues/187
*/
type ExecTestSuite struct {
	suite.Suite
}

func TestExecTestSuite(t *testing.T) {
	suite.Run(t, &ExecTestSuite{})
}

// Stores processed output of cobra.command
type CmdResult struct {
	Cmd         cobra.Command
	StdOut      []string
	StdErr      []string
	IsJson      bool
	Pid         int
	StdOutCount int
	StdErrCount int
}

// Spin up a test unix socket against local filesystem.
// While this is a "unit" test it should be expected to run containerized, in a
// controlled environment- even locally. because of this I think its acceptable
// to spin up a socket for some tests.
// Abusing pointers here for testing. From the socket's side we need to assert
// the data was send correctly and also send test data to listening clients
func createTestSocket(ts *ExecTestSuite, socketFile string, mockResponse *string, assertReceived *string) {
	// Start socket and listen in background
	os.Remove(socketFile)
	go func(mockResponse *string) {
		os.Remove(socketFile)
		socket, err := net.Listen("unix", socketFile)
		if err != nil {
			ts.FailNow("Failed to create test unix socket. ", err)
		}
		defer socket.Close()
		defer os.Remove(socketFile)

		// Read data sent to socket
		for {
			conn, err := socket.Accept()
			if err != nil {
				ts.FailNow("Failed to create test unix socket. ", err)
			}

			// Handle the connection in a separate goroutine.
			go func(conn net.Conn) {
				defer conn.Close()

				if *mockResponse != "" {
					// If mockResponse gets set return that text to the client
					_, err = conn.Write([]byte(*mockResponse))
					if err != nil {
						ts.FailNow("Failed to create test unix socket. ", err)
					}

				} else if *assertReceived != "" {
					// If assertReceived set read data from connection and assert msg was received
					buf := make([]byte, 4096)
					n, err := conn.Read(buf)
					if err != nil {
						ts.FailNow("Failed to create test unix socket. ", err)
					}
					response := string(buf[:n])
					ts.Equal(*assertReceived, response)
				}

			}(conn)
		}
	}(mockResponse)
}

// Reads stdout and stderr buffers and attempts to parse as json.
// Returns parsed output, number of lines, bool for if json or not
func readB(b *bytes.Buffer) ([]string, int, bool) {
	var stdOut []string
	scanner := bufio.NewScanner(b)
	line := configs.OutputFormat{}
	isJson := true
	for scanner.Scan() {
		l := scanner.Text()
		err := json.Unmarshal([]byte(l), &line)
		if err != nil {
			// Output was not structured
			stdOut = append(stdOut, l)
			isJson = false
		} else {
			stdOut = append(stdOut, line.Msg)
		}
	}
	return stdOut, len(stdOut), isJson
}

// Wrapper for executing the root cobra.command.
// Redirects cmd's stdout and stderr to buffers and parses their output
func (ts *ExecTestSuite) ExecuteCmd(args []string) (CmdResult, error) {
	o := bytes.NewBufferString("")
	e := bytes.NewBufferString("")

	cmd := *RootCmd(configs.FallbackLogger)
	cmd.SetOut(o)
	cmd.SetErr(e)
	cmd.SetArgs(args)

	r := *new(CmdResult)
	r.Cmd = cmd

	err := cmd.Execute()

	if err != nil {
		return r, err
	} else {
		r.StdOut, r.StdOutCount, r.IsJson = readB(o)
		r.StdErr, r.StdErrCount, r.IsJson = readB(e)
	}

	return r, err
}

func (ts *ExecTestSuite) TestNoArgs() {
	_, err := ts.ExecuteCmd([]string{""})
	ts.IsType(&paramSetValidationError{}, err)
}

func (ts *ExecTestSuite) TestStdoutStderr() {
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o", "--stderr=e"})
	ts.Equal(true, cmd.IsJson)
	ts.Equal("o", cmd.StdOut[0])
	ts.Equal("e", cmd.StdErr[0])
}

func (ts *ExecTestSuite) TestUnixSocket() {
	// Open the test unix socket
	socketFile := "/tmp/ExecTestSuite_TestUnixSocket.sock"
	socketArg := fmt.Sprintf("--socket=%s", socketFile)
	mockResponse := ""
	assertReceived := ""
	createTestSocket(ts, socketFile, &mockResponse, &assertReceived)

	// Send data to test socket
	// assertReceived pointer used by TestSocket to assert msg received
	assertReceived = "test_msg_01\n"
	_, err := ts.ExecuteCmd([]string{socketArg, "--socket_send=test_msg_01"})
	ts.NoError(err)
	assertReceived = "test_msg_02\n"
	_, err = ts.ExecuteCmd([]string{socketArg, "--socket_send=test_msg_02"})
	ts.NoError(err)

	// Read data from the socket's buffer and close client when we get the exit_message
	// mockResponse pointer used by TestSocket to send response to client
	mockResponse = "test_msg_exit"
	_, err = ts.ExecuteCmd([]string{socketArg, "--read_socket", "--socket_exit_msg=test_msg_exit"})
	ts.NoError(err)
}

func (ts *ExecTestSuite) TestRepeat() {
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o", "--stderr=e", "--repeat=2"})
	ts.Equal(2, cmd.StdErrCount)
	ts.Equal(2, cmd.StdOutCount)
}

func (ts *ExecTestSuite) TestDefaultInterpolator() {
	// Default is int_counter
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o__I__o", "--stderr=e__I__e", "--repeat=3"})
	ts.Equal("o2o", cmd.StdOut[2])
	ts.Equal("e2e", cmd.StdErr[2])
}

func (ts *ExecTestSuite) TestIntInterpolatorStartVal() {
	// Default is int_counter
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o__I__o", "--stderr=e__I__e", "--repeat=2", "--interpolate_val=5"})
	ts.Equal("o6o", cmd.StdOut[1])
	ts.Equal("e6e", cmd.StdErr[1])
}

func (ts *ExecTestSuite) TestStringInterpolator() {
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o__I__o", "--stderr=e__I__e", "--repeat=2", "--interpolator=string", "--interpolate_val=zzz"})
	ts.Equal("ozzzo", cmd.StdOut[1])
	ts.Equal("ezzze", cmd.StdErr[1])
}

func (ts *ExecTestSuite) TestTimeout() {
	now := time.Now().Round(time.Second)
	expectedTimeout := (time.Duration(3) * time.Second).Round(time.Second)
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o", "--repeat_forever", "--timeout=3"})
	ts.Equal("o", cmd.StdOut[0])
	ts.Equal(expectedTimeout, time.Since(now).Round(time.Second))
}

func (ts *ExecTestSuite) TestSigTermTimeout() {
	// Get a handling to current process to send ourself signals
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		panic(err)
	}

	sigtermTimeoutArg := 3

	// Start command in background setting sigterm_timeout
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ts.ExecuteCmd([]string{"--stdout=o", "--repeat_forever",
			fmt.Sprintf("--sigterm_timeout=%s", strconv.Itoa(sigtermTimeoutArg))})
	}()
	// Just to make sure the routine started
	time.Sleep(time.Millisecond * 500)

	// Start timer and send ourselves a sigterm
	now := time.Now().Round(time.Second)
	expectedTimeout := (time.Duration(sigtermTimeoutArg) * time.Second).Round(time.Second)
	p.Signal(syscall.SIGTERM)

	// Right after routine returns confirm sigterm_timeout was honored
	wg.Wait()
	ts.Equal(expectedTimeout, time.Since(now).Round(time.Second))
}

// This test highlights how --timeout is somewhat best effort
// If timeout is not divisible by repeat_interval it could timeout a second
// or two off
func (ts *ExecTestSuite) TestRepeatInterval() {
	timeoutArg := 2 // Must be Divisible by repeatIntervalArg for test to pass
	repeatIntervalArg := 1

	expectedTimeout := (time.Duration(timeoutArg) * time.Second).Round(time.Second)
	now := time.Now().Round(time.Second)
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o", "--repeat_forever",
		fmt.Sprintf("--repeat_interval=%s", strconv.Itoa(repeatIntervalArg)),
		fmt.Sprintf("--timeout=%s", strconv.Itoa(timeoutArg))})
	ts.Equal("o", cmd.StdOut[0])
	ts.Equal(expectedTimeout, time.Since(now).Round(time.Second))
}

func (ts *ExecTestSuite) TestOutputFormat() {
	// Should default to structured output
	cmd, _ := ts.ExecuteCmd([]string{"--stdout=o", "--stderr=e"})
	ts.Equal(true, cmd.IsJson)
	// Support switching to human readable
	cmd, _ = ts.ExecuteCmd([]string{"--stdout=o", "--stderr=e", "--output_format=human_readable"})
	ts.Equal(false, cmd.IsJson)
}

func SetExitCodeFlag() {
	ts := new(ExecTestSuite)
	ts.ExecuteCmd([]string{"--exitcode=123"})
}

// // TODO this decided to stop working for some reason
// // TODO IM HERE try running it through the dagger pipeline maybe it works there
// // For capturing exit code, call go test itself as a subprocess that then calls SetExitCode()
// // which calls the Cobra command passing it the --exitcode flag... whew!
// // https://stackoverflow.com/questions/26225513/how-to-test-os-exit-scenarios-in-go
// // https://go.dev/talks/2014/testing.slide
// func (ts *ExecTestSuite) TestExitCode() {
// 	if os.Getenv("FORCE_EXIT") == "1" {
// 		SetExitCodeFlag()
// 		return
// 	}
//
// 	// Im not entirely sure why but previously passing "" worked but pointing to the test itself did not
// 	// Now I can't get it to work at all :(((
// 	// gotestArgs := "-run ^TestExecTestSuite$ -testify.m ^(TestExitCode)"
// 	// gotestArgs := "'-test.run' '^TestExecTestSuite/TestExitCode$'"
// 	gotestArgs := ""
// 	cmd := exec.Command(os.Args[0], gotestArgs)
//
// 	cmd.Env = append(os.Environ(), "FORCE_EXIT=1")
//
// 	fmt.Printf("%+v\n", cmd.Args)
// 	/////////////////////////////
// 	cmd.Stdout = os.Stdout
// 	cmd.Run()
//
// 	// out, err := cmd.CombinedOutput()
// 	// if err != nil {
// 	// 	exitCode := err.(*exec.ExitError).ExitCode()
// 	// 	ts.Equal(123, exitCode)
// 	// 	if exitCode != 123 {
// 	// 		ts.T().Log("TestExitCode() failed. Output from exec: " + string(out))
// 	// 	}
// 	// } else {
// 	// 	ts.T().Log("TestExitCode() failed. Output from exec: " + string(out))
// 	// }
// }
