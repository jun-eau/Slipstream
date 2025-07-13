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
	"time" // Added for BakkesMod launch delay

	"github.com/google/uuid"
	"github.com/ncruces/zenity"
)

// --- Constants ---
const (
	currentVersion = "v1.5.1"

	// API Configuration
	epicAPIURL       = "https://account-public-service-prod.ak.epicgames.com/account/api"
	epicLauncherAuth = "basic MzRhMDJjZjhmNDQxNGUyOWIxNTkyMTg3NmRhMzZmOWE6ZGFhZmJjY2M3Mzc3NDUwMzlkZmZlNTNkOTRmYzc2Y2Y="
	// This URL structure forces the login prompt, even if the user is already logged in on their browser.
	epicLoginRedirect = "https://www.epicgames.com/id/login?redirectUrl=https%3A//www.epicgames.com/id/api/redirect%3FclientId%3D34a02cf8f4414e29b15921876da36f9a%26responseType%3Dcode"
	tokenPath         = "/oauth/token"
	exchangePath      = "/oauth/exchange"

	// Local file configuration
	configFileName = "config.json"
)

// --- Data Structures ---

// Config holds all application settings.
type Config struct {
	RocketLeaguePath       string `json:"rocket_league_path"`
	EpicToken              string `json:"epic_token,omitempty"`
	BakkesModEnabled       bool   `json:"bakkesmod_enabled"` // No omitempty, so it defaults to false in JSON
	BakkesModPath          string `json:"bakkesmod_path,omitempty"`
	BakkesModLaunchDelay   int    `json:"bakkesmod_launch_delay,omitempty"`
	BakkesModSetupDeclined bool   `json:"bakkesmod_setup_declined"` // No omitempty, so it defaults to false
	LastNotifiedVersion    string `json:"last_notified_version,omitempty"`
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
	// Updated to pass the full cfg object
	if err := launchGame(cfg, creds, os.Args[1:]); err != nil {
		detailedMsg := "Failed to Launch Rocket League.\n\n" +
			"Please ensure the Rocket League path is correctly set in 'config.json' and that the game executable is not missing or corrupted.\n\n" +
			"Details: " + err.Error()
		showError("Failed to Launch Rocket League", detailedMsg)
		return
	}

	log.Println("Game process started successfully.")

	// 5. Check for updates in the background.
	// Pass a pointer to cfg so the goroutine can modify it
	go checkForUpdates(&cfg)

}

// --- Update Checker ---

// GitHubRelease represents the structure of a release from the GitHub API.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// isNewerVersion compares two version strings (e.g., "v1.5.1", "v1.6.0").
// It returns true if the latest version is newer than the current version.
func isNewerVersion(current, latest string) bool {
	// Simple string comparison works for "vX.Y.Z" format
	return strings.TrimPrefix(latest, "v") > strings.TrimPrefix(current, "v")
}

// checkForUpdates fetches the latest release from GitHub and notifies the user if it's a new version.
// It runs in a goroutine to avoid blocking the main application flow.
func checkForUpdates(cfg *Config) {
	log.Println("Checking for application updates...")
	// Use a longer timeout for the update check
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/jun-eau/Slipstream/releases/latest")
	if err != nil {
		log.Printf("Update check failed (network error): %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Update check failed (status code: %d)", resp.StatusCode)
		return
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		log.Printf("Update check failed (JSON parsing error): %v", err)
		return
	}

	latestVersion := release.TagName
	log.Printf("Current version: %s, Latest version: %s", currentVersion, latestVersion)

	if isNewerVersion(currentVersion, latestVersion) {
		log.Printf("A new version is available: %s", latestVersion)
		// Check if we've already notified the user about this specific version
		if cfg.LastNotifiedVersion != latestVersion {
			log.Println("Notifying user about the new version.")
			// Use a separate function to show the dialog to keep this clean
			showUpdateNotification(latestVersion)

			// Update the config and save it
			cfg.LastNotifiedVersion = latestVersion
			if err := saveConfig(*cfg); err != nil {
				log.Printf("Warning: failed to save last notified version: %v", err)
			}
		} else {
			log.Printf("Already notified user about version %s. Skipping.", latestVersion)
		}
	} else {
		log.Println("Application is up to date.")
	}
}

