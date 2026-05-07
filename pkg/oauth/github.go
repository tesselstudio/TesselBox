package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubUser represents a GitHub user profile
type GitHubUser struct {
	ID       int    `json:"id"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar_url"`
	Location string `json:"location"`
	Bio      string `json:"bio"`
}

// GitHubTokenResponse represents the response from GitHub's token endpoint
type GitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// GitHubOAuth handles GitHub OAuth authentication
type GitHubOAuth struct {
	config *Config
}

// NewGitHubOAuth creates a new GitHub OAuth handler
func NewGitHubOAuth(config *Config) *GitHubOAuth {
	return &GitHubOAuth{config: config}
}

// GetAuthURL returns the GitHub authorization URL
func (g *GitHubOAuth) GetAuthURL(state string) string {
	params := url.Values{}
	params.Add("client_id", g.config.GitHubClientID)
	params.Add("redirect_uri", g.config.GitHubRedirectURI)
	params.Add("scope", "user:email")
	params.Add("state", state)
	params.Add("response_type", "code")

	return fmt.Sprintf("https://github.com/login/oauth/authorize?%s", params.Encode())
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (g *GitHubOAuth) ExchangeCodeForToken(code string) (*GitHubTokenResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	
	data := url.Values{}
	data.Set("client_id", g.config.GitHubClientID)
	data.Set("client_secret", g.config.GitHubClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", g.config.GitHubRedirectURI)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp GitHubTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserProfile fetches the user's GitHub profile
func (g *GitHubOAuth) GetUserProfile(accessToken string) (*GitHubUser, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s"))
	req.Header.Set("User-Agent", "TesselBox")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("profile fetch failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user profile: %w", err)
	}

	// Fetch user email if not public
	if user.Email == "" {
		email, err := g.getUserEmail(accessToken)
		if err == nil {
			user.Email = email
		}
	}

	return &user, nil
}

// getUserEmail fetches the user's primary email from GitHub
func (g *GitHubOAuth) getUserEmail(accessToken string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s"))
	req.Header.Set("User-Agent", "TesselBox")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch user emails: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("email fetch failed: status %d", resp.StatusCode)
	}

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", fmt.Errorf("failed to decode emails: %w", err)
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	return "", fmt.Errorf("no primary email found")
}
