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
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/benorgil/exectester/configs"
)

/*
For the Cobra app's output we are also using loggers to get structured output,
BUT Cobra makes using loggers like this a pita. In our unit tests we use
cmd.SetOut() and cmd.SetErr() to mock the output. In order for these functions
to work though in the app itself we need to pass all output through OutOrStdout()
and ErrOrStderr()- ie we can't just log it directly.

So we must:
  - Create a separate logger and buffer for stdout and stderr
  - Log to those loggers and capture that output into the buffers (not sending to console)
  - Grab the output from the buffers and pass to OutOrStdout() and ErrOrStderr()
*/
type OutputFormatter struct {
	// To isolate regular log output from app output
	// this logger is used for log messages
	Logger *slog.Logger

	// Buffers for Cobra output
	BuffOut *bytes.Buffer
	BuffErr *bytes.Buffer
	// Logger to control stdout
	CobraLoggerStdout *slog.Logger
	// Logger to control stderr
	CobraLoggerStderr *slog.Logger
	// The structure of output fields
	LogSchema *configs.OutputFormat
}

// By default slog prints literal '\n' new line chars
// Using this function to swallow log lines and manually print with Println()
func humanReadableReplaceAttr() func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "msg" {
			re := regexp.MustCompile("^msg=")
			s := re.ReplaceAllString(a.String(), "")
			fmt.Println(s)
		}
		a = slog.Attr{}
		return a
	}
}

// Write to stdout via cobra method
func (a *OutputFormatter) cobraStdout(cmd *cobra.Command, output string) {
	a.CobraLoggerStdout.Info(output)
	out, _ := io.ReadAll(a.BuffOut)
	fmt.Fprint(cmd.OutOrStdout(), string(out))
}

// Write to stderr via cobra method
func (a *OutputFormatter) cobraStderr(cmd *cobra.Command, output string) {
	a.CobraLoggerStderr.Info(output)
	out, _ := io.ReadAll(a.BuffErr)
	fmt.Fprint(cmd.ErrOrStderr(), string(out))
}

// This should be called after the config is loaded to get the right logger
// If loggerType is "" a default is set and the logger config field is checked
// from env var
func (a *OutputFormatter) getLogger(loggerType string) OutputFormatter {
	logger := OutputFormatter{
		BuffOut:   bytes.NewBufferString(""),
		BuffErr:   bytes.NewBufferString(""),
		LogSchema: new(configs.OutputFormat),
	}

	// Default to structured logging everywhere
	if loggerType == "" && loggerType != "human_readable" && loggerType != "structured" {
		loggerType = "structured"
	}

	if loggerType == "human_readable" {
		logger.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			ReplaceAttr: humanReadableReplaceAttr(),
		}))
		logger.CobraLoggerStdout = slog.New(slog.NewTextHandler(logger.BuffOut, &slog.HandlerOptions{
			ReplaceAttr: humanReadableReplaceAttr(),
		}))
		logger.CobraLoggerStderr = slog.New(slog.NewTextHandler(logger.BuffErr, &slog.HandlerOptions{
			ReplaceAttr: humanReadableReplaceAttr(),
		}))
	} else if loggerType == "structured" {
		logger.Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
		logger.CobraLoggerStdout = slog.New(slog.NewJSONHandler(logger.BuffOut, nil))
		logger.CobraLoggerStderr = slog.New(slog.NewJSONHandler(logger.BuffErr, nil))
	}

	return logger
}
