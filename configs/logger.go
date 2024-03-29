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
# Initial config

I want to consolidate things like logger initialization and build/test
config in a single place. While Viper is used for config handling of the
cli app itself the config here needs to be initialized before the Cobra
app itself- so I don't think I can use Viper here.
*/
package configs

import (
	"log/slog"
	"os"
)

// This is a Cobra + Viper app that does all its config with Viper including
// the app's logging config. If there is an issue loading config it attempts
// to fallback to this logger
var FallbackLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// The default structured log format
type OutputFormat struct {
	Time  string `json:"time"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}
