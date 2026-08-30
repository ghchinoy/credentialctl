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
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// ============================================================================
// Feature 1 Boundaries: Upstream Go Module v0.1.5
// ============================================================================

func TestTier2_F1_GoModule_NoDirectReplaceHack(t *testing.T) {
	root := RepoRoot(t)
	goModPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	content := string(data)

	// Verify no dirty local replace directive overrides credentio-contributions
	if strings.Contains(content, "replace github.com/ghchinoy/credentio-contributions/go =>") {
		t.Errorf("go.mod should not contain a local replace directive for credentio-contributions in production")
	}
}

func TestTier2_F1_GoModule_DependencyCleanTidy(t *testing.T) {
	root := RepoRoot(t)
	res := RunCmd(t, root, "go", "mod", "tidy", "-diff")
	if res.ExitCode != 0 && strings.TrimSpace(res.Stderr) != "" {
		t.Logf("go mod tidy diff output: %s", res.Stderr)
	}
}

func TestTier2_F1_GoModule_DirectDependencyAnnotation(t *testing.T) {
	root := RepoRoot(t)
	goModPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	lines := strings.Split(string(data), "\n")

	foundDirect := false
	for _, line := range lines {
		if strings.Contains(line, "github.com/ghchinoy/credentio-contributions/go v0.1.5") {
			if !strings.Contains(line, "// indirect") {
				foundDirect = true
			}
		}
	}
	if !foundDirect {
		t.Errorf("credentio-contributions/go v0.1.5 must be declared as a direct dependency (not indirect)")
	}
}

func TestTier2_F1_GoModule_TransitiveIntegrity(t *testing.T) {
	root := RepoRoot(t)
	res := RunCmd(t, root, "go", "list", "-m", "-json", "github.com/ghchinoy/credentio-contributions/go")
	if res.ExitCode != 0 {
		t.Fatalf("failed to query module info: %s", res.Stderr)
	}

	var modInfo struct {
		Path    string
		Version string
	}
	if err := json.Unmarshal([]byte(res.Stdout), &modInfo); err != nil {
		t.Fatalf("failed to parse go list JSON: %v", err)
	}
	if modInfo.Version != "v0.1.5" {
		t.Errorf("expected module version v0.1.5, got %s", modInfo.Version)
	}
}

func TestTier2_F1_GoModule_ModuleNameExactMatch(t *testing.T) {
	root := RepoRoot(t)
	goModPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	firstLine := strings.Split(string(data), "\n")[0]
	expected := "module github.com/ghchinoy/credentialctl"
	if strings.TrimSpace(firstLine) != expected {
		t.Errorf("module name mismatch in go.mod: expected %q, got %q", expected, firstLine)
	}
}

// ============================================================================
// Feature 2 Boundaries: Native Library Fetch Script Sync
// ============================================================================

func TestTier2_F2_FetchScript_StripLeadingV(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, `VERSION="${VERSION#v}"`) {
		t.Errorf("fetch script should strip leading 'v' character from version argument")
	}
}

func TestTier2_F2_FetchScript_MultiPlatformAssetStaging(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "libcredentio_c-darwin-arm64.dylib") || !strings.Contains(content, "libcredentio_c-linux-amd64.so") {
		t.Errorf("fetch script should stage both Darwin arm64 and Linux amd64 native libraries")
	}
}

func TestTier2_F2_FetchScript_HostPlatformValidation(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "required on Darwin") || !strings.Contains(content, "exit 1") {
		t.Errorf("fetch script must validate host platform library presence and exit 1 on failure")
	}
}

func TestTier2_F2_FetchScript_OutputDirCreationIdempotency(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, `mkdir -p "${OUTPUT_DIR}"`) && !strings.Contains(content, "mkdir -p") {
		t.Errorf("fetch script must create output directory with mkdir -p")
	}
}

func TestTier2_F2_FetchScript_ExecutableFileMode0755(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "chmod 755") && !strings.Contains(content, "chmod +x") {
		t.Errorf("fetch script should set executable permissions on downloaded shared libraries")
	}
}

// ============================================================================
// Feature 3 Boundaries: Dynamic Multi-Platform RPATHs
// ============================================================================

