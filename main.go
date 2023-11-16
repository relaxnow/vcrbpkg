package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	var repoFolder string
	var rubyVersion Version

	// Check if at least one command-line argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: vcrbpkg /folder/to/clone OR vcrbpkg https://github.com/user/repo")
		return
	}

	ensureRubyIsInstalledGlobally()
	ensureRvmIsInstalledGlobally()
	repoFolder = cloneRepo(os.Args[1])
	ensureIsSupportedRailsVersion(repoFolder)
	rubyVersion = determineRubyVersion(repoFolder)
	ensureIsSupportedRubyVersion(rubyVersion)
	rvmInstallRuby(repoFolder, rubyVersion)
	installVeracodeGem(repoFolder)
	testRailsServe(repoFolder)
	runVeracodePrepare(repoFolder)
}

func ensureRubyIsInstalledGlobally() {
	command := "ruby"

	// LookPath returns the complete path to the binary or an error if not found
	path, err := exec.LookPath(command)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Printf("%s is available at %s\n", command, path)

	// Define the command and arguments
	cmd := exec.Command("ruby", "--version")

	// Run the command and capture its output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		return
	}

	// Print the output
	fmt.Printf("Ruby version:\n%s", output)
}

func ensureRvmIsInstalledGlobally() {
	command := "rvm"

	// LookPath returns the complete path to the binary or an error if not found
	path, err := exec.LookPath(command)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return
	}

	fmt.Printf("%s is available at %s\n", command, path)

	// Define the command and arguments
	cmd := exec.Command("rvm", "version")

	// Run the command and capture its output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		return
	}

	// Print the output
	fmt.Printf("Ruby version:\n%s", output)
}

func cloneRepo(urlOrFolder string) string {
	// Check if the given path is an existing directory
	if fi, err := os.Stat(urlOrFolder); err == nil && fi.IsDir() {
		fmt.Printf("Folder '%s' already exists. Skipping clone.\n", urlOrFolder)
		return urlOrFolder
	}

	// Check if 'git' is installed
	_, err := exec.LookPath("git")
	if err != nil {
		fmt.Printf("git is not installed: %v\n", err)
		os.Exit(1)
	}

	// Get the system's temporary directory
	tempDir := os.TempDir()

	// Create a temporary directory with a specific prefix
	prefix := "vcrbpkg"
	temporaryDir, err := os.MkdirTemp(tempDir, prefix)
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err)
		os.Exit(1)
	}

	// Use the temporary directory
	fmt.Println("Temporary directory:", temporaryDir)

	// Run 'git clone' command
	cmd := exec.Command("git", "clone", urlOrFolder, temporaryDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("Cloning repository from %s...\n", urlOrFolder)

	err = cmd.Run()
	if err != nil {
		fmt.Printf("failed to clone repository: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Repository cloned successfully.")
	return temporaryDir
}

func ensureIsSupportedRailsVersion(repoFolder string) {
	// TODO: Run bundle show rails and parse version
}

func determineRubyVersion(repoFolder string) Version {
	// TODO: try:
	// 1. .rails-version
	// 2. .rails-version.sample
	// 3. bundle platform
	// 4. .github/workflows ruby-version: '3.0'
	return Version{Major: 3, Minor: 0, Patch: 0}
}

func ensureIsSupportedRubyVersion(rubyVersion Version) {
	// TODO: check against versions in https://docs.veracode.com/r/compilation_ruby
}

func rvmInstallRuby(repoFolder string, rubyVersion Version) {
	// TODO: Get commonly required libraries
}

func testRailsServe(repoFolder string) {
	// TODO: Test if `rails serve` doesn't crash
}

func installVeracodeGem(repoFolder string) {
	// # If Ruby < 2.4
	// source 'https://rubygems.org'
	// gem 'rubyzip', '~>1.0'
	// bundle add veracode
}

func runVeracodePrepare(repoFolder string) {
	// TODO: veracode prepare -vD
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func (a *Version) IsGreaterThanOrEqual(b *Version) bool {
	// 3.0.0 > 2.0.0
	if a.Major > b.Major {
		return true
	}
	// 3.0.0 < 4.0.0
	if a.Major < b.Major {
		return false
	}
	// 2.5.0 > 2.4.0
	if a.Minor > b.Minor {
		return true
	}
	// 2.5.0 < 2.6.0
	if a.Minor < b.Minor {
		return false
	}
	// 2.4.11 > 2.4.0
	if a.Patch > b.Patch {
		return true
	}
	// 2.4.11 < 2.4.36
	if a.Patch < b.Patch {
		return false
	}
	// Exactly the same, return false.
	return true
}
