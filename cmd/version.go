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
	"runtime"
	"strings"

	"github.com/ghchinoy/credentialctl/internal/engine"
	"github.com/ghchinoy/credentialctl/internal/ui/theme"
	"github.com/spf13/cobra"
)

var (
	version     = "0.1.5-dev"
	commit      = "dev"
	date        = "unknown"
	versionJSON bool
)

// VersionInfo contains build, compiler, platform, and runtime engine metadata.
type VersionInfo struct {
	Version                string `json:"version"`
	GitCommit              string `json:"git_commit"`
	BuildDate              string `json:"build_date"`
	GoVersion              string `json:"go_version"`
	Compiler               string `json:"compiler"`
	Platform               string `json:"platform"`
	CredentioEngineVersion string `json:"credentio_engine_version"`
}

// GetVersionInfo constructs and returns the populated VersionInfo struct.
func GetVersionInfo() VersionInfo {
	engineVer := engine.CredentioCVersion()
	if engineVer == "" {
		engineVer = "unavailable"
	}
	return VersionInfo{
		Version:                version,
		GitCommit:              commit,
		BuildDate:              date,
		GoVersion:              runtime.Version(),
		Compiler:               runtime.Compiler,
		Platform:               fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		CredentioEngineVersion: engineVer,
	}
}

func formatHumanVersion(info VersionInfo) string {
	var sb strings.Builder
	title := fmt.Sprintf("credentialctl v%s", info.Version)
	subtitle := "(C2PA Content Credentials CLI & TUI)"

	sb.WriteString(fmt.Sprintf("%s %s\n\n", theme.StyleAccent.Render(title), theme.StyleMuted.Render(subtitle)))

	renderRow := func(label, val string) {
		paddedLabel := fmt.Sprintf("%-14s", label+":")
		sb.WriteString(fmt.Sprintf("  %s %s\n", theme.StyleMuted.Render(paddedLabel), val))
	}

	renderRow("Git Commit", theme.StyleID.Render(info.GitCommit))
	renderRow("Build Date", info.BuildDate)
	renderRow("Go Runtime", fmt.Sprintf("%s (%s/%s)", info.GoVersion, info.Compiler, info.Platform))

	var engineDisplay string
	if info.CredentioEngineVersion == "unavailable" || info.CredentioEngineVersion == "" {
		engineDisplay = theme.StyleWarn.Render("unavailable")
	} else {
		engineDisplay = fmt.Sprintf("%s %s", theme.StylePass.Render("v"+info.CredentioEngineVersion), theme.StyleMuted.Render("(C-ABI Shared Library)"))
	}
	renderRow("Credentio", engineDisplay)

	return sb.String()
}

var versionCmd = &cobra.Command{
	Use:     "version",
	GroupID: "inspection",
	Short:   "Print version and runtime build metadata",
	Long: `Print comprehensive build, compiler, platform, and runtime C2PA C-ABI engine version information.

Supports both human-readable styled terminal output and structured JSON for automated pipelines and agents.`,
	Example: `  # Print human-readable version information
  credentialctl version

  # Output structured JSON for agent and pipeline consumption
  credentialctl version --json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		info := GetVersionInfo()

		if versionJSON {
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetIndent("", "  ")
			return encoder.Encode(info)
		}

		out := formatHumanVersion(info)
		fmt.Fprint(cmd.OutOrStdout(), out)
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Output version information in structured JSON format")
	rootCmd.AddCommand(versionCmd)
}
