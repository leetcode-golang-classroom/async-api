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
	// build the websocket server
	err := sh.Run("go", "build", "-o", "./bin/server", "./cmd/server/main.go")
	if err != nil {
		return err
	}
	// build the websocket client
	err = sh.Run("go", "build", "-o", "./bin/client", "./cmd/client/main.go")
	if err != nil {
		return err
	}
	return nil
}

// LaunchServer start the server
func LaunchServer() error {
	mg.Deps(Build)
	err := sh.RunV("./bin/server")
	if err != nil {
		return err
	}
	return nil
}

// LaunchClient start the client
func LaunchClient() error {
	mg.Deps(Build)
	err := sh.RunV("./bin/client")
	if err != nil {
		return err
	}
	return nil
}

// run the test
func Test() error {
	err := sh.RunV("go", "test", "-v", "./...")
	if err != nil {
		return err
	}
	return nil
}
