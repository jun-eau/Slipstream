# Slipstream: A Rocket League Launcher

Slipstream is a standalone, cross-platform app that launches the Epic Games version of Rocket League **without the Epic Games Launcher**.

This project builds upon the original [RocketLeagueLauncher](https://github.com/LittleScripterBoy/RocketLeagueLauncher) by **LittleScripterBoy**.

### Key Benefits

*   **Integrate with Steam**: Seamlessly add the Epic Games version of Rocket League to Steam. Enables full support for the Steam Overlay, controller configurations, and the Steam Deck.
*   **Skip the Epic Launcher**: Play Rocket League without the Epic Launcher running. Perfect for standalone use or with other launchers like Playnite and Lutris.
*   **Optional BakkesMod**: Automatically launch BakkesMod with Rocket League (Windows & Linux).
*   **Simple Multi-Account**: Easily switch between Epic accounts in Rocket League.
*   **No Dependencies**: A single, dependency-free executable.

## Installation & Setup

> **Prerequisite:** Rocket League must be installed and kept up-to-date via a game manager (e.g., Epic Games Launcher, Heroic Games Launcher). Slipstream only *launches* the game.

1.  **Download**: Go to the [**Releases page**](https://github.com/jun-eau/Slipstream/releases/latest) and download the executable for your platform.
2.  **Create a Folder**: Place the downloaded file in a new, dedicated folder. Slipstream will store its configuration file (`config.json`) there.
3.  **Run Slipstream**:
    *   **Windows**: Double-click `Slipstream.exe`.
    *   **Linux / Steam Deck**: The recommended method is to add `Slipstream.exe` to Steam as a non-Steam game and force the latest Proton version in its compatibility settings.
        *   **Steam Deck users must do this in Desktop Mode.**
        *   If the recommended method fails, use the native Linux binary (`chmod +x Slipstream && ./Slipstream`) to run the initial setup first.
4.  **Initial Configuration (One-Time Setup)**:
    *   **Locate Files:** The app will prompt you to select `RocketLeague.exe`. If you enable BakkesMod, it will also prompt for `BakkesMod.exe` (see "Optional: BakkesMod Setup" under Usage for more details).
    *   **Epic Games Login:** Your browser will open to log in. Copy the 32-character `authorizationCode` from the final page and paste it into Slipstream's dialog.

The game will launch, and your settings will be saved in the `config.json` file.

**To add Slipstream to Steam (Windows & Linux/Proton):**
1. In Steam: **Add a Game** -> **Add a Non-Steam Game...**
2. **Browse...** to `Slipstream.exe` and select it.
3. Click **Add Selected Programs**.

## Usage

*   **Updating Slipstream**: Slipstream will automatically notify you about new versions. To update, simply replace your executable with the latest one from the [Releases page](https://github.com/jun-eau/Slipstream/releases/latest). Your `config.json` is preserved.
*   **Custom Launch Options**:
    1.  In Steam, right-click Slipstream -> **Properties...**
    2.  Under **General**, enter options in **Launch Options** (e.g., `-nomovie -high`). These are passed to Rocket League.
*   **Multiple Accounts**:
    1.  Create a **new, separate folder** for each account.
    2.  Copy the Slipstream executable into each new folder.
    3.  Run it from the new folder for that account's first-time setup.
    4.  If using Steam, add each Slipstream instance as a separate non-Steam game.

<details>
<summary>Optional: BakkesMod Setup</summary>

Slipstream can automatically launch BakkesMod. If enabled during initial setup, you'll be prompted for `BakkesMod.exe`.

**Windows:**
1. Install BakkesMod from [bakkesmod.com](https://bakkesmod.com/).
2. When Slipstream asks, locate `BakkesMod.exe` (usually `C:\Program Files\BakkesMod\BakkesMod.exe`).

**Linux (using Wine/Proton):**
BakkesMod is a Windows application, so it runs within Wine/Proton.
1. Download `BakkesModSetup.exe` from [bakkesmod.com](https://bakkesmod.com/).
2. Install it using your Wine/Proton environment:
    * **Proton (via Steam):** Add `BakkesModSetup.exe` as a non-Steam game, force the same Proton version as Slipstream/Rocket League, and run it once.
    * **Wine (standalone):** `wine /path/to/BakkesModSetup.exe`.
3. Point Slipstream to the installed `BakkesMod.exe` within your Wine/Proton prefix (e.g., `~/.wine/drive_c/Program Files/BakkesMod/BakkesMod.exe` or `~/.steam/steam/steamapps/compatdata/<AppID>/pfx/drive_c/Program Files/BakkesMod/BakkesMod.exe`).

> **If "Mod is out of date, waiting for an update" appears:** In the BakkesMod window (once running with Rocket League), go to "Settings", uncheck "Enable safe mode", and click "Yes" on the warning.

> **Steam Deck Users:** Navigating the BakkesMod window in Gaming Mode may require using the `Steam` button to access window controls.

For detailed Linux help, see the [BakkesLinux guide](https://github.com/CrumblyLiquid/BakkesLinux) (Setup/Installation sections). Additionally, for a step-by-step walkthrough of using Slipstream with Heroic (including auto-updates), see the [bakkeslinux guide](https://github.com/beidoubagel/bakkeslinux) by @beidoubagel.
</details>

<details>
<summary>FAQ & Troubleshooting</summary>

#### Q: Do I still need the Epic Games Launcher installed?
Yes, for installing and updating Rocket League. Slipstream lets you play without running the Epic Launcher.

#### Q: Does this improve in-game FPS?
It can speed up game boot time but shouldn't affect in-game FPS.

#### Q: What's the difference between this and Heroic/Legendary?
Slipstream is minimal, focused only on launching Rocket League via other launchers (like Steam) without extra dependencies. Heroic/Legendary manage entire game libraries.

#### Q: Does Slipstream modify game files?
No. It only reads your game path to launch the game.

#### Q: I'm getting a "version mismatch" error when I try to play online.
This means your game is out of date. Since Slipstream bypasses the launcher, it also bypasses the automatic update check. Run the Epic Games Launcher or your launcher of choice to make sure Rocket League is fully updated, then try launching with Slipstream again.

#### Q: My game is in the wrong language, how do I change it?
The Epic Launcher normally passes a language argument to the game. You can do this yourself in Slipstream's launch options. To force English, add `-language=INT`. Other common codes include `DEU` (German), `FRA` (French), and `ESN` (Spanish).
</details>

<details>
<summary>Building from Source</summary>

Requires **Go toolchain** (v1.24+).

1.  **Clone:** `git clone https://github.com/jun-eau/Slipstream.git`
2.  **Navigate:** `cd Slipstream`
3.  **Build:**
    *   **Windows (64-bit):** `go build -o Slipstream.exe .`
        *   Cross-compile on Linux/macOS: `GOOS=windows GOARCH=amd64 go build -o Slipstream.exe .`
    *   **Linux (64-bit):** `go build -o Slipstream .`
        *   Cross-compile on Windows: `$env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o Slipstream .`

The executable will be in the project directory.
</details>

## License and Credits

This project is open source (see `LICENSE` file).

It's a derivative of [RocketLeagueLauncher](https://github.com/LittleScripterBoy/RocketLeagueLauncher) by **LittleScripterBoy**. Original project was unlicensed; an [issue](https://github.com/LittleScripterBoy/RocketLeagueLauncher/issues/1) requests a permissive license. Slipstream is distributed in good faith.
