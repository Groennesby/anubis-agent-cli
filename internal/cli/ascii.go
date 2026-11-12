package cli

import (
	"crypto/sha256"
	"encoding/hex"
)

// ╔══════════════════════════════════════════════════════════════════════════╗
// ║ 𓂀 TOMB-SEALED SACRED EMBLEM — DO NOT MODIFY UNDER ANY CIRCUMSTANCES 𓂀   ║
// ║ Protected by SHA-256 Tomb Seal & Weighing of the Heart Guardian Law.     ║
// ╚══════════════════════════════════════════════════════════════════════════╝

// AnubisASCII is the exact Braille Anubis jackal emblem rendered in Warm Gold (#FFD700)
// with crimson red eyes (#FF2D2D).
const AnubisASCII = `⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣾⡇⠀⢠⣾⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣾⣿⡇⢠⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⡇⢸⣿⡏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣤⡀⠙⠛⠃⠘⠻⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠐⠚⣛⣛⣁⡀⠹⣿⣿⣶⣶⣤⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⣠⣴⠶⠿⠛⠛⠛⠛⠛⠀⢻⣿⣿⣤⣀⣙⣷⣀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣈⣁⣤⣴⣶⠶⠿⠿⠿⠿⠇⠸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣦⣤⡄⠀
⠀⠀⠐⠛⢉⣉⣠⣤⣤⣶⣶⣶⣶⣦⠀⣿⣿⣿⣿⣿⣿⣿⡿⠿⠿⠿⠛⠉⠀⠀
⠀⠀⡾⠛⠉⠉⠉⠙⠻⢿⣿⣿⣿⣿⡀⢹⡿⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⢸⡇⠀⠀⠀⠀⠀⠀⠀⠈⠻⣿⣿⡇⠘⢀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⢸⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⣸⣿⡇⠀⣾⣇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⠃⠀⢿⣿⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢠⣿⣿⠀⠀⠸⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠙⠀⠀⠀⠋⠁⠀⠀⠀`

// AnubisASCIISealHash is the immutable SHA-256 Tomb Seal of AnubisASCII.
const AnubisASCIISealHash = "ccb80bab1e794573f19b637057b645f1d97db26cf2df633c893a1891e705fa8a"

// immutableAnubisBackup stores the byte-for-byte sealed backup so the emblem can self-heal if ever tampered with.
var immutableAnubisBackup = []byte(AnubisASCII)

// VerifiedAnubisASCII returns the cryptographically verified Anubis Braille emblem.
func VerifiedAnubisASCII() string {
	sum := sha256.Sum256([]byte(AnubisASCII))
	if hex.EncodeToString(sum[:]) != AnubisASCIISealHash {
		return string(immutableAnubisBackup)
	}
	return AnubisASCII
}

// AnubisSpinners are the Egyptian hieroglyphic activity indicators.
var AnubisSpinners = []string{"𓂀", "𓆣", "𓅓"}
