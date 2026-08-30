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
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// ============================================================================
// Feature 1: Upstream Go Module v0.1.5 (ORIGINAL_REQUEST §R1)
// ============================================================================

func TestTier1_F1_UpstreamGoModule_RequirementDirective(t *testing.T) {
	root := RepoRoot(t)
	goModPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}
	content := string(data)

	expectedDep := "github.com/ghchinoy/credentio-contributions/go v0.1.5"
	if !strings.Contains(content, expectedDep) {
		t.Errorf("go.mod does not contain expected requirement '%s'. Found:\n%s", expectedDep, content)
	}
}

func TestTier1_F1_UpstreamGoModule_GoSumChecksums(t *testing.T) {
	root := RepoRoot(t)
	goSumPath := filepath.Join(root, "go.sum")
	data, err := os.ReadFile(goSumPath)
	if err != nil {
		t.Fatalf("failed to read go.sum: %v", err)
	}
	content := string(data)

	expectedEntry := "github.com/ghchinoy/credentio-contributions/go v0.1.5"
	if !strings.Contains(content, expectedEntry) {
		t.Errorf("go.sum does not contain checksum entry for '%s'", expectedEntry)
	}
}

func TestTier1_F1_UpstreamGoModule_GoListResolution(t *testing.T) {
	root := RepoRoot(t)
	res := RunCmd(t, root, "go", "list", "-m", "github.com/ghchinoy/credentio-contributions/go")
	if res.ExitCode != 0 {
		t.Fatalf("go list failed: %v, stderr: %s", res.Err, res.Stderr)
	}

	expected := "github.com/ghchinoy/credentio-contributions/go v0.1.5"
	if !strings.Contains(strings.TrimSpace(res.Stdout), expected) {
		t.Errorf("go list output mismatch: expected %q, got %q", expected, strings.TrimSpace(res.Stdout))
	}
}

func TestTier1_F1_UpstreamGoModule_GoVersionMinimum(t *testing.T) {
	root := RepoRoot(t)
	goModPath := filepath.Join(root, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		t.Fatalf("failed to read go.mod: %v", err)
	}

	re := regexp.MustCompile(`go\s+1\.(2[2-9]|[3-9][0-9])`)
	if !re.Match(data) {
		t.Errorf("go.mod should require at least Go 1.22 for credentio-contributions v0.1.5 compatibility")
	}
}

func TestTier1_F1_UpstreamGoModule_ModVerifyClean(t *testing.T) {
	root := RepoRoot(t)
	res := RunCmd(t, root, "go", "mod", "verify")
	if res.ExitCode != 0 {
		t.Errorf("go mod verify failed: %v, stderr: %s", res.Err, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "all modules verified") {
		t.Logf("go mod verify stdout: %s", res.Stdout)
	}
}

// ============================================================================
// Feature 2: Native Library Fetch Script Sync (ORIGINAL_REQUEST §R1)
// ============================================================================

func TestTier1_F2_FetchScript_ExecutablePermissions(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	fi, err := os.Stat(scriptPath)
	if err != nil {
		t.Fatalf("fetch script not found at %s: %v", scriptPath, err)
	}
	if fi.Mode()&0111 == 0 {
		t.Errorf("fetch script %s is not executable (mode: %v)", scriptPath, fi.Mode())
	}
}

func TestTier1_F2_FetchScript_DefaultVersionParameter(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read fetch script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, `VERSION="${1:-0.1.5}"`) && !strings.Contains(content, `VERSION="${1:-v0.1.5}"`) {
		t.Errorf("fetch script default version must be 0.1.5, found:\n%s", content)
	}
}

func TestTier1_F2_FetchScript_UpstreamRepoTarget(t *testing.T) {
	root := RepoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "fetch-credentio-lib.sh")
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("failed to read fetch script: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "ghchinoy/credentio-contributions") {
		t.Errorf("fetch script does not target ghchinoy/credentio-contributions")
	}
}

