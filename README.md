# Slipstream

**Slipstream** is a standalone, cross-platform app that launches the Epic Games version of Rocket League without the Epic Games Launcher.

This project would not exist without the work of **LittleScripterBoy** on the original [RocketLeagueLauncher](https://github.com/LittleScripterBoy/RocketLeagueLauncher). All credit for the authentication flow and the original concept belongs entirely to him. Slipstream is a rewrite that builds upon the foundation he created, aiming to make the tool more accessible and straightforward.

### Key Benefits

*   **Integrate Anywhere**: Add Slipstream to Steam for its overlay and controller support, run it standalone, or integrate it with other game launchers like Playnite or Lutris.
*   **Skip Epic Launcher**: Play Rocket League without the Epic Games Launcher running in the background.
*   **Simple Multi-Account Support**: Easily switch between multiple Epic Games accounts.
*   **No Dependencies**: Download and run a single executable; no Python or other libraries needed.

## Installation

> **Prerequisites:** Rocket League must be installed and kept up-to-date using a game manager like the Epic Games Launcher or Heroic Games Launcher. Slipstream only *launches* the game; it does not install, update, or manage game files.

1.  Go to the [**Releases** page](https://github.com/jun-eau/Slipstream/releases) and download the executable for your platform.
2.  Place the downloaded file in a dedicated folder, as Slipstream will generate configuration files in the same directory.
3.  Continue with the setup instructions for your operating system below.

### Windows Setup

For the best experience, it's recommended to add Slipstream to Steam to get the overlay and controller support.

1.  **Run the Launcher**: Double-click `Slipstream.exe`.
2.  **Locate Game File**: A file dialog will open. Navigate to your Rocket League installation and select `RocketLeague.exe` (usually in the `Binaries/Win64` folder).
3.  **Log In to Epic Games**: Your browser will open. Log in to the Epic account you want to use.
4.  **Get Authorization Code**: After logging in, you'll be redirected to a page with a 32-character `authorizationCode`. Copy this code.
5.  **Enter Code**: Paste the code into the launcher's dialog box and click OK.

The game will now launch. A `config.json` file is created, so you won't have to repeat this.

**To add Slipstream to Steam:**
1. In Steam: **Add a Game** -> **Add a Non-Steam Game...**
2. **Browse...** to where you saved `Slipstream.exe` and select it.
3. Click **Add Selected Programs**.

### Linux Setup

The recommended method for Linux is to use the Windows version (`Slipstream.exe`) with Proton, as this provides the best compatibility with gamepads and the Steam Overlay.

> For **Steam Deck users**: do this setup process in **Desktop Mode** to ensure the file browser and Epic login work; the game will appear in Gaming Mode after setup.

1.  **Add to Steam**: In your Steam library, click **Add a Game** -> **Add a Non-Steam Game...** and select the `Slipstream.exe` file.
2.  **Force Proton**: Right-click on Slipstream in Steam -> **Properties...** -> **Compatibility**. Check the box to **"Force the use of a specific Steam Play compatibility tool"** and choose the latest Proton version.
3.  **Run and Configure**: Launch Slipstream from Steam. It will guide you through the one-time setup (locating `RocketLeague.exe`, browser login, etc.) just like on Windows.

Once configured, Slipstream will launch Rocket League correctly using Proton.

> **Using the Native Binary:** The native Linux binary is also provided. Its primary purpose is for initial setup without needing Steam/Proton running, or for use with other tools like Lutris. If you're having trouble with the file picker or authentication process under Proton, using the native binary first for setup can be a reliable alternative. To use it:
> 1. Make it executable with `chmod +x Slipstream` and run with `./Slipstream`. 
> 2. Run through the setup as usual.
> 3. After it saves your `config.json`, it will display a confirmation message, as it cannot launch the Windows game directly. 
> 4. You can then use `Slipstream.exe` with your preferred compatibility layer. Ensure it is in the same directory as `config.json`.

## Usage & Configuration

### Updating Slipstream

Replace your current Slipstream executable with the latest one from the [Releases page](https://github.com/jun-eau/Slipstream/releases).

### Custom Launch Options

Add Rocket League launch options via Steam:

1.  In Steam, right-click Slipstream -> **Properties...**
2.  Under the **General** tab, find **Launch Options**.
3.  Enter options, space-separated (e.g., `-nomovie -high -USEALLAVAILABLECORES`). Slipstream forwards these to the game.

This should also work well with other launchers.

### Using Multiple Accounts

1.  Create a **new, separate folder** for each additional account.
2.  Copy the Slipstream executable into each new folder.
3.  Run it from the new folder and complete the first-time setup for that account. Each folder will have its own `config.json`, keeping accounts isolated. If using Steam, create separate library entries for each.

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
