# Technical notes: Redcat Superkarts `rcskspel.dat` widescreen + FPS patch

This document explains how the Redcat Superkarts racing executable, `rcskspel.dat`, was modified so the game can render correctly on a 16:9 display without horizontally stretching the 3D world.

The patcher created for this project is:

```text
RedcatSuperkartsWidescreenFPSPatcher.exe
```

The game file it patches is:

```text
rcskspel.dat
```

Even though the extension is `.dat`, `rcskspel.dat` is a normal Windows PE executable. It is launched by the game’s menu executable and contains the DirectDraw/Direct3D racing portion of the game.

## Goal of the patch

The original game is built around 4:3-era display modes and appears to run at roughly 31/32 FPS by default.

The final patch does three things:

1. patches the game’s internal display-mode table to the requested resolution, such as `2560x1440`;
2. patches the Direct3D horizontal clip range so the 3D world becomes proper **Hor+ widescreen** instead of stretched;
3. removes the old 25 ms frame limiter by setting it to `0 ms`, producing the same behavior as the working unlocked-FPS test build.

## What “Hor+” means

The correct widescreen behavior is:

```text
Original 4:3:
  normal vertical view
  limited horizontal view

Patched 16:9:
  same vertical view
  wider horizontal view
```

This is called **Hor+**.

The incorrect behavior would be simply stretching the old 4:3 image to fill 16:9. That makes the 3D world look wide and distorted. The patch avoids that by expanding the horizontal Direct3D clip range while leaving the vertical clip range alone.

## Important discovery during testing

Several earlier attempts only affected the HUD or the 2D overlay. Those did not change the 3D road/karts/world.

The successful path was found by changing the Direct3D viewport clip values. The important test result was:

- changing the vertical clip value caused vertical zoom/cropping;
- changing the horizontal clip range expanded the 3D view correctly;
- the working versions were known as the **T** and **U** Hor+ builds;
- the patcher uses the **U-style default + runtime horizontal clip range** approach.

## PE layout notes

For this executable, file offsets are used directly by the patcher. The values below are file offsets, not virtual addresses.

If you are viewing the file in a disassembler, the corresponding virtual addresses may depend on the executable image base and section mapping. In a hex editor, use the file offsets exactly as listed.

## Patch part 1: resolution / display-mode table

The game stores several display-mode table entries as 32-bit little-endian integers.

The patcher writes the user-specified width and height into the relevant table entries.

### Width offsets

```text
0x0FE930
0x0FE984
0x0FE9FC
0x0FEA9C
```

### Height offsets

```text
0x0FE934
0x0FE988
0x0FEA00
0x0FEAA0
```

### Example: 2560×1440

For `2560x1440`, write:

```text
2560 decimal = 0x00000A00
little endian = 00 0A 00 00
```

and:

```text
1440 decimal = 0x000005A0
little endian = A0 05 00 00
```

So each width offset receives:

```text
00 0A 00 00
```

and each height offset receives:

```text
A0 05 00 00
```

### Example: 1920×1080

For `1920x1080`, write:

```text
1920 decimal = 0x00000780
little endian = 80 07 00 00
```

and:

```text
1080 decimal = 0x00000438
little endian = 38 04 00 00
```

## Patch part 2: Hor+ Direct3D horizontal clip range

The original game uses a horizontal clip range equivalent to:

```text
-1.0 to +1.0
```

That is represented by:

```text
clip x     = -1.0
clip width =  2.0
```

The working U-style widescreen patch changes this to approximately:

```text
-1.3333334 to +1.3333334
```

That is represented by:

```text
clip x     = -1.3333334
clip width =  2.6666667
```

This is the 4:3-to-16:9 horizontal expansion factor:

```text
(16 / 9) / (4 / 3) = 1.3333333...
```

The vertical values are intentionally left alone. This preserves the original vertical framing.

### Float constants

IEEE-754 little-endian bytes:

```text
-1.0       = 00 00 80 BF
-1.3333334 = AB AA AA BF

 2.0       = 00 00 00 40
 2.6666667 = AB AA 2A 40
```

## Patch part 3: default horizontal clip setup

At file offset:

```text
0x0ACEBD
```

replace:

```text
C7 41 14 00 00 80 BF
```

with:

```text
C7 41 14 AB AA AA BF
```

This changes the default horizontal clip X value from `-1.0` to approximately `-1.3333334`.

At file offset:

```text
0x0ACEE7
```

replace:

```text
C7 41 1C 00 00 00 40
```

with:

```text
C7 41 1C AB AA 2A 40
```

This changes the default horizontal clip width from `2.0` to approximately `2.6666667`.

## Patch part 4: runtime horizontal clip setup

