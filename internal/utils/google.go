package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type GoogleTokenInfo struct {
	Issuer        string `json:"iss"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Error         string `json:"error"`
}

func VerifyGoogleIDToken(idToken string) (*GoogleTokenInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to verify google token: status " + resp.Status)
	}

	var info GoogleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.Error != "" {
		return nil, errors.New("google auth error: " + info.Error)
	}

	if info.Email == "" {
		return nil, errors.New("email not provided in google token")
	}

	return &info, nil
}
