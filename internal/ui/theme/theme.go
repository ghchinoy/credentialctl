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

package theme

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/ghchinoy/credentio-contributions/go"
)

// Semantic Color Palette
var (
	ColorAccent = lipgloss.Color("#7D56F4") // Vibrant Indigo / Purple
	ColorPass   = lipgloss.Color("#04B575") // Emerald Green
	ColorWarn   = lipgloss.Color("#FFB300") // Amber Yellow
	ColorFail   = lipgloss.Color("#FF4D4D") // Crimson Red
	ColorMuted  = lipgloss.Color("#626262") // Slate Gray
	ColorID     = lipgloss.Color("#00D7D7") // Cyan / Electric Blue
	ColorBg     = lipgloss.Color("#1B1B26") // Deep Background
	ColorFg     = lipgloss.Color("#EEEEEE") // Primary Foreground
)

// Semantic Styles
var (
	// Text styles
	StyleAccent = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
	StylePass   = lipgloss.NewStyle().Foreground(ColorPass).Bold(true)
	StyleWarn   = lipgloss.NewStyle().Foreground(ColorWarn).Bold(true)
	StyleFail   = lipgloss.NewStyle().Foreground(ColorFail).Bold(true)
	StyleMuted  = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleID     = lipgloss.NewStyle().Foreground(ColorID)
	StyleBold   = lipgloss.NewStyle().Bold(true)

	// Container & Layout styles
	StyleHeader = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(ColorAccent).
			Bold(true).
			Padding(0, 1)

	StyleSubHeader = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	StyleCard = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)

	StyleActiveTab = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorAccent).
			Foreground(ColorAccent).
			Bold(true).
			Padding(0, 2)

	StyleInactiveTab = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(ColorMuted).
				Foreground(ColorMuted).
				Padding(0, 2)

	// Status Badges
	BadgeSignedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorPass).
				Bold(true).
				Padding(0, 1)

	BadgeUnsignedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#000000")).
				Background(ColorWarn).
				Bold(true).
				Padding(0, 1)

	BadgeInvalidStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(ColorFail).
				Bold(true).
				Padding(0, 1)
)

// RenderBadge returns a formatted, semantic badge for a C2PA BadgeState.
func RenderBadge(state credentio.BadgeState) string {
	switch state {
	case credentio.BadgeSigned:
		return BadgeSignedStyle.Render("✓ SIGNED")
	case credentio.BadgeUnsigned:
		return BadgeUnsignedStyle.Render("⊘ UNSIGNED")
	case credentio.BadgeInvalid:
		return BadgeInvalidStyle.Render("✕ INVALID")
	default:
		return StyleMuted.Render(string(state))
	}
}

// RenderSeverity returns a formatted severity string.
func RenderSeverity(sev credentio.Severity) string {
	switch sev {
	case credentio.SeverityError:
		return StyleFail.Render("ERROR")
	case credentio.SeverityWarning:
		return StyleWarn.Render("WARN")
	case credentio.SeverityInfo:
		return StylePass.Render("INFO")
	default:
		return StyleMuted.Render(string(sev))
	}
}
