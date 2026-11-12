package cli

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Groennesby/anubis-agent-core/pkg/anubis"
)

// DetectTerminalSize queries ioctl TIOCGWINSZ, environment overrides (COLUMNS/LINES), or stty fallback.
func DetectTerminalSize() (width int, height int) {
	if c := os.Getenv("COLUMNS"); c != "" {
		if w, err := strconv.Atoi(c); err == nil && w > 20 {
			width = w
		}
	}
	if l := os.Getenv("LINES"); l != "" {
		if h, err := strconv.Atoi(l); err == nil && h > 10 {
			height = h
		}
	}
	if width > 0 && height > 0 {
		return width, height
	}

	if w, h := ioctlTermSize(); w > 0 && h > 0 {
		if width == 0 {
			width = w
		}
		if height == 0 {
			height = h
		}
		return width, height
	}

	if out, err := exec.Command("stty", "size").Output(); err == nil {
		parts := strings.Fields(string(out))
		if len(parts) == 2 {
			h, _ := strconv.Atoi(parts[0])
			w, _ := strconv.Atoi(parts[1])
			if width == 0 && w > 0 {
				width = w
			}
			if height == 0 && h > 0 {
				height = h
			}
		}
	}
	if width <= 0 {
		width = 120
	}
	if height <= 0 {
		height = 36
	}
	return width, height
}

// LayoutMode classifies the responsive breakpoint based on terminal width & height.
type LayoutMode string

const (
	LayoutDesktopFull  LayoutMode = "DESKTOP_FULL_140+"
	LayoutLaptopStd    LayoutMode = "LAPTOP_STANDARD_100_139"
	LayoutCompactSplit LayoutMode = "COMPACT_SPLIT_80_99"
	LayoutUltraMinimal LayoutMode = "ULTRA_MINIMAL_<80"
)

func ClassifyLayout(width, height int) (LayoutMode, bool) {
	compactHeight := height < 30
	switch {
	case width >= 140:
		return LayoutDesktopFull, compactHeight
	case width >= 100:
		return LayoutLaptopStd, compactHeight
	case width >= 80:
		return LayoutCompactSplit, compactHeight
	default:
		return LayoutUltraMinimal, compactHeight
	}
}

