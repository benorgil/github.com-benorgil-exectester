/*
# A Dagger build pipeline for this project
*/
package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"dagger.io/dagger"
)

func main() {
	bc := GetConfig()

	// Capture all unhandled panics and output in structured log format
	// https://stackoverflow.com/questions/60516923/logging-unhandled-golang-panics
	defer func() {
		if r := recover(); r != nil {
			bc.logger.Error("Captured panic: %v", r, string(debug.Stack()))
			os.Exit(1)
		}
	}()

	if err := build(context.Background(), bc); err != nil {
		panic(err)
	}

	if err := test(context.Background(), bc); err != nil {
		bc.logger.Error(err.Error())
		panic(err)
	}

	// TODO: IM HERE sonar
	// if err := lint(context.Background(), bc); err != nil {
	// 	bc.logger.Error(err.Error())
	// 	panic(err)
	// }

}

func build(ctx context.Context, bc *BuildConfig) error {
	bc.logger.Info("Building")

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

	for _, goos := range bc.oses {
		for _, goarch := range bc.arches {
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
	bc.logger.Info("Generating Test Reports")

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
	// Install testing tools
	test = test.WithExec([]string{"go", "install", "gotest.tools/gotestsum@latest"})
	// Get unit test run report
	test = test.WithExec([]string{"gotestsum", "--junitfile=" + bc.unitTestReport})
	// Get coverage report
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

func lint(ctx context.Context, bc *BuildConfig) error {
	bc.logger.Info("Linting")

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// get reference to the local project
	src := client.Host().Directory(".")

	// get official sonar scanner image
	golang := client.Container().From(bc.sonarScannerImage)
	// mount cloned repository into `golang` image
	golang = golang.WithDirectory(bc.workDir, src).WithWorkdir(bc.workDir)

	// Run sonar scan, cwd is already set to the project root
	golang.WithExec([]string{
		//"/opt/sonar-scanner/bin/sonar-scanner " +
		"-Dproject.settings=" + bc.workDir + "/sonar-project.properties",
	}).Sync(ctx)

	// TODO: IM HERE
	return nil
}
