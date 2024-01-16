/*
# Pipeline config

The pipeline inject a config struct into all the pipeline funcs.
It gets config to generate the struct from env vars and even this
apps own shared packages.

Inside an organization components of this pipeline could be broken out
into shared packages and reused.
*/
package main

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/benorgil/exectester/configs"
)

// The fields of this struct could probably be shared
// across many repos
type SharedBuildConfig struct {
	// The container image used to build this project
	builderImage string
	// The container image used run sonar scans
	sonarScannerImage string
	// TODO IM HERE sonar scanner token
	// The scanner token will be injected via env var during build pipeline
	sonarScannerTokenEnvVar string
	// The directory where artifacts will be copied
	buildDir string
	// The working dir project will be copied to inside container
	workDir string
	// Unit test execution report (passed/failed)
	unitTestReport string
	// Coverage report path
	coverageReport string
}

// The fields of this struct are specific to this project's build
// pipeline
type ExecTesterBuildConfig struct {
	// OSs to generate golang exe for
	oses []string
	// Architectures for each OS to generate golang exe for
	arches []string
	// Name of compiled exe
	exeName string
	// The env var integration tests look for to get the path of exe
	testArgExePathEnvVar string
	// The actual path of the exe
	testArgExePath string
	// The logger build automation will use for output.
	// Uses the "fallback" logger because we don't need anything custom here
	logger *slog.Logger
}

// All build config is finally collected in this struct
// for organizational purposes
type BuildConfig struct {
	SharedBuildConfig
	ExecTesterBuildConfig
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
	bc.sonarScannerImage = "sonarsource/sonar-scanner-cli:5"
	bc.buildDir = "build"
	bc.workDir = "/src"
	bc.unitTestReport = "./build/unit-tests.xml"
	bc.coverageReport = "./build/cov.out"

	bc.oses = []string{"linux", "darwin", "windows"}
	bc.arches = []string{"amd64", "arm64"}
	bc.exeName = "et"
	bc.logger = configs.FallbackLogger
	bc.testArgExePathEnvVar = configs.TestArgExePath
	bc.testArgExePath = getTestExePath()

	return bc
}
