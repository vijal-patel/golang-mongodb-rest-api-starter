package captcha

import (
	"encoding/json"
	"fmt"
	"golang-mongodb-rest-api-starter/internal/config"
	"golang-mongodb-rest-api-starter/internal/constants"
	"io"
	"net/http"
	"net/url"
)

func VerifyCaptcha(token string, config config.CaptchaConfig) error {
	return verifyHcaptcha(token, config.Secret, config.CaptchaUrl, constants.EmptyString)
}

type hcaptchaHttpResponse struct {
	Success     bool     `json:"success"`
	ChallengeTs string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	Credit      bool     `json:"credit"`
	ErrorCodes  []string `json:"error-codes"`
}

func verifyHcaptcha(response, secret, captchaUrl, remoteip string) error {
	values := url.Values{"secret": {secret}, "response": {response}}
	if remoteip != constants.EmptyString {
		values.Set("remoteip", remoteip)
	}
	resp, err := http.PostForm(captchaUrl, values)
	if err != nil {
		return fmt.Errorf("HTTP error: %w", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("HTTP read error: %w", err)
	}

	r := hcaptchaHttpResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return fmt.Errorf("JSON error: %w", err)
	}

	return nil
}
