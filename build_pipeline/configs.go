/*
# Pipeline config

The pipeline inject a config struct into all the pipeline funcs.
It gets config to generate the struct from env vars and even this
apps own shared packages.
*/
package main

import (
	"log/slog"

	"github.com/benorgil/exectester/configs"
)

type BuildConfig struct {
	TestArgExePath string
	Logger         *slog.Logger
}

func GetConfig() *BuildConfig {
	bc := new(BuildConfig)
	// TODO IM HERE build out config loading for build pipeline here- maybe viper lol
	bc.Logger = configs.FallbackLogger
	bc.TestArgExePath = configs.TestArgExePath

	return bc
}
