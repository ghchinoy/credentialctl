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
	"path/filepath"
	"strings"

	"github.com/ghchinoy/credentio-contributions/go"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
	"github.com/spf13/cobra"
)

var (
	keygenOutDir string
	keygenOrg    string
	keygenJSON   bool
)

type jsonKeygenOutput struct {
	OutputDir       string `json:"output_dir"`
	Organization    string `json:"organization"`
	RootCACert      string `json:"root_ca_cert"`
	ClaimSignerKey  string `json:"claim_signer_key"`
	ClaimSignerCert string `json:"claim_signer_cert"`
	TSAKey          string `json:"tsa_key"`
	TSACert         string `json:"tsa_cert"`
}

var keygenCmd = &cobra.Command{
	Use:     "keygen",
	GroupID: "signing",
	Short:   "Generate a local C2PA-compliant test PKI (Root CA, Claim Signer, TSA)",
	Long: `Generate an EC P-256 test Public Key Infrastructure (PKI) bundle conforming to the
C2PA X.509 certificate profile (C2PA claim signing EKU 1.3.6.1.4.1.62558.2.1 and
RFC 3161 critical timeStamping EKU) for local signing and verification workflows.`,
	Example: `  # Generate test keys and certificates in ./test_pki
  credentialctl keygen --out-dir ./test_pki

  # Generate with custom organization name and structured JSON output
  credentialctl keygen --out-dir ./my_pki --org "Acme Media Studio" --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(keygenOutDir)
		if err != nil {
			return fmt.Errorf("invalid output directory: %w", err)
		}

		pki, err := credentio.GenerateTestPKI(keygenOrg)
		if err != nil {
			return fmt.Errorf("failed to generate C2PA test PKI: %w", err)
		}

		if err := pki.WriteToDir(absDir); err != nil {
			return fmt.Errorf("failed to write PKI bundle to '%s': %w", absDir, err)
		}

		out := jsonKeygenOutput{
			OutputDir:       absDir,
			Organization:    keygenOrg,
			RootCACert:      filepath.Join(absDir, "root_ca_cert.pem"),
			ClaimSignerKey:  filepath.Join(absDir, "claim_signer_key.pem"),
			ClaimSignerCert: filepath.Join(absDir, "claim_signer_cert.pem"),
			TSAKey:          filepath.Join(absDir, "tsa_key.pem"),
			TSACert:         filepath.Join(absDir, "tsa_cert.pem"),
		}

		if keygenJSON {
			data, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		var sb strings.Builder
		sb.WriteString(theme.StyleHeader.Render(" C2PA TEST PKI GENERATED ") + "\n\n")
		renderRow := func(label, val string) {
			lbl := theme.StyleMuted.Width(18).Bold(true).Render(label + ":")
			sb.WriteString(fmt.Sprintf("  %s %s\n", lbl, val))
		}
		renderRow("Directory", theme.StyleAccent.Render(out.OutputDir))
		renderRow("Organization", out.Organization)
		renderRow("Root CA Cert", out.RootCACert)
		renderRow("Claim Signer Key", out.ClaimSignerKey)
		renderRow("Claim Signer Cert", out.ClaimSignerCert)
		renderRow("TSA Key", out.TSAKey)
		renderRow("TSA Cert", out.TSACert)
		fmt.Fprintln(cmd.OutOrStdout(), sb.String())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(keygenCmd)
	keygenCmd.Flags().StringVar(&keygenOutDir, "out-dir", "./test_pki", "Directory to write generated PEM files")
	keygenCmd.Flags().StringVar(&keygenOrg, "org", "Credentio Test", "Organization name for X.509 certificate subjects")
	keygenCmd.Flags().BoolVar(&keygenJSON, "json", false, "Output structured JSON for machine/agent parsing")
}
