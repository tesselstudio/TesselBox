package oauth

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
)

// AuthServer handles OAuth authentication endpoints
type AuthServer struct {
	config      *Config
	githubOAuth *GitHubOAuth
	// In production, use Redis or database for session storage
	sessions map[string]string      // state -> userID
	users    map[string]*GitHubUser // userID -> user
}

// NewAuthServer creates a new OAuth authentication server
func NewAuthServer(config *Config) *AuthServer {
	return &AuthServer{
		config:      config,
		githubOAuth: NewGitHubOAuth(config),
		sessions:    make(map[string]string),
		users:       make(map[string]*GitHubUser),
	}
}

// StartServer starts the OAuth server
func (s *AuthServer) StartServer() error {
	mux := http.NewServeMux()

	// OAuth endpoints
	mux.HandleFunc("/auth/github/login", s.handleGitHubLogin)
	mux.HandleFunc("/auth/github/callback", s.handleGitHubCallback)

	// API endpoints
	mux.HandleFunc("/api/user", s.handleUserProfile)
	mux.HandleFunc("/api/logout", s.handleLogout)

	// Static files for OAuth success/failure pages
	mux.HandleFunc("/auth/success", s.handleAuthSuccess)
	mux.HandleFunc("/auth/error", s.handleAuthError)

	addr := fmt.Sprintf("%s:%s", s.config.ServerHost, s.config.ServerPort)
	log.Printf("OAuth server starting on %s", addr)

	return http.ListenAndServe(addr, mux)
}

// handleGitHubLogin initiates GitHub OAuth flow
func (s *AuthServer) handleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state for CSRF protection
	state, err := GenerateState()
	if err != nil {
		http.Error(w, "Failed to generate state", http.StatusInternalServerError)
		return
	}

	// Store state (in production, use Redis with TTL)
	s.sessions[state] = "pending"

	// Redirect to GitHub
	authURL := s.githubOAuth.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleGitHubCallback handles GitHub OAuth callback
func (s *AuthServer) handleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		log.Printf("GitHub OAuth error: %s", errorParam)
		http.Redirect(w, r, "/auth/error?error=oauth_failed", http.StatusFound)
		return
	}

	// Validate state
	if _, exists := s.sessions[state]; !exists {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	tokenResp, err := s.githubOAuth.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("Failed to exchange code: %v", err)
		http.Redirect(w, r, "/auth/error?error=token_exchange_failed", http.StatusFound)
		return
	}

	// Get user profile
	user, err := s.githubOAuth.GetUserProfile(tokenResp.AccessToken)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
		http.Redirect(w, r, "/auth/error?error=profile_fetch_failed", http.StatusFound)
		return
	}

	// Generate JWT
	userID := fmt.Sprintf("github_%d", user.ID)
	token, err := GenerateJWT(userID, user.Login, s.config)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		http.Redirect(w, r, "/auth/error?error=jwt_generation_failed", http.StatusFound)
		return
	}

	// Store user and clean up state
	s.users[userID] = user
	delete(s.sessions, state)

	// Set cookie and redirect
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   s.config.JWTExpiryHours * 3600,
		Secure:   false, // Set to true in production with HTTPS
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect to success page with token
	redirectURL := fmt.Sprintf("/auth/success?token=%s", url.QueryEscape(token))
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// handleUserProfile returns the current user's profile
func (s *AuthServer) handleUserProfile(w http.ResponseWriter, r *http.Request) {
	// Get token from cookie
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	// Validate JWT
	claims, err := ValidateJWT(cookie.Value, s.config)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	user, exists := s.users[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleLogout logs the user out
func (s *AuthServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}

// handleAuthSuccess shows OAuth success page
func (s *AuthServer) handleAuthSuccess(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	tmpl := template.Must(template.New("success").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #2ecc71; font-size: 24px; }
        .token { background: #f4f4f4; padding: 10px; margin: 20px; word-break: break-all; }
        .close { margin-top: 20px; }
    </style>
</head>
<body>
    <div class="success">✓ Authentication Successful!</div>
    <p>You can now close this window and return to the game.</p>
    <div class="token">Token: {{.Token}}</div>
    <script>
        setTimeout(function() {
            window.close();
        }, 3000);
    </script>
</body>
</html>
`))

	tmpl.Execute(w, map[string]string{"Token": token})
}

// handleAuthError shows OAuth error page
func (s *AuthServer) handleAuthError(w http.ResponseWriter, r *http.Request) {
	errorParam := r.URL.Query().Get("error")

	tmpl := template.Must(template.New("error").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Failed</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .error { color: #e74c3c; font-size: 24px; }
        .close { margin-top: 20px; }
    </style>
</head>
<body>
    <div class="error">✗ Authentication Failed</div>
    <p>Error: {{.Error}}</p>
    <p>Please try again or contact support.</p>
    <script>
        setTimeout(function() {
            window.close();
        }, 5000);
    </script>
</body>
</html>
`))

	tmpl.Execute(w, map[string]string{"Error": errorParam})
}
