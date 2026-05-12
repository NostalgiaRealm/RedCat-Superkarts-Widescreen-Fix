package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	gameFileName = "rcskspel.dat"
	backupDir   = "4x3_backup"
)

type clipPatch struct {
	offset int
	prefix []byte
	label  string
	isX    bool
}

var clipPatches = []clipPatch{
	{0x0ACEBE, []byte{0xC7, 0x41, 0x14}, "default horizontal clip x", true},
	{0x0ACEE8, []byte{0xC7, 0x41, 0x1C}, "default horizontal clip width", false},
	{0x0ACFBA, []byte{0xC7, 0x46, 0x14}, "runtime horizontal clip x", true},
	{0x0ACFC6, []byte{0xC7, 0x46, 0x1C}, "runtime horizontal clip width", false},
}

func main() {
	fmt.Println("Redcat Superkarts Hor+ Widescreen + FPS Unlock Patcher")
	fmt.Println("------------------------------------------------------")
	fmt.Println("Run this patcher from the folder containing rcskspel.dat.")
	fmt.Println("This version does NOT change or force the game's resolution.")
	fmt.Println("It patches the Hor+ 3D view according to an aspect ratio you enter,")
	fmt.Println("and also applies the 0 ms FPS unlock.")
	fmt.Println()

	aspect, label, err := askAspectRatio()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := patchInPlace(aspect); err != nil {
		fmt.Println()
		fmt.Println("Patch failed:", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Done. Created patched %s for %s Hor+ widescreen.\n", gameFileName, label)
	fmt.Println("Original file was moved to the 4x3_backup folder.")
}

func askAspectRatio() (float64, string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter desired aspect ratio, for example 16:9, 16:10, 21:9, 32:9, or 1.7777: ")

	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return 0, "", err
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return 0, "", fmt.Errorf("no aspect ratio entered")
	}

	aspect, label, err := parseAspectRatio(line)
	if err != nil {
		return 0, "", err
	}

	if aspect <= 0 {
		return 0, "", fmt.Errorf("aspect ratio must be positive")
	}

	if aspect < 1.2 || aspect > 4.0 {
		return 0, "", fmt.Errorf("aspect ratio %.4f is outside the accepted range 1.2 to 4.0", aspect)
	}

	if aspect < 1.70 || aspect > 1.82 {
		fmt.Printf("Warning: %s is not 16:9. The patch will still create a Hor+ view for that aspect ratio.\n", label)
	}

	return aspect, label, nil
}

func parseAspectRatio(input string) (float64, string, error) {
	s := strings.TrimSpace(strings.ToLower(input))
	s = strings.ReplaceAll(s, " ", "")

	reRatio := regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(?:\:|/)([0-9]+(?:\.[0-9]+)?)$`)
	if m := reRatio.FindStringSubmatch(s); m != nil {
		a, err := strconv.ParseFloat(m[1], 64)
		if err != nil {
			return 0, "", err
		}

		b, err := strconv.ParseFloat(m[2], 64)
		if err != nil {
			return 0, "", err
		}

		if b == 0 {
			return 0, "", fmt.Errorf("aspect ratio denominator cannot be zero")
		}

		return a / b, fmt.Sprintf("%g:%g", a, b), nil
	}

	aspect, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid aspect ratio format. Use 16:9, 21:9, 32:9, or a decimal such as 1.7777")
	}

	return aspect, fmt.Sprintf("%.6g:1", aspect), nil
}

func patchInPlace(aspect float64) error {
	inputPath := filepath.Join(".", gameFileName)

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", gameFileName, err)
	}

	if len(data) < 0x0ACFCD {
		return fmt.Errorf("%s is smaller than expected; wrong file?", gameFileName)
	}

	if len(data) < 2 || data[0] != 'M' || data[1] != 'Z' {
		return fmt.Errorf("%s does not look like a Windows PE executable", gameFileName)
	}

	patched := make([]byte, len(data))
	copy(patched, data)

	if err := applyPatch(patched, aspect); err != nil {
		return err
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("could not create %s folder: %w", backupDir, err)
	}

	backupPath := filepath.Join(backupDir, gameFileName)
	if _, err := os.Stat(backupPath); err == nil {
		stamp := time.Now().Format("20060102_150405")
		backupPath = filepath.Join(backupDir, "rcskspel_"+stamp+".dat")
	}

	if err := os.Rename(inputPath, backupPath); err != nil {
		return fmt.Errorf("could not move original %s to %s: %w", gameFileName, backupPath, err)
	}

	if err := os.WriteFile(inputPath, patched, 0644); err != nil {
		_ = os.Rename(backupPath, inputPath)
		return fmt.Errorf("could not write patched %s; original was restored if possible: %w", gameFileName, err)
	}

	return nil
}

func applyPatch(data []byte, aspect float64) error {
	baseAspect := 4.0 / 3.0
	multiplier := aspect / baseAspect

	clipX := float32(-multiplier)
	clipWidth := float32(multiplier * 2.0)

	for _, p := range clipPatches {
		value := clipWidth
		if p.isX {
			value = clipX
		}

		if err := patchFloatInstruction(data, p.offset, p.prefix, p.label, value); err != nil {
			return err
		}
	}

	if 0x043321 >= len(data) {
		return fmt.Errorf("FPS limiter offset 0x043321 is outside the file")
	}

	if data[0x043321] != 0x19 && data[0x043321] != 0x00 {
		return fmt.Errorf("FPS wait byte did not match expected 0x19 or 0x00 at 0x043321. This may be the wrong rcskspel.dat version")
	}
	data[0x043321] = 0x00

	return nil
}

func patchFloatInstruction(data []byte, offset int, prefix []byte, label string, value float32) error {
	if offset < 0 || offset+7 > len(data) {
		return fmt.Errorf("%s offset 0x%X is outside the file", label, offset)
	}

	currentPrefix := data[offset : offset+3]
	if !equalBytes(currentPrefix, prefix) {
		return fmt.Errorf("%s instruction prefix did not match expected pattern at 0x%X. This may be the wrong rcskspel.dat version", label, offset)
	}

	bits := math.Float32bits(value)
	binary.LittleEndian.PutUint32(data[offset+3:offset+7], bits)
	return nil
}

func equalBytes(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
