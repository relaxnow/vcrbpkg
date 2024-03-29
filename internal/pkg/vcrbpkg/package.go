package vcrbpkg

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"github.com/relaxnow/vcrbpkg/internal/pkg/logger"
)

func Package(args []string, outFile string) error {
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

	var input string
	if len(args) == 0 {
		input = "."
	} else {
		input = args[0]
	}
	if isAlreadyDirectory(input) {
		repoFolder = input
	} else {
		repoFolder, err = cloneRepo(input)
		if err != nil {
			return err
		}
	}

	if err = ensureHasRailsStructure(repoFolder); err != nil {
		return err
	}

	rubyVersion = determineRubyVersion(repoFolder)
	checkIsSupportedRubyVersion(rubyVersion)
	if err = rvmInstallRuby(repoFolder, rubyVersion); err != nil {
		return err
	}
	checkIsSupportedRailsVersion(repoFolder, rubyVersion)
	if err = installVeracodeGem(repoFolder, rubyVersion); err != nil {
		return err
	}
	railsEnv := testForBestEnv(repoFolder, rubyVersion)

	packagedFile, err := runVeracodePrepare(repoFolder, rubyVersion, railsEnv)
	if err != nil {
		return err
	}
	if outFile != "" {
		copyFile(packagedFile, outFile)
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

func ensureHasRailsStructure(repoFolder string) error {
	requiredFiles := []string{"app", "config", "public", "Gemfile"}

	// Check if the required folders exist inside repoFolder
	for _, requiredFile := range requiredFiles {
		path := filepath.Join(repoFolder, requiredFile)
		_, err := os.Stat(path)
		if err != nil {
			logger.WithError(err).Errorf("Directory does not have Rails structure, file %s not found", requiredFile)
			return fmt.Errorf("directory does not have Rails structure, can only package Rails apps. file %s not found", requiredFile)
		}
	}
	return nil
}

func checkIsSupportedRailsVersion(repoFolder string, rubyVersion Version) {
	logger.Info("Detecting Rails version with Bundler")

	cmd := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "show", "rails")
	cmd.Dir = repoFolder
	var so saveOutput
	cmd.Stdout = &so
	cmd.Stderr = &so
	err := cmd.Run()

	if err != nil {
		logger.WithError(err).Errorf("failed to bundle show rails")
		logger.Warn("failed to run bundle show rails, unable to verify rails version, hoping for the best and continuing")
		return
	}

	// Define a regular expression pattern to match the version number
	re := regexp.MustCompile(`rails-(\d+\.\d+\.\d+)`)

	// Find the first match in the input string
	match := re.FindStringSubmatch(string(so.savedOutput))

	// Check if a match is found
	if len(match) < 2 {
		logger.Warnf("Version number not found in the input string")
		logger.Warn("failed to parse output of bundle show rails, unable to verify rails version, hoping for the best and continuing")
		return
	}

	// The version number is captured in the first submatch group
	versionNumber := match[1]
	railsVersion := parseRubyVersion(versionNumber)
	if railsVersion.Major >= 3 && railsVersion.Major <= 6 || (railsVersion.Major == 7 && railsVersion.Minor == 0) {
		logger.Infof("Veracode Static Analysis supported Rails version: %s", versionNumber)
	} else {
		logger.Warnf("Veracode Static Analysis unsupported Rails version %s, hoping for the best and continuing", versionNumber)
	}
}

func rvmInstallRuby(repoFolder string, rubyVersion Version) error {
	var rvmInstallCmd *exec.Cmd

	// https://wiki.archlinux.org/title/RVM#RVM_uses_wrong_OpenSSL_version
	if rubyVersion.LowerThan(parseRubyVersion("3.0.0")) {
		// Install OpenSSL in RVM because the system OpenSSL might be incompatible
		opensslInstallCmd := exec.Command("rvm", "pkg", "install", "openssl")
		opensslInstallCmd.Dir = repoFolder
		var opensslInstallSavedOutput saveOutput
		opensslInstallCmd.Stdout = &opensslInstallSavedOutput
		opensslInstallCmd.Stderr = &opensslInstallSavedOutput

		logger.Info("Installing OpenSSL for RVM")

		err := opensslInstallCmd.Run()
		if err != nil {
			logger.WithError(err).Warnf("failed to install openssl for rvm")
		}

		re := regexp.MustCompile(`Installing openssl to (.+)\.\.\.`)
		match := re.FindStringSubmatch(string(opensslInstallSavedOutput.savedOutput))

		var filePath string
		if len(match) < 2 {
			logger.Warn("No path after OpenSSL install? Guessing '/usr/local/rvm/usr/'")
			filePath = "/usr/local/rvm/usr/"
		} else {
			filePath = match[1]
			// Run 'rvm install' command
			// TODO: make with-openssl-dir use output of prev command
			logger.Infof("Running rvm install --autolibs=disabled --with-openssl-dir=%s %s", filePath, rubyVersion.String())
		}
		rvmInstallCmd = exec.Command("rvm", "install", "--autolibs=disabled", "--with-openssl-dir="+filePath, rubyVersion.String())
	} else {
		rvmInstallCmd = exec.Command("rvm", "install", rubyVersion.String())
	}
	rvmInstallCmd.Dir = repoFolder
	var rvmInstallSavedOutput saveOutput
	rvmInstallCmd.Stdout = &rvmInstallSavedOutput
	rvmInstallCmd.Stderr = &rvmInstallSavedOutput

	logger.Info("Installing Ruby version with RVM, this may take a while")

	err := rvmInstallCmd.Run()

	if err != nil {
		logger.WithError(err).Error("failed to  rvm install")

		re := regexp.MustCompile(`please read (.+\.log)`)
		match := re.FindStringSubmatch(string(rvmInstallSavedOutput.savedOutput))

		if len(match) < 2 {
			logger.Warn("No log file to read?")
		} else {
			filePath := match[1]

			logFileContents, err := os.ReadFile(filePath)
			if err != nil {
				logger.Fatalf("unable to read file %s: %v", filePath, err)
			}
			logger.Errorf("Make output: %s", logFileContents)
		}

		return fmt.Errorf("failed to rvm install %s", rubyVersion.String())
	}

	logger.Info("Creating a veracode gemset")

	// Run 'rvm install' command
	cmd3 := exec.Command("rvm", rubyVersion.String(), "do", "rvm", "gemset", "create", "veracode")
	cmd3.Dir = repoFolder
	cmd3.Stdout = os.Stdout
	cmd3.Stderr = os.Stderr

	err = cmd3.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to create gemset")
		return fmt.Errorf("failed to create gemset for ruby version: %s", rubyVersion.String())
	}

	return nil
}

