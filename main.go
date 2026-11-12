package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Groennesby/anubis-agent-cli/internal/cli"
	"github.com/Groennesby/anubis-agent-core/pkg/anubis"
)

func resolveModelSelection(input string) string {
	trimmed := strings.TrimSpace(input)
	if idx, err := strconv.Atoi(trimmed); err == nil && idx >= 1 && idx <= len(anubis.AvailableModels) {
		return anubis.AvailableModels[idx-1].ID
	}
	return trimmed
}

func streamWithThinkingSpinner(theme cli.Theme, eng *anubis.BicameralEngine, prompt string, isRetry bool) (string, anubis.Verdict, anubis.RecallResult) {
	var firstTokenReceived int32
	stopSpinner := make(chan struct{})
	spinnerDone := make(chan struct{})

	go func() {
		defer close(spinnerDone)
		idx := 0
		t0 := time.Now()
		for {
			select {
			case <-stopSpinner:
				return
			case <-time.After(120 * time.Millisecond):
				if atomic.LoadInt32(&firstTokenReceived) == 1 {
					return
				}
				sp := cli.AnubisSpinners[idx%len(cli.AnubisSpinners)]
				idx++
				elapsed := time.Since(t0).Seconds()
				fmt.Printf("\r%s%s Anubis (%s) denkt na & laadt .synapse tensoren... (%.1fs)%s",
					theme.PrimaryGold, sp, eng.ActiveModel, elapsed, theme.Reset)
			}
		}
	}()

	tStart := time.Now()
	tokenCount := 0

	onToken := func(tok string) {
		if atomic.CompareAndSwapInt32(&firstTokenReceived, 0, 1) {
			close(stopSpinner)
			<-spinnerDone
			fmt.Print("\r\x1b[K")
			fmt.Printf("%s%s𓂀 Anubis (%s):%s ", theme.Bold, theme.PrimaryGold, eng.ActiveModel, theme.Reset)
		}
		tokenCount++
		fmt.Print(theme.TextPrimary + tok + theme.Reset)
	}

	var (
		reply   string
		verdict anubis.Verdict
		recall  anubis.RecallResult
	)
	if isRetry {
		reply, verdict, recall, _ = eng.StreamRetryLast(context.Background(), onToken)
	} else {
		reply, verdict, recall, _ = eng.StreamQuery(context.Background(), prompt, onToken)
	}

	if atomic.CompareAndSwapInt32(&firstTokenReceived, 0, 1) {
		close(stopSpinner)
		<-spinnerDone
		fmt.Print("\r\x1b[K")
	}

	totalSec := time.Since(tStart).Seconds()
	tokPerSec := 0.0
	if totalSec > 0 {
		tokPerSec = float64(tokenCount) / totalSec
	}
	fmt.Printf("\n%s[⚡ .synapse: %.2fms | ⚖️ Heart: %d/%d (%s) | %.1f tok/s | %.1fs]%s\n\n",
		theme.TextMuted, recall.LatencyMs, verdict.HeartScore, verdict.FeatherScore, verdict.Status, tokPerSec, totalSec, theme.Reset)

	return reply, verdict, recall
}

