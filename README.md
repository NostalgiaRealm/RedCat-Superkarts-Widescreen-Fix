# Redcat Superkarts Aspect-Ratio Hor+ + FPS Unlock Patcher

<img width="2560" height="1080" alt="Screenshot" src="https://github.com/user-attachments/assets/9f2fcd27-b026-44c6-855e-884638fe3804" />

Note: In order to get a 16 by 9 picture, you will still need to use a wrapper like dgVoodoo2 to force the actual output resolution. On Linux you will need to use dgVoodoo2 together with DXVK.

This is a Redcat Superkarts patcher for:

```text
rcskspel.dat
```

It does **not** patch or force the game's display resolution.

Testing showed that changing the game's resolution table did not meaningfully affect the actual output resolution on the tested setup. So this patcher only applies the parts that mattered:

1. **Hor+ Direct3D projection/clip correction**
2. **Unlocked FPS / 0 ms wait patch**

The difference from the earlier 16:9-only patcher is that this version asks for an **aspect ratio**, then calculates the Hor+ patch for that ratio.

## Usage on Windows

Put the patcher in the same folder as:

```text
rcskspel.dat
```

Run:

```text
RedcatSuperkartsAspectRatioHorPlusFPSPatcher.exe
```

Enter an aspect ratio, for example:

```text
16:9
```

Other accepted examples:

```text
16:10
21:9
32:9
1.7777
2.3333
```

The patcher then:

1. creates a backup folder:
   ```text
   4x3_backup
   ```
2. moves the original `rcskspel.dat` into that folder;
3. writes a new patched `rcskspel.dat` in the game folder.

## Usage on Linux

```bash
chmod +x RedcatSuperkartsAspectRatioHorPlusFPSPatcher_linux_x86_64
./RedcatSuperkartsAspectRatioHorPlusFPSPatcher_linux_x86_64
```

Enter an aspect ratio such as:

```text
16:9
```

## What the aspect ratio does

The original game is built around a 4:3 projection.

The patcher calculates:

```text
Hor+ multiplier = requested aspect / original 4:3 aspect
```

So:

```text
16:9  -> (16/9) / (4/3) = 1.3333333
16:10 -> (16/10) / (4/3) = 1.2
21:9  -> (21/9) / (4/3) = 1.75
32:9  -> (32/9) / (4/3) = 2.6666667
```

It then changes the Direct3D horizontal clip range from:

```text
-1.0 to +1.0
```

to:

```text
-multiplier to +multiplier
```

For example, for 16:9:

```text
-1.3333333 to +1.3333333
```

This gives Hor+ widescreen behavior:

```text
same vertical view
more horizontal view
no horizontal stretching
```

## Technical details

The patcher modifies the following Direct3D horizontal clip instructions:

```text
0x0ACEBE
0x0ACEE8
0x0ACFBA
0x0ACFC6
```

It verifies the instruction prefixes and then replaces only the float constants.

For 16:9, the values become the same as the previously working U-style patch:

```text
-1.3333334 = AB AA AA BF
 2.6666667 = AB AA 2A 40
```

For another aspect ratio, the patcher writes different float constants using the same formula.

## FPS unlock

The original game waits until about 25 ms have passed before allowing the next frame, which effectively locks the game around 31/32 FPS.

The FPS patch changes:

```text
0x043321: 19 -> 00
```

This changes the wait threshold from:

```text
25 ms
```

to:

```text
0 ms
```

## Build from source

The patcher is written in Go.

### Build Windows EXE

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsAspectRatioHorPlusFPSPatcher.exe RedcatSuperkartsAspectRatioHorPlusFPSPatcher.go
```

### Build native Linux binary

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsAspectRatioHorPlusFPSPatcher_linux_x86_64 RedcatSuperkartsAspectRatioHorPlusFPSPatcher.go
chmod +x RedcatSuperkartsAspectRatioHorPlusFPSPatcher_linux_x86_64
```

## Notes

- Back up your game files before patching.
- The patcher makes its own backup in `4x3_backup`.
- This patcher intentionally does not patch resolution values.
- You still need to force the actual output resolution externally if the game does not select it on its own.
- The patch is based on the previously working U-style Hor+ fix and the `0 ms` FPS unlock.
