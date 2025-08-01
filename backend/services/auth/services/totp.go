package services

import (
	"github.com/pquerna/otp/totp"
)

func GenerateTOTPSecret(username string) (string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SusiApp",
		AccountName: username,
	})
	if err != nil {
		return "", err
	}
	return key.Secret(), nil
}

func ValidateTOTP(code, secret string) bool {
	return totp.Validate(code, secret)
}
