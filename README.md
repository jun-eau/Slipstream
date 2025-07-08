# Slipstream

**Slipstream** is a cross-platform app that launches the Epic Games version of Rocket League without the Epic Games Launcher.

This project is a rewrite of [RocketLeagueLauncher](https://github.com/LittleScripterBoy/RocketLeagueLauncher) by **LittleScripterBoy**, who deserves credit for the original concept and authentication. Slipstream aims for easier use by removing dependencies like Python.

## Key Benefits

*   **Launch via Steam**: Use Steam Overlay features (performance monitor, browser, SteamInput).
*   **Skip Epic Launcher**: Play Rocket League without the Epic Games Launcher running.
*   **Multi-Account Support**: Easily switch between Epic Games accounts.
*   **No Dependencies**: Download and run a single executable; no Python or other libraries needed.

## Installation

1.  Go to the [**Releases** page](https://github.com/jun-eau/Slipstream/releases).
2.  Download the correct file for your OS (`Slipstream.exe` for Windows, `Slipstream` for Linux) and place it in a dedicated folder.

## First-Time Setup

Run the launcher once per account to set it up:

1.  **Run Slipstream**: Double-click `Slipstream.exe` (or run `./Slipstream` on Linux).
2.  **Locate Game**: A file dialog will prompt you to select `RocketLeague.exe` (usually in "Binaries/Win64/" of your game installation).
3.  **Epic Games Login**: Your browser will open to the Epic Games login. Sign in.
4.  **Authorization Code**: After login, you'll be redirected to a page displaying a 32-character `authorizationCode`. Copy it.
5.  **Enter Code in Launcher**: Paste the code into the launcher's dialog box and click OK.

Rocket League will launch. A `config.json` storing your game path and session token is created in the launcher's folder, so you won't need to repeat this setup.

## Adding to Steam

1.  In Steam: **Add a Game** -> **Add a Non-Steam Game...**
2.  **Browse...** to your `Slipstream.exe` (Windows) or `Slipstream` (Linux native) and select it.
3.  Click **Add Selected Programs**.

> **Linux Users Note:** For optimal compatibility with Steam Overlay and Proton, it's often better to add the **`Slipstream.exe`** (Windows version) to Steam, not the native Linux binary. After adding it, right-click Slipstream in Steam -> **Properties...** -> **Compatibility** -> check **"Force the use of a specific Steam Play compatibility tool"** and choose the latest Proton version.

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

## License and Credits

This project is a derivative of [RocketLeagueLauncher](https://github.com/LittleScripterBoy/RocketLeagueLauncher) by **LittleScripterBoy**. Credit for the original concept and authentication flow goes to him.

The original project was unlicensed. An [issue](https://github.com/LittleScripterBoy/RocketLeagueLauncher/issues/1) requests a permissive license. Slipstream is distributed hoping the original author supports its continuation and improvement.
