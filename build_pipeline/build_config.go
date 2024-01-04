/*
# Pipeline config

The pipeline inject a config struct into all the pipeline funcs.
It gets config to generate the struct from env vars and even this
apps own shared packages.
*/
package main

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/benorgil/exectester/configs"
)

var (
	oses   = []string{"linux", "darwin", "windows"}
	arches = []string{"amd64", "arm64"}
)

// All build config is collected in this struct for organizational
// purposes
type BuildConfig struct {
	// The container image to use to build this project
	builderImage string
	// Name of compiled exe
	exeName string
	// Build artifact directory
	buildDir string
	// The working dir project will be copied to inside container
	workDir string
	// Coverage report path
	coverageReport string
	// The env var integration tests look for to get the path of exe
	testArgExePathEnvVar string
	// The actual path of the exe
	testArgExePath string
	// The logger build automation will use for output.
	// Uses the "fallback" logger because we don't need anything custom here
	logger *slog.Logger
}

// Checks arch to get location of compiled exe some integration tests need
// to run. The cwd of the test run will be the package folder, go up a dir
// to get at "./build"
func getTestExePath() string {
	// Relative to dagger's cwd inside the project
	buildDir := "../build"
	// Will always be linux because it runs containerized
	os := "linux"
	// Arch is dependant on the host system
	arch := runtime.GOARCH

	return fmt.Sprintf("%s/%s/%s/et", buildDir, os, arch)
}

func GetConfig() *BuildConfig {
	bc := new(BuildConfig)

	bc.builderImage = "golang:latest"
	bc.exeName = "et"
	bc.buildDir = "build"
	bc.buildDir = "/src"
	bc.coverageReport = "./build/cov.out"
	bc.logger = configs.FallbackLogger
	bc.testArgExePathEnvVar = configs.TestArgExePath
	bc.testArgExePath = getTestExePath()

	return bc
}
