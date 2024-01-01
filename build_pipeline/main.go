/*
# A Dagger build pipeline for this project
*/
package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
)

// Define build and test params matrix
var (
	oses   = []string{"linux", "darwin", "windows"}
	arches = []string{"amd64", "arm64"}
	// Name of compiled exe
	exe_name = "et"
	// Build artifact directory
	build_dir = "build"
	// Tell integration tests where to find compiled exe
	// The cwd of the test run will be the package folder, go up a dir to get at "./build"
	TestArgExePath = "../build/linux/arm64/et"
	// Coverage report path
	coverageReport = "./build/cov.out"
)

func main() {
	bc := GetConfig()

	if err := build(context.Background(), bc); err != nil {
		bc.Logger.Error(err.Error())
	}

	if err := test(context.Background(), bc); err != nil {
		bc.Logger.Error(err.Error())
	}
}

func build(ctx context.Context, bc *BuildConfig) error {
	bc.Logger.Info("Building with Dagger")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// get reference to the local project
	src := client.Host().Directory(".")

	// create empty directory to put build outputs
	outputs := client.Directory()

	// get `golang` image
	golang := client.Container().From("golang:latest")

	// mount cloned repository into `golang` image
	golang = golang.WithDirectory("/src", src).WithWorkdir("/src")

	for _, goos := range oses {
		for _, goarch := range arches {
			// create a directory for each os and arch
			path := fmt.Sprintf("%s/%s/%s/", build_dir, goos, goarch)
			exePath := path + exe_name

			// set GOARCH and GOOS in the build environment
			build := golang.WithEnvVariable("GOOS", goos)
			build = build.WithEnvVariable("GOARCH", goarch)

			// build application
			build = build.WithExec([]string{"go", "build", "-o", exePath})

			// get reference to build output directory in container
			outputs = outputs.WithDirectory(path, build.Directory(path))
		}
	}
	// write build artifacts to host
	_, err = outputs.Export(ctx, ".")
	if err != nil {
		return err
	}

	return nil
}

func test(ctx context.Context, bc *BuildConfig) error {
	bc.Logger.Info("Testing with Dagger")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// get reference to the local project
	src := client.Host().Directory(".")

	// create empty directory to put build outputs
	outputs := client.Directory()

	// get `golang` image
	golang := client.Container().From("golang:latest")

	// mount cloned repository into `golang` image
	golang = golang.WithDirectory("/src", src).WithWorkdir("/src")

	// Run test suite
	test := golang.WithEnvVariable(bc.TestArgExePath, TestArgExePath)
	test = test.WithExec([]string{"go", "test", "./...", "-coverprofile=" + coverageReport})

	// get reference to build output directory in container
	path := build_dir + "/"
	outputs = outputs.WithDirectory(path, test.Directory(path))

	// write build artifacts to host
	_, err = outputs.Export(ctx, ".")
	if err != nil {
		return err
	}

	return nil
}