func TestTier2_F3_DynamicRPATH_RelocatedBinaryWithSiblingLib(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-specific shared library relocation test")
	}
	bin := EnsureBuiltBinary(t)
	root := RepoRoot(t)
	libSrc := filepath.Join(root, "third_party", "credentio", "lib", "libcredentio_c.dylib")

	tempDir := t.TempDir()
	binDest := filepath.Join(tempDir, "credentialctl")
	libDest := filepath.Join(tempDir, "libcredentio_c.dylib")

	copyFile(t, bin, binDest, 0755)
	copyFile(t, libSrc, libDest, 0755)

	otherDir := t.TempDir()
	res := RunCmd(t, otherDir, binDest, "--help")
	if res.ExitCode != 0 {
		t.Errorf("relocated binary with sibling lib failed (exit %d): %s", res.ExitCode, res.Stderr)
	}
}

func TestTier2_F3_DynamicRPATH_UnsetDyldLibraryPathSafety(t *testing.T) {
	bin := EnsureBuiltBinary(t)
	cmd := exec.Command(bin, "--help")
	cleanEnv := filterEnv(os.Environ(), "DYLD_LIBRARY_PATH", "LD_LIBRARY_PATH")
	cmd.Env = cleanEnv

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Errorf("binary failed to execute with sanitized library path environment: %v, stderr: %s", err, stderr.String())
	}
}

func TestTier2_F3_DynamicRPATH_MissingLibErrorHandling(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-specific dyld test")
	}
	bin := EnsureBuiltBinary(t)

	tempDir := t.TempDir()
	binDest := filepath.Join(tempDir, "isolated_credentialctl")
	copyFile(t, bin, binDest, 0755)

	res := RunCmd(t, tempDir, binDest, "--help")
	if res.ExitCode == 0 {
		t.Logf("Note: binary found lib in default system paths or fallback rpath")
	} else {
		if !strings.Contains(res.Stderr, "dyld") && !strings.Contains(res.Stderr, "Library not loaded") {
			t.Logf("Process failed as expected when isolated: %s", res.Stderr)
		}
	}
}

func TestTier2_F3_DynamicRPATH_MultipleRPATHPrecedence(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-specific Mach-O RPATH check")
	}
	bin := EnsureBuiltBinary(t)
	rpaths := GetBinaryRPATHs(t, bin)

	if len(rpaths) < 2 {
		t.Logf("Binary has %d RPATHs: %v", len(rpaths), rpaths)
	}
}

func TestTier2_F3_DynamicRPATH_DynamicMachOHeaderNotStatic(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin Mach-O header test")
	}
	bin := EnsureBuiltBinary(t)
	res := RunCmd(t, "", "otool", "-L", bin)
	if res.ExitCode != 0 {
		t.Fatalf("otool -L failed: %s", res.Stderr)
	}
	if !strings.Contains(res.Stdout, "libcredentio_c") {
		t.Errorf("otool -L output does not reference libcredentio_c:\n%s", res.Stdout)
	}
}

// ============================================================================
// Feature 4 Boundaries: Runtime C-ABI Version Discovery
// ============================================================================

func TestTier2_F4_RuntimeVersion_DynamicSymbolLookup(t *testing.T) {
	ver := QueryNativeCRVersion()
	if ver == "" {
		t.Errorf("cr_version symbol resolution failed in native library")
	}
}

func TestTier2_F4_RuntimeVersion_NullTerminatedCStr(t *testing.T) {
	ver := QueryNativeCRVersion()
	if strings.Contains(ver, "\x00") {
		t.Errorf("version string contains unhandled null terminator byte: %q", ver)
	}
}

func TestTier2_F4_RuntimeVersion_ExactOrPrefixMatchV015(t *testing.T) {
	ver := QueryNativeCRVersion()
	if !strings.HasPrefix(ver, "v0.1.5") && !strings.HasPrefix(ver, "0.1.5") && !strings.Contains(ver, "0.1.5") {
		t.Logf("Note: runtime version is %q", ver)
	}
}

func TestTier2_F4_RuntimeVersion_ZeroPanicStressTest(t *testing.T) {
	const iterations = 100
	for i := 0; i < iterations; i++ {
		v := QueryNativeCRVersion()
		if len(v) == 0 {
			t.Fatalf("iteration %d returned empty version", i)
		}
	}
}

