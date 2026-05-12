# Technical notes: Redcat Superkarts `rcskspel.dat` aspect-ratio Hor+ + FPS patch

This document explains how the latest Redcat Superkarts patcher works:

```text
RedcatSuperkartsAspectRatioHorPlusFPSPatcher_release
```

The patcher modifies the racing executable used by Redcat Superkarts:

```text
rcskspel.dat
```

Even though the file extension is `.dat`, `rcskspel.dat` is a Windows PE executable. The game menu launches this file to run the DirectDraw/Direct3D racing portion of the game.

## What the latest patcher does

The latest patcher applies two fixes:

1. **Aspect-ratio-aware Hor+ 3D correction**
2. **Unlocked FPS / 0 ms wait patch**

It intentionally **does not patch the game’s internal resolution table** anymore.

Earlier testing showed that changing the display-mode table did not meaningfully affect the actual output resolution in the tested setup. The confirmed useful patches were:

```text
Direct3D horizontal clip range
FPS wait threshold
```

If you need to force the actual output resolution, use an external method such as dgVoodoo2 or another DirectDraw/Direct3D wrapper. This patcher controls the 3D aspect behavior, not the game’s output-resolution selection.

## Why aspect-ratio patching is needed

The original game was built around a 4:3 view.

If the old 4:3 frame is simply stretched to fill a widescreen output, the 3D image becomes too wide. Karts, roads, and world geometry look horizontally distorted.

The correct behavior is **Hor+**:

```text
Original 4:3:
  normal vertical view
  limited horizontal view

Patched widescreen:
  same vertical view
  wider horizontal view
```

The patcher achieves this by expanding the Direct3D horizontal clip range while leaving the vertical range alone.

## Formula used by the patcher

The original game is treated as a 4:3 game:

```text
baseAspect = 4 / 3 = 1.3333333...
```

The user enters a target aspect ratio, for example:

```text
16:9
16:10
21:9
32:9
1.7777
2.3333
```

The patcher calculates:

```text
Hor+ multiplier = targetAspect / baseAspect
```

Then it changes the horizontal clip range from:

```text
-1.0 to +1.0
```

to:

```text
-multiplier to +multiplier
```

The clip width becomes:

```text
2 × multiplier
```

## Common examples

| Target aspect ratio | Decimal aspect | Hor+ multiplier | Clip X | Clip width |
|---:|---:|---:|---:|---:|
| 4:3 | 1.333333 | 1.000000 | -1.000000 | 2.000000 |
| 16:10 | 1.600000 | 1.200000 | -1.200000 | 2.400000 |
| 16:9 | 1.777778 | 1.333333 | -1.333333 | 2.666667 |
| 21:9 | 2.333333 | 1.750000 | -1.750000 | 3.500000 |
| 32:9 | 3.555556 | 2.666667 | -2.666667 | 5.333333 |

For 16:9, this reproduces the previously working U-style patch.

## Modified locations in `rcskspel.dat`

The patcher modifies four Direct3D horizontal clip setup instructions and one FPS limiter byte.

The horizontal clip instructions are at these file offsets:

```text
0x0ACEBE
0x0ACEE8
0x0ACFBA
0x0ACFC6
```

The FPS limiter byte is at:

```text
0x043321
```

These are file offsets, not virtual addresses.

## Direct3D clip patch details

The patcher verifies the first three bytes of each instruction, then replaces the 32-bit floating-point constant that follows.

### 1. Default horizontal clip X

File offset:

```text
0x0ACEBE
```

Instruction prefix:

```text
C7 41 14
```

Original full instruction:

```text
C7 41 14 00 00 80 BF
```

This writes:

```text
-1.0
```

For 16:9, the replacement is:

```text
C7 41 14 AB AA AA BF
```

This writes approximately:

```text
-1.3333334
```

For other aspect ratios, the patcher keeps `C7 41 14` and writes a different calculated float.

### 2. Default horizontal clip width

File offset:

```text
0x0ACEE8
```

Instruction prefix:

```text
C7 41 1C
```

Original full instruction:

```text
C7 41 1C 00 00 00 40
```

This writes:

```text
2.0
```

For 16:9, the replacement is:

```text
C7 41 1C AB AA 2A 40
```

This writes approximately:

```text
2.6666667
```

### 3. Runtime horizontal clip X

File offset:

```text
0x0ACFBA
```

Instruction prefix:

```text
C7 46 14
```

Original full instruction:

```text
C7 46 14 00 00 80 BF
```

For 16:9:

```text
C7 46 14 AB AA AA BF
```

### 4. Runtime horizontal clip width

File offset:

```text
0x0ACFC6
```

Instruction prefix:

```text
C7 46 1C
```

