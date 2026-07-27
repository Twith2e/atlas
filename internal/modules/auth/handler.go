package auth

import (
	"atlas/internal/response"

	appErr "atlas/internal/errors"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns access/refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegistrationRequest  true  "Registration payload"
// @Success      200   {object}  RegistrationResponse
// @Failure      400   {object}  response.APIError
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req RegistrationRequest
	if err := c.BindJSON(&req); err != nil {
		mapped := response.MapError(appErr.ErrInvalidRequestBody)
		c.AbortWithStatusJSON(mapped.Status, response.APIResponse[any]{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	resp, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.ConfirmPassword, req.FirstName, req.LastName)
	if err != nil {
		mapped := response.MapError(err)
		c.AbortWithStatusJSON(mapped.Status, response.APIResponse[any]{
			Status: "error",
			Error:  &mapped.Error,
		})
		return
	}

	dto := RegistrationResponse{
		User: &UserResponse{
			PublicID:  resp.User.PublicID,
			FirstName: resp.User.FirstName,
			LastName:  resp.User.LastName,
			Email:     resp.User.Email,
		},
		Tokens: resp.Tokens,
	}

	c.JSON(200, response.APIResponse[RegistrationResponse]{
		Status:  "success",
		Message: "Registration was successful",
		Data:    &dto,
	})

}
