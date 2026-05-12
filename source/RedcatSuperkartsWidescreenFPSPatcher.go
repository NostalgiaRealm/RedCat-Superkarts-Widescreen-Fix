package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
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

// These offsets come from the working U Hor+ build.
// The game stores the selectable display mode table as 32-bit little-endian
// width/height values. Only the mode entries used by this DirectDraw/D3D module
// need to be changed.
var widthOffsets = []int{
	0x0FE930,
	0x0FE984,
	0x0FE9FC,
	0x0FEA9C,
}

var heightOffsets = []int{
	0x0FE934,
	0x0FE988,
	0x0FEA00,
	0x0FEAA0,
}

func main() {
	fmt.Println("Redcat Superkarts Widescreen + FPS Patcher")
	fmt.Println("-------------------------------------------")
	fmt.Println("Run this patcher from the folder containing rcskspel.dat.")
	fmt.Println()

	width, height, err := askResolution()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	if err := patchInPlace(width, height); err != nil {
		fmt.Println()
		fmt.Println("Patch failed:", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("Done. Created patched %s for %dx%d.\n", gameFileName, width, height)
	fmt.Println("Original file was moved to the 4x3_backup folder.")
}

func askResolution() (int, int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter desired 16:9 resolution, for example 2560x1440: ")

	line, err := reader.ReadString('\n')
	if err != nil && len(line) == 0 {
		return 0, 0, err
	}

	line = strings.TrimSpace(line)

	re := regexp.MustCompile(`(?i)^\s*([0-9]{3,5})\s*x\s*([0-9]{3,5})\s*$`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return 0, 0, fmt.Errorf("invalid resolution format. Use something like 2560x1440")
	}

	width, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, 0, err
	}

	height, err := strconv.Atoi(matches[2])
	if err != nil {
		return 0, 0, err
	}

	if width <= 0 || height <= 0 {
		return 0, 0, fmt.Errorf("resolution values must be positive")
	}

	// The display table stores dimensions as 32-bit values, but absurdly large
	// numbers are still rejected to avoid accidental input mistakes.
	if width > 32768 || height > 32768 {
		return 0, 0, fmt.Errorf("resolution is unusually large; refusing to patch")
	}

	aspect := float64(width) / float64(height)
	if aspect < 1.70 || aspect > 1.82 {
		fmt.Printf("Warning: %dx%d is not very close to 16:9. Continuing anyway.\n", width, height)
	}

	return width, height, nil
}

func patchInPlace(width int, height int) error {
	inputPath := filepath.Join(".", gameFileName)

	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("could not read %s: %w", gameFileName, err)
	}

	if len(data) < 0x0FEAA4 {
		return fmt.Errorf("%s is smaller than expected; wrong file?", gameFileName)
	}

	if len(data) < 2 || data[0] != 'M' || data[1] != 'Z' {
		return fmt.Errorf("%s does not look like a Windows PE executable", gameFileName)
	}

	patched := make([]byte, len(data))
	copy(patched, data)

	if err := applyPatch(patched, width, height); err != nil {
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
		// Try to restore the original if writing fails.
		_ = os.Rename(backupPath, inputPath)
		return fmt.Errorf("could not write patched %s; original was restored if possible: %w", gameFileName, err)
	}

	return nil
}

func applyPatch(data []byte, width int, height int) error {
	putU32 := func(offset int, value uint32) error {
		if offset < 0 || offset+4 > len(data) {
			return fmt.Errorf("offset 0x%X is outside the file", offset)
		}
		binary.LittleEndian.PutUint32(data[offset:offset+4], value)
		return nil
	}

	patchBytes := func(offset int, original []byte, replacement []byte, label string) error {
		if offset < 0 || offset+len(replacement) > len(data) {
			return fmt.Errorf("%s offset 0x%X is outside the file", label, offset)
		}

		current := data[offset : offset+len(replacement)]

		if bytes.Equal(current, replacement) {
			return nil // already patched
		}

		if !bytes.Equal(current, original) {
			return fmt.Errorf("%s bytes did not match expected original pattern at 0x%X. This may be the wrong rcskspel.dat version", label, offset)
		}

		copy(current, replacement)
		return nil
	}

	for _, offset := range widthOffsets {
		if err := putU32(offset, uint32(width)); err != nil {
			return err
		}
	}

	for _, offset := range heightOffsets {
		if err := putU32(offset, uint32(height)); err != nil {
			return err
		}
	}

	// U-style Hor+ Direct3D horizontal clip range.
	// These are the exact horizontal view corrections from the working U build:
	// -1.0      -> -1.3333334  (BF800000 -> BFAAAAAB)
	//  2.0      ->  2.6666667  (40000000 -> 402AAAAB)
	// This expands the horizontal view without changing the vertical framing.
	if err := patchBytes(
		0x0ACEBD,
		[]byte{0xC7, 0x41, 0x14, 0x00, 0x00, 0x80, 0xBF},
		[]byte{0xC7, 0x41, 0x14, 0xAB, 0xAA, 0xAA, 0xBF},
		"default horizontal clip x",
	); err != nil {
		return err
	}

	if err := patchBytes(
		0x0ACEE7,
		[]byte{0xC7, 0x41, 0x1C, 0x00, 0x00, 0x00, 0x40},
		[]byte{0xC7, 0x41, 0x1C, 0xAB, 0xAA, 0x2A, 0x40},
		"default horizontal clip width",
	); err != nil {
		return err
	}

	if err := patchBytes(
		0x0ACFB9,
		[]byte{0xC7, 0x46, 0x14, 0x00, 0x00, 0x80, 0xBF},
		[]byte{0xC7, 0x46, 0x14, 0xAB, 0xAA, 0xAA, 0xBF},
		"runtime horizontal clip x",
	); err != nil {
		return err
	}

	if err := patchBytes(
		0x0ACFC5,
		[]byte{0xC7, 0x46, 0x1C, 0x00, 0x00, 0x00, 0x40},
		[]byte{0xC7, 0x46, 0x1C, 0xAB, 0xAA, 0x2A, 0x40},
		"runtime horizontal clip width",
	); err != nil {
		return err
	}

	// FPS unlock from the working rcskspel_U_horplus_fps_unlocked_0ms.dat build.
	// The original main-loop limiter waits until 25 ms have passed:
	//   cmp eax, 0x19
	// This patch changes the wait threshold to 0 ms:
	//   cmp eax, 0x00
	// Result: the hard 31/32 FPS cap is removed.
	if err := patchBytes(
		0x043321,
		[]byte{0x19},
		[]byte{0x00},
		"0 ms FPS wait",
	); err != nil {
		return err
	}

	return nil
}
