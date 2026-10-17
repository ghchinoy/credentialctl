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
	signOutputPath   string
	signKeyPath      string
	signCertPath     string
	signTSAKeyPath   string
	signTSACertPath  string
	signMediaType    string
	signSourceType   string
	signGenerator    string
	signAIOptOut     bool
	signJSON         bool
)

type jsonSignOutput struct {
	InputPath         string               `json:"input_path"`
	OutputPath        string               `json:"output_path"`
	ByteSize          int                  `json:"byte_size"`
	MediaType         string               `json:"media_type"`
	DigitalSourceType string               `json:"digital_source_type,omitempty"`
	Generator         string               `json:"generator"`
	SigningCoreSec    float64              `json:"signing_core_seconds"`
	SigningElapsedSec float64              `json:"signing_elapsed_seconds"`
	VerifiedBadge     string               `json:"verified_badge"`
	ActiveManifest    *credentio.Manifest  `json:"active_manifest,omitempty"`
}

var signCmd = &cobra.Command{
	Use:     "sign <input_path> -o <output_path>",
	GroupID: "signing",
	Short:   "Embed and sign C2PA content credentials into a media asset",
	Long: `Sign embeds a C2PA manifest store into a media asset (JPEG, PNG, GIF, WebP, TIFF, WAV,
MP3, FLAC, MP4/MOV/HEIC/AVIF, PDF, or OOXML document), signs it in-process using the
Google Credentio C-ABI engine, and immediately verifies the signed output.`,
	Example: `  # Sign an image using an ephemeral C2PA test key/certificate
  credentialctl sign photo.png -o signed.png

  # Sign with explicit PKI credentials and AI generation source type
  credentialctl sign art.png -o art_signed.png \
    --key ./test_pki/claim_signer_key.pem \
    --cert ./test_pki/claim_signer_cert.pem \
    --source-type ai \
    --ai-opt-out

  # Sign and emit structured JSON for automated pipelines
  credentialctl sign doc.pdf -o doc_signed.pdf --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		inPath, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("invalid input path: %w", err)
		}
		if _, err := os.Stat(inPath); err != nil {
			return fmt.Errorf("input file not found at '%s'.\nHint: Provide a valid path to an existing media asset.", inPath)
		}

		if strings.TrimSpace(signOutputPath) == "" {
			return fmt.Errorf("output path is required.\nHint: Pass '-o <output_path>' (must differ from input path).")
		}
		outPath, err := filepath.Abs(signOutputPath)
		if err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}
		if inPath == outPath {
			return fmt.Errorf("in-place signing is not supported ('%s').\nHint: Specify a distinct output filename with '-o'.", inPath)
		}

		var keyPEM, certPEM, tsaKeyPEM, tsaCertPEM, rootCAPEM string
		if signKeyPath != "" || signCertPath != "" {
			if signKeyPath == "" || signCertPath == "" {
				return fmt.Errorf("both --key and --cert must be provided together.\nHint: Run 'credentialctl keygen' to create a local test PKI bundle.")
			}
			kb, err := os.ReadFile(signKeyPath)
			if err != nil {
				return fmt.Errorf("failed to read --key '%s': %w", signKeyPath, err)
			}
			cb, err := os.ReadFile(signCertPath)
			if err != nil {
				return fmt.Errorf("failed to read --cert '%s': %w", signCertPath, err)
			}
			keyPEM = string(kb)
			certPEM = string(cb)

			if signTSAKeyPath != "" && signTSACertPath != "" {
				tkb, err := os.ReadFile(signTSAKeyPath)
				if err != nil {
					return fmt.Errorf("failed to read --tsa-key '%s': %w", signTSAKeyPath, err)
				}
				tcb, err := os.ReadFile(signTSACertPath)
				if err != nil {
					return fmt.Errorf("failed to read --tsa-cert '%s': %w", signTSACertPath, err)
				}
				tsaKeyPEM = string(tkb)
				tsaCertPEM = string(tcb)
			}
		} else {
			pki, err := credentio.GenerateTestPKI("credentialctl")
			if err != nil {
				return fmt.Errorf("failed to generate ephemeral signing PKI: %w", err)
			}
			keyPEM = pki.ClaimSignerKeyPEM
			certPEM = pki.ClaimSignerCertPEM
			tsaKeyPEM = pki.TSAKeyPEM
			tsaCertPEM = pki.TSACertPEM
			rootCAPEM = pki.RootCACertPEM
		}

		var signerOpts []credentio.SignerOption
		signerOpts = append(signerOpts, credentio.WithGeneratorInfo(signGenerator, version))
		if tsaKeyPEM != "" && tsaCertPEM != "" {
			signerOpts = append(signerOpts, credentio.WithTSACredentials(tsaKeyPEM, tsaCertPEM))
		}

		signer, err := credentio.NewSigner(keyPEM, certPEM, signerOpts...)
		if err != nil {
			return fmt.Errorf("failed to initialize native signer: %w", err)
		}
		defer signer.Close()

		resolvedSourceType := expandDigitalSourceType(signSourceType)
		var customAssertions map[string]any
		if signAIOptOut {
			customAssertions = map[string]any{
				"c2pa.training-mining": map[string]any{
					"entries": map[string]any{
						"c2pa.ai_generative_training": map[string]any{"use": "notAllowed"},
						"c2pa.data_mining":            map[string]any{"use": "notAllowed"},
					},
				},
			}
		}

		mediaType := signMediaType
		if mediaType == "" {
			mediaType = engine.DetectMediaType(inPath)
		}

		signRes, err := signer.SignFile(inPath, outPath, &credentio.SignParams{
			MediaType:         mediaType,
			DigitalSourceType: resolvedSourceType,
			CustomAssertions:  customAssertions,
		})
		if err != nil {
			return fmt.Errorf("signing failed: %w", err)
		}

		// Verify the newly signed file immediately
		var valOpts []credentio.Option
		if rootCAPEM != "" {
			valOpts = append(valOpts,
				credentio.WithClaimSignerTrust(rootCAPEM),
				credentio.WithTSATrust(rootCAPEM),
				credentio.WithSkipTrustChecks(false),
			)
		} else {
			valOpts = append(valOpts, credentio.WithSkipTrustChecks(true))
		}

		validator, err := engine.NewValidatorService(valOpts...)
		if err != nil {
			return fmt.Errorf("post-sign validator initialization failed: %w", err)
		}
		defer validator.Close()

		report, err := validator.ValidateFile(outPath, signRes.MediaType)
		if err != nil {
			return fmt.Errorf("post-sign verification failed: %w", err)
		}

		if signJSON {
			out := jsonSignOutput{
				InputPath:         inPath,
				OutputPath:        outPath,
				ByteSize:          signRes.ByteSize,
				MediaType:         report.MediaType,
				DigitalSourceType: resolvedSourceType,
				Generator:         signGenerator,
				SigningCoreSec:    signRes.CoreSeconds,
				SigningElapsedSec: signRes.ElapsedSeconds,
				VerifiedBadge:     string(report.Badge()),
				ActiveManifest:    report.ActiveManifest,
			}
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		var sb strings.Builder
		sb.WriteString(theme.StyleHeader.Render(" C2PA ASSET SIGNED & VERIFIED ") + "\n\n")
		renderRow := func(label, val string) {
			lbl := theme.StyleMuted.Width(16).Bold(true).Render(label + ":")
			sb.WriteString(fmt.Sprintf("  %s %s\n", lbl, val))
		}

		sizeMB := float64(signRes.ByteSize) / (1024.0 * 1024.0)
		renderRow("Input", filepath.Base(inPath))
		renderRow("Output", fmt.Sprintf("%s (%.2f MB, %s)", outPath, sizeMB, report.MediaType))
		renderRow("Status", theme.RenderBadge(report.Badge()))
		if report.ActiveManifest != nil {
			renderRow("Generator", theme.StyleAccent.Render(report.ActiveManifest.ClaimGenerator))
			if report.ActiveManifest.Signature != nil && report.ActiveManifest.Signature.Issuer != "" {
				renderRow("Signer", theme.StylePass.Render(report.ActiveManifest.Signature.Issuer))
			}
			renderRow("Assertions", fmt.Sprintf("%d attached", len(report.ActiveManifest.Assertions)))
		}
		renderRow("Sign Core Time", fmt.Sprintf("%.2f ms", signRes.CoreSeconds*1000.0))
		renderRow("Verify Time", fmt.Sprintf("%.2f ms", report.CoreSeconds*1000.0))

		fmt.Fprintln(cmd.OutOrStdout(), sb.String())
		return nil
	},
}

func expandDigitalSourceType(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "", "none":
		return ""
	case "capture", "digitalcapture":
		return credentio.DigitalSourceTypeDigitalCapture
	case "ai", "trainedalgorithmicmedia":
		return credentio.DigitalSourceTypeTrainedAlgorithmicMedia
	case "composite-ai":
		return credentio.DigitalSourceTypeCompositeWithTrainedAI
	case "algorithmic":
		return credentio.DigitalSourceTypeAlgorithmicMedia
	case "creation", "digitalcreation":
		return credentio.DigitalSourceTypeDigitalCreation
	default:
		return v
	}
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVarP(&signOutputPath, "output", "o", "", "Output path for signed media asset (required)")
	signCmd.Flags().StringVar(&signKeyPath, "key", "", "Path to claim signer EC P-256 private key PEM")
	signCmd.Flags().StringVar(&signCertPath, "cert", "", "Path to claim signer X.509 certificate chain PEM")
	signCmd.Flags().StringVar(&signTSAKeyPath, "tsa-key", "", "Optional path to TSA EC P-256 private key PEM")
	signCmd.Flags().StringVar(&signTSACertPath, "tsa-cert", "", "Optional path to TSA X.509 certificate PEM")
	signCmd.Flags().StringVar(&signMediaType, "media-type", "", "Explicit IANA media type (e.g. image/png)")
	signCmd.Flags().StringVar(&signSourceType, "source-type", "capture", "IPTC digitalSourceType URI or shorthand (capture, ai, composite-ai)")
	signCmd.Flags().StringVar(&signGenerator, "generator", "credentialctl", "Claim generator software name")
	signCmd.Flags().BoolVar(&signAIOptOut, "ai-opt-out", false, "Attach c2pa.training-mining opt-out assertion")
	signCmd.Flags().BoolVar(&signJSON, "json", false, "Output structured JSON for machine/agent parsing")
}
