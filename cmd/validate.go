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

	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
	"github.com/spf13/cobra"
)

var (
	validateJSON      bool
	validateMediaType string
)

type jsonValidateOutput struct {
	AssetPath           string                     `json:"asset_path"`
	ByteSize            int64                      `json:"byte_size"`
	MediaType           string                     `json:"media_type"`
	EngineID            string                     `json:"engine_id"`
	HasCredentials      bool                       `json:"has_credentials"`
	Badge               string                     `json:"badge"`
	SpecVersion         string                     `json:"spec_version,omitempty"`
	ElapsedSeconds      float64                    `json:"elapsed_seconds"`
	CoreSeconds         float64                    `json:"core_seconds,omitempty"`
	ActiveManifest      *credentio.Manifest        `json:"active_manifest,omitempty"`
	IngredientManifests []credentio.Manifest       `json:"ingredient_manifests,omitempty"`
}

var validateCmd = &cobra.Command{
	Use:     "validate <file_path>",
	GroupID: "validation",
	Short:   "Validate C2PA content credentials in a media asset",
	Long: `Validate verifies the C2PA manifest, digital signatures, certificate chains,
and assertions embedded within a single target media asset file.`,
	Example: `  # Validate a photo with human-readable semantic output
  credentialctl validate photo.jpg

  # Validate with structured JSON output for machine / agent consumption
  credentialctl validate asset.mp4 --json

  # Explicitly specify the MIME media type
  credentialctl validate custom_image.bin --media-type image/webp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("invalid file path: %w\nHint: Verify that the path does not contain illegal characters.", err)
		}

		fi, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("file not found at '%s'.\nHint: Check the filename spelling or provide an absolute path.", absPath)
			}
			return fmt.Errorf("cannot access file '%s': %w", absPath, err)
		}

		if fi.IsDir() {
			return fmt.Errorf("path '%s' is a directory, not a file.\nHint: Use 'credentialctl folder %s' to validate directories.", absPath, filePath)
		}

		if validateMediaType == "" {
			validateMediaType = engine.DetectMediaType(absPath)
		}

		var opts []credentio.Option
		opts = append(opts, credentio.WithSkipTrustChecks(skipTrustChecks))

		validator, err := engine.NewValidatorService(opts...)
		if err != nil {
			return fmt.Errorf("validator initialization error: %w", err)
		}
		defer validator.Close()

		report, err := validator.ValidateFile(absPath, validateMediaType)
		if err != nil {
			return fmt.Errorf("validation execution failed: %w", err)
		}

		if validateJSON {
			out := jsonValidateOutput{
				AssetPath:           absPath,
				ByteSize:            fi.Size(),
				MediaType:           report.MediaType,
				EngineID:            report.EngineID,
				HasCredentials:      report.HasCredentials,
				Badge:               string(report.Badge()),
				SpecVersion:         report.SpecVersion,
				ElapsedSeconds:      report.ElapsedSeconds,
				CoreSeconds:         report.CoreSeconds,
				ActiveManifest:      report.ActiveManifest,
				IngredientManifests: report.IngredientManifests,
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		} else {
			renderHumanValidation(report, absPath, fi.Size())
		}

		// Exit with status code corresponding to validity state
		switch report.Badge() {
		case credentio.BadgeSigned:
			os.Exit(0)
		case credentio.BadgeUnsigned:
			os.Exit(1)
		case credentio.BadgeInvalid:
			os.Exit(2)
		default:
			os.Exit(0)
		}
		return nil
	},
}

func renderHumanValidation(report *credentio.ProvenanceReport, absPath string, sizeBytes int64) {
	sizeMB := float64(sizeBytes) / (1024.0 * 1024.0)

	var sb strings.Builder
	sb.WriteString(theme.StyleHeader.Render(" C2PA VALIDATION REPORT ") + "\n\n")

	renderRow := func(label, value string) {
		lbl := theme.StyleMuted.Width(16).Bold(true).Render(label + ":")
		sb.WriteString(fmt.Sprintf("  %s %s\n", lbl, value))
	}

	renderRow("Asset", fmt.Sprintf("%s (%.2f MB, %s)", filepath.Base(absPath), sizeMB, report.MediaType))
	renderRow("Path", absPath)
	renderRow("Status", theme.RenderBadge(report.Badge()))

	if report.HasCredentials && report.ActiveManifest != nil {
		m := report.ActiveManifest
		gen := m.ClaimGenerator
		if gen == "" {
			gen = "—"
		}
		issuer := "—"
		if m.Signature != nil && m.Signature.Issuer != "" {
			issuer = m.Signature.Issuer
		}
		spec := report.SpecVersion
		if spec == "" {
			spec = "—"
		}
		renderRow("Generator", theme.StyleAccent.Render(gen))
		renderRow("Signer", theme.StylePass.Render(issuer))
		renderRow("Format/Spec", fmt.Sprintf("%s (C2PA %s)", m.Format, spec))
		renderRow("Assertions", fmt.Sprintf("%d attached", len(m.Assertions)))
		renderRow("Validation", fmt.Sprintf("%d reported", len(m.ValidationStatuses)))
	}

	if report.CoreSeconds > 0 {
		renderRow("Core Time", fmt.Sprintf("%.2f ms", report.CoreSeconds*1000.0))
	}
	renderRow("Wall Time", fmt.Sprintf("%.2f ms", report.ElapsedSeconds*1000.0))

	fmt.Println(sb.String())
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolVar(&validateJSON, "json", false, "Output structured JSON for machine/agent parsing")
	validateCmd.Flags().StringVar(&validateMediaType, "media-type", "", "Explicit IANA media type (e.g. image/jpeg)")
}