func TestTier1_F2_FetchScript_StagedSharedLibraryExtension(t *testing.T) {
	root := RepoRoot(t)
	libDir := filepath.Join(root, "third_party", "credentio", "lib")

	expectedLib := "libcredentio_c.dylib"
	if runtime.GOOS == "linux" {
		expectedLib = "libcredentio_c.so"
	}

	targetPath := filepath.Join(libDir, expectedLib)
	fi, err := os.Stat(targetPath)
	if err != nil {
		t.Fatalf("expected native shared library %s does not exist in %s: %v", expectedLib, libDir, err)
	}
	if fi.Size() < 1024*1024 { // at least 1MB
		t.Errorf("native library size %d bytes seems too small", fi.Size())
	}
}

func TestTier1_F2_FetchScript_HeaderDeclarations(t *testing.T) {
	root := RepoRoot(t)
	headerPath := filepath.Join(root, "third_party", "credentio", "include", "credentio_c.h")
	data, err := os.ReadFile(headerPath)
	if err != nil {
		t.Fatalf("failed to read C-ABI header %s: %v", headerPath, err)
	}
	content := string(data)

	requiredDeclarations := []string{
		"cr_validator_create",
		"cr_validate_file",
		"cr_validate_bytes",
		"cr_version",
		"cr_string_free",
	}

	for _, decl := range requiredDeclarations {
		if !strings.Contains(content, decl) {
			t.Errorf("credentio_c.h missing required C-ABI declaration '%s'", decl)
		}
	}
}

// ============================================================================
// Feature 3: Dynamic Multi-Platform RPATHs (ORIGINAL_REQUEST §R2)
// ============================================================================

func TestTier1_F3_DynamicRPATH_LoaderPathEmbedded(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Mach-O LC_RPATH checks are Darwin-specific")
	}
	bin := EnsureBuiltBinary(t)
	rpaths := GetBinaryRPATHs(t, bin)

	foundLoaderPath := false
	for _, rp := range rpaths {
		if rp == "@loader_path" || rp == "@executable_path" {
			foundLoaderPath = true
			break
		}
	}
	if !foundLoaderPath {
		t.Errorf("binary at %s missing @loader_path or @executable_path in RPATHs. Found: %v", bin, rpaths)
	}
}

func TestTier1_F3_DynamicRPATH_ExecutablePathEmbedded(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Mach-O LC_RPATH checks are Darwin-specific")
	}
	bin := EnsureBuiltBinary(t)
	rpaths := GetBinaryRPATHs(t, bin)

	found := false
	for _, rp := range rpaths {
		if rp == "@executable_path" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("binary at %s missing @executable_path in RPATHs. Found: %v", bin, rpaths)
	}
}

func TestTier1_F3_DynamicRPATH_ExecutablePathCellarLibEmbedded(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Mach-O LC_RPATH checks are Darwin-specific")
	}
	bin := EnsureBuiltBinary(t)
	rpaths := GetBinaryRPATHs(t, bin)

	found := false
	for _, rp := range rpaths {
		if rp == "@executable_path/../lib" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("binary at %s missing @executable_path/../lib in RPATHs for Homebrew Cellar support. Found: %v", bin, rpaths)
	}
}

func TestTier1_F3_DynamicRPATH_MakefileLDFLAGSConfiguration(t *testing.T) {
	root := RepoRoot(t)
	makefilePath := filepath.Join(root, "Makefile")
	data, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("failed to read Makefile: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "-Wl,-rpath,@loader_path") && !strings.Contains(content, "-Wl,-rpath,@executable_path") {
		t.Errorf("Makefile CREDENTIO_LDFLAGS must configure dynamic RPATHs. Found:\n%s", content)
	}
}

func TestTier1_F3_DynamicRPATH_ArbitraryCwdExecution(t *testing.T) {
	tempDir := t.TempDir()
	res := RunCredentialctl(t, tempDir, "--help")
	if res.ExitCode != 0 {
		t.Errorf("running credentialctl from arbitrary directory %s failed with exit code %d: %s", tempDir, res.ExitCode, res.Stderr)
	}
}

// ============================================================================
// Feature 4: Runtime C-ABI Version Discovery (ORIGINAL_REQUEST §R4)
// ============================================================================

func TestTier1_F4_RuntimeVersion_ExportedSymbolPresent(t *testing.T) {
	ver := QueryNativeCRVersion()
	if ver == "" {
		t.Fatalf("QueryNativeCRVersion() returned empty string")
	}
}

func TestTier1_F4_RuntimeVersion_NonEmptyReturn(t *testing.T) {
	ver := QueryNativeCRVersion()
	if len(ver) < 3 {
		t.Errorf("QueryNativeCRVersion() returned unexpectedly short string: %q", ver)
	}
}

