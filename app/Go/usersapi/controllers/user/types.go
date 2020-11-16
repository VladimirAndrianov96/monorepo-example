package user_controller

type RegistrationRequest struct{
	EmailAddress string `json:"email_address"`
	Password string `json:"password"`
}

type RegistrationSuccessResponse struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
}

type StatusResponse struct{
	Message string `json:"response"`
}