package models

type LoginRequest struct {
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8,alphanum,max=255"`
	CaptchaToken string `json:"captchaToken" validate:"required"`
}

type RegisterRequest struct {
	Email            string `json:"email" validate:"required,email"`
	Password         string `json:"password" validate:"required,min=8,alphanum,max=255"`
	UserName         string `json:"userName" validate:"required,nourl,max=255"`
	OrganizationName string `json:"organizationName" validate:"required,max=255"`
	CaptchaToken     string `json:"captchaToken" validate:"required"`
}

type OtpRequest struct {
	Otp string `json:"otp" validate:"required"`
}

type RefreshRequest struct {
	Token string `json:"token" validate:"required"`
}

type PasswordChangeRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8,alphanum,max=255"`
}

type PasswordRecoverRequest struct {
	Email        string `json:"email" validate:"required"`
	CaptchaToken string `json:"captchaToken" validate:"required"`
}

type PasswordResetRequest struct {
	Email       string `json:"email" validate:"required"`
	Otp         string `json:"otp" validate:"required"`
	NewPassword string `json:"password" validate:"required,min=8,alphanum,max=255"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	Exp          int64  `json:"exp"`
}

func NewLoginResponse(token, refreshToken string, exp int64) *LoginResponse {
	return &LoginResponse{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Exp:          exp,
	}
}
