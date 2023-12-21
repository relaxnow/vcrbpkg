/*
Copyright © 2023 Boy Baukema <vcrbpkg@relaxnow.nl>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"net/url"
	"os"

	"github.com/relaxnow/vcrbpkg/internal/pkg/logger"
	"github.com/relaxnow/vcrbpkg/internal/pkg/vcrbpkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vcrbpkg [url or filepath]",
	Short: "Package Ruby on Rails applications for Veracode Static Analysis",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs, validateURLorFilePath),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Your main command logic here
		fmt.Println("Hello from your CLI tool!")
		return vcrbpkg.Package(args)
	},
	Example: "vcrbpkg /folder/to/clone OR vcrbpkg https://github.com/user/repo",
}

func validateURLorFilePath(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("requires a single argument")
	}

	input := args[0]

	// Check if it's a valid URL
	_, err := url.ParseRequestURI(input)
	if err == nil {
		return nil // It's a valid URL
	}

	// Check if it's a valid file path
	_, err = os.Stat(input)
	if err == nil {
		return nil // It's a valid file path
	}

	// If neither URL nor file path, return an error
	return fmt.Errorf("invalid input: %s must be either a URL or a file path", input)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var logLevel string

func init() {
	// Add a flag to set the log level
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error, fatal, panic)")
}

func configureLogger() {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logger.Fatalf("Error setting log level: %v", err)
	}
	logger.SetLevel(level)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	configureLogger()

	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}
