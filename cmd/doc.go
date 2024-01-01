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

/*
# Cmd package

This is a mostly standard Cobra + Viper app created with the default
templates:

  - `go mod init exectester`
  - `cobra-cli init`

There are no subcommands just the root command.

The root cobra Command ([github.com/spf13/cobra.Command]) is wrapped
in a function ([cmd.RootCmd]) to make it testable. It calls another
function ([cmd.ExecTester]) where the entire app's functionality is
defined.

# Custom type flags

These are are Cobra flags built from a custom type.

  - [cmd.outputFormatterEnum]
  - [cmd.interpolatorEnum]

It takes an obnoxious amount of scaffolding to get Cobra + Viper to
support flags from custom types.
*/
package cmd
