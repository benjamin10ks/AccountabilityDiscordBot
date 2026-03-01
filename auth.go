package main

import (
	"crypto/rand"
	"encoding/hex"
)

// TODO: updated with salt and hashing for security
func generateStateToken() string {
	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func exchangeCodeForToken(code string) (string, error) {
	return "", nil
}

// TODO: implement this function to set up GitHub webhooks for the registered repositories
func createWebhook(accessToken, owner, repo, baseURL string) error {
	return nil
}
