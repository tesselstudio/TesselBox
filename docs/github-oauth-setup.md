# GitHub OAuth Integration Setup

This document explains how to set up GitHub OAuth authentication for TesselBox.

## Overview

TesselBox now supports GitHub OAuth login, allowing users to authenticate using their GitHub accounts. The OAuth flow includes:

- GitHub OAuth button on login screen
- Secure token generation and validation
- User profile retrieval from GitHub API
- JWT-based session management

## Setup Instructions

### 1. Create GitHub OAuth App

1. Go to GitHub → Settings → Developer settings → OAuth Apps
2. Click "New OAuth App"
3. Fill in the application details:
   - **Application name**: TesselBox
   - **Homepage URL**: `http://localhost:8080`
   - **Authorization callback URL**: `http://localhost:8080/auth/github/callback`
4. Click "Register application"
5. Note down the **Client ID** and generate a **Client Secret**

### 2. Configure Environment Variables

Copy the example environment file:
```bash
cp .env.example .env
```

Edit `.env` and add your GitHub OAuth credentials:
```bash
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
GITHUB_REDIRECT_URI=http://localhost:8080/auth/github/callback
JWT_SECRET=your_jwt_secret_here_make_it_long_and_random
```

### 3. Run the Application

```bash
# Load environment variables
export $(cat .env | xargs)

# Run TesselBox
make run
```

## OAuth Flow

1. **User clicks "Login with GitHub"** button
2. **OAuth server starts** on `http://localhost:8080`
3. **Browser opens** to GitHub authorization page
4. **User authorizes** the application
5. **GitHub redirects** back to callback URL
6. **Server exchanges** authorization code for access token
7. **User profile** is fetched from GitHub API
8. **JWT token** is generated and stored
9. **User is logged in** and redirected to main menu

## Security Features

- **CSRF Protection**: Random state strings prevent CSRF attacks
- **JWT Tokens**: Secure token-based authentication
- **Token Expiration**: Tokens expire after 24 hours
- **Input Validation**: All user inputs are validated and sanitized
- **HTTPS Ready**: Production deployment should use HTTPS

## API Endpoints

### OAuth Endpoints
- `GET /auth/github/login` - Initiate GitHub OAuth flow
- `GET /auth/github/callback` - Handle GitHub OAuth callback

### API Endpoints
- `GET /api/user` - Get current user profile
- `POST /api/logout` - Logout user

### Utility Pages
- `GET /auth/success` - OAuth success page
- `GET /auth/error` - OAuth error page

## Token Storage

The implementation supports two storage backends:

### File Storage (Default)
- Tokens stored in `oauth_tokens.json`
- Persistent across restarts
- Automatic cleanup of expired tokens

### Memory Storage (Testing)
- Tokens stored in memory
- Lost on restart
- Useful for testing and development

## Development Notes

### Testing Without GitHub Credentials

You can test the UI without real GitHub credentials:

1. Use the test environment file:
   ```bash
   cp .env.test .env
   ```

2. The OAuth button will show an error message when clicked

### Production Deployment

For production deployment:

1. **Use HTTPS** for all OAuth URLs
2. **Set secure cookie flags**: `Secure: true, HttpOnly: true`
3. **Use Redis or database** for token storage instead of files
4. **Implement proper logging** and monitoring
5. **Set up rate limiting** for OAuth endpoints

### Customization

You can customize the OAuth flow by modifying:

- `pkg/oauth/config.go` - Configuration settings
- `pkg/oauth/github.go` - GitHub API interactions
- `pkg/oauth/handlers.go` - HTTP endpoint handlers
- `game_content/ui/login.html` - Login screen UI

## Troubleshooting

### Common Issues

1. **"OAuth not configured" error**
   - Check environment variables are set
   - Verify `.env` file exists and is readable

2. **"Invalid state" error**
   - Clear browser cookies
   - Restart the application

3. **"Token exchange failed" error**
   - Verify GitHub Client ID and Secret are correct
   - Check callback URL matches GitHub OAuth app settings

4. **"Profile fetch failed" error**
   - Verify access token is valid
   - Check GitHub API rate limits

### Debug Mode

Enable debug logging by setting:
```bash
DEBUG=true
LOG_LEVEL=debug
```

This will show detailed OAuth flow information in the console.

## Security Considerations

- Never commit `.env` files to version control
- Use strong, random JWT secrets
- Implement proper session management
- Regularly rotate GitHub OAuth secrets
- Monitor for suspicious authentication attempts
