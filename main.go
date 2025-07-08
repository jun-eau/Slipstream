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
	// API Configuration
	epicAPIURL        = "https://account-public-service-prod.ak.epicgames.com/account/api"
	epicLauncherAuth  = "basic MzRhMDJjZjhmNDQxNGUyOWIxNTkyMTg3NmRhMzZmOWE6ZGFhZmJjY2M3Mzc3NDUwMzlkZmZlNTNkOTRmYzc2Y2Y="
	epicLoginRedirect = "https://www.epicgames.com/id/api/redirect?clientId=34a02cf8f4414e29b15921876da36f9a&responseType=code&prompt=login"
	tokenPath         = "/oauth/token"
	exchangePath      = "/oauth/exchange"

	// Local file configuration
	configFileName = "config.json"
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
	// Initialize file logging
	logFile, err := os.OpenFile(filepath.Join(getExecutableDir(), "slipstream.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: Failed to open log file: %v", err)
	} else {
		defer logFile.Close()
		// Create a multi-writer to write to both stdout and the log file
		mw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(mw)
	}

	// 1. Load configuration.
	cfg, err := loadConfig()
	if err != nil {
		detailedMsg := "Failed to load configuration.\n\n" +
			"Please ensure that the program has permissions to read and write 'config.json' in its directory, and that the file is not corrupted.\n" +
			"If the problem persists, you can try deleting 'config.json' and the program will attempt to recreate it.\n\n" +
			"Details: " + err.Error()
		showError("Configuration Error", detailedMsg)
		return
	}

	// 2. Authenticate with Epic Games to get launch credentials.
	auth := NewAuthenticator()
	creds, newEpicToken, err := auth.GetLaunchCredentials(cfg.EpicToken)
	if err != nil {
		detailedMsg := "Authentication Failed.\n\n" +
			"Your session may have expired or the authentication details are incorrect. The simplest fix is often to delete the 'config.json' file and run Slipstream again to log in from scratch.\n\n" +
			"Details: " + err.Error()
		showError("Authentication Failed", detailedMsg)
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

	// 4. Launch Rocket League with the obtained credentials and any extra args.
	log.Println("Successfully authenticated. Launching Rocket League...")
	// os.Args[0] is the program name, os.Args[1:] is all subsequent arguments.
	if err := launchGame(cfg.RocketLeaguePath, creds, os.Args[1:]); err != nil {
		detailedMsg := "Failed to Launch Rocket League.\n\n" +
			"Please ensure the Rocket League path is correctly set in 'config.json' and that the game executable is not missing or corrupted.\n\n" +
			"Details: " + err.Error()
		showError("Failed to Launch Rocket League", detailedMsg)
		return
	}

	log.Println("Game process started successfully.")
}

// --- Core Functions ---

// GetLaunchCredentials orchestrates the entire authentication flow.
// It takes the current token and returns the launch credentials and the new token to be saved.
func (a *Authenticator) GetLaunchCredentials(currentToken string) (LaunchCredentials, string, error) {
	var creds LaunchCredentials
	var newRefreshToken string // This will store the latest refresh token to be saved.
	var currentRefreshToken string

	tokenFromConfig := strings.TrimSpace(currentToken)

	if tokenFromConfig == "" {
		// No token in config, perform first-time setup to get a refresh token.
		log.Println("No token found in config, performing first-time setup...")
		initialRefreshToken, err := a.performFirstTimeSetup() // This now returns a refresh token
		if err != nil {
			return creds, "", fmt.Errorf("initial setup failed: %w", err)
		}
		currentRefreshToken = initialRefreshToken
		// This initial refresh token will also be the newRefreshToken to be saved by main().
		newRefreshToken = currentRefreshToken
	} else {
		// Token found in config, assume it's a refresh token.
		currentRefreshToken = tokenFromConfig
	}

	// Use the current refresh token (either from config or first-time setup) to get a new access token.
	log.Println("Acquiring new access token using refresh token...")
	tokenResp, err := a.exchangeRefreshToken(currentRefreshToken)
	if err != nil {
		// If exchanging refresh token fails, it might be expired.
		// Try to perform first-time setup again as a recovery mechanism.
		log.Printf("Failed to exchange refresh token (%v). Attempting first-time setup again.", err)
		recoveredRefreshToken, setupErr := a.performFirstTimeSetup()
		if setupErr != nil {
			return creds, "", fmt.Errorf("failed to exchange refresh token and subsequent first-time setup also failed: %w (original error: %v)", setupErr, err)
		}
		log.Println("Successfully obtained a new refresh token via recovery setup.")
		currentRefreshToken = recoveredRefreshToken
		newRefreshToken = currentRefreshToken // This is the new token to save.

		// Retry exchanging the newly obtained refresh token for an access token.
		log.Println("Retrying: Acquiring new access token with newly recovered refresh token...")
		tokenResp, err = a.exchangeRefreshToken(currentRefreshToken)
		if err != nil {
			return creds, "", fmt.Errorf("could not get access token even after recovery via first-time setup: %w", err)
		}
	}
	// Always update newRefreshToken with the latest one from the exchange, as it might have been rotated.
	newRefreshToken = tokenResp.RefreshToken

	// Finally, exchange the access token for the game launch code.
	log.Println("Acquiring game launch exchange code...")
	exchangeResp, err := a.getExchangeCode(tokenResp.AccessToken)
	if err != nil {
		return creds, newRefreshToken, fmt.Errorf("could not get game launch code: %w", err)
	}

	creds.ExchangeCode = exchangeResp.Code
	creds.AccountID = tokenResp.AccountID
	return creds, newRefreshToken, nil
}

// The function now accepts extraArgs to pass to the game.
func launchGame(path string, creds LaunchCredentials, extraArgs []string) error {
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

	// Append the extra launch options from Steam.
	args = append(args, extraArgs...)

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

	authCodeStr, err := askForInput("Enter Authorization Code", "Paste the 32-character authorization code here:")
	if err != nil {
		return "", fmt.Errorf("user cancelled input")
	}
	if len(authCodeStr) != 32 {
		return "", fmt.Errorf("invalid authorization code: must be 32 characters long")
	}

	log.Println("Exchanging authorization code for refresh token...")
	resp, err := a.exchangeAuthCode(authCodeStr)
	if err != nil {
		return "", fmt.Errorf("could not exchange authorization code: %w", err)
	}
	if resp.RefreshToken == "" {
		return "", fmt.Errorf("did not receive a refresh token after exchanging authorization code")
	}
	return resp.RefreshToken, nil
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
	err := a.apiRequest("POST", tokenPath, data, epicLauncherAuth, &resp)
	if err != nil {
		return resp, fmt.Errorf("token request failed: %w", err)
	}
	if resp.ErrorCode != "" {
		return resp, fmt.Errorf("API error: %s", resp.ErrorMessage)
	}
	return resp, nil
}

func (a *Authenticator) getExchangeCode(accessToken string) (apiResponse, error) {
	var resp apiResponse
	authHeader := "bearer " + accessToken
	err := a.apiRequest("GET", exchangePath, nil, authHeader, &resp)
	if err != nil {
		return resp, fmt.Errorf("exchange code request failed: %w", err)
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
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode json response: %w", err)
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