// RenderResponsiveDashboard renders the Hermes-inspired Anubis TUI frame adapted to any terminal resolution.
func RenderResponsiveDashboard(t Theme, eng *anubis.BicameralEngine, lastVerdict anubis.Verdict, messages []string, showLaunchArt bool) string {
	width, height := DetectTerminalSize()
	mode, compactHeight := ClassifyLayout(width, height)
	shortHeight := height < 24

	var sb strings.Builder
	recall := eng.RightHemisphere.HopfieldAssociativeReload("Guardian Law REA Kali OPSEC", 4)

	// 1. ULTRA-MINIMAL (<80 cols) or SHORT (<24 rows after first message)
	if mode == LayoutUltraMinimal || (shortHeight && len(messages) > 0) {
		div := t.BorderGold + strings.Repeat("─", maxInt(20, width)) + t.Reset
		sb.WriteString(fmt.Sprintf("%s%s𓂀 ANUBIS%s %s| %s | .synapse %.2fms | ⚖️ %d/%d%s\n",
			t.Bold, t.PrimaryGold, t.Reset, t.TextMuted, eng.ActiveModel, recall.LatencyMs, lastVerdict.HeartScore, lastVerdict.FeatherScore, t.Reset))
		sb.WriteString(div + "\n")
		for _, m := range tailMessages(messages, maxInt(3, height-6)) {
			for _, wrapped := range wordWrap(m, maxInt(30, width-2)) {
				sb.WriteString(t.TextPrimary + wrapped + t.Reset + "\n")
			}
		}
		sb.WriteString(div + "\n")
		sb.WriteString(fmt.Sprintf("%s𓂀 /new /model /renew /retry /theme /memory /tomb /judgement /rea /opsec%s\n", t.LapisBlue, t.Reset))
		return sb.String()
	}

	innerW := width - 4
	if innerW < 40 {
		innerW = 40
	}

	topBorder := t.BorderGold + "╭─ " + t.Bold + t.PrimaryGold + "𓂀 ANUBIS AGENT v1.0" + t.Reset + t.BorderGold + " " +
		strings.Repeat("─", maxInt(2, innerW-21)) + "╮" + t.Reset
	midBorder := t.BorderGold + "├" + strings.Repeat("─", innerW+2) + "┤" + t.Reset
	botBorder := t.BorderGold + "╰" + strings.Repeat("─", innerW+2) + "╯" + t.Reset

	sb.WriteString(topBorder + "\n")

	// 2. HEADER SECTION BY BREAKPOINT
	artLines := RenderColorizedAnubis(t)
	infoLines := []string{
		t.Bold + t.PrimaryGold + "ANUBIS AGENT" + t.Reset + " " + t.TextMuted + "— Holographic Bicameral Sovereign AI" + t.Reset,
		t.TextMuted + "Model      : " + t.TextPrimary + eng.ActiveModel + t.Reset,
		t.TextMuted + "GPU Engine : " + t.LapisBlue + eng.BackendType + t.Reset,
		fmt.Sprintf("%sMemory     : %s%d .synapse pages (<%.2fms) | L1: %d tok%s",
			t.TextMuted, t.SuccessEmerald, recall.TotalPages, recall.LatencyMs, eng.RightHemisphere.Adaptive.L1TokenHorizon, t.Reset),
		fmt.Sprintf("%sJudgement  : %s⚖️  Heart %d/100 vs Feather %d (%s)%s",
			t.TextMuted, verdictColor(t, lastVerdict), lastVerdict.HeartScore, lastVerdict.FeatherScore, lastVerdict.Status, t.Reset),
		t.TextMuted + "Tomb Seal  : " + t.SuccessEmerald + "ACTIVE (Zero-Telemetry + ChainKeys OPSEC + REA v6.8.0)" + t.Reset,
		t.TextMuted + "Du'at Mode : " + t.PrimaryGold + "READY (3 Overnight Self-Healing Scrolls)" + t.Reset,
		t.TextMuted + "Censorship : " + t.HighlightRed + "0% (100% Local Sovereign & Loyal Best Friend)" + t.Reset,
	}

	switch mode {
	case LayoutDesktopFull:
		// Width >= 140: Full Anubis ASCII art on left of header + telemetry on right
		rows := len(artLines)
		if !compactHeight {
			for i := 0; i < rows; i++ {
				left := artLines[i]
				right := ""
				if i >= 2 && (i-2) < len(infoLines) {
					right = infoLines[i-2]
				}
				lineContent := fmt.Sprintf("  %s    %s", left, right)
				sb.WriteString(padBoxLine(t, lineContent, innerW) + "\n")
			}
		} else {
			for _, inf := range infoLines[:5] {
				sb.WriteString(padBoxLine(t, " "+inf, innerW) + "\n")
			}
		}

	case LayoutLaptopStd:
		// Width 100-139: Centered Anubis ASCII art at top (unless height < 30), compact status bar below
		if !compactHeight && showLaunchArt {
			for _, al := range artLines {
				sb.WriteString(padBoxLine(t, "      "+al, innerW) + "\n")
			}
			sb.WriteString(midBorder + "\n")
		}
		for _, inf := range infoLines {
			sb.WriteString(padBoxLine(t, " "+inf, innerW) + "\n")
		}

	case LayoutCompactSplit:
		// Width 80-99: Mini Anubis header; full ASCII art only on launch if height permits
		if showLaunchArt && !compactHeight {
			for _, al := range artLines {
				sb.WriteString(padBoxLine(t, "  "+al, innerW) + "\n")
			}
			sb.WriteString(midBorder + "\n")
		}
		sb.WriteString(padBoxLine(t, " "+infoLines[0], innerW) + "\n")
		sb.WriteString(padBoxLine(t, " "+infoLines[1]+" | "+infoLines[3], innerW) + "\n")
		sb.WriteString(padBoxLine(t, " "+infoLines[4], innerW) + "\n")
	}

	sb.WriteString(midBorder + "\n")

	// 3. BODY SECTION (2-Column on >=140, Single-Column on <140)
	maxChatLines := maxInt(5, height-len(artLines)-12)
	if compactHeight {
		maxChatLines = maxInt(4, height-12)
	}
	recent := tailMessages(messages, maxChatLines)

	if mode == LayoutDesktopFull {
		leftW := (innerW * 68) / 100
		rightW := innerW - leftW - 3

		rightPanel := []string{
			t.Bold + t.SecondaryGold + "⚖️  WEIGHING OF THE HEART" + t.Reset,
			verdictColor(t, lastVerdict) + fmt.Sprintf("Heart Risk : %d / 100", lastVerdict.HeartScore) + t.Reset,
			t.SuccessEmerald + fmt.Sprintf("Ma'at Feather: %d / 100", lastVerdict.FeatherScore) + t.Reset,
			t.TextMuted + "Verdict: " + lastVerdict.Status + t.Reset,
			"",
			t.Bold + t.LapisBlue + "🧠 HOLOGRAPHIC .SYNAPSE VAULT" + t.Reset,
			fmt.Sprintf("%sHopfield Energy: %.4f%s", t.TextPrimary, recall.HopfieldEnergy, t.Reset),
			fmt.Sprintf("%sHebbian Norm   : %.4f%s", t.TextPrimary, recall.HebbianFrobenius, t.Reset),
		}
		for _, p := range recall.ReloadedPages {
			rightPanel = append(rightPanel, fmt.Sprintf("%s• #%d %s (%.2f)%s", t.TextMuted, p.CrystalID, truncateVis(p.Title, rightW-12), p.CosineResonance, t.Reset))
		}

		var wrappedLeft []string
		for _, m := range recent {
			wrappedLeft = append(wrappedLeft, wordWrap(m, leftW-2)...)
		}
		totalRows := maxInt(len(wrappedLeft), len(rightPanel))
		for r := 0; r < totalRows; r++ {
			lTxt := ""
			if r < len(wrappedLeft) {
				lTxt = wrappedLeft[r]
			}
			rTxt := ""
			if r < len(rightPanel) {
				rTxt = rightPanel[r]
			}
			rowStr := padPlain(lTxt, leftW) + t.BorderGold + " │ " + t.Reset + padPlain(rTxt, rightW)
			sb.WriteString(padBoxLine(t, rowStr, innerW) + "\n")
		}
	} else {
		for _, m := range recent {
			for _, wline := range wordWrap(m, innerW-2) {
				sb.WriteString(padBoxLine(t, " "+t.TextPrimary+wline+t.Reset, innerW) + "\n")
			}
		}
	}

	sb.WriteString(midBorder + "\n")
	statusLine := fmt.Sprintf(" %s𓂀 [Model: %s]%s  %s[Memory: %d .synapse <%.2fms]%s  %s[⚖️ Heart: %d vs Feather: %d]%s  %s[Tomb: SEALED]%s",
		t.PrimaryGold, eng.ActiveModel, t.Reset,
		t.LapisBlue, recall.TotalPages, recall.LatencyMs, t.Reset,
		verdictColor(t, lastVerdict), lastVerdict.HeartScore, lastVerdict.FeatherScore, t.Reset,
		t.SuccessEmerald, t.Reset)
	sb.WriteString(padBoxLine(t, statusLine, innerW) + "\n")
	cmdHint := fmt.Sprintf(" %sCommands: /new  /model  /renew  /retry  /theme  /memory  /tomb  /judgement  /rea  /opsec  /duat  /quit%s", t.TextMuted, t.Reset)
	sb.WriteString(padBoxLine(t, cmdHint, innerW) + "\n")
	sb.WriteString(botBorder + "\n")

	return sb.String()
}