The game also sets clip values at runtime, so the same style of patch is applied to the runtime path.

At file offset:

```text
0x0ACFB9
```

replace:

```text
C7 46 14 00 00 80 BF
```

with:

```text
C7 46 14 AB AA AA BF
```

At file offset:

```text
0x0ACFC5
```

replace:

```text
C7 46 1C 00 00 00 40
```

with:

```text
C7 46 1C AB AA 2A 40
```

The combination of the default and runtime clip patches is what matched the working **U** build.

## Patch part 5: FPS unlock / 0 ms wait

The original game has a frame limiter that waits until about 25 ms have passed before allowing the next frame. This produces the default 31/32 FPS behavior.

The unlocked-FPS test build changed that wait threshold to `0 ms`.

At file offset:

```text
0x043321
```

replace:

```text
19
```

with:

```text
00
```

`0x19` is decimal `25`, so this changes the limiter from:

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
- CFF Explorer
- PE-bear

### Step 1: back up the file

Copy:

```text
rcskspel.dat
```

to something like:

```text
rcskspel_original_backup.dat
```

Never patch your only copy.

### Step 2: confirm it is the right file

Open `rcskspel.dat` in a hex editor or PE tool.

The file should begin with:

```text
MZ
```

This means it is a Windows PE executable.

### Step 3: patch the resolution table

Choose your target resolution.

For `2560x1440`:

- write `00 0A 00 00` at each width offset;
- write `A0 05 00 00` at each height offset.

Width offsets:

```text
0x0FE930
0x0FE984
0x0FE9FC
0x0FEA9C
```

Height offsets:

```text
0x0FE934
0x0FE988
0x0FEA00
0x0FEAA0
```

### Step 4: patch the Hor+ clip values

Patch the default clip setup:

```text
0x0ACEBD:
C7 41 14 00 00 80 BF
->
C7 41 14 AB AA AA BF
```

```text
0x0ACEE7:
C7 41 1C 00 00 00 40
->
C7 41 1C AB AA 2A 40
```

Patch the runtime clip setup:

```text
0x0ACFB9:
C7 46 14 00 00 80 BF
->
C7 46 14 AB AA AA BF
```

```text
0x0ACFC5:
C7 46 1C 00 00 00 40
->
C7 46 1C AB AA 2A 40
```

### Step 5: patch the FPS limiter

At file offset:

```text
0x043321
```

replace:

```text
19
```

with:

```text
00
```

### Step 6: save the file

Save the patched file as:

```text
rcskspel.dat
```

Place it back in the game folder.

## Manual patching on Linux

### Recommended tools

Useful Linux tools:

```text
python3
xxd
hexdump
objdump
Ghidra
radare2 / rizin
Bless Hex Editor
Okteta
```

### Step 1: back up the file

```bash
cp rcskspel.dat rcskspel_original_backup.dat
```

### Step 2: patch with Python

Save this as:

```text
patch_rcskspel_widescreen_fps.py
```

```python
from pathlib import Path
import struct
import sys

if len(sys.argv) != 4:
    print("Usage: python3 patch_rcskspel_widescreen_fps.py <input_rcskspel.dat> <output_rcskspel.dat> <width>x<height>")
    raise SystemExit(1)

input_path = Path(sys.argv[1])
output_path = Path(sys.argv[2])
resolution = sys.argv[3].lower()

width_text, height_text = resolution.split("x")
width = int(width_text)
height = int(height_text)

data = bytearray(input_path.read_bytes())

if data[:2] != b"MZ":
    raise RuntimeError("Input file does not look like a Windows PE executable.")

width_offsets = [
    0x0FE930,
    0x0FE984,
    0x0FE9FC,
    0x0FEA9C,
]

height_offsets = [
    0x0FE934,
    0x0FE988,
    0x0FEA00,
    0x0FEAA0,
]

for offset in width_offsets:
    data[offset:offset + 4] = struct.pack("<I", width)

for offset in height_offsets:
    data[offset:offset + 4] = struct.pack("<I", height)

def patch_bytes(offset, original, replacement, label):
    current = data[offset:offset + len(replacement)]

    if current == replacement:
        return

    if current != original:
        raise RuntimeError(
            f"{label}: unexpected bytes at 0x{offset:X}. "
            "This may be the wrong rcskspel.dat version."
        )

    data[offset:offset + len(replacement)] = replacement

# U-style Hor+ Direct3D horizontal clip patch.
patch_bytes(
    0x0ACEBD,
    bytes.fromhex("C7 41 14 00 00 80 BF"),
    bytes.fromhex("C7 41 14 AB AA AA BF"),
    "default clip x",
)

patch_bytes(
    0x0ACEE7,
    bytes.fromhex("C7 41 1C 00 00 00 40"),
    bytes.fromhex("C7 41 1C AB AA 2A 40"),
    "default clip width",
)

patch_bytes(
    0x0ACFB9,
    bytes.fromhex("C7 46 14 00 00 80 BF"),
    bytes.fromhex("C7 46 14 AB AA AA BF"),
    "runtime clip x",
)

patch_bytes(
    0x0ACFC5,
    bytes.fromhex("C7 46 1C 00 00 00 40"),
    bytes.fromhex("C7 46 1C AB AA 2A 40"),
    "runtime clip width",
)

# FPS unlock:
# 25 ms wait threshold -> 0 ms wait threshold.
patch_bytes(
    0x043321,
    bytes.fromhex("19"),
    bytes.fromhex("00"),
    "FPS wait threshold",
)

output_path.write_bytes(data)

print(f"Patched {input_path} -> {output_path}")
print(f"Resolution: {width}x{height}")
print("Applied Hor+ 16:9 correction.")
print("Applied 0 ms FPS unlock.")
```

