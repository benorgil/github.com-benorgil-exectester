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
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const SocketDialTimeout int = 2

// Custom type flags
var interpolatorEnumVal interpolatorEnum
var outputFormatterEnumVal OutputFormatter

// Holds all the viper args that were retrieved and parsed
type viperArgs struct {
	outputFormatter        OutputFormatter
	interpolatorEnumVal    interpolatorEnum
	outputFormatterEnumVal OutputFormatter
	stdout                 string
	stderr                 string
	socket                 string
	socketSend             string
	readSocket             bool
	socketExitMsg          string
	exitcode               int
	repeat                 int
	repeatInterval         time.Duration
	repeatForever          bool
	timeout                int
	sigtermTimeout         int
	interpolateKey         string
	interpolator           string
	interpolateVal         string
}

// Environment variable support is totally broken (sans hard coded config file)
// Just doing it manually
// https://stackoverflow.com/questions/67608629/viper-automatic-environment-does-not-read-from-environment
func bindEnvToFlags() {
	for k := range viper.AllSettings() {
		envName := viper.GetEnvPrefix() + strings.ToUpper(k)
		viper.BindEnv(k, envName)
	}
}

// Error wrapper for validation errors. Tests check the returned Error
// type to know if there were validation errors
type paramSetValidationError struct {
	message string
}

func (e *paramSetValidationError) Error() string {
	return fmt.Sprintf("(paramSetValidationError) %v", e.message)
}

// Coerce flags into bools for validation checking.
// Feels like I shouldn't need to do this myself.
func paramSet(m map[string]any, k string) bool {
	if v, ok := m[k]; ok {
		if v == "" {
			return false
		} else if v == 0 {
			return false
		} else {
			return true
		}
	} else {
		return false
	}
}

// Validation of parameter grouping is defined in this function
// There might be support for doing this with flag groups
// (https://github.com/spf13/cobra/issues/1936) but it seems like more
// trouble then it was worth.
func validateParamSets(cmd *cobra.Command) error {
	m := viper.AllSettings()
	switch {
	case !paramSet(m, "stderr") && !paramSet(m, "stdout") && !paramSet(m, "socket") && !paramSet(m, "exitcode"):
		return &paramSetValidationError{"you must specify at least stderr | stdout | socket | exitcode"}
	case paramSet(m, "socket") && (!paramSet(m, "socket_send") && !paramSet(m, "read_socket")):
		return &paramSetValidationError{"if socket specified must also set socket_send and or read_socket"}
	default:
		return nil
	}
}

// Reading user supplied input string and if interpolator token is found
// replaces it per user interpolate input
func interpolate(interpolateKeyArg string, interpolator string, outputText string, counter int, interpolateVal string) (string, error) {
	var interpolated string
	var errReturn error

	// If no interpolator token is not found in outputText do nothing
	if !strings.Contains(outputText, interpolateKeyArg) {
		return outputText, nil
	}

	switch i := interpolator; i {
	case "int_counter":
		var startCounter int
		if v, err := strconv.Atoi(interpolateVal); err == nil {
			startCounter = v + counter
			interpolated = strings.ReplaceAll(outputText, interpolateKeyArg, strconv.Itoa(startCounter))
		} else {
			errReturn = fmt.Errorf(fmt.Sprintf("'interpolate_val' of '%v' cannot be converted to a number! Defaulting to '0'",
				interpolateVal))
			startCounter = v + counter
			interpolated = strings.ReplaceAll(outputText, interpolateKeyArg, strconv.Itoa(startCounter))
		}
	case "string":
		interpolated = strings.ReplaceAll(outputText, interpolateKeyArg, interpolateVal)
	default:
		interpolated = strings.ReplaceAll(outputText, interpolateKeyArg, strconv.Itoa(counter))
	}

	return interpolated, errReturn
}

