package auth

type RegistrationRequest struct {
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegistrationResponse struct {
	User   *UserResponse `json:"user"`
	Tokens Tokens        `json:"tokens"`
}

type LoginResponse struct {
	User   *UserResponse `json:"user"`
	Tokens Tokens        `json:"tokens"`
}

type RefreshResponse struct {
	Tokens Tokens `json:"tokens"`
}

type UserResponse struct {
	PublicID  string `json:"public_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// Tokens carries both tokens internally between the service and handler, but
// only the access token is ever serialised into a response body — the refresh
// token is delivered as an HttpOnly cookie, so it is hidden from the docs.
type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty" swaggerignore:"true"`
}
