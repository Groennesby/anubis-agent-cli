package cli

import (
	"fmt"
	"os"
	"strings"
)

// Theme defines the 24-bit TrueColor ANSI palette inspired by Hermes Classic + Egyptian Gold.
type Theme struct {
	Name            string
	BgDark          string // #1A1A2E / #0F0F1A
	PrimaryGold     string // #FFD700
	SecondaryGold   string // #E6C200
	BorderGold      string // #C9A227
	TextPrimary     string // #FFF8DC (Cornsilk)
	TextMuted       string // #D4AF37 / #8B7355
	HighlightRed    string // #FF2D2D (Anubis Red eyes & alerts)
	LapisBlue       string // #1E90FF (Royal Lapis Blue for MCP/tools)
	SuccessEmerald  string // #2ECC71 (Feather lighter than Heart)
	WarningAmber    string // #FFA500
	DangerCrimson   string // #FF4444 (Heart heavier than Feather)
	Reset           string
	Bold            string
}

func fgHex(r, g, b int) string {
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func bgHex(r, g, b int) string {
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

// LoadTheme returns the active Hermes-inspired Anubis theme, respecting ANUBIS_TUI_THEME.
func LoadTheme(name string) Theme {
	if name == "" {
		name = os.Getenv("ANUBIS_TUI_THEME")
	}
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "lapis-night":
		return Theme{
			Name:           "lapis-night",
			BgDark:         bgHex(15, 15, 26),
			PrimaryGold:    fgHex(30, 144, 255),
			SecondaryGold:  fgHex(65, 105, 225),
			BorderGold:     fgHex(30, 144, 255),
			TextPrimary:    fgHex(255, 248, 220),
			TextMuted:      fgHex(212, 175, 55),
			HighlightRed:   fgHex(255, 45, 45),
			LapisBlue:      fgHex(255, 215, 0),
			SuccessEmerald: fgHex(46, 204, 113),
			WarningAmber:   fgHex(255, 165, 0),
			DangerCrimson:  fgHex(255, 68, 68),
			Reset:          "\x1b[0m",
			Bold:           "\x1b[1m",
		}
	case "duat-blood":
		return Theme{
			Name:           "duat-blood",
			BgDark:         bgHex(18, 10, 16),
			PrimaryGold:    fgHex(255, 45, 45),
			SecondaryGold:  fgHex(255, 215, 0),
			BorderGold:     fgHex(201, 162, 39),
			TextPrimary:    fgHex(255, 248, 220),
			TextMuted:      fgHex(212, 175, 55),
			HighlightRed:   fgHex(255, 45, 45),
			LapisBlue:      fgHex(30, 144, 255),
			SuccessEmerald: fgHex(46, 204, 113),
			WarningAmber:   fgHex(255, 165, 0),
			DangerCrimson:  fgHex(255, 68, 68),
			Reset:          "\x1b[0m",
			Bold:           "\x1b[1m",
		}
	default: // "anubis-gold" (Hermes Classic Warm Gold + Lapis Blue)
		return Theme{
			Name:           "anubis-gold",
			BgDark:         bgHex(26, 26, 46),    // #1A1A2E
			PrimaryGold:    fgHex(255, 215, 0),   // #FFD700
			SecondaryGold:  fgHex(230, 194, 0),   // #E6C200
			BorderGold:     fgHex(201, 162, 39),  // #C9A227
			TextPrimary:    fgHex(255, 248, 220), // #FFF8DC
			TextMuted:      fgHex(212, 175, 55),  // #D4AF37
			HighlightRed:   fgHex(255, 45, 45),   // #FF2D2D
			LapisBlue:      fgHex(30, 144, 255),  // #1E90FF
			SuccessEmerald: fgHex(46, 204, 113),  // #2ECC71
			WarningAmber:   fgHex(255, 165, 0),   // #FFA500
			DangerCrimson:  fgHex(255, 68, 68),   // #FF4444
			Reset:          "\x1b[0m",
			Bold:           "\x1b[1m",
		}
	}
}

// RenderColorizedAnubis renders the Tomb-Sealed Braille Anubis ASCII art in Warm Gold (#FFD700)
// and highlights the eye row (row index 5) with Anubis Crimson Red (#FF2D2D).
func RenderColorizedAnubis(t Theme) []string {
	rawLines := strings.Split(VerifiedAnubisASCII(), "\n")
	out := make([]string, 0, len(rawLines))
	for idx, line := range rawLines {
		if idx == 5 {
			// Highlight the glowing eye region in Anubis Red (#FF2D2D) while keeping the snout/head in Gold
			out = append(out, t.PrimaryGold+strings.Replace(line, "⢻⣿⣿", t.HighlightRed+"⢻⣿⣿"+t.PrimaryGold, 1)+t.Reset)
		} else {
			out = append(out, t.PrimaryGold+line+t.Reset)
		}
	}
	return out
}