func TestTier1_F4_RuntimeVersion_SemverPatternMatch(t *testing.T) {
	ver := QueryNativeCRVersion()
	re := regexp.MustCompile(`^v?\d+\.\d+\.\d+`)
	if !re.MatchString(ver) {
		t.Errorf("runtime version %q does not match semver pattern", ver)
	}
}

func TestTier1_F4_RuntimeVersion_IdempotentCalls(t *testing.T) {
	v1 := QueryNativeCRVersion()
	v2 := QueryNativeCRVersion()
	v3 := QueryNativeCRVersion()
	if v1 != v2 || v2 != v3 {
		t.Errorf("QueryNativeCRVersion() is not idempotent: %q vs %q vs %q", v1, v2, v3)
	}
}

func TestTier1_F4_RuntimeVersion_ConcurrentQueries(t *testing.T) {
	const goroutines = 20
	var wg sync.WaitGroup
	results := make([]string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = QueryNativeCRVersion()
		}(i)
	}
	wg.Wait()

	base := results[0]
	for i, r := range results {
		if r != base {
			t.Errorf("goroutine %d got different version %q (expected %q)", i, r, base)
		}
	}
}

// ============================================================================
// Feature 5: Version Command & CLI Flags (ORIGINAL_REQUEST §R4)
// ============================================================================

func TestTier1_F5_VersionCommand_HumanOutput(t *testing.T) {
	res := RunCredentialctl(t, "", "version")
	if res.ExitCode != 0 {
		t.Fatalf("credentialctl version exited with code %d, stderr: %s", res.ExitCode, res.Stderr)
	}

	out := res.Stdout
	if !strings.Contains(out, "credentialctl") || !strings.Contains(out, "Credentio") {
		t.Errorf("human version output missing expected CLI and engine labels. Got:\n%s", out)
	}
}

func TestTier1_F5_VersionCommand_JSONStructuredOutput(t *testing.T) {
	res := RunCredentialctl(t, "", "version", "--json")
	if res.ExitCode != 0 {
		t.Fatalf("credentialctl version --json exited with code %d, stderr: %s", res.ExitCode, res.Stderr)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(res.Stdout), &payload); err != nil {
		t.Fatalf("credentialctl version --json did not return valid JSON: %v. Output:\n%s", err, res.Stdout)
	}

	requiredKeys := []string{"version", "git_commit", "build_date", "go_version", "compiler", "platform", "credentio_engine_version"}
	for _, k := range requiredKeys {
		if _, ok := payload[k]; !ok {
			t.Errorf("JSON output missing required key %q", k)
		}
	}
}

func TestTier1_F5_VersionCommand_LongFlagVersion(t *testing.T) {
	res := RunCredentialctl(t, "", "--version")
	if res.ExitCode != 0 {
		t.Fatalf("credentialctl --version exited with code %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	if !strings.Contains(res.Stdout, "credentialctl") && !strings.Contains(res.Stdout, "version") {
		t.Errorf("credentialctl --version output unexpected: %s", res.Stdout)
	}
}

func TestTier1_F5_VersionCommand_ShortFlagVersion(t *testing.T) {
	res := RunCredentialctl(t, "", "-v")
	if res.ExitCode != 0 {
		t.Fatalf("credentialctl -v exited with code %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	if len(strings.TrimSpace(res.Stdout)) == 0 {
		t.Errorf("credentialctl -v returned empty output")
	}
}

func TestTier1_F5_VersionCommand_AgentAwareHelpThreePillars(t *testing.T) {
	res := RunCredentialctl(t, "", "version", "--help")
	if res.ExitCode != 0 {
		t.Fatalf("credentialctl version --help exited with code %d, stderr: %s", res.ExitCode, res.Stderr)
	}
	out := res.Stdout
	if !strings.Contains(out, "Usage:") || !strings.Contains(out, "Flags:") {
		t.Errorf("help output missing Usage or Flags sections:\n%s", out)
	}
}

// ============================================================================
// Feature 6: GoReleaser Configuration & Brews (ORIGINAL_REQUEST §R3)
// ============================================================================

func TestTier1_F6_GoReleaser_SchemaVersion2(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "version: 2") {
		t.Errorf(".goreleaser.yaml must specify schema version: 2")
	}
}

func TestTier1_F6_GoReleaser_CGOFlagsAndRPATHs(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "CGO_ENABLED=1") {
		t.Errorf(".goreleaser.yaml missing CGO_ENABLED=1")
	}
	if !strings.Contains(content, "@executable_path") || !strings.Contains(content, "@executable_path/../lib") {
		t.Errorf(".goreleaser.yaml missing required multi-path RPATHs in CGO_LDFLAGS")
	}
}

