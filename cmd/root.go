// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
	"github.com/spf13/cobra"
)

var (
	skipTrustChecks bool
)

var rootCmd = &cobra.Command{
	Use:   "credentialctl [command|path]",
	Short: "C2PA Content Credentials validation and inspection tool",
	Long: `credentialctl is an agent-aware CLI and interactive Bubble Tea terminal user interface
for validating C2PA (Coalition for Content Provenance and Authenticity) manifests in media files.

Powered by Google Credentio high-performance native C-ABI bindings.`,
	Example: `  # Launch interactive TUI in the current directory
  credentialctl

  # Launch interactive TUI reviewing a specific folder
  credentialctl tui /path/to/media

  # Validate a single file and output human report
  credentialctl validate image.jpg

  # Validate a file and output structured JSON for agent consumption
  credentialctl validate photo.png --json

  # Scan and validate all media assets in a directory
  credentialctl folder ./samples --recursive`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		fi, err := os.Stat(targetPath)
		if err != nil {
			return fmt.Errorf("target path '%s' does not exist.\nHint: Provide a valid file or directory path.", targetPath)
		}

		var opts []credentio.Option
		opts = append(opts, credentio.WithSkipTrustChecks(skipTrustChecks))

		validator, err := engine.NewValidatorService(opts...)
		if err != nil {
			return fmt.Errorf("failed to initialize native validator: %w\nHint: Verify that libcredentio_c shared library is accessible in your library path.", err)
		}
		defer validator.Close()

		var appModel ui.AppModel
		if !fi.IsDir() {
			item := engine.FileItem{
				Path:      targetPath,
				Filename:  fi.Name(),
				SizeBytes: fi.Size(),
				MediaType: engine.DetectMediaType(targetPath),
			}
			validator.ValidateItem(&item)
			appModel = ui.NewInspectAppModel(item, validator)
		} else {
			appModel = ui.NewAppModel(targetPath, false, validator)
		}

		p := tea.NewProgram(appModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}
		return nil
	},
}

func init() {
	// Command Groups adhering to Agent-Aware standards
	rootCmd.AddGroup(&cobra.Group{
		ID:    "validation",
		Title: "Validation Commands:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "inspection",
		Title: "Inspection Commands:",
	})
	rootCmd.AddGroup(&cobra.Group{
		ID:    "interactive",
		Title: "Interactive TUI Commands:",
	})

	rootCmd.Version = version
	rootCmd.SetVersionTemplate("credentialctl version {{.Version}}\n")
	rootCmd.Flags().BoolP("version", "v", false, "Show credentialctl version")
	rootCmd.SilenceErrors = true

	rootCmd.PersistentFlags().BoolVar(&skipTrustChecks, "skip-trust-checks", true, "Skip certificate trust checks for local verification")
}

// Execute runs the root CLI command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", theme.StyleFail.Render("Error: "+err.Error()))
		os.Exit(1)
	}
}
