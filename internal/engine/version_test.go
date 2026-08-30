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

package engine

import (
	"strings"
	"sync"
	"testing"
)

func TestCredentioCVersion_NonEmpty(t *testing.T) {
	ver := CredentioCVersion()
	if ver == "" {
		t.Fatalf("CredentioCVersion() returned empty string")
	}
	if len(ver) < 3 {
		t.Errorf("CredentioCVersion() returned unexpectedly short version: %q", ver)
	}
}

func TestCredentioCVersion_ExpectedFormat(t *testing.T) {
	ver := CredentioCVersion()
	expected := "0.1.0-credentio-c"
	if ver != expected {
		t.Logf("Note: CredentioCVersion() returned %q (expected %q)", ver, expected)
	}
	if !strings.Contains(ver, "0.1.0") && !strings.Contains(ver, "credentio") {
		t.Errorf("CredentioCVersion() %q does not match expected credentio pattern", ver)
	}
}

func TestCredentioCVersion_NoNullByte(t *testing.T) {
	ver := CredentioCVersion()
	if strings.Contains(ver, "\x00") {
		t.Errorf("CredentioCVersion() returned string containing embedded null byte: %q", ver)
	}
}

func TestCredentioCVersion_Idempotent(t *testing.T) {
	v1 := CredentioCVersion()
	v2 := CredentioCVersion()
	v3 := CredentioCVersion()

	if v1 != v2 || v2 != v3 {
		t.Errorf("CredentioCVersion() is not idempotent: v1=%q, v2=%q, v3=%q", v1, v2, v3)
	}
}

func TestCredentioCVersion_Concurrent(t *testing.T) {
	const goroutines = 50
	var wg sync.WaitGroup
	results := make([]string, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = CredentioCVersion()
		}(i)
	}
	wg.Wait()

	base := results[0]
	if base == "" {
		t.Fatalf("concurrent CredentioCVersion() returned empty string")
	}
	for i, r := range results {
		if r != base {
			t.Errorf("goroutine %d returned %q, expected %q", i, r, base)
		}
	}
}
