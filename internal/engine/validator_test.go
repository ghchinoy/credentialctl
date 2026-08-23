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
	"os"
	"path/filepath"
	"testing"

	"github.com/ghchinoy/credentio-contributions/go"
)

func TestValidatorService(t *testing.T) {
	vs, err := NewValidatorService(credentio.WithSkipTrustChecks(true))
	if err != nil {
		t.Fatalf("failed to create validator service: %v", err)
	}
	defer vs.Close()

	tempDir := t.TempDir()
	dummyFile := filepath.Join(tempDir, "sample.jpg")
	if err := os.WriteFile(dummyFile, []byte("not a real jpeg"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}

	report, err := vs.ValidateFile(dummyFile, "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error validating file: %v", err)
	}

	if report.Badge() != credentio.BadgeUnsigned {
		t.Errorf("expected badge to be unsigned, got %s", report.Badge())
	}
}
