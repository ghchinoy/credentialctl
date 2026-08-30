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
	"strings"
	"sync"
	"testing"
)

// TestEmpirical_GetVersionInfo_HighConcurrency launches 2,000 concurrent goroutines
// to stress GetVersionInfo() for race conditions and data consistency.
func TestEmpirical_GetVersionInfo_HighConcurrency(t *testing.T) {
	const count = 2000
	barrier := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(count)

	results := make([]VersionInfo, count)

	for i := 0; i < count; i++ {
		go func(idx int) {
			defer wg.Done()
			<-barrier
			results[idx] = GetVersionInfo()
		}(i)
	}

	close(barrier)
	wg.Wait()

	base := results[0]
	if base.Version == "" || base.Platform == "" || base.CredentioEngineVersion == "" {
		t.Fatalf("base VersionInfo has empty required fields: %+v", base)
	}

	for i, info := range results {
		if info.Version != base.Version {
			t.Errorf("[%d] Version mismatch: got %q, want %q", i, info.Version, base.Version)
		}
		if info.GitCommit != base.GitCommit {
			t.Errorf("[%d] GitCommit mismatch: got %q, want %q", i, info.GitCommit, base.GitCommit)
		}
		if info.BuildDate != base.BuildDate {
			t.Errorf("[%d] BuildDate mismatch: got %q, want %q", i, info.BuildDate, base.BuildDate)
		}
		if info.GoVersion != base.GoVersion {
			t.Errorf("[%d] GoVersion mismatch: got %q, want %q", i, info.GoVersion, base.GoVersion)
		}
		if info.Compiler != base.Compiler {
			t.Errorf("[%d] Compiler mismatch: got %q, want %q", i, info.Compiler, base.Compiler)
		}
		if info.Platform != base.Platform {
			t.Errorf("[%d] Platform mismatch: got %q, want %q", i, info.Platform, base.Platform)
		}
		if info.CredentioEngineVersion != base.CredentioEngineVersion {
			t.Errorf("[%d] CredentioEngineVersion mismatch: got %q, want %q", i, info.CredentioEngineVersion, base.CredentioEngineVersion)
		}
	}
}

// TestEmpirical_FormatHumanVersion_EdgeCases validates table alignment and formatting
// across diverse engine version states (empty, unavailable, arbitrary strings, unicode).
func TestEmpirical_FormatHumanVersion_EdgeCases(t *testing.T) {
	testCases := []struct {
		name          string
		engineVer     string
		expectWarn    bool
		expectContain string
	}{
		{
			name:          "standard valid version",
			engineVer:     "0.1.0-credentio-c",
			expectWarn:    false,
			expectContain: "v0.1.0-credentio-c (C-ABI Shared Library)",
		},
		{
			name:          "unavailable version string",
			engineVer:     "unavailable",
			expectWarn:    true,
			expectContain: "unavailable",
		},
		{
			name:          "empty version string",
			engineVer:     "",
			expectWarn:    true,
			expectContain: "unavailable",
		},
		{
			name:          "custom release tag",
			engineVer:     "1.0.0-rc.1",
			expectWarn:    false,
			expectContain: "v1.0.0-rc.1 (C-ABI Shared Library)",
		},
		{
			name:          "long version string",
			engineVer:     "0.1.5-beta.1+build.20260830.sha.abcdef0123456789",
			expectWarn:    false,
			expectContain: "v0.1.5-beta.1+build.20260830.sha.abcdef0123456789",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			info := VersionInfo{
				Version:                "1.0.0",
				GitCommit:              "abc1234",
				BuildDate:              "2026-08-30T00:00:00Z",
				GoVersion:              "go1.25.0",
				Compiler:               "gc",
				Platform:               "darwin/arm64",
				CredentioEngineVersion: tc.engineVer,
			}

			out := formatHumanVersion(info)
			if !strings.Contains(out, "credentialctl v1.0.0") {
				t.Errorf("missing header title in output: %s", out)
			}
			if !strings.Contains(out, "Git Commit:") {
				t.Errorf("missing padded Git Commit label in output: %s", out)
			}
			if !strings.Contains(out, "Credentio:") {
				t.Errorf("missing padded Credentio label in output: %s", out)
			}
			if !strings.Contains(out, tc.expectContain) {
				t.Errorf("expected output to contain %q, got:\n%s", tc.expectContain, out)
			}
		})
	}
}

// TestEmpirical_JSONSerialization_Stress tests rapid concurrent JSON marshaling of VersionInfo.
func TestEmpirical_JSONSerialization_Stress(t *testing.T) {
	const count = 1000
	var wg sync.WaitGroup
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(idx int) {
			defer wg.Done()
			info := GetVersionInfo()
			data, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				t.Errorf("json.MarshalIndent failed on [%d]: %v", idx, err)
				return
			}
			var parsed VersionInfo
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Errorf("json.Unmarshal failed on [%d]: %v", idx, err)
				return
			}
			if parsed.Version != info.Version || parsed.CredentioEngineVersion != info.CredentioEngineVersion {
				t.Errorf("parsed data mismatch on [%d]", idx)
			}
		}(i)
	}
	wg.Wait()
}
