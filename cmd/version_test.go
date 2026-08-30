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
	"bytes"
	"encoding/json"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestGetVersionInfo(t *testing.T) {
	info := GetVersionInfo()

	if info.Version == "" {
		t.Errorf("expected non-empty Version, got empty string")
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("expected GoVersion %q, got %q", runtime.Version(), info.GoVersion)
	}
	if info.Compiler != runtime.Compiler {
		t.Errorf("expected Compiler %q, got %q", runtime.Compiler, info.Compiler)
	}
	expectedPlatform := runtime.GOOS + "/" + runtime.GOARCH
	if info.Platform != expectedPlatform {
		t.Errorf("expected Platform %q, got %q", expectedPlatform, info.Platform)
	}
	if info.CredentioEngineVersion == "" {
		t.Errorf("expected non-empty CredentioEngineVersion, got empty string")
	}
}

func TestVersionInfo_JSONTags(t *testing.T) {
	typ := reflect.TypeOf(VersionInfo{})
	expectedTags := map[string]string{
		"Version":                "version",
		"GitCommit":              "git_commit",
		"BuildDate":              "build_date",
		"GoVersion":              "go_version",
		"Compiler":               "compiler",
		"Platform":               "platform",
		"CredentioEngineVersion": "credentio_engine_version",
	}

	for fieldName, expectedJSONKey := range expectedTags {
		field, ok := typ.FieldByName(fieldName)
		if !ok {
			t.Fatalf("VersionInfo missing field %q", fieldName)
		}
		tag := field.Tag.Get("json")
		if tag != expectedJSONKey {
			t.Errorf("field %q json tag = %q, want %q", fieldName, tag, expectedJSONKey)
		}
	}
}

func TestVersionCommand_HumanOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)
	versionJSON = false

	err := versionCmd.RunE(versionCmd, []string{})
	if err != nil {
		t.Fatalf("versionCmd.RunE failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "credentialctl") {
		t.Errorf("human output missing 'credentialctl', got:\n%s", out)
	}
	if !strings.Contains(out, "Credentio") {
		t.Errorf("human output missing 'Credentio', got:\n%s", out)
	}
	if !strings.Contains(out, "Git Commit") {
		t.Errorf("human output missing 'Git Commit', got:\n%s", out)
	}
	if !strings.Contains(out, "Go Runtime") {
		t.Errorf("human output missing 'Go Runtime', got:\n%s", out)
	}
}

func TestVersionCommand_JSONOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	versionCmd.SetOut(buf)
	versionCmd.SetErr(buf)
	versionJSON = true
	defer func() { versionJSON = false }()

	err := versionCmd.RunE(versionCmd, []string{})
	if err != nil {
		t.Fatalf("versionCmd.RunE with --json failed: %v", err)
	}

	raw := buf.String()
	// Must not contain ANSI escape codes
	if strings.Contains(raw, "\x1b[") || strings.Contains(raw, "\033[") {
		t.Errorf("version --json output must not contain ANSI escape codes: %q", raw)
	}

	var parsed VersionInfo
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		t.Fatalf("failed to decode JSON output into VersionInfo: %v. Raw:\n%s", err, raw)
	}

	if parsed.Version == "" {
		t.Errorf("parsed Version is empty")
	}
	if parsed.Platform == "" {
		t.Errorf("parsed Platform is empty")
	}
	if parsed.CredentioEngineVersion == "" {
		t.Errorf("parsed CredentioEngineVersion is empty")
	}
}

func TestVersionCommand_ThreePillarsDocs(t *testing.T) {
	if versionCmd.GroupID != "inspection" {
		t.Errorf("expected GroupID 'inspection', got %q", versionCmd.GroupID)
	}
	if strings.TrimSpace(versionCmd.Short) == "" {
		t.Errorf("expected non-empty Short documentation")
	}
	if strings.TrimSpace(versionCmd.Long) == "" {
		t.Errorf("expected non-empty Long documentation")
	}
	if strings.TrimSpace(versionCmd.Example) == "" {
		t.Errorf("expected non-empty Example documentation")
	}
}

func TestFormatHumanVersion_UnavailableEngine(t *testing.T) {
	info := VersionInfo{
		Version:                "0.1.5-test",
		GitCommit:              "abc1234",
		BuildDate:              "2026-08-30T00:00:00Z",
		GoVersion:              "go1.22.0",
		Compiler:               "gc",
		Platform:               "darwin/arm64",
		CredentioEngineVersion: "unavailable",
	}

	rendered := formatHumanVersion(info)
	if !strings.Contains(rendered, "unavailable") {
		t.Errorf("expected formatHumanVersion to display 'unavailable', got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "credentialctl v0.1.5-test") {
		t.Errorf("expected title in rendered output, got:\n%s", rendered)
	}
}

func TestRootCmd_VersionTemplate(t *testing.T) {
	tmpl := rootCmd.VersionTemplate()
	if !strings.Contains(tmpl, "{{.Version}}") {
		t.Errorf("expected VersionTemplate to contain {{.Version}}, got: %s", tmpl)
	}
	flag := rootCmd.Flags().Lookup("version")
	if flag == nil {
		t.Fatalf("expected rootCmd to have 'version' flag")
	}
	if flag.Shorthand != "v" {
		t.Errorf("expected shorthand 'v', got %q", flag.Shorthand)
	}
}

func TestFormatHumanVersion_ValidEngine(t *testing.T) {
	info := VersionInfo{
		Version:                "0.1.5",
		GitCommit:              "abcdef0",
		BuildDate:              "2026-08-30T00:00:00Z",
		GoVersion:              "go1.25.0",
		Compiler:               "gc",
		Platform:               "darwin/arm64",
		CredentioEngineVersion: "0.1.5",
	}

	rendered := formatHumanVersion(info)
	if !strings.Contains(rendered, "v0.1.5") {
		t.Errorf("expected formatHumanVersion to display 'v0.1.5', got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "C-ABI Shared Library") {
		t.Errorf("expected formatHumanVersion to display '(C-ABI Shared Library)', got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "credentialctl v0.1.5") {
		t.Errorf("expected title in rendered output, got:\n%s", rendered)
	}
}

func TestVersionCommand_RejectsArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"version", "unexpected-arg"})
	err := rootCmd.Execute()
	if err == nil {
		t.Errorf("expected error when passing extra args to version command, got nil")
	}
}
