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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ============================================================================
// Tier 4: Real-World Application Scenarios
// ============================================================================

// Scenario 1: Standalone Developer Build & Run (F1, F2, F3, F4, F5)
// Simulates developer workflow: clean build from source, verify dynamic linkage, run CLI inspection
func TestTier4_Scenario1_StandaloneDeveloperBuildAndRun(t *testing.T) {
	root := RepoRoot(t)

	// Step 1: Ensure binary is built
	bin := EnsureBuiltBinary(t)

	// Step 2: Verify binary exists and has execute permissions
	fi, err := os.Stat(bin)
	if err != nil {
		t.Fatalf("built binary not found: %v", err)
	}
	if fi.Mode()&0111 == 0 {
		t.Fatalf("built binary is not executable: mode %v", fi.Mode())
	}

	// Step 3: Run credentialctl from neutral working directory
	neutralDir := t.TempDir()
	res := RunCmd(t, neutralDir, bin, "--help")
	if res.ExitCode != 0 {
		t.Fatalf("binary failed running --help from neutral dir (exit %d): %s", res.ExitCode, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "C2PA") {
		t.Errorf("binary --help missing expected description")
	}

	// Step 4: Run version command and verify output
	verRes := RunCmd(t, neutralDir, bin, "version")
	if verRes.ExitCode != 0 {
		t.Fatalf("version command failed (exit %d): %s", verRes.ExitCode, verRes.Stderr)
	}
	if !strings.Contains(verRes.Stdout, "credentialctl") || !strings.Contains(verRes.Stdout, "Credentio") {
		t.Errorf("version output missing required labels:\n%s", verRes.Stdout)
	}

	// Step 5: Run version --json and verify structured fields
	jsonRes := RunCmd(t, neutralDir, bin, "version", "--json")
	if jsonRes.ExitCode != 0 {
		t.Fatalf("version --json failed (exit %d): %s", jsonRes.ExitCode, jsonRes.Stderr)
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonRes.Stdout), &data); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if v, ok := data["credentio_engine_version"].(string); !ok || len(v) == 0 {
		t.Errorf("missing or empty credentio_engine_version in JSON payload: %+v", data)
	}

	t.Logf("Scenario 1 verified: standalone developer build and run completed successfully.")
	_ = root
}

// Scenario 2: Simulated Homebrew Cellar Layout Execution (F3, F4, F5, F6)
// Simulates Homebrew Cellar directory structure with symlinked binary in prefix/bin
func TestTier4_Scenario2_SimulatedHomebrewCellarLayoutExecution(t *testing.T) {
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		t.Skip("Homebrew Cellar layout simulation requires POSIX symlinks")
	}

	root := RepoRoot(t)
	bin := EnsureBuiltBinary(t)

	libName := "libcredentio_c.dylib"
	if runtime.GOOS == "linux" {
		libName = "libcredentio_c.so"
	}
	libSrc := filepath.Join(root, "third_party", "credentio", "lib", libName)

	// Create simulated Homebrew environment:
	// /mock_brew/
	//   Cellar/credentialctl/0.1.5/
	//     bin/credentialctl
	//     lib/libcredentio_c.dylib
	//   bin/credentialctl -> ../Cellar/credentialctl/0.1.5/bin/credentialctl
	brewRoot := t.TempDir()
	cellarBinDir := filepath.Join(brewRoot, "Cellar", "credentialctl", "0.1.5", "bin")
	cellarLibDir := filepath.Join(brewRoot, "Cellar", "credentialctl", "0.1.5", "lib")
	prefixBinDir := filepath.Join(brewRoot, "bin")

	if err := os.MkdirAll(cellarBinDir, 0755); err != nil {
		t.Fatalf("failed to create cellar bin dir: %v", err)
	}
	if err := os.MkdirAll(cellarLibDir, 0755); err != nil {
		t.Fatalf("failed to create cellar lib dir: %v", err)
	}
	if err := os.MkdirAll(prefixBinDir, 0755); err != nil {
		t.Fatalf("failed to create prefix bin dir: %v", err)
	}

	cellarBin := filepath.Join(cellarBinDir, "credentialctl")
	cellarLib := filepath.Join(cellarLibDir, libName)
	symlinkedBin := filepath.Join(prefixBinDir, "credentialctl")

	copyFile(t, bin, cellarBin, 0755)
	copyFile(t, libSrc, cellarLib, 0755)

	// Create relative or absolute symlink from prefix/bin to Cellar/.../bin/credentialctl
	if err := os.Symlink(cellarBin, symlinkedBin); err != nil {
		t.Fatalf("failed to create symlink: %v", err)
	}

	// Execute via symlink from an arbitrary working directory with sanitized environment
	neutralDir := t.TempDir()
	cmd := exec.Command(symlinkedBin, "version", "--json")
	cmd.Dir = neutralDir
	cmd.Env = filterEnv(os.Environ(), "DYLD_LIBRARY_PATH", "LD_LIBRARY_PATH")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Simulated Homebrew Cellar execution failed: %v, output:\n%s", err, string(out))
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("failed to parse JSON output from Homebrew cellar execution: %v", err)
	}

	if ver, ok := parsed["credentio_engine_version"].(string); !ok || len(ver) == 0 {
		t.Errorf("expected valid credentio_engine_version from Homebrew cellar binary, got %+v", parsed)
	}

	t.Logf("Scenario 2 verified: Homebrew Cellar layout execution resolved @executable_path/../lib successfully.")
}

