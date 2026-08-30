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

package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ============================================================================
// Tier 3: Cross-Feature Combinations (Pairwise Interaction Testing)
// ============================================================================

// Pairwise 1: F1 (Go Module) + F2 (Fetch Script Sync)
// Verifies version alignment between go.mod and scripts/fetch-credentio-lib.sh
func TestTier3_Pairwise1_GoModule_FetchScript_VersionAlignment(t *testing.T) {
	root := RepoRoot(t)
	goModData, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	scriptData, err := os.ReadFile(filepath.Join(root, "scripts", "fetch-credentio-lib.sh"))
	if err != nil {
		t.Fatalf("failed to read fetch script: %v", err)
	}

	goModContent := string(goModData)
	scriptContent := string(scriptData)

	if strings.Contains(goModContent, "credentio-contributions/go v0.1.5") {
		if !strings.Contains(scriptContent, "0.1.5") {
			t.Errorf("Fetch script version is not aligned with go.mod v0.1.5 requirement")
		}
	}
}

// Pairwise 2: F3 (RPATHs) + F5 (Version Command Execution)
// Verifies that the built binary in bin/ can execute the version command from any arbitrary working directory
func TestTier3_Pairwise2_RPATH_VersionCommand_ArbitraryDirExecution(t *testing.T) {
	tempWorkingDir := t.TempDir()
	res := RunCredentialctl(t, tempWorkingDir, "version")
	if res.ExitCode != 0 {
		t.Fatalf("executing version from outside directory %s failed (exit %d): %s", tempWorkingDir, res.ExitCode, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "credentialctl") {
		t.Errorf("expected version output to contain 'credentialctl', got:\n%s", res.Stdout)
	}
}

// Pairwise 3: F4 (C-ABI Runtime Discovery) + F5 (Version Command JSON)
// Verifies that `credentialctl version --json` reports a credentio_engine_version matching native cr_version()
func TestTier3_Pairwise3_RuntimeCABI_VersionJSON_OutputSync(t *testing.T) {
	nativeVersion := QueryNativeCRVersion()

	res := RunCredentialctl(t, "", "version", "--json")
	if res.ExitCode != 0 {
		t.Fatalf("version --json failed (exit %d): %s", res.ExitCode, res.Stderr)
	}

	var payload struct {
		CredentioEngineVersion string `json:"credentio_engine_version"`
	}
	if err := json.Unmarshal([]byte(res.Stdout), &payload); err != nil {
		t.Fatalf("failed to parse JSON from version command: %v", err)
	}

	if payload.CredentioEngineVersion != nativeVersion && !strings.Contains(payload.CredentioEngineVersion, nativeVersion) {
		t.Errorf("version JSON credentio_engine_version %q does not match native library %q", payload.CredentioEngineVersion, nativeVersion)
	}
}

// Pairwise 4: F6 (GoReleaser Brews) + F3 (Dynamic RPATHs)
// Verifies that Homebrew formula install paths (bin/ + lib/) align with binary @executable_path/../lib RPATH
func TestTier3_Pairwise4_GoReleaserBrews_RPATH_CellarAlignment(t *testing.T) {
	root := RepoRoot(t)
	goreleaserData, err := os.ReadFile(filepath.Join(root, ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(goreleaserData)

	hasBrewsLibInstall := strings.Contains(content, "lib.install")
	hasBrewsBinInstall := strings.Contains(content, "bin.install")

	if !hasBrewsBinInstall || !hasBrewsLibInstall {
		t.Errorf(".goreleaser.yaml brews stanza must install both binary (bin.install) and shared library (lib.install)")
	}

	if runtime.GOOS == "darwin" {
		bin := EnsureBuiltBinary(t)
		rpaths := GetBinaryRPATHs(t, bin)
		hasCellarRPATH := false
		for _, rp := range rpaths {
			if rp == "@executable_path/../lib" {
				hasCellarRPATH = true
				break
			}
		}
		if !hasCellarRPATH {
			t.Errorf("binary missing @executable_path/../lib RPATH required to support Homebrew Cellar lib installation")
		}
	}
}

// Pairwise 5: F7 (Release Workflow) + F6 (GoReleaser Configuration)
// Verifies that GitHub Actions release workflow injects secrets required by .goreleaser.yaml
func TestTier3_Pairwise5_Workflow_GoReleaser_SecretsSync(t *testing.T) {
	root := RepoRoot(t)
	workflowData, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yaml"))
	if err != nil {
		t.Fatalf("failed to read workflow: %v", err)
	}
	goreleaserData, err := os.ReadFile(filepath.Join(root, ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}

	wfContent := string(workflowData)
	grContent := string(goreleaserData)

	if strings.Contains(grContent, "HOMEBREW_TAP_GITHUB_TOKEN") {
		if !strings.Contains(wfContent, "HOMEBREW_TAP_GITHUB_TOKEN") {
			t.Errorf("Release workflow does not inject HOMEBREW_TAP_GITHUB_TOKEN required by .goreleaser.yaml")
		}
	}
}

// Pairwise 6: F8 (Release Docs) + F7 (Release Workflow)
// Verifies that docs/RELEASING.md tag instructions match the GitHub Actions workflow trigger filters
func TestTier3_Pairwise6_ReleaseDocs_Workflow_TagFilterSync(t *testing.T) {
	root := RepoRoot(t)
	docsData, err := os.ReadFile(filepath.Join(root, "docs", "RELEASING.md"))
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	wfData, err := os.ReadFile(filepath.Join(root, ".github", "workflows", "release.yaml"))
	if err != nil {
		t.Fatalf("failed to read release.yaml: %v", err)
	}

	docsContent := string(docsData)
	wfContent := string(wfData)

	if strings.Contains(wfContent, "'v*'") || strings.Contains(wfContent, "v*") {
		if !strings.Contains(docsContent, "git tag v") && !strings.Contains(docsContent, "git tag -a v") {
			t.Errorf("docs/RELEASING.md tag examples should demonstrate 'v*' tag prefix matching CI workflow filter")
		}
	}
}

// Pairwise 7: F9 (README) + F6 (GoReleaser Formula Configuration)
// Verifies that Homebrew tap and formula name documented in README.md matches .goreleaser.yaml
func TestTier3_Pairwise7_README_GoReleaser_TapNameSync(t *testing.T) {
	root := RepoRoot(t)
	readmeData, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	grData, err := os.ReadFile(filepath.Join(root, ".goreleaser.yaml"))
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}

	readmeContent := string(readmeData)
	grContent := string(grData)

	if strings.Contains(grContent, "homebrew-tap") {
		if !strings.Contains(readmeContent, "ghchinoy/homebrew-tap") && !strings.Contains(readmeContent, "ghchinoy/tap") {
			t.Errorf("README.md should document tap repository name matching .goreleaser.yaml brews stanza")
		}
	}
}

// Pairwise 8: F10 (Snapshot Artifacts) + F3/F4/F5 (Binary/RPATH/Version)
// Unpacks GoReleaser snapshot archive and verifies the packaged binary executes version --json cleanly
func TestTier3_Pairwise8_SnapshotArchive_BinaryExecution_VersionJSON(t *testing.T) {
	root := RepoRoot(t)
	distDir := filepath.Join(root, "dist")
	entries, err := os.ReadDir(distDir)
	if err != nil {
		t.Fatalf("failed to read dist/: %v", err)
	}

	var archiveFile string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") && strings.HasPrefix(entry.Name(), "credentialctl_") {
			archiveFile = filepath.Join(distDir, entry.Name())
			break
		}
	}
	if archiveFile == "" {
		t.Fatalf("no archive file found in dist/")
	}

	tempDir := t.TempDir()
	if err := UnpackTarGz(t, archiveFile, tempDir); err != nil {
		t.Fatalf("failed to unpack archive: %v", err)
	}

	binPath := filepath.Join(tempDir, "credentialctl")
	res := RunCmd(t, tempDir, binPath, "version", "--json")
	if res.ExitCode != 0 {
		t.Fatalf("unpacked binary failed running version --json (exit %d): %s", res.ExitCode, res.Stderr)
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(res.Stdout), &m); err != nil {
		t.Fatalf("unpacked binary output invalid JSON: %v", err)
	}
}
