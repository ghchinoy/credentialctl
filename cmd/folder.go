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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
	"github.com/spf13/cobra"
)

var (
	folderJSON      bool
	folderRecursive bool
)

var folderCmd = &cobra.Command{
	Use:     "folder <directory_path>",
	Aliases: []string{"scan", "dir"},
	GroupID: "validation",
	Short:   "Validate all media assets in a directory",
	Long: `Folder scans a target directory for supported C2PA media assets (JPEG, PNG, WebP,
AVIF, MP4, MOV, etc.), validates each asset in batch, and produces a summary report.`,
	Example: `  # Scan a folder of images
  credentialctl folder ./photos

  # Scan recursively including subdirectories
  credentialctl folder /path/to/media -r

  # Output full batch report as structured JSON
  credentialctl folder ./downloads --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dirPath := args[0]
		absPath, err := filepath.Abs(dirPath)
		if err != nil {
			return fmt.Errorf("invalid directory path: %w", err)
		}

		fi, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("directory not found at '%s'.\nHint: Check the path or ensure the folder exists.", absPath)
			}
			return fmt.Errorf("cannot access directory '%s': %w", absPath, err)
		}

		if !fi.IsDir() {
			return fmt.Errorf("path '%s' is a file, not a directory.\nHint: Use 'credentialctl validate %s' for single file validation.", absPath, dirPath)
		}

		var opts []credentio.Option
		opts = append(opts, credentio.WithSkipTrustChecks(skipTrustChecks))

		validator, err := engine.NewValidatorService(opts...)
		if err != nil {
			return fmt.Errorf("validator initialization error: %w", err)
		}
		defer validator.Close()

		summary, err := validator.ScanFolder(absPath, folderRecursive, nil)
		if err != nil {
			return fmt.Errorf("failed scanning folder: %w", err)
		}

		if folderJSON {
			data, err := json.MarshalIndent(summary, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			renderHumanFolderSummary(summary)
		}

		if summary.InvalidCount > 0 {
			os.Exit(2)
		} else if summary.SignedCount == 0 && summary.TotalFiles > 0 {
			os.Exit(1)
		}
		return nil
	},
}

func renderHumanFolderSummary(summary *engine.FolderScanSummary) {
	var sb strings.Builder
	sb.WriteString(theme.StyleHeader.Render(fmt.Sprintf(" FOLDER VALIDATION SUMMARY: %s ", summary.Directory)) + "\n\n")

	stats := fmt.Sprintf(" Total Files: %d  •  Signed: %d  •  Unsigned: %d  •  Invalid: %d  •  Duration: %.2fs",
		summary.TotalFiles, summary.SignedCount, summary.UnsignedCount, summary.InvalidCount, summary.DurationSec)
	sb.WriteString(theme.StyleCard.Padding(0, 1).Render(stats) + "\n\n")

	if summary.TotalFiles == 0 {
		sb.WriteString(theme.StyleMuted.Render(fmt.Sprintf("  No supported media files found.\n  Supported extensions: %s\n", engine.SupportedExtensionsList())))
		fmt.Println(sb.String())
		return
	}

	colStatus := lipgloss.NewStyle().Width(14).Bold(true).Render("STATUS")
	colName := lipgloss.NewStyle().Width(28).Bold(true).Render("FILENAME")
	colSize := lipgloss.NewStyle().Width(10).Bold(true).Render("SIZE")
	colMime := lipgloss.NewStyle().Width(14).Bold(true).Render("TYPE")
	colSigner := lipgloss.NewStyle().Width(24).Bold(true).Render("SIGNER / GENERATOR")

	tableHeader := lipgloss.JoinHorizontal(lipgloss.Left, colStatus, colName, colSize, colMime, colSigner)
	sb.WriteString(" " + theme.StyleMuted.Render(tableHeader) + "\n")
	sb.WriteString(" " + theme.StyleMuted.Render(strings.Repeat("─", 90)) + "\n")

	for _, item := range summary.Files {
		statusCell := theme.RenderBadge(item.Badge())

		nameStr := item.Filename
		if len(nameStr) > 26 {
			nameStr = nameStr[:23] + "..."
		}
		nameCell := lipgloss.NewStyle().Width(28).Render(nameStr)
		sizeCell := lipgloss.NewStyle().Width(10).Render(item.HumanSize())

		mimeStr := item.MediaType
		if len(mimeStr) > 12 {
			mimeStr = mimeStr[:11] + "…"
		}
		mimeCell := theme.StyleMuted.Width(14).Render(mimeStr)

		signerGen := item.Signer()
		if signerGen == "—" {
			signerGen = item.Generator()
		}
		if len(signerGen) > 22 {
			signerGen = signerGen[:20] + "..."
		}
		signerCell := theme.StyleID.Width(24).Render(signerGen)

		rowStr := lipgloss.JoinHorizontal(lipgloss.Left,
			lipgloss.NewStyle().Width(14).Render(statusCell),
			nameCell,
			sizeCell,
			mimeCell,
			signerCell,
		)
		sb.WriteString("  " + rowStr + "\n")
	}

	fmt.Println(sb.String())
}

func init() {
	rootCmd.AddCommand(folderCmd)
	folderCmd.Flags().BoolVar(&folderJSON, "json", false, "Output structured JSON for batch results")
	folderCmd.Flags().BoolVarP(&folderRecursive, "recursive", "r", false, "Recursively scan subdirectories")
}
