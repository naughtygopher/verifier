package verifier

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	alphaNumericList = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
	numericList      = []rune("0123456789")
)

func randRune(runes []rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

// randomString returns a random alpha numeric string of length n
func randomString(n int) string {
	return randRune(alphaNumericList, n)
}

func randomNumericString(n int) string {
	return randRune(numericList, n)
}

// validateEmailAddress offline validation of email.
func validateEmailAddress(email string) error {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("[1] invalid email provided")
	}

	parts = strings.Split(parts[1], ".")
	if len(parts) < 2 {
		return errors.New("[2] invalid email provided")
	}
	return nil
}

func validateMobile(mobile string) error {
	return nil
}