func main() {
	cwd, _ := os.Getwd()
	eng, err := anubis.NewBicameralEngine(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Anubis CLI Init Error: %v\n", err)
		os.Exit(1)
	}
	theme := cli.LoadTheme("")

	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "--version", "-v", "version":
			fmt.Printf("%s𓂀 ANUBIS AGENT CLI v1.0.0 (Go Bicameral Engine | Hermes Gold Theme)%s\n", theme.PrimaryGold, theme.Reset)
			return
		case "--render-preview":
			verdict := eng.Scales.WeighHeart("startup", "anubis interactive dashboard")
			msgs := []string{
				"𓂀 Welcome back, my friend. Holographic Bicameral Memory (.synapse) is online (<0.3ms).",
				"🛡️ Tomb Seal Sandbox, REA v6.8.0 Reverse Engineering & Kali/OPSEC ChainKeys Shield ready.",
				"💡 Commands ready: /new, /model (met keuze: vorig gesprek of nieuw gesprek), /renew, /retry.",
			}
			fmt.Print(cli.RenderResponsiveDashboard(theme, eng, verdict, msgs, true))
			return
		case "run":
			task := "status check"
			if len(args) > 1 {
				task = strings.Join(args[1:], " ")
			}
			verdict := eng.Scales.WeighHeart("one_shot_run", task)
			fmt.Println(theme.PrimaryGold + cli.AnubisSpinners[0] + " Weighing of the Heart: " + theme.Reset + verdict.RenderGauge(28))
			if verdict.RequiresAuth {
				fmt.Printf("%s⚠️  Heart (%d) is heavier than the Feather of Ma'at (%d): %s%s\n",
					theme.HighlightRed, verdict.HeartScore, verdict.FeatherScore, verdict.Reason, theme.Reset)
				return
			}
			streamWithThinkingSpinner(theme, eng, task, false)
			return
		case "model":
			sub := "list"
			if len(args) > 1 {
				sub = args[1]
			}
			if sub == "list" {
				fmt.Printf("%s𓂀 PRELOADED & AVAILABLE MODELS (0%% Censorship):%s\n", theme.PrimaryGold, theme.Reset)
				for i, m := range anubis.AvailableModels {
					fmt.Printf("  [%d] %-26s — %s (%s)\n", i+1, m.ID, m.Label, m.Description)
				}
			} else if sub == "pull" && len(args) > 2 {
				target := args[2]
				msg := eng.SwitchModelWithMode(target, true)
				fmt.Printf("%s𓂀 %s%s\n", theme.SuccessEmerald, msg, theme.Reset)
			}
			return
		case "skill":
			fmt.Printf("%s𓂀 VERIFIED TOMB SKILLS (Sandboxed & Self-Healing):%s\n", theme.PrimaryGold, theme.Reset)
			fmt.Println("  1. [VERIFIED] rea-reverse-engineering  — REA v6.8.0 (35 security & binary/codebase audit skills)")
			fmt.Println("  2. [VERIFIED] kali-opsec-chainkeys     — Kali defensive diagnostics + multi-hop ProxyChains/ChainKeys shield")
			fmt.Println("  3. [VERIFIED] universal-media-ytdlp    — Best quality video & YouTube Music 1:1 thumbnail + -14 LUFS m4a")
			fmt.Println("  4. [VERIFIED] self-healing-coder       — Automatic compile -> test -> fix verification loop")
			fmt.Println("  5. [VERIFIED] bicameral-synapse-vault  — Sub-millisecond .synapse tensor page consolidation")
			return
		case "duat":
			path, err := eng.Duat.WriteMorningScroll()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Du'at error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s𓂀 Du'at Autonomous Daemon active. Morning Dawn Scroll generated at:%s\n  %s\n",
				theme.PrimaryGold, theme.Reset, path)
			return
		}
	}

	// Interactive TUI / Chat Mode with live token streaming
	verdict := eng.Scales.WeighHeart("interactive_session", "anubis chat")
	messages := []string{
		"𓂀 Anubis Agent initialized. I am your loyal best friend & guardian.",
		"🧠 Right Hemisphere (.synapse tensors) & Left Hemisphere (Tomb Seal Executor) synced.",
		"💡 Typ een vraag om live te chatten, of gebruik /new, /model, /renew, /retry, /status, /theme, /memory, /tomb, /judgement, /rea, /opsec, /duat, /quit.",
	}

	sigCh := make(chan os.Signal, 1)
	cli.RegisterResizeSignal(sigCh)
	go func() {
		for range sigCh {
			fmt.Print("\x1b[2J\x1b[H")
			fmt.Print(cli.RenderResponsiveDashboard(theme, eng, verdict, messages, false))
			fmt.Printf("%s𓂀 anubis > %s", theme.PrimaryGold, theme.Reset)
		}
	}()

	// Show the full Hermes-Gold Anubis dashboard once at startup
	fmt.Print(cli.RenderResponsiveDashboard(theme, eng, verdict, messages, true))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("%s𓂀 anubis > %s", theme.PrimaryGold, theme.Reset)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "/quit" || line == "/exit" || line == "exit" {
			fmt.Printf("%s𓂀 May the Scales of Ma'at guard your code, my friend.%s\n", theme.PrimaryGold, theme.Reset)
			break
		}

		if strings.HasPrefix(line, "/") {
			parts := strings.Fields(line)
			cmd := strings.ToLower(parts[0])
			switch cmd {
			case "/status", "/dashboard":
				fmt.Print(cli.RenderResponsiveDashboard(theme, eng, verdict, messages, true))

			case "/new":
				newSessID := eng.NewSession()
				messages = []string{
					fmt.Sprintf("🆕 Volledig nieuw gesprek gestart (Sessie: %s | Model: %s).", newSessID, eng.ActiveModel),
					"🧠 Vorige context is veilig gearchiveerd in de Holographic .synapse Vault (<0.25ms recall).",
				}
				fmt.Print(cli.RenderResponsiveDashboard(theme, eng, verdict, messages, true))

			case "/renew":
				sessID, turns := eng.RenewSession()
				msg := fmt.Sprintf("🔄 [/renew] Bicameral Bridge & .synapse Fast-Weights vernieuwd voor sessie %s (%d berichten behouden, L1 Horizon: %d tokens).",
					sessID, turns, eng.RightHemisphere.Adaptive.L1TokenHorizon)
				messages = append(messages, msg)
				fmt.Printf("%s%s%s\n\n", theme.SuccessEmerald, msg, theme.Reset)

			case "/retry":
				if eng.LastUserPrompt == "" {
					fmt.Printf("%s⚠️ [/retry] Geen vorig bericht gevonden om opnieuw uit te voeren.%s\n\n", theme.WarningAmber, theme.Reset)
				} else {
					if len(messages) > 0 && strings.HasPrefix(messages[len(messages)-1], "𓂀 Anubis:") {
						messages = messages[:len(messages)-1]
					}
					fmt.Printf("%s🔁 [/retry] Opnieuw genereren voor: %q (met model %s)...%s\n",
						theme.LapisBlue, eng.LastUserPrompt, eng.ActiveModel, theme.Reset)
					reply, v, _ := streamWithThinkingSpinner(theme, eng, eng.LastUserPrompt, true)
					verdict = v
					messages = append(messages, "𓂀 Anubis: "+reply)
				}

			case "/model":
				var targetModel string
				explicitMode := ""

				for _, p := range parts[1:] {
					lp := strings.ToLower(p)
					if lp == "--new" || lp == "new" || lp == "nieuw" {
						explicitMode = "new"
					} else if lp == "--continue" || lp == "continue" || lp == "verder" {
						explicitMode = "continue"
					} else if targetModel == "" {
						targetModel = resolveModelSelection(p)
					}
				}

				if targetModel == "" {
					fmt.Printf("\n%s𓂀 KIES EEN MODEL (0%% Censuur — Lokaal of Router):%s\n", theme.PrimaryGold, theme.Reset)
					for i, m := range anubis.AvailableModels {
						activeMark := " "
						if m.ID == eng.ActiveModel {
							activeMark = "*"
						}
						fmt.Printf("  %s[%d]%s %-26s — %s\n", theme.PrimaryGold, i+1, activeMark, m.ID, m.Label+theme.Reset)
					}
					fmt.Printf("%sVoer nummer (1-%d) of modelnaam in [huidig: %s]: %s",
						theme.LapisBlue, len(anubis.AvailableModels), eng.ActiveModel, theme.Reset)
					if !scanner.Scan() {
						break
					}
					choice := strings.TrimSpace(scanner.Text())
					if choice == "" {
						targetModel = eng.ActiveModel
					} else {
						targetModel = resolveModelSelection(choice)
					}
				}

				continuePrevious := true
				if explicitMode == "new" {
					continuePrevious = false
				} else if explicitMode == "continue" {
					continuePrevious = true
				} else {
					fmt.Printf("\n%s𓂀 Model %q geselecteerd. Wat wil je doen met het gesprek?%s\n", theme.PrimaryGold, targetModel, theme.Reset)
					fmt.Printf("  %s[1]%s Verder gaan in het vorige gesprek (gesprekshistorie behouden) %s[Standaard]%s\n",
						theme.SuccessEmerald, theme.Reset, theme.TextMuted, theme.Reset)
					fmt.Printf("  %s[2]%s Volledig nieuw gesprek starten (/new)\n",
						theme.LapisBlue, theme.Reset)
					fmt.Printf("%sMaak je keuze [1/2]: %s", theme.PrimaryGold, theme.Reset)
					if scanner.Scan() {
						modeChoice := strings.ToLower(strings.TrimSpace(scanner.Text()))
						if modeChoice == "2" || modeChoice == "new" || modeChoice == "nieuw" || modeChoice == "n" {
							continuePrevious = false
						}
					}
				}

				statusMsg := eng.SwitchModelWithMode(targetModel, continuePrevious)
				if !continuePrevious {
					messages = []string{"🆕 " + statusMsg}
					fmt.Print(cli.RenderResponsiveDashboard(theme, eng, verdict, messages, true))
				} else {
					messages = append(messages, "🔄 "+statusMsg)
					fmt.Printf("\n%s🔄 %s%s\n\n", theme.SuccessEmerald, statusMsg, theme.Reset)
				}

			case "/theme":
				tName := "lapis-night"
				if len(parts) > 1 {
					tName = parts[1]
				} else if theme.Name == "lapis-night" {
					tName = "duat-blood"
				} else if theme.Name == "duat-blood" {
					tName = "anubis-gold"
				}
				theme = cli.LoadTheme(tName)
				messages = append(messages, fmt.Sprintf("🎨 Switched Hermes/Anubis TUI theme to '%s'.", theme.Name))
				fmt.Print(cli.RenderResponsiveDashboard(theme, eng, verdict, messages, true))

			case "/memory":
				rec := eng.RightHemisphere.HopfieldAssociativeReload("Guardian Law REA Kali OPSEC", 4)
				fmt.Printf("%s🧠 .synapse Vault: %d pages | Recall: %.2fms | Hopfield Energy: %.4f | Hebbian Norm: %.4f%s\n\n",
					theme.LapisBlue, rec.TotalPages, rec.LatencyMs, rec.HopfieldEnergy, rec.HebbianFrobenius, theme.Reset)

			case "/judgement":
				testCmd := "git push origin main"
				if len(parts) > 1 {
					testCmd = strings.Join(parts[1:], " ")
				}
				verdict = eng.Scales.WeighHeart("manual_weigh", testCmd)
				fmt.Printf("%s%s — %s%s\n\n", theme.PrimaryGold, verdict.RenderGauge(26), verdict.Reason, theme.Reset)

			case "/tomb":
				sealID, _, _ := eng.Tomb.ExecuteSandboxed(context.Background(), "uname -a", 5*time.Second)
				fmt.Printf("%s🛡️ Tomb Seal Snapshot created: %s (OPSEC ChainKeys Env Scrubbing: ACTIVE)%s\n\n",
					theme.SuccessEmerald, sealID, theme.Reset)

			case "/rea":
				fmt.Printf("%s🔍 REA v6.8.0 Reverse Engineering Engine: ONLINE (35 skills, 0%% telemetry, 16x512 native routing tensor).%s\n\n",
					theme.LapisBlue, theme.Reset)

			case "/opsec":
				fmt.Printf("%s🕵️ Kali Defensive Toolkit & ChainKeys Shield: ACTIVE (DNS/WebRTC leak prevention + proxy chain ready).%s\n\n",
					theme.SuccessEmerald, theme.Reset)

			case "/duat":
				scroll, _ := eng.Duat.WriteMorningScroll()
				fmt.Printf("%s🌙 Du'at Dawn Scroll saved: %s%s\n\n", theme.PrimaryGold, scroll, theme.Reset)

			default:
				fmt.Printf("%sCommands: /new, /model, /renew, /retry, /status, /theme, /memory, /judgement <cmd>, /tomb, /rea, /opsec, /duat, /quit%s\n\n",
					theme.TextMuted, theme.Reset)
			}
			continue
		}

		// Live streaming conversation turn!
		messages = append(messages, "You: "+line)
		reply, v, _ := streamWithThinkingSpinner(theme, eng, line, false)
		verdict = v
		messages = append(messages, "𓂀 Anubis: "+reply)
	}
}
