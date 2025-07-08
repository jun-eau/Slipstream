// main.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/ncruces/zenity"
)

// --- Constants ---
const (
	epicAPIURL        = "https://account-public-service-prod.ak.epicgames.com/account/api"
	epicLauncherAuth  = "basic MzRhMDJjZjhmNDQxNGUyOWIxNTkyMTg3NmRhMzZmOWE6ZGFhZmJjY2M3Mzc3NDUwMzlkZmZlNTNkOTRmYzc2Y2Y="
	epicLoginRedirect = "https://www.epicgames.com/id/api/redirect?clientId=34a02cf8f4414e29b15921876da36f9a&responseType=code&prompt=login"
	configFileName    = "config.json"
)

// --- Data Structures ---

// Config holds all application settings.
type Config struct {
	RocketLeaguePath string `json:"rocket_league_path"`
	EpicToken        string `json:"epic_token,omitempty"`
}

// LaunchCredentials holds the final codes needed to start the game.
type LaunchCredentials struct {
	ExchangeCode string
	AccountID    string
}

// apiResponse is used to decode all token/code responses from the Epic API.
type apiResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	AccountID    string `json:"account_id"`
	Code         string `json:"code"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

// Authenticator handles the Epic Games authentication flow.
type Authenticator struct {
	client *http.Client
}

// NewAuthenticator creates a new authenticator instance.
func NewAuthenticator() *Authenticator {
	return &Authenticator{
		client: &http.Client{},
	}
}

// --- Main Application Logic ---

func main() {
	// 1. Load configuration.
	cfg, err := loadConfig()
	if err != nil {
		showError("Configuration Error", err.Error())
		return
	}

	// 2. Authenticate with Epic Games to get launch credentials.
	auth := NewAuthenticator()
	creds, newEpicToken, err := auth.GetLaunchCredentials(cfg.EpicToken)
	if err != nil {
		showError("Authentication Failed", err.Error())
		return
	}

	// 3. Save the new token if it has changed.
	if newEpicToken != "" && newEpicToken != cfg.EpicToken {
		log.Println("Saving new session token.")
		cfg.EpicToken = newEpicToken
		if err := saveConfig(cfg); err != nil {
			log.Printf("Warning: could not save new session token: %v", err)
		}
	}

	// 4. Launch Rocket League with the obtained credentials.
	log.Println("Successfully authenticated. Launching Rocket League...")
	if err := launchGame(cfg.RocketLeaguePath, creds); err != nil {
		showError("Failed to Launch Rocket League", err.Error())
		return
	}

	log.Println("Game process started successfully.")
}

// --- Core Functions ---

// GetLaunchCredentials orchestrates the entire authentication flow.
// It takes the current token and returns the launch credentials and the new token to be saved.
func (a *Authenticator) GetLaunchCredentials(currentToken string) (LaunchCredentials, string, error) {
	var creds LaunchCredentials
	var newRefreshToken string

	tokenStr := strings.TrimSpace(currentToken)
	if tokenStr == "" {
		// If no token exists, start the first-time setup.
		authCode, err := a.performFirstTimeSetup()
		if err != nil {
			return creds, "", fmt.Errorf("initial setup failed: %w", err)
		}
		tokenStr = authCode
	}

	var refreshToken string
	if len(tokenStr) == 32 {
		// It's an authorization code, exchange it for a refresh token.
		log.Println("Exchanging authorization code for refresh token...")
		resp, err := a.exchangeAuthCode(tokenStr)
		if err != nil {
			return creds, "", err
		}
		refreshToken = resp.RefreshToken
	} else {
		// It's already a refresh token.
		refreshToken = tokenStr
	}

	// Use the refresh token to get a new access token.
	log.Println("Acquiring new access token...")
	tokenResp, err := a.exchangeRefreshToken(refreshToken)
	if err != nil {
		return creds, "", fmt.Errorf("your session may have expired. Please delete '%s' to log in again. Original error: %w", configFileName, err)
	}
	newRefreshToken = tokenResp.RefreshToken // This is the token we want to save.

	// Finally, exchange the access token for the game launch code.
	log.Println("Acquiring game launch exchange code...")
	exchangeResp, err := a.getExchangeCode(tokenResp.AccessToken)
	if err != nil {
		return creds, newRefreshToken, err
	}

	creds.ExchangeCode = exchangeResp.Code
	creds.AccountID = tokenResp.AccountID
	return creds, newRefreshToken, nil
}

func launchGame(path string, creds LaunchCredentials) error {
	args := []string{
		"-AUTH_LOGIN=unused",
		"-AUTH_PASSWORD=" + creds.ExchangeCode,
		"-AUTH_TYPE=exchangecode",
		"-epicapp=Sugar",
		"-epicenv=Prod",
		"-EpicPortal",
		"-epicusername=\"\"",
		"-epicuserid=" + creds.AccountID,
	}

	cmd := exec.Command(path, args...)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start the game executable at:\n%s\n\nError: %w\n\nPlease ensure the path is correct in %s", path, err, configFileName)
	}
	return nil
}

// --- Authentication Steps ---

func (a *Authenticator) performFirstTimeSetup() (string, error) {
	showInfo("Authorization Required", "A browser window will now open. Please log in to your Epic Games account, then copy the 'authorizationCode' value from the page you are redirected to.")
	openBrowser(epicLoginRedirect)

	authCode, err := askForInput("Enter Authorization Code", "Paste the 32-character authorization code here:")
	if err != nil {
		return "", fmt.Errorf("user cancelled input")
	}
	if len(authCode) != 32 {
		return "", fmt.Errorf("invalid authorization code: must be 32 characters long")
	}
	return authCode, nil
}

func (a *Authenticator) exchangeAuthCode(code string) (apiResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	return a.makeTokenRequest(data)
}

func (a *Authenticator) exchangeRefreshToken(token string) (apiResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token)
	data.Set("token_type", "eg1")
	return a.makeTokenRequest(data)
}

func (a *Authenticator) makeTokenRequest(data url.Values) (apiResponse, error) {
	var resp apiResponse
	err := a.apiRequest("POST", "/oauth/token", data, epicLauncherAuth, &resp)
	if err != nil {
		return resp, err
	}
	if resp.ErrorCode != "" {
		return resp, fmt.Errorf("API error: %s", resp.ErrorMessage)
	}
	return resp, nil
}

func (a *Authenticator) getExchangeCode(accessToken string) (apiResponse, error) {
	var resp apiResponse
	authHeader := "bearer " + accessToken
	err := a.apiRequest("GET", "/oauth/exchange", nil, authHeader, &resp)
	if err != nil {
		return resp, err
	}
	if resp.ErrorCode != "" {
		return resp, fmt.Errorf("API error: %s", resp.ErrorMessage)
	}
	return resp, nil
}

// --- Generic API Request Helper ---

func (a *Authenticator) apiRequest(method, path string, data url.Values, authHeader string, target interface{}) error {
	var reqBody io.Reader
	if data != nil {
		reqBody = strings.NewReader(data.Encode())
	}

	req, err := http.NewRequest(method, epicAPIURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("User-Agent", "UELauncher/16.12.1-36115220+++Portal+Release-Live")
	req.Header.Set("X-Epic-Correlation-ID", "UE4-"+strings.ToUpper(uuid.New().String()))

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// --- Configuration Helpers ---

func loadConfig() (Config, error) {
	var cfg Config
	path := filepath.Join(getExecutableDir(), configFileName)

	file, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(file, &cfg)
	}

	// If the path is missing, always prompt for it.
	if cfg.RocketLeaguePath == "" {
		showInfo("Rocket League Path Setup", "Please locate your Rocket League executable (e.g., RocketLeague.exe). This will only be asked once.")
		selectedPath, err := zenity.SelectFile(
			zenity.Title("Select Rocket League Executable"),
			zenity.FileFilters{
				{Name: "Rocket League Executable", Patterns: []string{"RocketLeague.exe", "RocketLeague"}, CaseFold: true},
				{Name: "All Files", Patterns: []string{"*"}},
			},
		)
		if err != nil {
			return cfg, fmt.Errorf("you must select a path to continue")
		}
		cfg.RocketLeaguePath = selectedPath
		if err := saveConfig(cfg); err != nil {
			return cfg, fmt.Errorf("could not save the configuration file: %w", err)
		}
	}
	return cfg, nil
}

func saveConfig(cfg Config) error {
	path := filepath.Join(getExecutableDir(), configFileName)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// --- Utility Helpers ---

func getExecutableDir() string {
	ex, err := os.Executable()
	if err != nil {
		dir, _ := os.Getwd()
		return dir
	}
	return filepath.Dir(ex)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		log.Printf("Failed to open browser: %v", err)
	}
}

// --- GUI Dialog Functions ---

func showError(title, message string) {
	log.Printf("ERROR: %s - %s", title, message)
	zenity.Error(message, zenity.Title(title), zenity.ErrorIcon)
}

func showInfo(title, message string) {
	log.Printf("INFO: %s - %s", title, message)
	zenity.Info(message, zenity.Title(title), zenity.InfoIcon)
}

func askForInput(title, message string) (string, error) {
	log.Printf("PROMPT: %s", title)
	return zenity.Entry(message, zenity.Title(title))
}
