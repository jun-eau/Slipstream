# RocketLeagueLauncher

This is a Python script which will launch the Epic Games version of Rocket League. 

Benefits of using this are:
* Launch Rocket League via Steam to have all of the benefits of the Steam overlay, recording, FPS counter etc. 
* You no longer need to run the Epic Games Launcher at all
* You can have multiple accounts configured and easily launch them without having to sign out and back in to a different account via Epic Games

# Requirements

* Python 3.7+
* Epic Games version of Rocket League
* Steam

To use it, you'll need a valid authorization code. Full instructions on how to configure this script are as follows:

# Configure the script

1. Login to your Epic Games account at https://www.epicgames.com/id/login
2. Once logged in, go to https://www.epicgames.com/id/api/redirect?clientId=34a02cf8f4414e29b15921876da36f9a&responseType=code&prompt=login& and copy the `authorizationCode` value (e.g. `33c56a17870a110ea5955c133f5e64c2`)

**You must now do the following steps quickly because the authorizationCode is only valid for a couple of minutes max**:

3. Edit the `.epicenv` file - paste the `authorizationCode` value into it and save it
4. Edit `rocketleague_launcher.py` and modify the `rlpath` variable to point to wherever your Rocket League executable is
5. Run the script so it gets an initial refresh code and you can test that it's working (`py rocketleague_launcher.py`) - this should run Rocket League and it should be signed into your account, if not then you did not configure something correctly in steps 1-4

# Configure Steam

1. In your Steam Library click on **Add a Game** -> **Add a Non-Steam Game...**
2. Wait for it to finish loading and then click on **Browse...** and find your Rocket League executable
3. Once created, edit the game you just added by right clicking and choosing **Properties...**
4. Change the `Target` to your Python's `pythonw.exe` path with `rocketleague_launcher.py` as the argument (e.g. `"C:\Python312\pythonw.exe" rocketleague_launcher.py`) - NOTE: Using `pythonw.exe` means Python will launch with no visible windows, which is what you want since Steam will assume the window is the game and try to inject the overlay into it
5. Change the `Start In` value to wherever you saved `rocketleague_launcher.py` (e.g. `C:\Users\yourname\Documents\RocketLeagueLauncher\rocketleague_launcher.py`)

Now you're done and can begin launching Rocket League via Steam!

# How to use multiple accounts

1. Copy `rocketleague_launcher.py` to something like `rocketleague_launcher_alt.py`
2. Modify the new copy of the script and change the `envfile` variable value to something like `.epicenv_alt`
3. Follow steps 1-3 of **Configure the script** but instead of modifying `.epicenv` you will be modifying whatever name you specified above (e.g.`.epicenv_alt`)
4. Follow steps 1-5 of **Configure Steam** but using `rocketleague_launcher_alt.py` (or whatever name you used when you copied the script)

You can do this for as many accounts as you want. Just make sure you copy the script to a new name and configure it with a new `envfile` path to use.

##

# Troubleshooting

If it works to begin with and then stops working, it's likely that your refresh code has expired. Refresh codes will last for 23 days, so if you don't launch the script for longer than that then it will expire. It's also possible for them to be invalidated by Epic Games. Just follow steps 1-3 of **Configure the script** and it will work again.

# How does it work?

The script makes use of Epic's Epic Launcher credentials to make API requests to the Epic Games OAuth backend. We first need an Epic Games launcher authorization code which we can get via a browser. Once we have this code, we're allowed to perform API requests as the Epic Games Launcher so long as we're also using valid HTTP Basic Auth credentials associated with the Epic Games Launcher. We can then ask the API to give us an initial, generic OAuth refresh code which we'll then exchange for an `eg1` access code and refresh code. This means we now have game launching permissions, but to launch a game we'll need to exchange this access code for a launcher code. Finally, once we have the launcher code we can start the game by providing it with this launcher code, and the game will verify this code with Epic Games to obtain all of your account details and authorize you into the game itself. 
