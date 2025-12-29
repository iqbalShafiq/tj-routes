package handler

import (
	"context"
	"errors"
	"net/http"

	"tj-routes/internal/cache"
	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/service"
	"tj-routes/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	userService     service.UserService
	config          *config.Config
	oauthConfig     *oauth2.Config
	oauthStateStore *utils.OAuthStateStore
}

// secureCookieOptions returns cookie options with security settings
func secureCookieOptions(cfg *config.Config) http.Cookie {
	return http.Cookie{
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.Server.Environment == "production",
		SameSite: http.SameSiteStrictMode,
	}
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"` // min=8 for basic length check, stronger validation in service
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

func NewAuthHandler(userService service.UserService, cfg *config.Config, cacheInstance cache.Cache) *AuthHandler {
	oauthConfig := utils.GetGoogleOAuthConfig(&cfg.OAuth)
	oauthStateStore := utils.NewOAuthStateStore(cacheInstance)
	return &AuthHandler{
		userService:     userService,
		config:          cfg,
		oauthConfig:     oauthConfig,
		oauthStateStore: oauthStateStore,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	user, err := h.userService.Register(req.Email, req.Username, req.Password)
	if err != nil {
		BadRequest(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err)
		return
	}

	user, accessToken, refreshToken, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		Unauthorized(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *AuthHandler) OAuthInitiate(c *gin.Context) {
	provider := c.Param("provider")
	if provider != "google" {
		BadRequest(c, errors.New("unsupported OAuth provider"))
		return
	}

	// Generate redirect URL for state storage
	redirectURL := c.Query("redirect_url")
	if redirectURL == "" {
		redirectURL = "/auth/callback"
	}

	// Generate secure state token with redirect URL stored server-side
	state, err := h.oauthStateStore.GenerateState(c.Request.Context(), redirectURL)
	if err != nil {
		InternalServerError(c, errors.New("failed to generate state token"))
		return
	}

	// Set secure cookie with SameSite protection
	cookie := secureCookieOptions(h.config)
	cookie.Name = "oauth_state"
	cookie.Value = state
	cookie.MaxAge = 600 // 10 minutes
	http.SetCookie(c.Writer, &cookie)

	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	providerParam := c.Param("provider")
	if providerParam != "google" {
		BadRequest(c, errors.New("unsupported OAuth provider"))
		return
	}

	// Validate state token (CSRF protection) - server-side validation
	state := c.Query("state")
	if state == "" {
		BadRequest(c, errors.New("state parameter is required"))
		return
	}

	// Validate against server-side storage
	_, valid, err := h.oauthStateStore.ValidateState(c.Request.Context(), state)
	if err != nil || !valid {
		BadRequest(c, errors.New("invalid or expired state token"))
		return
	}

	code := c.Query("code")
	if code == "" {
		BadRequest(c, errors.New("authorization code is required"))
		return
	}

	ctx := context.Background()
	token, err := h.oauthConfig.Exchange(ctx, code)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	userInfo, err := utils.GetGoogleUserInfo(ctx, token)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	provider := models.OAuthProviderGoogle
	user, accessToken, refreshToken, err := h.userService.AuthenticateOAuth(
		provider,
		userInfo.ID,
		userInfo.Email,
		userInfo.Name,
	)
	if err != nil {
		InternalServerError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, LoginResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
