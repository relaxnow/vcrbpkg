package app

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func determineRubyVersion(repoFolder string) Version {
	// TODO: try:
	// 1. .ruby-version
	// 2. .ruby-version.sample
	// 3. bundle platform
	// 4. .github/workflows ruby-version: '3.0'
	return Version{Major: 3, Minor: 0, Patch: 0}
}

func checkRubyVersionFile(filePath string) (Version, error) {
	// Check if the file exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return Version{}, fmt.Errorf("File %s does not exist", filePath)
	} else if err != nil {
		return Version{}, err
	}

	// Check if the file is readable
	content, err := readRubyVersionFile(filePath)
	if err != nil {
		return Version{}, err
	}

	// Use regular expression to validate version format
	rubyVersionRegex := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	match := rubyVersionRegex.FindStringSubmatch(content)
	if match == nil {
		return Version{}, fmt.Errorf("Invalid version format in %s", filePath)
	}

	// Parse version string into Version struct
	version, err := parseRubyVersion(content)
	if err != nil {
		return Version{}, err
	}

	return version, nil
}

func readRubyVersionFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(content)), nil
}

func parseRubyVersion(versionStr string) (Version, error) {
	// Split version string into major, minor, and patch parts
	versionParts := strings.Split(versionStr, ".")

	// Convert parts to integers
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return Version{}, err
	}

	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return Version{}, err
	}

	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return Version{}, err
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
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

func ensureIsSupportedRubyVersion(rubyVersion Version) {
	// TODO: check against versions in https://docs.veracode.com/r/compilation_ruby
}
