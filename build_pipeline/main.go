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

func main() {
	bc := GetConfig()

	if err := build(context.Background(), bc); err != nil {
		bc.logger.Error(err.Error())
	}

	if err := test(context.Background(), bc); err != nil {
		bc.logger.Error(err.Error())
	}
}

func build(ctx context.Context, bc *BuildConfig) error {
	bc.logger.Info("Building with Dagger")

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
	golang := client.Container().From(bc.builderImage)

	// mount cloned repository into `golang` image
	golang = golang.WithDirectory(bc.workDir, src).WithWorkdir(bc.workDir)

	for _, goos := range oses {
		for _, goarch := range arches {
			// create a directory for each os and arch
			path := fmt.Sprintf("%s/%s/%s/", bc.buildDir, goos, goarch)
			exePath := path + bc.exeName

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
	bc.logger.Info("Testing with Dagger")

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
	golang := client.Container().From(bc.builderImage)

	// mount cloned repository into `golang` image
	golang = golang.WithDirectory(bc.workDir, src).WithWorkdir(bc.workDir)

	// Run test suite
	test := golang.WithEnvVariable(bc.testArgExePathEnvVar, bc.testArgExePath)
	test = test.WithExec([]string{"go", "test", "./...", "-coverprofile=" + bc.coverageReport})

	// get reference to build output directory in container
	path := bc.buildDir + "/"
	outputs = outputs.WithDirectory(path, test.Directory(path))

	// write build artifacts to host
	_, err = outputs.Export(ctx, ".")
	if err != nil {
		return err
	}

	return nil
}
