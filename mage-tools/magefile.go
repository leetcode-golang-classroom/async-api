//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

var Default = Build

// clean the build binary
func Clean() error {
	return sh.Rm("bin")
}

// update the dependency
func Update() error {
	return sh.Run("go", "mod", "download")
}

// build Creates the binary in the current directory.
func Build() error {
	mg.Deps(Clean)
	mg.Deps(Update)
	// build the http server
	return sh.Run("go", "build", "-o", "./bin/apiserver", "./cmd/apiserver/main.go")
}

// buildWorker Create the binary worker
func BuildWorker() error {
	mg.Deps(Clean)
	mg.Deps(Update)
	// build the worker
	return sh.Run("go", "build", "-o", "./bin/worker", "./cmd/worker/main.go")
}

// LaunchServer start the server
func LaunchServer() error {
	mg.Deps(Build)
	return sh.RunV("./bin/apiserver")
}

// LaunchWorker start the worker
func LaunchWorker() error {
	mg.Deps(BuildWorker)
	return sh.RunV("./bin/worker")
}

// run the test
func Test() error {
	return sh.RunV("go", "test", "-v", "./...")
}
