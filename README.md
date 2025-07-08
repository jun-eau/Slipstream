# Rocket League Launcher (Go Edition)

This is a standalone, cross-platform application that launches the Epic Games version of Rocket League without needing the Epic Games Launcher.

This project is a complete rewrite of the original Python-based [RocketLeagueLauncher](https://github.com/LittleScriptorBoy/RocketLeagueLauncher) by **LittleScriptorBoy**. All credit for the original concept and authentication flow goes to him. This version aims to make the tool more accessible by removing the need for any dependencies like Python.

## Key Benefits

* **Launch via Steam**: Easily add the launcher to Steam to get the full benefits of the Steam Overlay, including FPS counter, web browser, and controller support.
* **Skip the Epic Launcher**: You no longer need to have the Epic Games Launcher running to play Rocket League.
* **Simple Multi-Account Support**: Easily switch between multiple Epic Games accounts.
* **No Dependencies**: No need to install Python or any other libraries. Just download and run a single executable.

## Installation

1.  Go to the [**Releases** page](https://github.com/YOUR_USERNAME/YOUR_REPOSITORY/releases) on the right-hand sidebar of this GitHub repository.
2.  Download the correct file for your operating system:
    * For Windows: `RocketLeagueLauncher.exe`
    * For Linux: `RocketLeagueLauncher`
3.  Place the downloaded file in a folder where you want to keep it.

## First-Time Setup

The setup process is now fully interactive and only needs to be done once per account.

1.  **Run the Launcher**: Double-click the `RocketLeagueLauncher.exe` (or run `./RocketLeagueLauncher` on Linux).
2.  **Locate Game File**: A file selection window will open. Navigate to your Rocket League installation folder and select `RocketLeague.exe` (or the Linux equivalent). This should be located in the "Binaries/Win64/" path in your installation folder.
3.  **Log In to Epic Games**: Your web browser will open to the Epic Games login page. Log in to the account you want to use.
4.  **Get Authorization Code**: After logging in, you will be redirected to a page with a long `authorizationCode`. Copy this 32-character code.
5.  **Enter Code**: A dialog box will appear from the launcher. Paste the `authorizationCode` into the box and click OK.

The game will now launch. A `config.json` file will be created in the same folder as the launcher. This file stores your game path and session token so you don't have to repeat this process.

## How to Add to Steam

1.  In your Steam Library, click **Add a Game** -> **Add a Non-Steam Game...**
2.  Click **Browse...** and navigate to the folder where you saved `RocketLeagueLauncher.exe` and select it.
3.  Click **Add Selected Programs**.

You can now launch the game directly from your Steam library!

## Using Multiple Accounts

If you want to use a second account:

1.  Create a **new, separate folder** for your alternate account.
2.  Copy the `RocketLeagueLauncher.exe` into this new folder.
3.  Run it from the new folder and complete the first-time setup for your alternate account.

The launcher will create a separate `config.json` in each folder, keeping your accounts isolated. You can create a separate Steam library entry for each account's launcher.

## Troubleshooting

* **Authentication Error / Expired Session**: If you get an error message, the simplest fix is to **delete the `config.json` file** and run the launcher again. This will restart the setup process and allow you to get a fresh login token.
* **Wrong Game Path**: If you accidentally selected the wrong game path, simply delete `config.json` to be prompted for the correct path on the next run.