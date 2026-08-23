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
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
	"github.com/spf13/cobra"
)

var (
	inspectJSON bool
	inspectRaw  bool
	inspectTUI  bool
)

var inspectCmd = &cobra.Command{
	Use:     "inspect <file_path>",
	GroupID: "inspection",
	Short:   "Deep inspection of C2PA manifest, claims, assertions, and signatures",
	Long: `Inspect provides a granular view into the C2PA manifest structure of a media asset,
including assertion breakdowns (actions, data hashes, AI training mining metadata),
cryptographic signature verification, and validation status explanations.`,
	Example: `  # Deep inspection of a media asset
  credentialctl inspect photo.jpg

  # Inspect interactively in the terminal UI
  credentialctl inspect photo.jpg --tui

  # Output full provenance report as JSON
  credentialctl inspect asset.png --json

  # Print raw crJSON payload extracted from the manifest
  credentialctl inspect asset.png --raw`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}

		fi, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found at '%s'.\nHint: Check the path or ensure the file exists.", absPath)
			}
			return fmt.Errorf("cannot access file '%s': %w", absPath, err)
		}

		if fi.IsDir() {
			return fmt.Errorf("path '%s' is a directory.\nHint: Use 'credentialctl folder %s' to view directories.", absPath, filePath)
		}

		var opts []credentio.Option
		opts = append(opts, credentio.WithSkipTrustChecks(skipTrustChecks))

		validator, err := engine.NewValidatorService(opts...)
		if err != nil {
			return fmt.Errorf("validator initialization error: %w", err)
		}
		defer validator.Close()

		mediaType := engine.DetectMediaType(absPath)
		report, err := validator.ValidateFile(absPath, mediaType)
		if err != nil {
			return fmt.Errorf("validation execution failed: %w", err)
		}

		if inspectTUI {
			item := engine.FileItem{
				Path:      absPath,
				Filename:  fi.Name(),
				SizeBytes: fi.Size(),
				MediaType: mediaType,
				Report:    report,
				Validated: true,
			}
			appModel := ui.NewInspectAppModel(item, validator)
			p := tea.NewProgram(appModel, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("TUI error: %w", err)
			}
			return nil
		}

		if inspectRaw {
			if report.RawJSON == "" {
				fmt.Println("{}")
				return nil
			}
			var unmarshaled interface{}
			if err := json.Unmarshal([]byte(report.RawJSON), &unmarshaled); err == nil {
				indented, _ := json.MarshalIndent(unmarshaled, "", "  ")
				fmt.Println(string(indented))
				return nil
			}
			fmt.Println(report.RawJSON)
			return nil
		}

		if inspectJSON {
			data, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
			return nil
		}

		renderHumanInspection(report, absPath, fi.Size())
		return nil
	},
}

func renderHumanInspection(rep *credentio.ProvenanceReport, absPath string, sizeBytes int64) {
	var sb strings.Builder
	sb.WriteString(theme.StyleHeader.Render(fmt.Sprintf(" C2PA MANIFEST INSPECTION: %s ", filepath.Base(absPath))) + "\n\n")

	renderRow := func(label, value string) {
		lbl := theme.StyleMuted.Width(18).Bold(true).Render(label + ":")
		sb.WriteString(fmt.Sprintf("  %s %s\n", lbl, value))
	}

	sizeMB := float64(sizeBytes) / (1024.0 * 1024.0)
	renderRow("Path", absPath)
	renderRow("Size", fmt.Sprintf("%.2f MB (%d bytes)", sizeMB, sizeBytes))
	renderRow("Media Type", rep.MediaType)
	renderRow("Status", theme.RenderBadge(rep.Badge()))

	if !rep.HasCredentials || rep.ActiveManifest == nil {
		sb.WriteString("\n" + theme.StyleWarn.Render("  [No C2PA Manifest Store found in this asset]") + "\n")
		fmt.Println(sb.String())
		return
	}

	mft := rep.ActiveManifest
	sb.WriteString("\n" + theme.StyleSubHeader.Render("MANIFEST METADATA") + "\n")
	renderRow("Label", theme.StyleID.Render(mft.Label))
	if mft.Title != "" {
		renderRow("Title", mft.Title)
	}
	if mft.Format != "" {
		renderRow("Format", mft.Format)
	}
	if mft.ClaimGenerator != "" {
		renderRow("Claim Generator", theme.StyleAccent.Render(mft.ClaimGenerator))
	}
	if rep.SpecVersion != "" {
		renderRow("Spec Version", fmt.Sprintf("C2PA %s", rep.SpecVersion))
	}

	if mft.Signature != nil {
		sb.WriteString("\n" + theme.StyleSubHeader.Render("SIGNATURE & TRUST") + "\n")
		sig := mft.Signature
		issuer := sig.Issuer
		if issuer == "" {
			issuer = "—"
		}
		renderRow("Issuer / Signer", theme.StylePass.Render(issuer))
		alg := sig.Algorithm
		if alg == "" {
			alg = "—"
		}
		renderRow("Algorithm", theme.StyleID.Render(alg))
		if sig.Time != nil {
			renderRow("Signing Time", sig.Time.Format(time.RFC1123))
		}
		if sig.CertChainSummary != "" {
			renderRow("Cert Serial", theme.StyleMuted.Render(sig.CertChainSummary))
		}
	}

	if len(mft.Assertions) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s\n", theme.StyleSubHeader.Render(fmt.Sprintf("ASSERTIONS (%d)", len(mft.Assertions)))))
		for i, a := range mft.Assertions {
			kindBadge := lipgloss.NewStyle().
				Background(lipgloss.Color("#2E2E3E")).
				Foreground(theme.ColorID).
				Padding(0, 1).
				Render(string(a.Kind))

			sb.WriteString(fmt.Sprintf("  %d. %s  %s\n", i+1, theme.StyleAccent.Bold(true).Render(a.Label), kindBadge))
			if a.Summary != "" {
				sb.WriteString(fmt.Sprintf("     %s %s\n", theme.StyleMuted.Render("Summary:"), a.Summary))
			}
		}
	}

	if len(mft.ValidationStatuses) > 0 {
		sb.WriteString(fmt.Sprintf("\n%s\n", theme.StyleSubHeader.Render(fmt.Sprintf("VALIDATION STATUSES (%d)", len(mft.ValidationStatuses)))))
		for i, st := range mft.ValidationStatuses {
			sevBadge := theme.RenderSeverity(st.Severity)
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s\n", i+1, sevBadge, theme.StyleBold.Render(st.Code)))
			if st.Explanation != "" {
				sb.WriteString(fmt.Sprintf("     %s\n", st.Explanation))
			}
			if st.URL != "" {
				sb.WriteString(fmt.Sprintf("     %s %s\n", theme.StyleMuted.Render("Spec:"), theme.StyleID.Render(st.URL)))
			}
		}
	}

	fmt.Println(sb.String())
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.Flags().BoolVar(&inspectJSON, "json", false, "Output structured JSON manifest")
	inspectCmd.Flags().BoolVar(&inspectRaw, "raw", false, "Output unformatted or pretty raw crJSON")
	inspectCmd.Flags().BoolVar(&inspectTUI, "tui", false, "Open directly in interactive terminal inspector UI")
}