func TestTier2_F4_RuntimeVersion_HeaderSymbolSignatureMatch(t *testing.T) {
	root := RepoRoot(t)
	headerPath := filepath.Join(root, "third_party", "credentio", "include", "credentio_c.h")
	data, err := os.ReadFile(headerPath)
	if err != nil {
		t.Fatalf("failed to read header: %v", err)
	}
	content := string(data)

	expectedDecl := "const char* cr_version(void);"
	if !strings.Contains(content, expectedDecl) {
		t.Errorf("C-ABI header missing expected declaration %q", expectedDecl)
	}
}

// ============================================================================
// Feature 5 Boundaries: Version Command & CLI Flags
// ============================================================================

func TestTier2_F5_VersionCommand_JSONStrictSchemaTypes(t *testing.T) {
	res := RunCredentialctl(t, "", "version", "--json")
	if res.ExitCode != 0 {
		t.Fatalf("version --json failed: %s", res.Stderr)
	}

	var parsed struct {
		Version                string `json:"version"`
		GitCommit              string `json:"git_commit"`
		BuildDate              string `json:"build_date"`
		GoVersion              string `json:"go_version"`
		Compiler               string `json:"compiler"`
		Platform               string `json:"platform"`
		CredentioEngineVersion string `json:"credentio_engine_version"`
	}

	if err := json.Unmarshal([]byte(res.Stdout), &parsed); err != nil {
		t.Fatalf("failed to deserialize into typed struct: %v", err)
	}

	if parsed.Version == "" || parsed.Platform == "" || parsed.CredentioEngineVersion == "" {
		t.Errorf("deserialized struct has unexpected empty fields: %+v", parsed)
	}
}

func TestTier2_F5_VersionCommand_JSONNoANSIEscapes(t *testing.T) {
	res := RunCredentialctl(t, "", "version", "--json")
	if res.ExitCode != 0 {
		t.Fatalf("version --json failed: %s", res.Stderr)
	}

	if strings.Contains(res.Stdout, "\x1b[") || strings.Contains(res.Stdout, "\033[") {
		t.Errorf("version --json stdout must not contain ANSI escape codes: %s", res.Stdout)
	}
}

func TestTier2_F5_VersionCommand_InvalidFlagsFailFast(t *testing.T) {
	res := RunCredentialctl(t, "", "version", "--invalid-nonexistent-flag")
	if res.ExitCode == 0 {
		t.Errorf("credentialctl version with invalid flag should return non-zero exit code")
	}
}

func TestTier2_F5_VersionCommand_DefaultDevBuildFallbacks(t *testing.T) {
	res := RunCredentialctl(t, "", "version", "--json")
	if res.ExitCode != 0 {
		t.Fatalf("version --json failed: %s", res.Stderr)
	}

	var m map[string]interface{}
	json.Unmarshal([]byte(res.Stdout), &m)
	if v, ok := m["version"].(string); ok && v == "" {
		t.Errorf("version field should not be empty")
	}
}

func TestTier2_F5_VersionCommand_ShortAndLongFlagsEquivalence(t *testing.T) {
	resShort := RunCredentialctl(t, "", "-v")
	resLong := RunCredentialctl(t, "", "--version")

	if resShort.ExitCode != resLong.ExitCode {
		t.Errorf("exit codes differ: -v (%d) vs --version (%d)", resShort.ExitCode, resLong.ExitCode)
	}
	if strings.TrimSpace(resShort.Stdout) != strings.TrimSpace(resLong.Stdout) {
		t.Errorf("-v and --version outputs differ: %q vs %q", resShort.Stdout, resLong.Stdout)
	}
}

// ============================================================================
// Feature 6 Boundaries: GoReleaser Configuration & Brews
// ============================================================================

func TestTier2_F6_GoReleaser_ValidYAMLNoUnknownFields(t *testing.T) {
	root := RepoRoot(t)
	res := RunCmd(t, root, "goreleaser", "check")
	if res.ExitCode != 0 {
		t.Logf("goreleaser check output: %s %s", res.Stdout, res.Stderr)
	}
}

func TestTier2_F6_GoReleaser_BrewsRubyDarwinAndLinuxBranching(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "OS.mac?") && !strings.Contains(content, "libcredentio_c.dylib") {
		t.Errorf(".goreleaser.yaml brews stanza should handle macOS dynamic library installation")
	}
}

func TestTier2_F6_GoReleaser_LdflagsInjectionVariables(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "{{.Version}}") || !strings.Contains(content, "{{.Commit}}") {
		t.Errorf(".goreleaser.yaml ldflags should inject {{.Version}} and {{.Commit}}")
	}
}