func TestTier1_F6_GoReleaser_SharedLibraryArchivePackaging(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "libcredentio_c.dylib") && !strings.Contains(content, "libcredentio_c") {
		t.Errorf(".goreleaser.yaml archives section must package libcredentio_c shared library")
	}
}

func TestTier1_F6_GoReleaser_HomebrewTapTarget(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "brews:") || !strings.Contains(content, "homebrew-tap") {
		t.Errorf(".goreleaser.yaml missing brews configuration targeting homebrew-tap")
	}
}

func TestTier1_F6_GoReleaser_BrewsInstallAndTestBlocks(t *testing.T) {
	root := RepoRoot(t)
	path := filepath.Join(root, ".goreleaser.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read .goreleaser.yaml: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, `bin.install "credentialctl"`) {
		t.Errorf(".goreleaser.yaml brews stanza missing bin.install \"credentialctl\"")
	}
	if !strings.Contains(content, `lib.install "libcredentio_c.dylib"`) && !strings.Contains(content, `lib.install`) {
		t.Errorf(".goreleaser.yaml brews stanza missing lib.install for shared library")
	}
}

// ============================================================================
// Feature 7: GitHub Actions Release Workflow (ORIGINAL_REQUEST §R3)
// ============================================================================

func TestTier1_F7_ReleaseWorkflow_FileExistsAndValidYAML(t *testing.T) {
	root := RepoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("release workflow file %s does not exist: %v", workflowPath, err)
	}
	if len(data) == 0 {
		t.Errorf("release workflow file %s is empty", workflowPath)
	}
}

func TestTier1_F7_ReleaseWorkflow_TagPushTriggerFilter(t *testing.T) {
	root := RepoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", workflowPath, err)
	}
	content := string(data)

	if !strings.Contains(content, "tags:") || !strings.Contains(content, "'v*'") {
		t.Errorf("release workflow must be triggered on 'v*' tag pushes. Content:\n%s", content)
	}
}

func TestTier1_F7_ReleaseWorkflow_ContentsWritePermissions(t *testing.T) {
	root := RepoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", workflowPath, err)
	}
	content := string(data)

	if !strings.Contains(content, "contents: write") {
		t.Errorf("release workflow must declare 'contents: write' permission for GitHub Releases")
	}
}

func TestTier1_F7_ReleaseWorkflow_FetchScriptStepOrder(t *testing.T) {
	root := RepoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", workflowPath, err)
	}
	content := string(data)

	fetchIdx := strings.Index(content, "fetch-credentio-lib.sh")
	goreleaserIdx := strings.Index(content, "goreleaser-action")
	if fetchIdx == -1 {
		t.Errorf("release workflow must execute fetch-credentio-lib.sh before release")
	}
	if goreleaserIdx != -1 && fetchIdx > goreleaserIdx {
		t.Errorf("fetch-credentio-lib.sh must execute before goreleaser-action")
	}
}

func TestTier1_F7_ReleaseWorkflow_SecretInjections(t *testing.T) {
	root := RepoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "release.yaml")
	data, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("failed to read %s: %v", workflowPath, err)
	}
	content := string(data)

	if !strings.Contains(content, "HOMEBREW_TAP_GITHUB_TOKEN") {
		t.Errorf("release workflow must inject HOMEBREW_TAP_GITHUB_TOKEN secret into GoReleaser step")
	}
	if !strings.Contains(content, "GITHUB_TOKEN") {
		t.Errorf("release workflow must inject GITHUB_TOKEN secret into GoReleaser step")
	}
}

// ============================================================================
// Feature 8: Release Documentation (docs/RELEASING.md) (ORIGINAL_REQUEST §R5)
// ============================================================================

