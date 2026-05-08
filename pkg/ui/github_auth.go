package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// GitHubAuth handles GitHub OAuth authentication
type GitHubAuth struct {
	app         fyne.App
	window      fyne.Window
	statusLabel *widget.Label
	callback    func(*GitHubUser, error)
}

// GitHubOAuthConfig holds GitHub OAuth configuration
type GitHubOAuthConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	Scopes       string `json:"scopes"`
}

// NewGitHubAuth creates a new GitHub authentication handler
func NewGitHubAuth(app fyne.App, callback func(*GitHubUser, error)) *GitHubAuth {
	return &GitHubAuth{
		app:      app,
		callback: callback,
	}
}

// Authenticate starts the GitHub OAuth flow
func (g *GitHubAuth) Authenticate() {
	// GitHub OAuth configuration (in production, these should be loaded from secure config)
	config := &GitHubOAuthConfig{
		ClientID:     "your_github_client_id",     // Replace with actual client ID
		ClientSecret: "your_github_client_secret", // Replace with actual client secret
		RedirectURI:  "http://localhost:8080/callback",
		Scopes:       "user:email",
	}

	// Create OAuth URL
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=%s",
		url.QueryEscape(config.ClientID),
		url.QueryEscape(config.RedirectURI),
		url.QueryEscape(config.Scopes),
	)

	fmt.Println(" Opening GitHub authentication in browser...")
	fmt.Println(" Please complete authentication in your browser")
	fmt.Println(" Waiting for callback...")

	// Open GitHub in browser
	err := openBrowser(authURL)
	if err != nil {
		g.updateStatus("Failed to open browser: " + err.Error())
		return
	}

	g.updateStatus("Waiting for GitHub authentication...")

	// Start HTTP server to handle callback
	go g.startCallbackServer(config)
}

// openBrowser opens GitHub authorization URL in the default browser
func openBrowser(url string) error {
	switch {
	case os.Getenv("DISPLAY") != "":
		// Linux with display
		return exec.Command("xdg-open", url).Start()
	case os.Getenv("WSL_DISTRO_NAME") != "":
		// WSL
		return exec.Command("explorer.exe", url).Start()
	default:
		// Try other methods
		return exec.Command("open", url).Start()
	}
}

// startCallbackServer starts a local HTTP server to handle OAuth callback
func (g *GitHubAuth) startCallbackServer(config *GitHubOAuthConfig) {
	server := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		g.handleCallback(w, r, config)
	})

	go func() {
		if err := server.ListenAndServe(); err != nil {
			g.updateStatus("Server error: " + err.Error())
		}
	}()
}

// handleCallback processes the OAuth callback from GitHub
func (g *GitHubAuth) handleCallback(w http.ResponseWriter, r *http.Request, config *GitHubOAuthConfig) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for access token
	token, err := g.exchangeCodeForToken(code, config)
	if err != nil {
		http.Error(w, "Failed to exchange code for token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user information
	user, err := g.getGitHubUser(token)
	if err != nil {
		http.Error(w, "Failed to get user info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Success response
	g.callback(user, nil)

	// Show success message
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<html>
		<head><title>Authentication Successful</title></head>
		<body>
			<h1>Authentication Successful!</h1>
			<p>You can close this window and return to TesselBox.</p>
			<script>
				setTimeout(function() {
					window.close();
				}, 2000);
			</script>
		</body>
		</html>
	`)
}

// exchangeCodeForToken exchanges authorization code for access token
func (g *GitHubAuth) exchangeCodeForToken(code string, config *GitHubOAuthConfig) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	data := url.Values{
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
		"code":          {code},
	}

	resp, err := client.PostForm("https://github.com/login/oauth/access_token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.AccessToken, nil
}

// getGitHubUser gets user information using the access token
func (g *GitHubAuth) getGitHubUser(token string) (*GitHubUser, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("User-Agent", "TesselBox")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// updateStatus updates the status label
func (g *GitHubAuth) updateStatus(message string) {
	if g.statusLabel != nil {
		g.statusLabel.SetText(message)
	}
}