// Test which environment works best to by running `rails server`
// production is best because it does not have all the develoment tooling
// but then typically production does not work without some setup.
func testForBestEnv(repoFolder string, rubyVersion Version) string {
	testEnvs := []string{"production", "development", "test"}
	for _, testEnv := range testEnvs {
		var cmd4 *exec.Cmd
		if testEnv == "production" {
			cmd4 = exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "install", "--without", "development", "test")
		} else {
			cmd4 = exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "bundle", "install")
		}
		cmd4.Dir = repoFolder
		cmd4.Stdout = os.Stdout
		cmd4.Stderr = os.Stderr

		logger.Info("Doing Bundle Install")

		err := cmd4.Run()
		if err != nil {
			logger.WithError(err).Warnf("failed to do bundle install, trying to run server anyway, will probably fail")
		}

		if testWithEnv(repoFolder, rubyVersion, testEnv) {
			logger.Infof("Successfully verfied Rails environment %s, using it for Veracode Prepare", testEnv)
			return testEnv
		}
	}

	logger.Warn("Testing failed for all known environments, trying our luck with production")
	return "production"
}

func testWithEnv(repoFolder string, rubyVersion Version, railsEnv string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rvm", rubyVersion.String()+"@veracode", "do", "rails", "server")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RAILS_ENV="+railsEnv)
	cmd.Dir = repoFolder
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	logger.Infof("Running rails server in %s", railsEnv)

	if err := cmd.Run(); err != nil {
		logger.WithError(err).Info("rails server error")
		if err.Error() == "signal: killed" {
			logger.Infof("Server ran until getting killed, nice!")
			return true
		} else {
			logger.WithError(err).Warn("Unknown error, server failed")
			return false
		}
	}
	logger.Warn("Rails server ran without error? That's unexpected.")
	return false
}

