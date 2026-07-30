package auth

import (
	"atlas/internal/middleware"
	"atlas/internal/response"
	"net/http"
	"strings"

	appErr "atlas/internal/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
	env string
}

func NewHandler(svc *Service, env string) *Handler {
	return &Handler{svc: svc, env: env}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a user account, opens a session, and returns an access token.
// @Description  The refresh token is returned as an HttpOnly, SameSite=Strict cookie scoped to /api/v1/auth — it is never included in the response body.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegistrationRequest  true  "Registration payload"
// @Success      201   {object}  response.APIResponse[auth.RegistrationResponse]
// @Failure      400   {object}  response.ErrorResponse  "Invalid request body"
// @Failure      409   {object}  response.ErrorResponse  "Email already registered"
// @Failure      422   {object}  response.ErrorResponse  "Passwords do not match, or password fails policy"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegistrationRequest
	if err := c.BindJSON(&req); err != nil {
		mapped := response.MapError(appErr.ErrInvalidRequestBody)
		c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.ConfirmPassword, req.FirstName, req.LastName)
	if err != nil {
		mapped := response.MapError(err)
		c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	setRefreshCookie(c, resp.Tokens.RefreshToken, h.env)

	dto := RegistrationResponse{
		User: &UserResponse{
			PublicID:  resp.User.PublicID,
			FirstName: resp.User.FirstName,
			LastName:  resp.User.LastName,
			Email:     resp.User.Email,
		},
		Tokens: Tokens{
			AccessToken: resp.Tokens.AccessToken,
		},
	}

	c.JSON(http.StatusCreated, response.APIResponse[RegistrationResponse]{
		Status:  "success",
		Message: "Registration was successful",
		Data:    &dto,
	})

}

// Login godoc
// @Summary      Log in
// @Description  Authenticates an email and password, opens a new session, and returns an access token.
// @Description  The refresh token is returned as an HttpOnly, SameSite=Strict cookie scoped to /api/v1/auth — it is never included in the response body.
// @Description  An unknown email and a wrong password return the same 401 so the endpoint cannot be used to discover registered addresses.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Login payload"
// @Success      200   {object}  response.APIResponse[auth.LoginResponse]
// @Failure      400   {object}  response.ErrorResponse  "Invalid request body"
// @Failure      401   {object}  response.ErrorResponse  "Invalid email or password"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.BindJSON(&req); err != nil {
		mapped := response.MapError(appErr.ErrInvalidRequestBody)
		c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	resp, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		mapped := response.MapError(err)
		c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	setRefreshCookie(c, resp.Tokens.RefreshToken, h.env)

	dto := LoginResponse{
		User: &UserResponse{
			PublicID:  resp.User.PublicID,
			FirstName: resp.User.FirstName,
			LastName:  resp.User.LastName,
			Email:     resp.User.Email,
		},
		Tokens: Tokens{
			AccessToken: resp.Tokens.AccessToken,
		},
	}

	c.JSON(200, response.APIResponse[LoginResponse]{
		Status:  "success",
		Message: "Login was successful",
		Data:    &dto,
	})
}

// Logout godoc
// @Summary      Log out
// @Description  Revokes the current session and expires the refresh token cookie. Requires the refresh token cookie.
// @Description  Revoking an already-revoked session is a no-op success, so repeat calls are safe.
// @Description  Clients should clear local auth state and redirect regardless of the response status — logout is best-effort by design.
// @Tags         auth
// @Produce      json
// @Success      200   {object}  response.MessageResponse  "Session revoked and cookie cleared"
// @Failure      401   {object}  response.ErrorResponse  "Missing or invalid refresh token cookie"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	sid := c.GetString(middleware.SessionIDContextKey)

	err := h.svc.Logout(c.Request.Context(), sid)
	if err != nil {
		mapped := response.MapError(err)
		c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	clearRefreshCookie(c, h.env)

	c.JSON(200, response.MessageResponse{
		Status:  "success",
		Message: "Logout was successful",
	})
}

// RefreshAccessToken godoc
// @Summary      Refresh the access token
// @Description  Issues a new access token and rotates the refresh token cookie. Requires the refresh token cookie.
// @Description  The session id (sid) is preserved across refreshes, so access tokens issued earlier in the session remain valid until they expire on their own.
// @Description
// @Description  NOTE — clients must serialize refresh calls. Every call rotates the refresh token, and presenting a token that has already been rotated away is treated as theft: the whole session is revoked and the user must log in again. Keep a single in-flight request to this endpoint and have all concurrent 401s await it. Firing two refreshes in parallel will log the user out.
// @Tags         auth
// @Produce      json
// @Success      200   {object}  response.APIResponse[auth.RefreshResponse]
// @Failure      401   {object}  response.ErrorResponse  "Missing/invalid cookie, inactive session, or token reuse detected (session revoked)"
// @Failure      500   {object}  response.ErrorResponse  "Internal server error"
// @Router       /auth/refresh [post]
func (h *Handler) RefreshAccessToken(c *gin.Context) {
	sid := strings.TrimSpace(c.GetString(middleware.SessionIDContextKey))
	pid := strings.TrimSpace(c.GetString(middleware.UserPublicIDContextKey))
	refreshToken := strings.TrimSpace(c.GetString(middleware.RefreshTokenContextKey))

	resp, err := h.svc.RefreshAccessToken(c.Request.Context(), pid, sid, refreshToken)
	if err != nil {
		mapped := response.MapError(err)
		c.AbortWithStatusJSON(mapped.Status, response.ErrorResponse{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	setRefreshCookie(c, resp.RefreshToken, h.env)

	dto := RefreshResponse{
		Tokens: Tokens{
			AccessToken: resp.AccessToken,
		},
	}

	c.JSON(http.StatusOK, response.APIResponse[RefreshResponse]{
		Status:  "success",
		Message: "Access token successfully refreshed",
		Data:    &dto,
	})
}