// Scenario 3: CI Release Snapshot Artifact & Archive Audit (F6, F7, F10)
// Audits GoReleaser snapshot release archives, checksums, and package manifests
func TestTier4_Scenario3_CIReleaseSnapshotArtifactAudit(t *testing.T) {
	root := RepoRoot(t)

	// Step 1: Execute snapshot release
	res := RunCmd(t, root, "goreleaser", "release", "--snapshot", "--clean")
	if res.ExitCode != 0 {
		t.Fatalf("goreleaser snapshot release failed (exit %d): %s", res.ExitCode, res.Stderr)
	}

	// Step 2: Inspect dist/
	distDir := filepath.Join(root, "dist")
	entries, err := os.ReadDir(distDir)
	if err != nil {
		t.Fatalf("failed to read dist/: %v", err)
	}

	var archiveName string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tar.gz") && strings.HasPrefix(e.Name(), "credentialctl_") {
			archiveName = e.Name()
			break
		}
	}
	if archiveName == "" {
		t.Fatalf("no release archive .tar.gz generated in dist/")
	}

	archivePath := filepath.Join(distDir, archiveName)

	// Step 3: Verify SHA256 in checksums.txt
	checksumsData, err := os.ReadFile(filepath.Join(distDir, "checksums.txt"))
	if err != nil {
		t.Fatalf("failed to read checksums.txt: %v", err)
	}

	// Calculate actual SHA256 of archive
	f, err := os.Open(archivePath)
	if err != nil {
		t.Fatalf("failed to open archive %s: %v", archivePath, err)
	}
	hasher := sha256.New()
	io.Copy(hasher, f)
	f.Close()
	actualHash := hex.EncodeToString(hasher.Sum(nil))

	if !strings.Contains(string(checksumsData), actualHash) {
		t.Errorf("checksums.txt does not contain matching SHA256 hash %s for %s", actualHash, archiveName)
	}

	// Step 4: Extract archive and audit packaged files
	unpackDir := t.TempDir()
	if err := UnpackTarGz(t, archivePath, unpackDir); err != nil {
		t.Fatalf("failed to unpack %s: %v", archivePath, err)
	}

	expectedFiles := []string{
		"credentialctl",
		"README.md",
		"LICENSE",
	}
	if runtime.GOOS == "darwin" {
		expectedFiles = append(expectedFiles, "libcredentio_c.dylib")
	} else if runtime.GOOS == "linux" {
		expectedFiles = append(expectedFiles, "libcredentio_c.so")
	}

	for _, ef := range expectedFiles {
		p := filepath.Join(unpackDir, ef)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("snapshot archive %s missing expected packaged file %s", archiveName, ef)
		}
	}

	// Step 5: Test extracted binary
	extractedBin := filepath.Join(unpackDir, "credentialctl")
	testRes := RunCmd(t, unpackDir, extractedBin, "version", "--json")
	if testRes.ExitCode != 0 {
		t.Errorf("extracted binary from snapshot archive failed running version (exit %d): %s", testRes.ExitCode, testRes.Stderr)
	}

	t.Logf("Scenario 3 verified: snapshot release archive and checksums validated successfully.")
}

