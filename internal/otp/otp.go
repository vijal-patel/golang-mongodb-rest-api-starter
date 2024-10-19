package otp

import (
	"crypto/rand"
	"fmt"
	"time"
)

const otpConfirmLength = 4
const otpLoginLength = 8
const otpAlphaNumChars = "abcdefghijklmnopqrstuvwxyz1234567890" // TODO find a better way to do this
const otpNumChars = "1234567890"

func GenerateConfirmOTP() string {
	return generateOTP(otpConfirmLength, otpNumChars)
}

func GenerateLoginOTP() string {
	return generateOTP(otpLoginLength, otpAlphaNumChars)
}

func generateOTP(otpLength int, otpChars string) string {
	buffer := make([]byte, otpLength)
	_, err := rand.Read(buffer)
	if err != nil {
		return fmt.Sprint(time.Now().Nanosecond())[:otpLength]
	}

	otpCharsLength := len(otpChars)
	for i := 0; i < otpLength; i++ {
		buffer[i] = otpChars[int(buffer[i])%otpCharsLength]
	}

	return string(buffer)
}
