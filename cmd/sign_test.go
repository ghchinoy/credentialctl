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
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestKeygenAndSignRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	pkiDir := filepath.Join(tmpDir, "test_pki")

	// 1. Run keygen --json
	keygenBuf := new(bytes.Buffer)
	keygenCmd.SetOut(keygenBuf)
	keygenCmd.SetErr(keygenBuf)
	keygenOutDir = pkiDir
	keygenOrg = "CredentialCtl Test"
	keygenJSON = true
	defer func() { keygenJSON = false }()

	if err := keygenCmd.RunE(keygenCmd, []string{}); err != nil {
		t.Fatalf("keygenCmd.RunE failed: %v", err)
	}

	var pkiOut jsonKeygenOutput
	if err := json.Unmarshal(keygenBuf.Bytes(), &pkiOut); err != nil {
		t.Fatalf("failed to decode keygen JSON output: %v", err)
	}

	// 2. Create a sample 16x16 PNG
	inPath := filepath.Join(tmpDir, "input.png")
	outPath := filepath.Join(tmpDir, "signed.png")
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: 50, G: 180, B: 100, A: 255})
		}
	}
	f, err := os.Create(inPath)
	if err != nil {
		t.Fatalf("failed to create input PNG: %v", err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatalf("failed to encode PNG: %v", err)
	}
	_ = f.Close()

	// 3. Run sign --json using generated PKI
	signBuf := new(bytes.Buffer)
	signCmd.SetOut(signBuf)
	signCmd.SetErr(signBuf)
	signOutputPath = outPath
	signKeyPath = pkiOut.ClaimSignerKey
	signCertPath = pkiOut.ClaimSignerCert
	signTSAKeyPath = pkiOut.TSAKey
	signTSACertPath = pkiOut.TSACert
	signSourceType = "ai"
	signAIOptOut = true
	signJSON = true
	defer func() {
		signJSON = false
		signAIOptOut = false
		signKeyPath = ""
		signCertPath = ""
		signTSAKeyPath = ""
		signTSACertPath = ""
	}()

	if err := signCmd.RunE(signCmd, []string{inPath}); err != nil {
		t.Fatalf("signCmd.RunE failed: %v", err)
	}

	var signOut jsonSignOutput
	if err := json.Unmarshal(signBuf.Bytes(), &signOut); err != nil {
		t.Fatalf("failed to decode sign JSON output: %v (raw: %s)", err, signBuf.String())
	}

	if signOut.VerifiedBadge != "signed" {
		t.Fatalf("expected verified_badge to be 'signed', got %q", signOut.VerifiedBadge)
	}
	if signOut.ByteSize == 0 {
		t.Errorf("expected non-zero signed file size")
	}
	if signOut.ActiveManifest == nil {
		t.Fatalf("expected active_manifest in sign output")
	}
}