Original full instruction:

```text
C7 46 1C 00 00 00 40
```

For 16:9:

```text
C7 46 1C AB AA 2A 40
```

## Why both default and runtime paths are patched

Earlier tests showed that patching only one path was not enough.

The game appears to set the Direct3D clip range both during setup and during runtime. The working U-style patch changed both paths. That is why the latest patcher writes four float constants instead of only two.

## FPS unlock

The original game waits until about 25 ms have passed before allowing the next frame. This effectively locks the game around 31/32 FPS.

At file offset:

```text
0x043321
```

the original byte is:

```text
19
```

`0x19` is decimal `25`.

The patcher changes this to:

```text
00
```

So the patch is:

```text
0x043321: 19 -> 00
```

This changes the limiter from:

```text
25 ms
```

to:

```text
0 ms
```

## Manual patching on Windows

### Recommended tools

You can use:

- HxD
- 010 Editor
- x32dbg
- Ghidra
- IDA Free
- PE-bear
- CFF Explorer

### Step 1: back up the file

Back up:

```text
rcskspel.dat
```

For example:

```text
rcskspel_original_backup.dat
```

### Step 2: confirm the file is correct

Open `rcskspel.dat` in a hex editor. It should begin with:

```text
MZ
```

That confirms it is a PE executable.

### Step 3: calculate the Hor+ values

Choose a target aspect ratio.

For 16:9:

```text
targetAspect = 16 / 9 = 1.7777778
baseAspect = 4 / 3 = 1.3333333
multiplier = targetAspect / baseAspect = 1.3333333
clipX = -1.3333333
clipWidth = 2.6666667
```

### Step 4: convert to little-endian float bytes

Common examples:

```text
16:9:
-1.3333334 = AB AA AA BF
 2.6666667 = AB AA 2A 40

16:10:
-1.2 = 9A 99 99 BF
 2.4 = 9A 99 19 40

21:9:
-1.75 = 00 00 E0 BF
 3.5  = 00 00 60 40

32:9:
-2.6666667 = AB AA 2A C0
 5.3333335 = AB AA AA 40
```

### Step 5: patch the four clip instructions

At `0x0ACEBE`, keep the prefix and replace the float:

```text
C7 41 14 00 00 80 BF
```

For 16:9:

```text
C7 41 14 AB AA AA BF
```

At `0x0ACEE8`:

```text
C7 41 1C 00 00 00 40
```

For 16:9:

```text
C7 41 1C AB AA 2A 40
```

At `0x0ACFBA`:

```text
C7 46 14 00 00 80 BF
```

For 16:9:

```text
C7 46 14 AB AA AA BF
```

At `0x0ACFC6`:

```text
C7 46 1C 00 00 00 40
```

For 16:9:

```text
C7 46 1C AB AA 2A 40
```

### Step 6: patch the FPS wait

At `0x043321`, change:

```text
19
```

to:

```text
00
```

### Step 7: save

Save the modified file as:

```text
rcskspel.dat
```

Place it back in the game folder.

## Manual patching on Linux

### Recommended tools

Useful Linux tools include:

```text
python3
xxd
hexdump
Bless Hex Editor
Okteta
Ghidra
radare2 / rizin
```

### Step 1: back up the file

```bash
cp rcskspel.dat rcskspel_original_backup.dat
```

### Step 2: use a Python patch script

Save this as:

```text
patch_rcskspel_aspect_horplus_fps.py
```