func TestTier1_F8_ReleaseDocs_FileExistsAndNonEmpty(t *testing.T) {
	root := RepoRoot(t)
	releasingPath := filepath.Join(root, "docs", "RELEASING.md")
	fi, err := os.Stat(releasingPath)
	if err != nil {
		t.Fatalf("docs/RELEASING.md does not exist: %v", err)
	}
	if fi.Size() < 500 {
		t.Errorf("docs/RELEASING.md is too short (%d bytes)", fi.Size())
	}
}

func TestTier1_F8_ReleaseDocs_DocumentedSecrets(t *testing.T) {
	root := RepoRoot(t)
	releasingPath := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(releasingPath)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "HOMEBREW_TAP_GITHUB_TOKEN") {
		t.Errorf("docs/RELEASING.md must document HOMEBREW_TAP_GITHUB_TOKEN secret")
	}
	if !strings.Contains(content, "GITHUB_TOKEN") {
		t.Errorf("docs/RELEASING.md must document GITHUB_TOKEN secret")
	}
}

func TestTier1_F8_ReleaseDocs_StepByStepInstructions(t *testing.T) {
	root := RepoRoot(t)
	releasingPath := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(releasingPath)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	requiredPhrases := []string{
		"git tag",
		"git push",
		"homebrew-tap",
	}
	for _, phrase := range requiredPhrases {
		if !strings.Contains(content, phrase) {
			t.Errorf("docs/RELEASING.md missing required instruction phrase %q", phrase)
		}
	}
}

func TestTier1_F8_ReleaseDocs_SnapshotTestingCommands(t *testing.T) {
	root := RepoRoot(t)
	releasingPath := filepath.Join(root, "docs", "RELEASING.md")
	data, err := os.ReadFile(releasingPath)
	if err != nil {
		t.Fatalf("failed to read docs/RELEASING.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "goreleaser release --snapshot --clean") && !strings.Contains(content, "make release-snapshot") {
		t.Errorf("docs/RELEASING.md must document local snapshot release verification command")
	}
}

func TestTier1_F8_ReleaseDocs_ZeroEmDashes(t *testing.T) {
	root := RepoRoot(t)
	releasingPath := filepath.Join(root, "docs", "RELEASING.md")
	audit := AuditEmDashes(t, releasingPath)
	if audit.TotalEmDashes > 0 {
		t.Errorf("docs/RELEASING.md violates editorial standard with %d em dashes:\n%s", audit.TotalEmDashes, strings.Join(audit.Occurrences, "\n"))
	}
}

// ============================================================================
// Feature 9: README Quality & Homebrew Instructions (ORIGINAL_REQUEST §R5)
// ============================================================================

func TestTier1_F9_README_HomebrewInstallationCommands(t *testing.T) {
	root := RepoRoot(t)
	readmePath := filepath.Join(root, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "brew tap ghchinoy/homebrew-tap") && !strings.Contains(content, "brew tap ghchinoy/tap") {
		t.Errorf("README.md missing Homebrew tap command (e.g. brew tap ghchinoy/homebrew-tap)")
	}
	if !strings.Contains(content, "brew install credentialctl") {
		t.Errorf("README.md missing brew install credentialctl command")
	}
}

func TestTier1_F9_README_ProjectDescriptionQuickStart(t *testing.T) {
	root := RepoRoot(t)
	readmePath := filepath.Join(root, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "C2PA") || !strings.Contains(content, "Content Credentials") {
		t.Errorf("README.md missing clear opening explanation of C2PA Content Credentials")
	}
}

func TestTier1_F9_README_CLIAndTUIDocumentation(t *testing.T) {
	root := RepoRoot(t)
	readmePath := filepath.Join(root, "README.md")
	data, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("failed to read README.md: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "validate") || !strings.Contains(content, "folder") || !strings.Contains(content, "inspect") {
		t.Errorf("README.md missing documentation for core CLI subcommands (validate, folder, inspect)")
	}
}

func TestTier1_F9_README_ZeroEmDashes(t *testing.T) {
	root := RepoRoot(t)
	readmePath := filepath.Join(root, "README.md")
	audit := AuditEmDashes(t, readmePath)
	if audit.TotalEmDashes > 0 {
		t.Errorf("README.md violates editorial standard with %d em dashes:\n%s", audit.TotalEmDashes, strings.Join(audit.Occurrences, "\n"))
	}
}

func TestTier1_F9_README_MarkAllenRubricScore(t *testing.T) {
	root := RepoRoot(t)
	readmePath := filepath.Join(root, "README.md")
	score := AuditMarkAllenRubric(t, readmePath)
	if score.TotalScore < 34 {
		t.Errorf("README.md Mark Allen rubric score is %d/40 (target >= 34). Feedback: %v", score.TotalScore, score.Feedback)
	}
}

// ============================================================================
// Feature 10: Snapshot Build & Archive Artifacts (ORIGINAL_REQUEST Acceptance)
// ============================================================================

func TestTier1_F10_SnapshotBuild_ExecutionSuccess(t *testing.T) {
	root := RepoRoot(t)
	res := RunCmd(t, root, "goreleaser", "release", "--snapshot", "--clean")
	if res.ExitCode != 0 {
		t.Fatalf("goreleaser release --snapshot --clean failed (exit %d): %s", res.ExitCode, res.Stderr)
	}
}

func TestTier1_F10_SnapshotBuild_ArchiveCreated(t *testing.T) {
	root := RepoRoot(t)
	distDir := filepath.Join(root, "dist")
	entries, err := os.ReadDir(distDir)
	if err != nil {
		t.Fatalf("failed to read dist directory %s: %v", distDir, err)
	}

	foundArchive := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") && strings.HasPrefix(entry.Name(), "credentialctl_") {
			foundArchive = true
			break
		}
	}
	if !foundArchive {
		t.Errorf("no credentialctl_*.tar.gz archive found in %s", distDir)
	}
}

