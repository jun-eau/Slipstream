import os, argparse, requests, uuid, subprocess, webbrowser
import tkinter as tk
from tkinter import simpledialog, messagebox

##########################

rlpath = 'C:\\Epic Games\\rocketleague\\Binaries\\Win64\\RocketLeague.exe'
envfile = '.epicenv'

# Configure this if you want to proxy the traffic
#proxy = {'http': '127.0.0.1:8888', 'https': '127.0.0.1:8888'}
#verify = False
proxy = {}
verify = True

##########################

class MyDialog(simpledialog.Dialog):
	def body(self, master):
		tk.Label(master, text="Enter authorization code:", anchor="w").pack(fill="x")
		self.text = tk.Text(master, width=40, height=1)
		self.text.pack()
		return self.text

	def buttonbox(self):
		box = tk.Frame(self)
		tk.Button(box, text="OK", width=10, command=self.ok, default=tk.ACTIVE)\
			.pack(side=tk.LEFT, padx=5, pady=5)
		box.pack()

	def apply(self):
		self.result = self.text.get("1.0", "end-1c")


def get_authorization_code():
	webbrowser.open('https://www.epicgames.com/id/login?redirectUrl=https%3A//www.epicgames.com/id/api/redirect%3FclientId%3D34a02cf8f4414e29b15921876da36f9a%26responseType%3Dcode%26prompt%3Dlogin%26', new=0, autoraise=True)
	root = tk.Tk()
	root.withdraw()
	dlg = MyDialog(root, title="Enter authorization code")
	code = dlg.result
	root.destroy()
	return code.strip()

##########################

# Perform an API request using the Epic Launcher authorization credentials
def api_request(method='post', path='/oauth/token', data='', auth='basic MzRhMDJjZjhmNDQxNGUyOWIxNTkyMTg3NmRhMzZmOWE6ZGFhZmJjY2M3Mzc3NDUwMzlkZmZlNTNkOTRmYzc2Y2Y='):
	epic_api_url = 'https://account-public-service-prod.ak.epicgames.com/account/api'

	if method == 'post':
		req = requests.post(epic_api_url + path, headers={
			'Accept': '*/*',
			'Accept-Encoding': 'deflate, gzip',
			'X-Epic-Correlation-ID': 'UE4-5dea7166457d308530e85a9b3333ff78-C9AD1AA540580FDB3CCEA5B0A22B3218-' + str(uuid.uuid4()).replace('-', '').upper(),
			'User-Agent': 'UELauncher/16.12.1-36115220+++Portal+Release-Live Windows/10.0.22631.1.768.64bit',
			'Content-Type': 'application/x-www-form-urlencoded',
			'Authorization': auth
		}, data=data, proxies=proxy, verify=verify)
	elif method == 'get':
		req = requests.get(epic_api_url + path, headers={
			'Accept': '*/*',
			'Accept-Encoding': 'deflate, gzip',
			'X-Epic-Correlation-ID': 'UE4-5dea7166457d308530e85a9b3333ff78-C9AD1AA540580FDB3CCEA5B0A22B3218-' + str(uuid.uuid4()).replace('-', '').upper(),
			'User-Agent': 'UELauncher/16.12.1-36115220+++Portal+Release-Live Windows/10.0.22631.1.768.64bit',
			'Authorization': auth
		}, proxies=proxy, verify=verify)

	return req

##########################

# Step 1 (browser) - GET https://www.epicgames.com/id/api/redirect?clientId=34a02cf8f4414e29b15921876da36f9a&responseType=code&prompt=login&
# Step 2 (get auth & refresh codes) - POST https://account-public-service-prod.ak.epicgames.com/account/api/oauth/token - grant_type=authorization_code&code=<code>
# Step 3 (convert refresh code to eg1 refresh & access token) - POST https://account-public-service-prod.ak.epicgames.com/account/api/oauth/token - grant_type=refresh_token&refresh_token=<refresh>&token_type=eg1
# Step 4 (use access token to get an exchange code) - GET https://account-public-service-prod.ak.epicgames.com/account/api/oauth/exchange - Authorization: bearer <access_token>
# Step 5 (launch game with -AUTH_PASSWORD=<exchangecode>)

if __name__ == '__main__':
	parser = argparse.ArgumentParser()
	parser.add_argument('-f', '--envfile', type=str, default=envfile, help='Credential environment file where your auth code or refresh token is store, which must be in the same directory as this script')
	parser.add_argument('-p', '--rlpath', type=str, default=rlpath, help='Path to your Rocket League executable')
	args = parser.parse_args()
	envfile = args.envfile
	rlpath = args.rlpath

	epicenv = ''

	# Get existing auth code or refresh token, or if none then open a browser to login and retrieve an auth code
	try:
		with open(f'{os.getcwd()}\\{envfile}', 'r') as f:
			epicenv = f.read().strip()
	except:
		pass

	if len(epicenv) == 0:
		messagebox.showwarning('Invalid credential file', f'Either {os.getcwd()}\\{envfile} was empty or was not found\r\n\r\n!!!!\r\n\r\nWe will now open a new browser window...please login and then copy your authorization code!')
		epicenv = get_authorization_code()
		if len(epicenv) == 32:
			with open(f'{os.getcwd()}\\{envfile}', 'w') as f:
				f.write(epicenv)
		else:
			messagebox.showerror('Invalid authorization code', 'Invalid authorization code, please run the script again and provide the correct code')
			exit()

	auth_code = ''
	refresh_token = ''
	if len(epicenv) == 0:
		messagebox.showerror('Invalid file', f'Your {os.getcwd()}\\{envfile} file does not contain anything! Please make sure this file exists and contains your authorization code!')
		exit()
	elif len(epicenv) == 32:
		auth_code = epicenv
	else:
		refresh_token = epicenv

	# If we only have an auth code let's convert it to an initial generic refresh code
	if auth_code:
		req = api_request(method='post', path='/oauth/token', data='grant_type=authorization_code&code=' + auth_code)
		refresh_token = req.json()['refresh_token']

	# Get a new eg1 access token and refresh code
	req = api_request(method='post', path='/oauth/token', data='grant_type=refresh_token&refresh_token=' + refresh_token + '&token_type=eg1')

	response = req.json()
	if 'errorMessage' in response:
		messagebox.showerror(response['errorCode'], response['errorMessage'])
		exit()

	access_token = response['access_token']
	refresh_token = response['refresh_token']
	account_id = response['account_id']

	# Save our refresh code for next time
	with open(f'{os.getcwd()}\\{envfile}', 'w') as f:
		f.write(refresh_token)

	# Exchange our access token for a launcher code
	req = api_request(method='get', path='/oauth/exchange', auth='bearer ' + access_token)

	response = req.json()
	if 'errorMessage' in response:
		messagebox.showerror(response['errorCode'], response['errorMessage'])
		exit()

	code = response['code']

	# Launch the game using our launcher exchange code!
	subprocess.Popen([rlpath, '-AUTH_LOGIN=unused', '-AUTH_PASSWORD=' + code, '-AUTH_TYPE=exchangecode', '-epicapp=Sugar', '-epicenv=Prod', '-EpicPortal', '-epicusername=""', '-epicuserid=' + account_id])
