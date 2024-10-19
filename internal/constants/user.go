package constants

const (
	UserCollection              string = "users"
	UserNotFound                string = "User not found"
	UserNotConfirmed            string = "Please confirm your email address to perform this action"
	UserAlreadyExists           string = "A user with this email address already exists"
	UserUpdated                 string = "User updated"
	UserDeleted                 string = "User deleted"
	UserPasswordRecoverResponse string = "You will receive a code via email if this email has an account with us"
	UserPasswordUpdated         string = "Password updated"
	UserConfirmOtpIntervalMilli int64  = 60000
)
