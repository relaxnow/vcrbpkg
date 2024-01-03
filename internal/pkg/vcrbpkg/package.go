package vcrbpkg

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

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
	checkIsSupportedRailsVersion(repoFolder, rubyVersion)
	err = installVeracodeGem(repoFolder, rubyVersion)
	if err != nil {
		return err
	}
	checkRailsServer(repoFolder, rubyVersion)
	runVeracodePrepare(repoFolder, rubyVersion)
	if err != nil {
		return err
	}
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
	logger.Infof("Temporary directory: %s", temporaryDir)

	// Run 'git clone' command
	cmd := exec.Command("git", "clone", "--depth", "1", urlOrFolder, temporaryDir)
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

func checkIsSupportedRailsVersion(repoFolder string, rubyVersion Version) {
	cmd := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "show", "rails")
	cmd.Dir = repoFolder

	logger.Info("Detecting Rails version with Bundler")

	output, err := cmd.CombinedOutput()

	if err != nil {
		logger.WithError(err).Errorf("failed to bundle show rails")
		logger.Warn("failed to run bundle show rails, unable to verify rails version, hoping for the best and continuing")
		return
	}

	logger.Infof("bundle show rails output: '%s'", output)

	// Define a regular expression pattern to match the version number
	re := regexp.MustCompile(`rails-(\d+\.\d+\.\d+)`)

	// Find the first match in the input string
	match := re.FindStringSubmatch(string(output))

	// Check if a match is found
	if len(match) < 2 {
		logger.Warnf("Version number not found in the input string: %s", output)
		logger.Warn("failed to parse output of bundle show rails, unable to verify rails version, hoping for the best and continuing")
		return
	}

	// The version number is captured in the first submatch group
	versionNumber := match[1]
	railsVersion := parseRubyVersion(versionNumber)
	if railsVersion.Major >= 3 && railsVersion.Major <= 6 || (railsVersion.Major == 7 && railsVersion.Minor == 0) {
		logger.Infof("Supported Rails version: %s", versionNumber)
	} else {
		logger.Warnf("Unsupported Rails version: " + versionNumber)
		logger.Warn("Unsupported Rails version, hoping for the best and continuing")
	}
}

func rvmInstallRuby(repoFolder string, rubyVersion Version) error {
	// Run 'rvm install' command
	cmd := exec.Command("rvm", "install", rubyVersion.String())
	cmd.Dir = repoFolder
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info("Installing Ruby version with RVM")

	err := cmd.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to  rvm install")
		return fmt.Errorf("failed to rvm install %s", rubyVersion.String())
	}

	logger.Info("Installing Ruby version with RVM")

	// Run 'rvm install' command
	cmd3 := exec.Command("rvm", rubyVersion.String(), "do", "rvm", "gemset", "create", "veracode")
	cmd3.Dir = repoFolder
	cmd3.Stdout = os.Stdout
	cmd3.Stderr = os.Stderr

	err = cmd3.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to rvm use")
		return fmt.Errorf("failed to rvm use %s@veracode --create", rubyVersion.String())
	}

	cmd2 := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "install")
	cmd2.Dir = repoFolder
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	logger.Info("Using RVM ruby version")

	err = cmd2.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to rvm use")
		return fmt.Errorf("failed to rvm use %s@veracode bundle install", rubyVersion.String())
	}

	return nil
}

func checkRailsServer(repoFolder string, rubyVersion Version) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rvm", rubyVersion.String()+"@veracode", "do", "rails", "server")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RAILS_ENV=development")
	cmd.Dir = repoFolder
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info("Running rails server in development")

	if err := cmd.Run(); err != nil {
		logger.WithError(err).Info("rails server error")
		if err.Error() == "signal: killed" {
			logger.Infof("Server ran until getting killed, nice!")
		} else {
			logger.WithError(err).Warn("Unknown error, server failed")
		}
	}
}

func installVeracodeGem(repoFolder string, rubyVersion Version) error {
	if rubyVersion.Major < 2 || (rubyVersion.Major == 2 && rubyVersion.Minor <= 4) {
		cmd := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "add", "rubyzip", "--version", "~>1.0", "--source", "https://rubygems.org")
		cmd.Dir = repoFolder
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		logger.Info("Ruby version < 2.4 detected, installing RubyZip 1.0")

		err := cmd.Run()
		if err != nil {
			logger.WithError(err).Errorf("failed to rvm add rubyzip")
			return fmt.Errorf("failed to bundle add rubyzip", rubyVersion.String())
		}
	}

	cmd2 := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "show", "veracode")
	cmd2.Dir = repoFolder

	logger.Info("Checking for existence of 'veracode' gem")

	output, err := cmd2.CombinedOutput()
	if err != nil {
		logger.WithError(err).Errorf("bundle show veracode failed, assuming it's not installed yet, output: %s", output)

		cmd := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "add", "veracode", "--source", "https://rubygems.org")
		cmd.Dir = repoFolder
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		logger.Info("Installing veracode gem with Bundler")

		err = cmd.Run()
		if err != nil {
			logger.WithError(err).Errorf("failed to rvm add veracode")
			return fmt.Errorf("failed to bundle add veracode")
		}
	} else {
		logger.Infof("Veracode gem already exists, skipping install, output of bundle show: %s", output)
	}

	return nil
}

func runVeracodePrepare(repoFolder string, rubyVersion Version) error {
	cmd := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "veracode", "prepare", "-vD")
	cmd.Dir = repoFolder
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RAILS_ENV=development")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Info("Running Veracode Prepare")

	err := cmd.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to run veracode prepare")
		return fmt.Errorf("failed run veracode prepare")
	}

	logger.Info("All done!")
	return nil
}