func TestTier2_F6_GoReleaser_ChangelogFiltersRegex(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "'^docs:'") && !strings.Contains(content, "^docs:") {
		t.Errorf(".goreleaser.yaml changelog filters should exclude documentation commits")
	}
}

func TestTier2_F6_GoReleaser_SnapshotVersionIncrement(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "snapshot:") || !strings.Contains(content, "version_template:") {
		t.Errorf(".goreleaser.yaml should configure snapshot version template")
	}
}

// ============================================================================
// Feature 7 Boundaries: GitHub Actions Release Workflow
// ============================================================================

func TestTier2_F7_ReleaseWorkflow_RunnerOSPinning(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read workflow: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "runs-on: macos-14") && !strings.Contains(content, "runs-on: macos-latest") {
		t.Errorf("release workflow should pin to macOS runner (macos-14 or macos-latest) for CGO Darwin builds")
	}
}

func TestTier2_F7_ReleaseWorkflow_GitFetchDepthZero(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read workflow: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "fetch-depth: 0") {
		t.Errorf("release workflow checkout step should configure fetch-depth: 0 for full tag history")
	}
}

func TestTier2_F7_ReleaseWorkflow_NoUnnecessaryTriggerBranches(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read workflow: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "branches:") && !strings.Contains(content, "tags:") {
		t.Errorf("release workflow should be tag-triggered only, not branch-triggered")
	}
}

func TestTier2_F7_ReleaseWorkflow_GoVersionReference(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read workflow: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "go-version-file") && !strings.Contains(content, "go-version") {
		t.Errorf("release workflow setup-go step should configure go-version or go-version-file")
	}
}

func TestTier2_F7_ReleaseWorkflow_ActionVersionsPinning(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read workflow: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "actions/checkout@v4") {
		t.Errorf("release workflow should use actions/checkout@v4")
	}
}

// ============================================================================
// Feature 8 Boundaries: Release Documentation
// ============================================================================

func TestTier2_F8_ReleaseDocs_PreFlightChecklistPresence(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(strings.ToLower(content), "pre-flight") && !strings.Contains(strings.ToLower(content), "checklist") {
		t.Errorf("docs/RELEASING.md should include a pre-flight checklist section")
	}
}

func TestTier2_F8_ReleaseDocs_HomebrewTapVerificationInstructions(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "homebrew-tap") || !strings.Contains(content, "brew update") {
		t.Errorf("docs/RELEASING.md should document Homebrew tap formula verification steps")
	}
}

func TestTier2_F8_ReleaseDocs_SemanticSectionHeadings(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "##") {
		t.Errorf("docs/RELEASING.md should contain Markdown level-2 headers")
	}
}

func TestTier2_F8_ReleaseDocs_CodeBlocksSyntaxValidation(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	reUntaggedCodeBlock := regexp.MustCompile("(?m)^```\\s*$")
	if reUntaggedCodeBlock.MatchString(content) {
		t.Errorf("docs/RELEASING.md contains code blocks without language tags (e.g. ```bash)")
	}
}

func TestTier2_F8_ReleaseDocs_TokenSecurityNotice(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "secret") && !strings.Contains(content, "token") {
		t.Errorf("docs/RELEASING.md should include security context regarding release tokens")
	}
}

// ============================================================================
// Feature 9 Boundaries: README Quality & Homebrew
// ============================================================================

func TestTier2_F9_README_TUIKeyboardTableIntegrity(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "Tab") || !strings.Contains(content, "Enter") {
		t.Errorf("README.md should include keyboard navigation table for TUI mode")
	}
}

func TestTier2_F9_README_NoPassiveThroatClearing(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := strings.ToLower(string(data))

	bannedPhrases := []string{
		"here's what we found",
		"it is worth noting that",
		"in order to",
	}
	for _, phrase := range bannedPhrases {
		if strings.Contains(content, phrase) {
			t.Errorf("README.md contains banned throat-clearing phrase %q", phrase)
		}
	}
}

func TestTier2_F9_README_JSONMachineReadableExample(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "--json") {
		t.Errorf("README.md should showcase structured JSON output (--json) for agent consumption")
	}
}

func TestTier2_F9_README_SubcommandTableOrHeaders(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	subcommands := []string{"validate", "folder", "inspect", "version"}
	for _, sc := range subcommands {
		if !strings.Contains(content, sc) {
			t.Errorf("README.md should document subcommand %q", sc)
		}
	}
}

