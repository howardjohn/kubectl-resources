// +build mage

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"

	"github.com/magefile/mage/sh"
	"github.com/rogpeppe/go-internal/modfile"
)

var Default = All

func init() {
	_ = os.Setenv("GO111MODULE", "on")
}

// Format, lint, test, and install
func All() {
	mg.SerialDeps(Format, Lint, Test, Install)
}

// Run all linters
func Lint() error {
	mg.Deps(Deps)
	return sh.Run("golangci-lint", "run", "--fix")
}

// Run goimports
func Format() error {
	mg.Deps(Deps)
	if err := sh.Run("go", "mod", "tidy"); err != nil {
		return err
	}
	module, err := parseGoModule()
	if err != nil {
		return err
	}
	return sh.Run("goimports", "-l", "-w", "-local", module, ".")
}

// Installs this module
func Install() error {
	return sh.Run("go", "install")
}

// Test this module
func Test() error {
	return sh.Run("go", "test", "-v", "./...")
}

// Helper to quickly run main
func Run() error {
	return sh.RunV("go", "run", "main.go")
}

// Installs all dependencies for building and testing
func Deps() error {
	modules := map[string]string{
		"golangci-lint": "github.com/golangci/golangci-lint/cmd/golangci-lint",
		"goimports":     "golang.org/x/tools/cmd/goimports",
	}
	for binary, mod := range modules {
		if !binaryExists(binary) {
			if err := sh.Run("go", "get", mod); err != nil {
				return err
			}
		}
	}
	return nil
}

// Fails if the repo is dirty
func GitDirty() error {
	o, err := sh.Output("git", "status", "--porcelain")
	if o != "" || err != nil {
		// Show the full status
		sh.Run("git", "status")
		sh.Run("git", "diff")
		return fmt.Errorf("git is dirty")
	}
	return nil
}

func parseGoModule() (string, error) {
	modContents, err := ioutil.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	mod, _ := modfile.Parse("", modContents, nil)
	if err != nil {
		return "", err
	}
	return mod.Module.Mod.Path, nil
}

func binaryExists(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}
