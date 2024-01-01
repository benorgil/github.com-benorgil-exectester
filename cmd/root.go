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
	"log/slog"
	"os"

	"github.com/benorgil/exectester/configs"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flags
var (
	cfgFile        string
	stdout         string
	stderr         string
	socket         string
	socketSend     string
	readSocket     bool
	socketExitMsg  string
	exitcode       int
	repeat         int
	repeatInterval int
	repeatForever  bool
	timeout        int
	sigtermTimeout int
	interpolateKey string
	interpolateVal string
)

// If setting config with env vars they must be prefixed with this string
const (
	EnvPrefix string = "ET_"
)

// rootCmd represents the base command when called without any subcommands
func RootCmd(fallbackLogger *slog.Logger) *cobra.Command {

	rootCmd := &cobra.Command{
		Use:   "et",
		Short: "A tool to send arbitrary text to stdout or stderr",
		Long: `A simple cli tool to write output to stdout and or stderr 
and also set the exit code. It supports dynamically generating output
which can also be interpolated with other values in various ways.

Example Usage
-------------
Send to stdout and stderr:
$ et --stdout='sending to stdout' --stderr='sending to stderr'

Send to stdout and stderr 3 times:
$ et --stdout='sending to stdout' --stderr='sending to stderr' --repeat=3

Send to stdout and stderr 3 times and interpolate __I__ with int counter:
$ et --stdout='stdout counter: __I__' --stderr='stderr counter: __I__' --repeat=3

Send to stdout and stderr 3 times and interpolate __I__ with int counter and start counter at 5:
$ et --stdout='stdout counter: __I__' --stderr='stderr counter: __I__' --repeat=3 --interpolate_val=zzz

Send to stdout and stderr 3 times and interpolate __I__ with a string 'zzz':
$ et --stdout='stdout counter: __I__' --stderr='stderr counter: __I__' --repeat=3 --interpolator=string --interpolate_val=zzz

Send to stdout for 5 seconds:
$ et --stdout='stdout counter: __I__' --repeat_forever --timeout=5

Send to stdout and stderr and then exit with code '123'
$ et --stdout='sending to stdout' --stderr='sending to stderr' --exitcode=123
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := ExecTester(cmd, fallbackLogger); err != nil {
				return err
			}
			return nil
		},
	}

	viper.SetEnvPrefix(EnvPrefix)

	//////// Set flags here
	//// Supported outputs
	rootCmd.PersistentFlags().StringVarP(&stdout, "stdout", "o", "", "Text to send to stdout")
	viper.BindPFlag("stdout", rootCmd.PersistentFlags().Lookup("stdout"))

	rootCmd.PersistentFlags().StringVarP(&stderr, "stderr", "e", "", "Text to send to stderr")
	viper.BindPFlag("stderr", rootCmd.PersistentFlags().Lookup("stderr"))

	rootCmd.PersistentFlags().StringVarP(&socket, "socket", "u", "", "Name of unix socket")
	viper.BindPFlag("socket", rootCmd.PersistentFlags().Lookup("socket"))

	rootCmd.PersistentFlags().StringVarP(&socketSend, "socket_send", "w", "", "Text to send to unix socket")
	viper.BindPFlag("socket_send", rootCmd.PersistentFlags().Lookup("socket_send"))

	rootCmd.PersistentFlags().BoolVarP(&readSocket, "read_socket", "q", false, "Poll the unix socket for output")
	viper.BindPFlag("read_socket", rootCmd.PersistentFlags().Lookup("read_socket"))

	rootCmd.PersistentFlags().StringVarP(&socketExitMsg, "socket_exit_msg", "l", "", "Close connection to socket if it returns this text")
	viper.BindPFlag("socket_exit_msg", rootCmd.PersistentFlags().Lookup("socket_exit_msg"))

	//// Custom type flags
	var outputFormatterEnumDefault = outputFormatterEnumStructured // Default value
	rootCmd.PersistentFlags().VarP(&outputFormatterEnumDefault, "output_format", "z", outputFormatterEnumValuesInfoMsg)
	viper.BindPFlag("output_format", rootCmd.PersistentFlags().Lookup("output_format"))

	var interpolatorEnumDefault = interpolatorEnumIntCounter // Default value
	rootCmd.PersistentFlags().VarP(&interpolatorEnumDefault, "interpolator", "i", interpolatorEnumValuesInfoMsg)
	viper.BindPFlag("interpolator", rootCmd.PersistentFlags().Lookup("interpolator"))

	//// Rest of flags
	rootCmd.PersistentFlags().IntVarP(&exitcode, "exitcode", "c", 0, "Exit with this exit code")
	viper.BindPFlag("exitcode", rootCmd.PersistentFlags().Lookup("exitcode"))

	rootCmd.PersistentFlags().IntVarP(&repeat, "repeat", "r", 1, "Number of times to repeat output")
	viper.BindPFlag("repeat", rootCmd.PersistentFlags().Lookup("repeat"))

	rootCmd.PersistentFlags().IntVarP(&repeatInterval, "repeat_interval", "p", 1, "Seconds to wait between repeated output")
	viper.BindPFlag("repeat_interval", rootCmd.PersistentFlags().Lookup("repeat_interval"))

	rootCmd.PersistentFlags().StringVarP(&interpolateKey, "interpolate_key", "k", "__I__", "Substring key to interpolate")
	viper.BindPFlag("interpolate_key", rootCmd.PersistentFlags().Lookup("interpolate_key"))

	rootCmd.PersistentFlags().StringVarP(&interpolateVal, "interpolate_val", "v", "", "The value to replace interpolate_key with")
	viper.BindPFlag("interpolate_val", rootCmd.PersistentFlags().Lookup("interpolate_val"))

	rootCmd.PersistentFlags().BoolVarP(&repeatForever, "repeat_forever", "f", false, "Run forever")
	viper.BindPFlag("repeat_forever", rootCmd.PersistentFlags().Lookup("repeat_forever"))

	rootCmd.PersistentFlags().IntVarP(&timeout, "timeout", "t", 0, "Exits when timeout (seconds) exceeded. '0' means no timeout set")
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))

	rootCmd.PersistentFlags().IntVarP(&sigtermTimeout, "sigterm_timeout", "x", 0, "If a sigterm is caught while running wait for X seconds because exiting")
	viper.BindPFlag("sigterm_timeout", rootCmd.PersistentFlags().Lookup("sigterm_timeout"))

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.exectester.yaml)")

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(cmd *cobra.Command) error {
	cobra.OnInitialize(initConfig)
	return cmd.Execute()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".exectester" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".exectester")
	}

	// This is useless without a config file that hard codes ALL your fields: https://github.com/spf13/viper/issues/584
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		configs.FallbackLogger.Error("Using config file:" + viper.ConfigFileUsed())
	}
}
