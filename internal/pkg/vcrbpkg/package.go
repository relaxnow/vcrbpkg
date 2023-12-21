package vcrbpkg

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/relaxnow/vcrbpkg/internal/pkg/logger"
)

func Package(args []string) error {
	var repoFolder string
	var rubyVersion Version

	// Prereqs
	err := ensureRubyIsInstalledGlobally()
	if err != nil {
		return err
	}
	err = ensureRvmIsInstalledGlobally()
	if err != nil {
		return err
	}

	if isAlreadyDirectory(args[0]) {
		repoFolder = args[0]
	} else {
		repoFolder, err = cloneRepo(args[0])
		if err != nil {
			return err
		}
	}

	rubyVersion = determineRubyVersion(repoFolder)
	checkIsSupportedRubyVersion(rubyVersion)
	err = rvmInstallRuby(repoFolder, rubyVersion)
	if err != nil {
		return err
	}
	ensureIsSupportedRailsVersion(repoFolder)
	installVeracodeGem(repoFolder)
	testRailsServe(repoFolder)
	runVeracodePrepare(repoFolder)
	return nil
}

func ensureRubyIsInstalledGlobally() error {
	command := "ruby"

	// LookPath returns the complete path to the binary or an error if not found
	path, err := exec.LookPath(command)
	if err != nil {
		logger.WithError(err).Error("Unable to run ruby command")
		return fmt.Errorf("unable to run ruby command, please ensure ruby is installed")
	}

	logger.Infof("%s is available at %s", command, path)

	// Define the command and arguments
	cmd := exec.Command("ruby", "--version")

	// Run the command and capture its output
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.WithError(err).Error("Unable to run ruby --version command")
		return fmt.Errorf("unable to run ruby --version command, please ensure ruby is installed correctly")
	}

	// Print the output
	logger.Infof("ruby version: %s", output)
	return nil
}

func ensureRvmIsInstalledGlobally() error {
	command := "rvm"

	// LookPath returns the complete path to the binary or an error if not found
	path, err := exec.LookPath(command)
	if err != nil {
		logger.WithError(err).Error("Unable to run rvm command")
		return fmt.Errorf("unable to run rvm command. RVM may not be available, please install with: curl -sSL https://get.rvm.io | bash")
	}

	logger.Infof("%s is available at %s", command, path)

	// Define the command and arguments
	cmd := exec.Command("rvm", "version")

	// Run the command and capture its output
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.WithError(err).Error("Unable to run rvm version command")
		return fmt.Errorf("unable to run rvm version command. Please reinstall RVM")
	}

	// Print the output
	logger.Infof("RVM version: %s", output)
	return nil
}

func isAlreadyDirectory(dirPath string) bool {
	// Check if the provided path is a directory
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Infof("Input %s does not exist.", dirPath)
			return false
		} else {
			logger.WithError(err).Error("Error checking directory, assuming it's not one")
			return false
		}
	}

	if fileInfo.IsDir() {
		logger.Infof("%s is a valid directory.", dirPath)
		return true
	} else {
		logger.Infof("%s is not a directory.", dirPath)
		return false
	}
}

func cloneRepo(urlOrFolder string) (string, error) {
	// Check if the given path is an existing directory
	if fi, err := os.Stat(urlOrFolder); err == nil && fi.IsDir() {
		logger.Infof("Folder '%s' already exists. Skipping clone.", urlOrFolder)
		return urlOrFolder, nil
	}

	// Check if 'git' is installed
	_, err := exec.LookPath("git")
	if err != nil {
		logger.WithError(err).Error("git is not installed")
		return "", fmt.Errorf("git is not installed, please install git")
	}

	// Get the system's temporary directory
	tempDir := os.TempDir()

	// Create a temporary directory with a specific prefix
	prefix := "vcrbpkg"
	temporaryDir, err := os.MkdirTemp(tempDir, prefix)
	if err != nil {
		logger.WithError(err).Errorf("eror creating temporary directory in '%s' with prefix '%s'", tempDir, prefix)
		return "", fmt.Errorf("eror creating temporary directory in '%s' with prefix '%s'", tempDir, prefix)
	}

	// Use the temporary directory
	logger.Infof("Temporary directory:", temporaryDir)

	// Run 'git clone' command
	cmd := exec.Command("git", "clone", urlOrFolder, temporaryDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Infof("Cloning repository from %s...\n", urlOrFolder)

	err = cmd.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to clone repository '%s' to '%s'", urlOrFolder, temporaryDir)
		return "", fmt.Errorf("failed to clone repository '%s' to '%s'", urlOrFolder, temporaryDir)
	}

	logger.Infof("Repository cloned successfully at '%s'.", temporaryDir)
	return temporaryDir, nil
}

func ensureIsSupportedRailsVersion(repoFolder string) {
	// TODO: Run bundle show rails and parse version
}

func rvmInstallRuby(repoFolder string, rubyVersion Version) error {
	// Run 'rvm install' command
	cmd := exec.Command("rvm", "install", rubyVersion.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info("Installing Ruby version with RVM")

	err := cmd.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to  rvm install")
		return fmt.Errorf("failed to rvm install %s", rubyVersion.String())
	}
	return nil
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