// Collect and parse all Viper args, returning a struct holding their values
func getViperArgs(fallbackLogger *slog.Logger) (viperArgs, error) {
	// Parse viper flags from custom types
	err := viper.UnmarshalKey("output_format", &outputFormatterEnumVal, viper.DecodeHook(outputFormatterEnumHookFunc))
	if err != nil {
		fallbackLogger.Error("Failed to decode 'output_format' flag! Error: ", err)
	}
	logger := outputFormatterEnumVal

	err = viper.UnmarshalKey("interpolator", &interpolatorEnumVal, viper.DecodeHook(interpolatorEnumHookFunc))
	if err != nil {
		logger.Logger.Error("Failed to decode 'interpolator' flag!")
	}

	args := &viperArgs{
		outputFormatter:        logger,
		outputFormatterEnumVal: outputFormatterEnumVal,
		interpolatorEnumVal:    interpolatorEnumVal,
		stdout:                 viper.GetString("stdout"),
		stderr:                 viper.GetString("stderr"),
		socket:                 viper.GetString("socket"),
		socketSend:             viper.GetString("socket_send"),
		readSocket:             viper.GetBool("read_socket"),
		socketExitMsg:          viper.GetString("socket_exit_msg"),
		exitcode:               viper.GetInt("exitcode"),
		repeat:                 viper.GetInt("repeat"),
		repeatInterval:         time.Duration(viper.GetInt("repeat_interval")) * time.Second,
		repeatForever:          viper.GetBool("repeat_forever"),
		timeout:                viper.GetInt("timeout"),
		sigtermTimeout:         viper.GetInt("sigterm_timeout"),
		interpolateKey:         viper.GetString("interpolate_key"),
		interpolator:           viper.GetString("interpolator"),
		interpolateVal:         viper.GetString("interpolate_val"),
	}

	return *args, err
}

// Send and or read from unix socket. This func also parses args to
// determine if sending or reading
func outputSocket(cmd *cobra.Command, args viperArgs, outputText string) {
	logger := args.outputFormatter

	s, err := getunixSocket(*logger.Logger, args.socket, SocketDialTimeout)
	if err != nil {
		logger.Logger.Error(err.Error())
	}
	defer s.close()

	if args.socketSend != "" {
		err := s.sendTounixSocket(outputText + "\n")
		if err != nil {
			logger.Logger.Error(err.Error())
		}
	}

	if args.readSocket {
		response, err := s.readFromunixSocket(*logger.Logger, args.socketExitMsg)
		if err != nil {
			logger.Logger.Error(err.Error())
		}
		logger.cobraStdout(cmd, response)
	}
}

// Sends text to supported output locations
func outputStream(cmd *cobra.Command, args viperArgs, outputStream string) {
	counter := 0
	startTime := time.Now()
	timeout := time.Duration(args.timeout) * time.Second
	logger := args.outputFormatter

	// Pull text to output from right cli arg per output stream type
	var outputText string
	switch o := outputStream; o {
	case "stdout":
		outputText = args.stdout
	case "stderr":
		outputText = args.stderr
	case "socket":
		outputText = args.socketSend
	}

	for ok := true; ok; {
		interpolated, err := interpolate(args.interpolateKey, args.interpolator, outputText, counter, interpolateVal)
		if err != nil {
			logger.Logger.Error(err.Error())
		}

		// Use correct method of output per output stream type
		switch o := outputStream; o {
		case "stdout":
			logger.cobraStdout(cmd, interpolated)
		case "stderr":
			logger.cobraStderr(cmd, interpolated)
		case "socket":
			outputSocket(cmd, args, interpolated)
		}

		time.Sleep(args.repeatInterval)
		counter++

		if args.timeout != 0 {
			if time.Since(startTime) > timeout {
				logger.Logger.Info(fmt.Sprintf("Timeout of '%v' was reached", strconv.Itoa(args.timeout)))
				break
			}
		}
		if counter == args.repeat && !args.repeatForever {
			break
		}
	}
}

// ExecTester() is called by the root cmd. The entire functionality of
// the app is defined here.
func ExecTester(cmd *cobra.Command, fallbackLogger *slog.Logger) error {

	bindEnvToFlags()

	if err := validateParamSets(cmd); err != nil {
		return err
	}

	args, err := getViperArgs(fallbackLogger)
	logger := args.outputFormatter

	c := make(chan bool)
	// Catch signals to support a post sigterm timeout
	// os.Interrupt is thrown when sending ctrl+c
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Send output to correct stream by checking cli args
	if args.stdout != "" {
		go func() {
			outputStream(cmd, args, "stdout")
			c <- true
		}()
	}
	if args.stderr != "" {
		go func() {
			outputStream(cmd, args, "stderr")
			c <- true
		}()
	}
	if args.socket != "" {
		go func() {
			outputStream(cmd, args, "socket")
			c <- true
		}()
	}

	select {
	case <-c:
	case <-ctx.Done():
		stop()
		logger.Logger.Info(
			fmt.Sprintf(
				"Caught signal. Starting Sigterm timer to wait for "+
					"'%s' seconds to shutdown. Output to console will "+
					"continue while this timer is in effect",
				strconv.Itoa(args.sigtermTimeout)))

		t := time.Duration(args.sigtermTimeout) * time.Second
		time.Sleep(t)
	}

	// If non zero exit immediately with that exit code
	if cmd.Flags().Lookup("exitcode").Changed {
		os.Exit(args.exitcode)
	}

	if err != nil {
		return err
	} else {
		return nil
	}

}