func installVeracodeGem(repoFolder string, rubyVersion Version) error {
	// TODO: What if rubyzip is already installed?
	if rubyVersion.Major < 2 || (rubyVersion.Major == 2 && rubyVersion.Minor <= 4) {
		cmd := exec.Command(
			"rvm", rubyVersion.String()+"@veracode", "do",
			"bundle", "add", "rubyzip",
			"--version", "~>1.0",
			"--source", "https://rubygems.org",
			"--skip-install")
		cmd.Dir = repoFolder
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		logger.Info("Ruby version < 2.4 detected, installing RubyZip 1.0")

		err := cmd.Run()
		if err != nil {
			logger.WithError(err).Errorf("failed to rvm add rubyzip")
			return fmt.Errorf("failed to bundle add rubyzip")
		}
	}

	logger.Info("Checking for existence of 'veracode' gem")
	cmd2 := exec.Command(
		"rvm", rubyVersion.String()+"@veracode", "do",
		"bundle", "show", "veracode")
	cmd2.Dir = repoFolder
	err := cmd2.Run()

	if err != nil {
		logger.WithError(err).Errorf("bundle show veracode failed, assuming it's not installed yet")

		logger.Info("Installing veracode gem with Bundler")
		cmd := exec.Command(
			"rvm", rubyVersion.String()+"@veracode", "do",
			"bundle", "add", "veracode",
			"--source", "https://rubygems.org",
			"--skip-install")
		cmd.Dir = repoFolder

		err = cmd.Run()
		if err != nil {
			logger.WithError(err).Errorf("failed to bundle add veracode")
			return fmt.Errorf("failed to bundle add veracode")
		}
	} else {
		logger.Infof("Veracode gem already exists, skipping install")
	}

	return nil
}

func runVeracodePrepare(repoFolder string, rubyVersion Version, railsEnv string) (string, error) {
	logger.Info("Running Veracode Prepare, this may take a while")

	cmd := exec.Command("rvm", rubyVersion.String()+"@veracode", "do", "veracode", "prepare", "-vD")
	cmd.Dir = repoFolder
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "RAILS_ENV="+railsEnv)
	var so saveOutput
	cmd.Stdout = &so
	cmd.Stderr = &so
	err := cmd.Run()
	if err != nil {
		logger.WithError(err).Errorf("failed to run veracode prepare")
		return "", fmt.Errorf("failed run veracode prepare")
	}

	veracodePrepareFile, err := extractFilePath(string(so.savedOutput))

	if err != nil {
		logger.WithError(err).Errorf("Did not find packaged file in veracode prepare output")
		return "", fmt.Errorf("did not find packaged file in veracode prepare output")
	}

	logger.Info("All done!")
	return filepath.Join(repoFolder, veracodePrepareFile), nil
}

func extractFilePath(output string) (string, error) {
	// Define a regular expression pattern to match the file path
	re := regexp.MustCompile(`Please upload\s+(\S+)`)

	// Find the first match in the output string
	match := re.FindStringSubmatch(output)

	// Check if a match is found
	if len(match) < 2 {
		return "", fmt.Errorf("file path not found in the output string")
	}

	// The file path is captured in the first submatch group
	filePath := match[1]

	return filePath, nil
}

func copyFile(packagedFile, outFile string) error {
	outFile, err := filepath.Abs(outFile)
	if err != nil {
		logger.WithError(err).Errorf("Unable to create absolute path for %s", outFile)
		return fmt.Errorf("unable to create absolute path for %s", outFile)
	}

	logger.Infof("Copying from %s => %s", packagedFile, outFile)
	// Open the source file
	src, err := os.Open(packagedFile)
	if err != nil {
		logger.WithError(err).Errorf("Error opening source file %s", packagedFile)
		return fmt.Errorf("error opening source file: %v", err)
	}
	defer src.Close()

	// Create or truncate the destination file
	dst, err := os.Create(outFile)
	if err != nil {
		logger.WithError(err).Errorf("Error creating destination file %s", outFile)
		return fmt.Errorf("error creating destination file: %s", outFile)
	}
	defer dst.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(dst, src)
	if err != nil {
		logger.WithError(err).Errorf("Error copying file contents")
		return fmt.Errorf("error copying file contents")
	}

	return nil
}