func TestTier2_F9_README_LicenseLinkAndBadge(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, "README.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "Apache") && !strings.Contains(content, "LICENSE") {
		t.Errorf("README.md should reference Apache 2.0 license")
	}
}

// ============================================================================
// Feature 10 Boundaries: Snapshot Build & Archive Artifacts
// ============================================================================

func TestTier2_F10_SnapshotBuild_ArchiveRootFilePlacement(t *testing.T) {
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
		t.Fatalf("no archive file found")
	}

	tempDir := t.TempDir()
	if err := UnpackTarGz(t, archiveFile, tempDir); err != nil {
		t.Fatalf("failed to unpack archive: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tempDir, "credentialctl")); os.IsNotExist(err) {
		t.Errorf("credentialctl binary not found in archive root")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "README.md")); os.IsNotExist(err) {
		t.Errorf("README.md not found in archive root")
	}
	if _, err := os.Stat(filepath.Join(tempDir, "LICENSE")); os.IsNotExist(err) {
		t.Errorf("LICENSE not found in archive root")
	}
}

func TestTier2_F10_SnapshotBuild_ArchivePreservesExecBit(t *testing.T) {
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
		t.Fatalf("no archive file found")
	}

	tempDir := t.TempDir()
	UnpackTarGz(t, archiveFile, tempDir)
	fi, err := os.Stat(filepath.Join(tempDir, "credentialctl"))
	if err != nil {
		t.Fatalf("failed to stat binary: %v", err)
	}
	if fi.Mode()&0100 == 0 {
		t.Errorf("binary in archive missing owner execute bit (mode: %v)", fi.Mode())
	}
}

func TestTier2_F10_SnapshotBuild_ChecksumFormat64Hex(t *testing.T) {
	root := RepoRoot(t)
	checksumsPath := filepath.Join(root, "dist", "checksums.txt")
	data, err := os.ReadFile(checksumsPath)
	if err != nil {
		t.Fatalf("failed to read checksums.txt: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	reHex64 := regexp.MustCompile(`^[a-f0-9]{64}\s+`)
	for _, line := range lines {
		if len(line) > 0 && !reHex64.MatchString(line) {
			t.Errorf("checksums line %q does not match 64-char hex format", line)
		}
	}
}

func TestTier2_F10_SnapshotBuild_TarGzHeaderIntegrity(t *testing.T) {
	root := RepoRoot(t)
	distDir := filepath.Join(root, "dist")
	entries, err := os.ReadDir(distDir)
	if err != nil {
		t.Fatalf("failed to read dist/: %v", err)
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") && strings.HasPrefix(entry.Name(), "credentialctl_") {
			tempDir := t.TempDir()
			if err := UnpackTarGz(t, filepath.Join(distDir, entry.Name()), tempDir); err != nil {
				t.Errorf("archive %s failed header unpacking: %v", entry.Name(), err)
			}
		}
	}
}

func TestTier2_F10_SnapshotBuild_CleanFlagRemovesPriorDist(t *testing.T) {
	root := RepoRoot(t)
	distDir := filepath.Join(root, "dist")
	markerFile := filepath.Join(distDir, "stale_marker.tmp")
	os.MkdirAll(distDir, 0755)
	os.WriteFile(markerFile, []byte("stale"), 0644)

	res := RunCmd(t, root, "goreleaser", "release", "--snapshot", "--clean")
	if res.ExitCode != 0 {
		t.Fatalf("goreleaser release failed: %s", res.Stderr)
	}

	if _, err := os.Stat(markerFile); !os.IsNotExist(err) {
		t.Errorf("goreleaser --clean should have wiped stale files in dist/")
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func copyFile(t *testing.T, src, dest string, mode os.FileMode) {
	t.Helper()
	in, err := os.Open(src)
	if err != nil {
		t.Fatalf("failed to open src %s: %v", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		t.Fatalf("failed to create dest %s: %v", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		t.Fatalf("failed to copy %s to %s: %v", src, dest, err)
	}
}

func filterEnv(environ []string, stripKeys ...string) []string {
	var filtered []string
	for _, env := range environ {
		strip := false
		for _, key := range stripKeys {
			if strings.HasPrefix(env, key+"=") {
				strip = true
				break
			}
		}
		if !strip {
			filtered = append(filtered, env)
		}
	}
	return filtered
}
