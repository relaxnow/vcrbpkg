package vcrbpkg

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/relaxnow/vcrbpkg/internal/pkg/logger"
)

func determineRubyVersion(repoFolder string) Version {
	version := checkRubyVersionFile(filepath.Join(repoFolder, ".ruby-version"))
	if version != (Version{}) {
		return version
	}

	version = checkRubyVersionFile(filepath.Join(repoFolder, ".ruby-version.sample"))
	if version != (Version{}) {
		return version
	}

	version = checkGemFile(repoFolder)
	if version != (Version{}) {
		return version
	}

	// TODO: try parsing .github/workflows for something like
	// https://github.com/ManageIQ/manageiq/blob/master/.github/workflows/ci.yaml#L16-L17

	logger.Warn("Unable to find Ruby version, giving it a try with latest 3.2.2")
	return Version{Major: 3, Minor: 2, Patch: 2}
}

func checkRubyVersionFile(filePath string) Version {
	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		logger.Infof("File does not exist: %s", filePath)
		return Version{}
	} else if err != nil {
		logger.WithError(err).Warnf("Error stat-ing file '%s', highly unusual but we'll try to keep on without", filePath)
		return Version{}
	}

	// Check if the file is readable
	content := readRubyVersionFile(filePath)
	if content == "" {
		logger.Warnf("File %s exists but cannot be read", filePath)
		return Version{}
	}

	// Use regular expression to validate version format
	rubyVersionRegex := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	match := rubyVersionRegex.FindStringSubmatch(content)
	if match == nil {
		logger.Warnf("Unrecognized format in %s: '%s' does not match x.y.z format", filePath, content)
		return Version{}
	}

	// Parse version string into Version struct
	return parseRubyVersion(content)
}

func readRubyVersionFile(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		logger.WithError(err).Warnf("Error reading file %s", filePath)
		return ""
	}

	return strings.TrimSpace(string(content))
}

func checkGemFile(filePath string) Version {
	// Open the Gemfile
	gemfilePath := filepath.Join(filePath, "Gemfile")
	file, err := os.Open(gemfilePath)
	if err != nil {
		logger.WithError(err).Warnf("Unable to find Gemfile on path %s", gemfilePath)
		return Version{}
	}
	defer file.Close()

	// Create a regular expression pattern for matching "ruby" and a version number
	rubyPattern := regexp.MustCompile(`^ruby.+(\d+)\.(\d+)\.(\d+).*$`)

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Loop through each line
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line starts with "ruby" and has a version number
		match := rubyPattern.FindStringSubmatch(line)
		if len(match) == 4 {
			// Extract version numbers
			major := match[1]
			minor := match[2]
			patch := match[3]

			// Print the version
			logger.Infof("Found Ruby version: %s.%s.%s\n", major, minor, patch)

			return parseRubyVersion(major + "." + minor + "." + patch)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		logger.WithError(err).Warn("Error reading Gemfile")
	}
	return Version{}
}

func parseRubyVersion(versionStr string) Version {
	// Split version string into major, minor, and patch parts
	versionParts := strings.Split(versionStr, ".")

	// Convert parts to integers
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		logger.WithError(err).Warnf("Conversion to int failed on 1/3 '%s'", versionParts[0])
		return Version{}
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		logger.WithError(err).Warnf("Conversion to int failed on 2/3 '%s'", versionParts[1])
		return Version{}
	}

	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		logger.WithError(err).Warnf("Conversion to int failed on 3/3 '%s'", versionParts[2])
		return Version{}
	}

	return Version{Major: major, Minor: minor, Patch: patch}
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v1 Version) Equals(v2 Version) bool {
	return v1.Major == v2.Major && v1.Minor == v2.Major && v1.Patch == v2.Patch
}

func checkIsSupportedRubyVersion(rubyVersion Version) {
	if rubyVersion.Equals(Version{Major: 1, Minor: 9, Patch: 3}) {
		logger.Infof("Supported Ruby 1 version %s", rubyVersion.String())
		return
	}

	if rubyVersion.Major == 2 && rubyVersion.Minor == 0 {
		logger.Infof("Supported Ruby 2.0 version %s", rubyVersion.String())
		return
	}

	if rubyVersion.Major == 2 && rubyVersion.Minor == 0 {
		logger.Infof("Supported Ruby 2.1 version %s", rubyVersion.String())
		return
	}

	if rubyVersion.Major == 2 && rubyVersion.Minor >= 3 && rubyVersion.Minor <= 7 {
		logger.Infof("Supported Ruby 2.3-2.7 version %s", rubyVersion.String())
		return
	}

	if rubyVersion.Major == 3 && rubyVersion.Minor == 0 {
		logger.Infof("Supported Ruby 3.0 version %s", rubyVersion.String())
		return
	}

	if rubyVersion.Major == 3 && rubyVersion.Minor == 1 {
		logger.Infof("Supported Ruby 3.1 version %s", rubyVersion.String())
		return
	}

	if rubyVersion.Major == 3 && rubyVersion.Minor == 2 {
		logger.Infof("Supported Ruby 3.2 version %s", rubyVersion.String())
		return
	}

	logger.Warnf("Ruby version %s is not supported! Trying anyway...", rubyVersion.String())
}
