# Slipstream

**Slipstream** is a standalone, cross-platform app that launches the Epic Games version of Rocket League without the Epic Games Launcher.

This project would not exist without the pioneering work of **LittleScripterBoy** on the original [RocketLeagueLauncher](https://github.com/LittleScripterBoy/RocketLeagueLauncher). All credit for discovering the authentication flow and the original concept belongs entirely to him. Slipstream is a rewrite that builds upon the solid foundation he created, aiming to make the tool more accessible.

## Key Benefits

*   **Integrate Anywhere**: Add Slipstream to Steam for its overlay and controller support, run it standalone, or integrate it with other game launchers like Playnite or Lutris.
*   **Skip Epic Launcher**: Play Rocket League without the Epic Games Launcher running in the background.
*   **Simple Multi-Account Support**: Easily switch between multiple Epic Games accounts.
*   **No Dependencies**: Download and run a single executable; no Python or other libraries needed.

## Installation

> **Prerequisites:** Rocket League must be installed and kept up-to-date using a game manager like the Epic Games Launcher or Heroic Games Launcher. Slipstream only *launches* the game; it does not install, update, or manage game files.

1.  Go to the [**Releases** page](https://github.com/jun-eau/Slipstream/releases).
2.  Download the correct file for your OS (`Slipstream.exe` for Windows, `Slipstream` for Linux) and place it in a dedicated folder.

> **Note for Linux Users:**
> Before running the native Linux binary, you must make it executable. Open a terminal, navigate to the folder where you saved `Slipstream`, and run the following command:
> ```sh
> chmod +x Slipstream
> ```
> You can then run the application with `./Slipstream`.

## First-Time Setup

Run the launcher once per account to set it up:

1.  **Run Slipstream**: Double-click `Slipstream.exe` (or run `./Slipstream` on Linux).
2.  **Locate Game**: A file dialog will prompt you to select `RocketLeague.exe` (usually in "/Binaries/Win64" of your game installation).
3.  **Epic Games Login**: Your browser will open to the Epic Games login. Sign in.
4.  **Authorization Code**: After login, you'll be redirected to a page displaying a 32-character `authorizationCode`. Copy it.
5.  **Enter Code in Launcher**: Paste the code into the launcher's dialog box and click OK.

Rocket League will launch. A `config.json` storing your game path and session token is created in the launcher's folder, so you won't need to repeat this setup.

## Adding to Steam

While Slipstream can be run on its own, adding it to a launcher like Steam unlocks features like the Steam Overlay (for performance monitoring, web browser) and SteamInput (for advanced controller configuration).

1.  In Steam: **Add a Game** -> **Add a Non-Steam Game...**
2.  **Browse...** to your `Slipstream.exe` (Windows) or `Slipstream` (Linux native) and select it.
3.  Click **Add Selected Programs**.

> **Note for Steam on Linux:** For optimal compatibility with Steam Overlay and Proton, it's often better to add the **`Slipstream.exe`** (Windows version) to Steam, not the native Linux binary. After adding it, right-click Slipstream in Steam -> **Properties...** -> **Compatibility** -> check **"Force the use of a specific Steam Play compatibility tool"** and choose the latest Proton version.

## Updating Slipstream

Replace your current Slipstream executable with the latest one from the [Releases page](https://github.com/jun-eau/Slipstream/releases).

## Custom Launch Options

Add Rocket League launch options (e.g., `-nomovie -high`) via Steam:

1.  In Steam, right-click Slipstream -> **Properties...**
2.  Under the **General** tab, find **Launch Options**.
3.  Enter options, space-separated (e.g., `-nomovie -high -USEALLAVAILABLECORES`). Slipstream forwards these to the game.

## Using Multiple Accounts

1.  Create a **new, separate folder** for each additional account.
2.  Copy the Slipstream executable into each new folder.
3.  Run it from the new folder and complete the first-time setup for that account.
    Each folder will have its own `config.json`, keeping accounts isolated. Create separate Steam library entries for each.

## Troubleshooting

*   **Authentication Error / Expired Session / Wrong Game Path**: Delete `config.json` in the launcher's folder and rerun Slipstream. This restarts the setup, allowing you to re-authenticate and/or correct the game path.

## Building from Source

If you prefer to compile the application yourself, you will need the **Go toolchain** (version 1.24 or newer) installed on your system.

1.  **Clone the repository:**
    ```sh
    git clone https://github.com/jun-eau/Slipstream.git
    ```

2.  **Navigate to the project directory:**
    ```sh
    cd Slipstream
    ```

3.  **Build the executable:**
    The Go toolchain makes it simple to compile for different operating systems. Run the command corresponding to your target platform.

    **For Windows (64-bit):**
    ```sh
    go build -o Slipstream.exe .
    ```
    *On Linux or macOS, you can cross-compile for Windows with:*
    ```sh
    GOOS=windows GOARCH=amd64 go build -o Slipstream.exe .
    ```

    **For Linux (64-bit):**
    ```sh
    go build -o Slipstream .
    ```
    *On Windows, you can cross-compile for Linux with:*
    ```powershell
    $env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -o Slipstream .
    ```

After running the command, the `Slipstream.exe` or `Slipstream` executable will be created in the project directory.

## License and Credits

This project is open source. See the `LICENSE` file for more details.

This project is a derivative of [RocketLeagueLauncher](https://github.com/LittleScripterBoy/RocketLeagueLauncher) by **LittleScripterBoy**. The original project was unlicensed; an [issue](https://github.com/LittleScripterBoy/RocketLeagueLauncher/issues/1) has been opened requesting a permissive license be added. Slipstream is distributed in the good-faith belief that the original author would not object to the continuation and improvement of their work.