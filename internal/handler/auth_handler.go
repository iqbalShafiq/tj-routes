package handler

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"

	"tj-routes/internal/config"
	"tj-routes/internal/models"
	"tj-routes/internal/service"
	"tj-routes/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	userService service.UserService
	config      *config.Config
	oauthConfig *oauth2.Config
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
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

func NewAuthHandler(userService service.UserService, cfg *config.Config) *AuthHandler {
	oauthConfig := utils.GetGoogleOAuthConfig(&cfg.OAuth)
	return &AuthHandler{
		userService: userService,
		config:      cfg,
		oauthConfig: oauthConfig,
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

	// Generate secure random state token for CSRF protection
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		InternalServerError(c, errors.New("failed to generate state token"))
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store state in session/cookie for validation in callback
	// For simplicity, we'll use a cookie (in production, consider using Redis or session store)
	c.SetCookie("oauth_state", state, 600, "/", "", h.config.Server.Environment == "production", true)

	url := h.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	providerParam := c.Param("provider")
	if providerParam != "google" {
		BadRequest(c, errors.New("unsupported OAuth provider"))
		return
	}

	// Validate state token (CSRF protection)
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")
	if err != nil || state == "" || state != storedState {
		BadRequest(c, errors.New("invalid state token"))
		return
	}

	// Clear the state cookie
	c.SetCookie("oauth_state", "", -1, "/", "", h.config.Server.Environment == "production", true)

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
