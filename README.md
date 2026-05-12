# Redcat Superkarts Widescreen + FPS Patcher

This repository contains a Windows command-line patcher for the DirectDraw/Direct3D racing executable used by **Redcat Superkarts**.

The relevant game file is:

```text
rcskspel.dat
```

Even though the file extension is `.dat`, this file is a PE32 Windows executable. The game menu executable starts this `.dat` file to run the actual racing portion of the game.

## What the patcher does

The patcher applies three changes to `rcskspel.dat`:

1. forces the game to use a chosen widescreen resolution, such as `2560x1440`;
2. applies the working **U-style Hor+ widescreen correction** so the 3D world is not stretched on 16:9 displays;
3. removes the built-in 31/32 FPS frame cap by changing the main-loop wait threshold from **25 ms** to **0 ms**.

## Usage

Put the patcher in the same folder as the original game file:

```text
rcskspel.dat
```

Run:

```text
RedcatSuperkartsWidescreenFPSPatcher.exe
```

The patcher asks for a resolution:

```text
Enter desired 16:9 resolution, for example 2560x1440:
```

Type for example:

```text
2560x1440
```

The patcher then:

1. creates a backup folder:
   ```text
   4x3_backup
   ```
2. moves the original `rcskspel.dat` into that folder;
3. writes a new patched `rcskspel.dat` in the game folder.

## Build from source

The patcher is written in **Go**.

### Build on Windows

Install Go, then run:

```cmd
go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsWidescreenFPSPatcher.exe RedcatSuperkartsWidescreenFPSPatcher.go
```

### Cross-compile from Linux

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsWidescreenFPSPatcher.exe RedcatSuperkartsWidescreenFPSPatcher.go
```


### Build native Linux version

The same Go source can also be compiled into a native Linux executable:

```bash
go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsWidescreenFPSPatcher_linux_x86_64 RedcatSuperkartsWidescreenFPSPatcher.go
chmod +x RedcatSuperkartsWidescreenFPSPatcher_linux_x86_64
```

You can also make the target explicit:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o RedcatSuperkartsWidescreenFPSPatcher_linux_x86_64 RedcatSuperkartsWidescreenFPSPatcher.go
chmod +x RedcatSuperkartsWidescreenFPSPatcher_linux_x86_64
```

### Usage on Linux

Place the Linux binary in the game folder next to `rcskspel.dat`, then run:

```bash
./RedcatSuperkartsWidescreenFPSPatcher_linux_x86_64
```

Enter a resolution such as:

```text
2560x1440
```

The Linux version performs the same patching process as the Windows version: it creates `4x3_backup`, moves the original `rcskspel.dat` into that folder, and writes a new widescreen + unlocked-FPS patched `rcskspel.dat`.

## Technical summary

### Resolution patch

The game contains a small display-mode table with several old 4:3 modes. The patcher updates the width and height values in the mode table to the resolution entered by the user.

The relevant width offsets are:

```text
0x0FE930
0x0FE984
0x0FE9FC
0x0FEA9C
```

The relevant height offsets are:

```text
0x0FE934
0x0FE988
0x0FEA00
0x0FEAA0
```

The values are written as 32-bit little-endian integers.

For example, `2560x1440` becomes:

```text
width  = 2560 = 00 0A 00 00
height = 1440 = A0 05 00 00
```

### Hor+ 16:9 correction

Simply changing the display mode is not enough. The 3D image can still stretch or crop incorrectly.

The working U-style widescreen patch changes the Direct3D horizontal clip range:

```text
-1.0 to +1.0
```

to approximately:

```text
-1.3333334 to +1.3333334
```

This gives proper **Hor+** behavior:

```text
same vertical view
more horizontal view
no horizontal stretching
```

The relevant patched constants are:

```text
BF800000 -> BFAAAAAB
40000000 -> 402AAAAB
```

These are applied in both the default and runtime D3D clip setup paths.

### FPS unlock

The original game waits until about **25 ms** have passed before allowing the next frame. That corresponds to roughly **31/32 FPS**.

The FPS patch changes the limiter threshold from:

```text
25 ms
```

to:

```text
0 ms
```

At file offset:

```text
0x043321
```

the byte is changed from:

```text
19
```

to:

```text
00
```

This removes the hard frame cap.

## Notes

- Back up your game files before patching.
- The patcher automatically backs up the original `rcskspel.dat` into `4x3_backup`.
- This patch was built from the earlier working `rcskspel_U_horplus_fps_unlocked_0ms.dat` test result.
- Because the FPS unlock removes a very old timing cap, gameplay behavior may vary by system.