func TestTier1_F10_SnapshotBuild_SharedLibraryIncluded(t *testing.T) {
	root := RepoRoot(t)
	distDir := filepath.Join(root, "dist")
	entries, err := os.ReadDir(distDir)
	if err != nil {
		t.Fatalf("failed to read dist directory: %v", err)
	}

	var archiveFile string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") && strings.HasPrefix(entry.Name(), "credentialctl_") {
			archiveFile = filepath.Join(distDir, entry.Name())
			break
		}
	}
	if archiveFile == "" {
		t.Fatalf("no archive file found to verify")
	}

	tempDir := t.TempDir()
	if err := UnpackTarGz(t, archiveFile, tempDir); err != nil {
		t.Fatalf("failed to unpack snapshot archive %s: %v", archiveFile, err)
	}

	expectedLib := "libcredentio_c.dylib"
	if runtime.GOOS == "linux" {
		expectedLib = "libcredentio_c.so"
	}

	libPath := filepath.Join(tempDir, expectedLib)
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		t.Errorf("archive %s does not contain bundled shared library %s in root", archiveFile, expectedLib)
	}
}

func TestTier1_F10_SnapshotBuild_ChecksumsFileGenerated(t *testing.T) {
	root := RepoRoot(t)
	checksumsPath := filepath.Join(root, "dist", "checksums.txt")
	data, err := os.ReadFile(checksumsPath)
	if err != nil {
		t.Fatalf("checksums.txt not found in dist/: %v", err)
	}
	content := string(data)
	if len(content) == 0 {
		t.Errorf("checksums.txt is empty")
	}
}

func TestTier1_F10_SnapshotBuild_ExecutablePermissions(t *testing.T) {
	root := RepoRoot(t)
	distDir := filepath.Join(root, "dist")
	entries, err := os.ReadDir(distDir)
	if err != nil {
		t.Fatalf("failed to read dist directory: %v", err)
	}

	var archiveFile string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tar.gz") && strings.HasPrefix(entry.Name(), "credentialctl_") {
			archiveFile = filepath.Join(distDir, entry.Name())
			break
		}
	}
	if archiveFile == "" {
		t.Fatalf("no archive file found to verify")
	}

	tempDir := t.TempDir()
	if err := UnpackTarGz(t, archiveFile, tempDir); err != nil {
		t.Fatalf("failed to unpack archive: %v", err)
	}

	binPath := filepath.Join(tempDir, "credentialctl")
	fi, err := os.Stat(binPath)
	if err != nil {
		t.Fatalf("credentialctl binary missing inside archive: %v", err)
	}
	if fi.Mode()&0111 == 0 {
		t.Errorf("credentialctl binary inside archive is not executable (mode: %v)", fi.Mode())
	}
}