func verdictColor(t Theme, v anubis.Verdict) string {
	if v.HeartScore > v.FeatherScore {
		return t.DangerCrimson
	}
	if v.HeartScore >= 25 {
		return t.WarningAmber
	}
	return t.SuccessEmerald
}

// visibleLen calculates approximate printable character width ignoring ANSI escape sequences.
func visibleLen(s string) int {
	inEsc := false
	count := 0
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		// Egyptian hieroglyphs take ~2 columns in most terminals
		if r >= 0x13000 && r <= 0x1342F {
			count += 2
		} else {
			count++
		}
	}
	return count
}

func padPlain(s string, targetW int) string {
	vLen := visibleLen(s)
	if vLen >= targetW {
		return s
	}
	return s + strings.Repeat(" ", targetW-vLen)
}

func padBoxLine(t Theme, content string, innerW int) string {
	vLen := visibleLen(content)
	pad := innerW - vLen
	if pad < 0 {
		pad = 0
	}
	return t.BorderGold + "│ " + t.Reset + content + strings.Repeat(" ", pad) + t.BorderGold + " │" + t.Reset
}

func wordWrap(s string, maxWidth int) []string {
	if maxWidth < 20 {
		maxWidth = 20
	}
	rawLines := strings.Split(s, "\n")
	var out []string
	for _, line := range rawLines {
		if visibleLen(line) <= maxWidth {
			out = append(out, line)
			continue
		}
		words := strings.Fields(line)
		cur := ""
		for _, w := range words {
			if cur == "" {
				cur = w
			} else if visibleLen(cur)+1+visibleLen(w) <= maxWidth {
				cur += " " + w
			} else {
				out = append(out, cur)
				cur = w
			}
		}
		if cur != "" {
			out = append(out, cur)
		}
	}
	return out
}

func tailMessages(msgs []string, n int) []string {
	if len(msgs) <= n {
		return msgs
	}
	return msgs[len(msgs)-n:]
}

func truncateVis(s string, max int) string {
	if max <= 4 || len(s) <= max {
		return s
	}
	return s[:max-2] + ".."
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