```python
from pathlib import Path
import struct
import sys
import re

if len(sys.argv) != 4:
    print("Usage: python3 patch_rcskspel_aspect_horplus_fps.py <input_rcskspel.dat> <output_rcskspel.dat> <aspect>")
    print("Example: python3 patch_rcskspel_aspect_horplus_fps.py rcskspel.dat rcskspel_16x9.dat 16:9")
    raise SystemExit(1)

input_path = Path(sys.argv[1])
output_path = Path(sys.argv[2])
aspect_text = sys.argv[3].strip().lower().replace(" ", "")

def parse_aspect_ratio(text):
    match = re.match(r"^([0-9]+(?:\.[0-9]+)?)(?:\:|/)([0-9]+(?:\.[0-9]+)?)$", text)
    if match:
        a = float(match.group(1))
        b = float(match.group(2))
        if b == 0:
            raise ValueError("aspect ratio denominator cannot be zero")
        return a / b
    return float(text)

target_aspect = parse_aspect_ratio(aspect_text)
base_aspect = 4 / 3

multiplier = target_aspect / base_aspect
clip_x = -multiplier
clip_width = multiplier * 2

data = bytearray(input_path.read_bytes())

if data[:2] != b"MZ":
    raise RuntimeError("Input file does not look like a Windows PE executable.")

def patch_float(offset, prefix, value, label):
    current_prefix = data[offset:offset + len(prefix)]
    if current_prefix != prefix:
        raise RuntimeError(
            f"{label}: unexpected instruction prefix at 0x{offset:X}. "
            "This may be the wrong rcskspel.dat version."
        )
    data[offset + len(prefix):offset + len(prefix) + 4] = struct.pack("<f", value)

patch_float(0x0ACEBE, bytes.fromhex("C7 41 14"), clip_x, "default horizontal clip x")
patch_float(0x0ACEE8, bytes.fromhex("C7 41 1C"), clip_width, "default horizontal clip width")
patch_float(0x0ACFBA, bytes.fromhex("C7 46 14"), clip_x, "runtime horizontal clip x")
patch_float(0x0ACFC6, bytes.fromhex("C7 46 1C"), clip_width, "runtime horizontal clip width")

if data[0x043321] not in (0x19, 0x00):
    raise RuntimeError("Unexpected FPS wait byte at 0x043321.")

data[0x043321] = 0x00

output_path.write_bytes(data)

print(f"Patched {input_path} -> {output_path}")
print(f"Target aspect: {target_aspect:.6f}")
print(f"Hor+ multiplier: {multiplier:.6f}")
print(f"Clip X: {clip_x:.6f}")
print(f"Clip width: {clip_width:.6f}")
print("Applied 0 ms FPS unlock.")
```

### Step 3: run the script

For 16:9:

```bash
python3 patch_rcskspel_aspect_horplus_fps.py rcskspel.dat rcskspel_16x9.dat 16:9
```

For 21:9:

```bash
python3 patch_rcskspel_aspect_horplus_fps.py rcskspel.dat rcskspel_21x9.dat 21:9
```

For 32:9:

```bash
python3 patch_rcskspel_aspect_horplus_fps.py rcskspel.dat rcskspel_32x9.dat 32:9
```

Then install the result:

```bash
cp rcskspel_16x9.dat /path/to/game/rcskspel.dat
```

### Step 4: verify the patch

```python
from pathlib import Path
import struct

data = Path("rcskspel_16x9.dat").read_bytes()

for offset, name in [
    (0x0ACEBE, "default clip x"),
    (0x0ACEE8, "default clip width"),
    (0x0ACFBA, "runtime clip x"),
    (0x0ACFC6, "runtime clip width"),
]:
    prefix = data[offset:offset+3]
    value = struct.unpack("<f", data[offset+3:offset+7])[0]
    print(hex(offset), name, prefix.hex(" "), value)

print("FPS wait byte:", data[0x043321])
```

For 16:9, the values should be approximately:

```text
default clip x      -1.333333
default clip width   2.666667
runtime clip x      -1.333333
runtime clip width   2.666667
FPS wait byte        0
```

## Building the Go patcher manually

### Windows build

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsAspectRatioHorPlusFPSPatcher.exe RedcatSuperkartsAspectRatioHorPlusFPSPatcher.go
```

### Linux build

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsAspectRatioHorPlusFPSPatcher_linux_x86_64 RedcatSuperkartsAspectRatioHorPlusFPSPatcher.go
chmod +x RedcatSuperkartsAspectRatioHorPlusFPSPatcher_linux_x86_64
```

## Troubleshooting

### The game is still stretched

Make sure the aspect ratio you entered matches the actual output aspect ratio.

Examples:

```text
1920x1080 -> 16:9
2560x1440 -> 16:9
3440x1440 -> about 2.3889:1
5120x1440 -> 32:9
```

If the actual output aspect ratio does not match the patch, the image will not look correct.

### The patcher says the instruction prefix does not match

This usually means the file at the expected offset does not contain the expected instruction prefix.

Possible causes:

- different `rcskspel.dat` version;
- already modified executable with different code at those locations;
- different regional release;
- patching the wrong file.

Use a clean original `rcskspel.dat` when possible.

### The game does not change resolution

This patcher intentionally does not patch resolution values. It only patches aspect behavior and FPS.

Use an external wrapper or renderer configuration if you need to force a specific output resolution.

### FPS unlock causes gameplay issues

The FPS unlock changes a 25 ms wait to 0 ms. On some old games, physics or input can be tied to frame timing.

To restore the original frame limiter manually:

```text
0x043321: 00 -> 19
```

## Final recommended patcher

Use:

```text
RedcatSuperkartsAspectRatioHorPlusFPSPatcher.exe
```

or on Linux:

```text
RedcatSuperkartsAspectRatioHorPlusFPSPatcher_linux_x86_64
```

Run it from the game folder next to:

```text
rcskspel.dat
```

Enter the aspect ratio that matches your actual output display.
