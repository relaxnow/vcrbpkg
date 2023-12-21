package app

import (
	"fmt"
	"os"
	"os/exec"
)

func Package() {
	var repoFolder string
	var rubyVersion Version

	// Check if at least one command-line argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: vcrbpkg /folder/to/clone OR vcrbpkg https://github.com/user/repo")
		return
	}

	// Prereqs
	ensureRubyIsInstalledGlobally()
	ensureRvmIsInstalledGlobally()

	if isAlreadyDirectory(os.Args[1]) {
		repoFolder = os.Args[1]
	} else {
		repoFolder = cloneRepo(os.Args[1])
	}

	rubyVersion = determineRubyVersion(repoFolder)
	ensureIsSupportedRubyVersion(rubyVersion)
	rvmInstallRuby(repoFolder, rubyVersion)
	ensureIsSupportedRailsVersion(repoFolder)
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
	}

	fmt.Printf("%s is available at %s\n", command, path)

	// Define the command and arguments
	cmd := exec.Command("rvm", "version")

	// Run the command and capture its output
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		fmt.Println("RVM may not be available, please install with: curl -sSL https://get.rvm.io | bash")
		os.Exit(1)
	}

	// Print the output
	fmt.Printf("Ruby version:\n%s", output)
}

func isAlreadyDirectory(dirPath string) bool {
	// Check if the provided path is a directory
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Input %s does not exist.\n", dirPath)
		} else {
			fmt.Printf("Error checking directory: %v\n", err)
		}
		return false
	}

	if fileInfo.IsDir() {
		fmt.Printf("%s is a valid directory.\n", dirPath)
		return true
	} else {
		fmt.Printf("%s is not a directory.\n", dirPath)
		return false
	}
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