// Scenario 4: JSON Machine-Consumption by Automated Agent (F4, F5)
// Simulates an autonomous coding agent parsing structured JSON output
func TestTier4_Scenario4_JSONMachineConsumptionByAgent(t *testing.T) {
	bin := EnsureBuiltBinary(t)

	res := RunCmd(t, "", bin, "version", "--json")
	if res.ExitCode != 0 {
		t.Fatalf("version --json returned non-zero exit code %d: %s", res.ExitCode, res.Stderr)
	}

	// Strict JSON unmarshaling into strongly-typed model
	type AgentVersionPayload struct {
		Version                string `json:"version"`
		GitCommit              string `json:"git_commit"`
		BuildDate              string `json:"build_date"`
		GoVersion              string `json:"go_version"`
		Compiler               string `json:"compiler"`
		Platform               string `json:"platform"`
		CredentioEngineVersion string `json:"credentio_engine_version"`
	}

	var payload AgentVersionPayload
	decoder := json.NewDecoder(strings.NewReader(res.Stdout))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&payload); err != nil {
		t.Fatalf("Agent JSON schema validation failed: %v. Raw stdout:\n%s", err, res.Stdout)
	}

	// Verify semantic properties expected by autonomous agents
	if payload.Version == "" {
		t.Errorf("Agent payload: missing Version")
	}
	if payload.Platform == "" {
		t.Errorf("Agent payload: missing Platform")
	}
	if payload.CredentioEngineVersion == "" {
		t.Errorf("Agent payload: missing CredentioEngineVersion")
	}

	t.Logf("Scenario 4 verified: Agent-Aware JSON contract fully conforms to schema.")
}

// Scenario 5: End-to-End Release & Documentation Audit (F8, F9, F10)
// Audits all project documentation against editorial standards and verifies release alignment
func TestTier4_Scenario5_EndToEndReleaseAndDocumentationAudit(t *testing.T) {
	root := RepoRoot(t)

	docFiles := []string{
		filepath.Join(root, "README.md"),
		filepath.Join(root, "docs", "RELEASING.md"),
		filepath.Join(root, "docs", "user_guide.md"),
		filepath.Join(root, "AGENTS.md"),
	}

	// Audit em dashes in all documentation
	for _, doc := range docFiles {
		if _, err := os.Stat(doc); os.IsNotExist(err) {
			t.Errorf("documentation file does not exist: %s", doc)
			continue
		}
		audit := AuditEmDashes(t, doc)
		if audit.TotalEmDashes > 0 {
			t.Errorf("Editorial violation in %s: found %d em dashes:\n%s", filepath.Base(doc), audit.TotalEmDashes, strings.Join(audit.Occurrences, "\n"))
		}
	}

	// Audit Mark Allen rubric for README.md
	readmePath := filepath.Join(root, "README.md")
	score := AuditMarkAllenRubric(t, readmePath)
	if score.TotalScore < 34 {
		t.Errorf("README.md failed Mark Allen rubric audit: score %d/40 (target >= 34). Feedback: %v", score.TotalScore, score.Feedback)
	} else {
		t.Logf("README.md Mark Allen rubric score: %d/40 (PASSED)", score.TotalScore)
	}

	t.Logf("Scenario 5 verified: Release and documentation editorial audit passed.")
}