Run it like this:

```bash
python3 patch_rcskspel_widescreen_fps.py rcskspel.dat rcskspel_2560x1440_unlocked.dat 2560x1440
```

Then install:

```bash
cp rcskspel_2560x1440_unlocked.dat /path/to/game/rcskspel.dat
```

### Step 3: verify with Python

You can verify the display-mode table like this:

```python
from pathlib import Path
import struct

data = Path("rcskspel_2560x1440_unlocked.dat").read_bytes()

for offset in [0x0FE930, 0x0FE984, 0x0FE9FC, 0x0FEA9C]:
    print(hex(offset), struct.unpack_from("<I", data, offset)[0])

for offset in [0x0FE934, 0x0FE988, 0x0FEA00, 0x0FEAA0]:
    print(hex(offset), struct.unpack_from("<I", data, offset)[0])
```

For `2560x1440`, the width offsets should print `2560`, and the height offsets should print `1440`.

You can also check the clip and FPS bytes:

```python
from pathlib import Path

data = Path("rcskspel_2560x1440_unlocked.dat").read_bytes()

checks = {
    0x0ACEBD: "C7 41 14 AB AA AA BF",
    0x0ACEE7: "C7 41 1C AB AA 2A 40",
    0x0ACFB9: "C7 46 14 AB AA AA BF",
    0x0ACFC5: "C7 46 1C AB AA 2A 40",
    0x043321: "00",
}

for offset, expected in checks.items():
    actual = data[offset:offset + len(bytes.fromhex(expected))]
    print(hex(offset), actual.hex(" ").upper(), "expected", expected)
```

## Adapting to other 16:9 resolutions

The Hor+ clip patch remains the same for common 16:9 resolutions.

Only the display-mode table values change.

Examples:

| Resolution | Width bytes | Height bytes |
|---:|---:|---:|
| 1280×720 | `00 05 00 00` | `D0 02 00 00` |
| 1600×900 | `40 06 00 00` | `84 03 00 00` |
| 1920×1080 | `80 07 00 00` | `38 04 00 00` |
| 2560×1440 | `00 0A 00 00` | `A0 05 00 00` |
| 3840×2160 | `00 0F 00 00` | `70 08 00 00` |
| 7680×4320 | `00 1E 00 00` | `E0 10 00 00` |

## Troubleshooting

### The game still opens at 800×600 or another old mode

The mode table may not have been patched in all required places. Check all width and height offsets listed above.

### The 3D world is stretched

Check the Hor+ clip patches. The horizontal clip constants must become:

```text
-1.3333334
2.6666667
```

The byte patterns should be:

```text
AB AA AA BF
AB AA 2A 40
```

### The image is vertically zoomed or cropped

Do not patch the vertical clip range. Earlier tests showed that changing the vertical clip path caused a vertical zoom/crop effect.

### HUD changes but the 3D world does not

That means the wrong path was patched. The final working patch targets the Direct3D horizontal clip range, not just DirectDraw or HUD/overlay dimensions.

### The game runs too fast or controls behave differently

The FPS patch removes the 25 ms frame limiter. This is the same behavior as the working `0 ms` unlocked-FPS test build, but very old games can have physics or input code tied to frame timing.

If needed, restore the original byte at `0x043321`:

```text
00 -> 19
```

to return to the original limiter.

## Final recommended patcher

Use:

```text
RedcatSuperkartsWidescreenFPSPatcher.exe
```

Run it from the game folder, enter a resolution such as:

```text
2560x1440
```

The patcher backs up the original file and writes a patched `rcskspel.dat` automatically.
