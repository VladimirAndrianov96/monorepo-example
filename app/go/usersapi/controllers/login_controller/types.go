package login_controller

type LoginRequest struct{
	EmailAddress string `json:"email_address"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}