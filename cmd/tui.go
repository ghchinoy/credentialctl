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
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui"
	"github.com/spf13/cobra"
)

var (
	tuiRecursive bool
)

var tuiCmd = &cobra.Command{
	Use:     "tui [path]",
	GroupID: "interactive",
	Short:   "Launch the interactive Bubble Tea terminal user interface",
	Long: `TUI launches a full-screen terminal interface for reviewing folder validation results,
filtering assets by signature status, and drilling down into individual file C2PA manifest details.`,
	Example: `  # Open TUI for current working directory
  credentialctl tui

  # Open TUI for a specific media directory
  credentialctl tui /path/to/media

  # Open TUI recursively scanning subdirectories
  credentialctl tui /path/to/media -r

  # Open TUI directly inspecting a single media asset
  credentialctl tui photo.jpg`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		targetPath := "."
		if len(args) > 0 {
			targetPath = args[0]
		}

		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}

		fi, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path '%s' not found.\nHint: Provide a valid file or folder path.", absPath)
			}
			return fmt.Errorf("cannot access path '%s': %w", absPath, err)
		}

		var opts []credentio.Option
		opts = append(opts, credentio.WithSkipTrustChecks(skipTrustChecks))

		validator, err := engine.NewValidatorService(opts...)
		if err != nil {
			return fmt.Errorf("validator initialization error: %w", err)
		}
		defer validator.Close()

		var appModel ui.AppModel
		if !fi.IsDir() {
			item := engine.FileItem{
				Path:      absPath,
				Filename:  fi.Name(),
				SizeBytes: fi.Size(),
				MediaType: engine.DetectMediaType(absPath),
			}
			validator.ValidateItem(&item)
			appModel = ui.NewInspectAppModel(item, validator)
		} else {
			appModel = ui.NewAppModel(absPath, tuiRecursive, validator)
		}

		p := tea.NewProgram(appModel, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI execution error: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().BoolVarP(&tuiRecursive, "recursive", "r", false, "Scan subdirectories recursively")
}