// showUpdateNotification displays the update dialog to the user.
func showUpdateNotification(version string) {
	message := fmt.Sprintf(
		"A new version of Slipstream is available!\n\n"+
			"You are on version: %s\n"+
			"The latest version is: %s\n\n"+
			"You can download the new version from the releases page.",
		currentVersion, version,
	)
	// We use a goroutine for the dialog itself to prevent any potential blocking
	// on the main update goroutine, although it's generally not an issue with zenity.
	go func() {
		err := zenity.Info(message,
			zenity.Title("Update Available"),
			zenity.ExtraButton("Open Download Page"),
			zenity.InfoIcon,
		)
		if err == zenity.ErrExtraButton {
			openBrowser("https://github.com/jun-eau/Slipstream/releases/latest")
		}
	}()
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

// launchGame starts the game with the provided credentials and arguments.
// On Linux, it detects if the target is a Windows executable and guides the user.
// The function signature now takes the full Config object.
func launchGame(cfg Config, creds LaunchCredentials, extraArgs []string) error {
	// 1. Preserve existing Linux .exe check.
	// This ensures the "Setup Complete" message appears correctly on Linux
	// before any launch attempt is made.
	if runtime.GOOS == "linux" && strings.HasSuffix(strings.ToLower(cfg.RocketLeaguePath), ".exe") {
		showInfo("Setup Complete!",
			"Your configuration and login token have been successfully saved to 'config.json'.\n\n"+
				"To play, please add 'Slipstream.exe' (the Windows version) to Steam or Lutris and run it using Proton or Wine. "+
				"It will use the config file you just created.")
		return nil // Expected outcome on Linux with .exe path
	}

	// 2. Launch Rocket League (asynchronously).
	log.Println("Launching Rocket League...")
	rlArgs := []string{
		"-AUTH_LOGIN=unused",
		"-AUTH_PASSWORD=" + creds.ExchangeCode,
		"-AUTH_TYPE=exchangecode",
		"-epicapp=Sugar",
		"-epicenv=Prod",
		"-EpicPortal",
		"-epicusername=\"\"", // Intentionally empty as per original args
		"-epicuserid=" + creds.AccountID,
	}
	rlArgs = append(rlArgs, extraArgs...)
	rlCmd := exec.Command(cfg.RocketLeaguePath, rlArgs...)

	if err := rlCmd.Start(); err != nil {
		return fmt.Errorf("failed to start Rocket League at %s: %w", cfg.RocketLeaguePath, err)
	}
	log.Println("Rocket League process started.")

	// 3. Conditional BakkesMod Launch.
	if !cfg.BakkesModEnabled || cfg.BakkesModPath == "" {
		log.Println("BakkesMod is not enabled or path is not set. Launch complete.")
		return nil // Standard launch finished successfully.
	}

	// 4. Execute BakkesMod launch sequence.
	// We need to import "time" for this. Ensure it's added if not already.
	// (It should be, due to existing logging, but good to double check)
	delay := time.Duration(cfg.BakkesModLaunchDelay) * time.Second
	log.Printf("BakkesMod is enabled. Waiting for %v before launching...", delay)
	time.Sleep(delay) // Requires "time" package

	log.Println("Launching BakkesMod...")
	bmCmd := exec.Command(cfg.BakkesModPath) // No arguments needed for BakkesMod.exe
	if err := bmCmd.Start(); err != nil {
		// Inform the user but do not treat it as a fatal error for the game itself.
		errorMsg := fmt.Sprintf(
			"Could not start BakkesMod.exe at the specified path:\n\n%s\n\nError: %v\n\nRocket League should still be running.",
			cfg.BakkesModPath, err,
		)
		showError("BakkesMod Launch Failed", errorMsg)
		log.Printf("Error launching BakkesMod: %v", err)
		// Non-fatal error, Rocket League is (presumably) running.
	} else {
		log.Println("BakkesMod process started.")
	}

	return nil // Two-process launch sequence finished (or attempted).
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
		rlPath, err := zenity.SelectFile(
			zenity.Title("Select Rocket League Executable"),
			zenity.FileFilters{
				{Name: "Rocket League Executable", Patterns: []string{"RocketLeague.exe", "RocketLeague"}, CaseFold: true},
				{Name: "All Files", Patterns: []string{"*"}},
			},
		)
		if err != nil {
			return cfg, fmt.Errorf("you must select a Rocket League path to continue")
		}
		cfg.RocketLeaguePath = rlPath
		// Save immediately after getting RL path, so it's there before BM setup
		if err := saveConfig(cfg); err != nil {
			// Log or show error, but try to continue to BM setup if possible
			log.Printf("Warning: could not save Rocket League path: %v", err)
		}
	}

	// BakkesMod Setup Prompt - only if RL path is set and BM not already configured or declined
	if cfg.RocketLeaguePath != "" && cfg.BakkesModPath == "" && !cfg.BakkesModSetupDeclined {
		log.Println("Prompting for BakkesMod setup.")
		err := zenity.Question("Would you like to enable automatic launching for BakkesMod?",
			zenity.Title("BakkesMod Setup"),
			zenity.DefaultCancel(), // Makes "No" the default if user just closes dialog
			zenity.OKLabel("Yes"),
			zenity.CancelLabel("No"),
		)

		if err == nil { // User clicked "Yes"
			log.Println("User opted to set up BakkesMod.")
			bmPath, err := zenity.SelectFile(
				zenity.Title("Select BakkesMod.exe"),
				zenity.FileFilters{
					{Name: "BakkesMod Executable", Patterns: []string{"BakkesMod.exe"}, CaseFold: true},
					{Name: "All Files", Patterns: []string{"*"}},
				},
			)
			if err == nil && bmPath != "" {
				log.Printf("BakkesMod path selected: %s", bmPath)
				cfg.BakkesModPath = bmPath
				cfg.BakkesModEnabled = true
				if cfg.BakkesModLaunchDelay == 0 { // Set default delay if not already set by user
					cfg.BakkesModLaunchDelay = 5
				}
				cfg.BakkesModSetupDeclined = false // Ensure this is false if they just set it up
			} else {
				log.Println("User did not select a BakkesMod path or cancelled.")
				// User cancelled BakkesMod selection, treat as "No" for this session, but don't set Declined.
				// They might want to try again next time.
				// If we want to treat this as a permanent "No", we'd set BakkesModSetupDeclined = true here.
				// For now, let's assume cancelling file dialog means they don't want it *now*.
			}
		} else { // User clicked "No" or closed the dialog
			log.Println("User declined BakkesMod setup.")
			cfg.BakkesModSetupDeclined = true
			cfg.BakkesModEnabled = false // Ensure it's disabled if they decline
		}
		// Save config after BM interaction (or lack thereof)
		if err := saveConfig(cfg); err != nil {
			return cfg, fmt.Errorf("could not save BakkesMod configuration: %w", err)
		}
	}

	// Ensure default launch delay if enabled and not set
	if cfg.BakkesModEnabled && cfg.BakkesModLaunchDelay == 0 {
		log.Println("BakkesMod enabled but launch delay is 0, setting to default (5s).")
		cfg.BakkesModLaunchDelay = 5
		if err := saveConfig(cfg); err != nil {
			// Log this, but it's not critical enough to halt the app
			log.Printf("Warning: could not save default BakkesMod launch delay: %v", err)
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
