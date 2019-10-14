package verifier

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	regexMobile = regexp.MustCompile(`^(\+)?([0-9]+){7,24}$`)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	// DefaultEmailOTPPayload is the default email body used
	DefaultEmailOTPPayload = `
    <html style="background: #fefefe; font-size: 14px; font-family: sans-serif; color: #333;">
    <body style="max-width: 780px; margin: 0 auto; padding: 2rem;">
      <div>Hello,</div>
      <p>
        Please click
        <a href="%s" style="font-weight: 700; text-decoration: underline">here</a>
        to verify your email.
      </p>
  
      <h5>Note: This link is valid only for %s.</h5>
      <p style="margin-top: 3rem; color: #999;"><em>
          Disclaimer: This is a system generated email, please do not reply to this address.
      </em></p>
    </body></html>
  `
	// DefaultSMSOTPPayload is the default text message body
	DefaultSMSOTPPayload = "%s is the OTP to verify your mobile number. It is valid only for %s."
)

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
		return ErrInvalidEmail
	}

	parts = strings.Split(parts[1], ".")
	if len(parts) < 2 {
		return ErrInvalidEmail
	}
	return nil
}

func validateMobile(mobile string) error {
	if !regexMobile.MatchString(mobile) {
		return ErrInvalidMobileNumber
	}
	return nil
}

// EmailCallbackURL adds the relevant query string parameters to the email callback URL
func EmailCallbackURL(baseurl, email, secret string) (string, error) {
	callbackURL, err := url.Parse(baseurl)
	if err != nil {
		return "", err
	}

	queryParms := callbackURL.Query()
	queryParms.Add("email", email)
	queryParms.Add("secret", secret)
	callbackURL.RawQuery = queryParms.Encode()

	return callbackURL.String(), nil
}

// emailBody if body is an empty string it parses and sends back the default email body
func emailBody(callbackURL, expiry string) string {
	return fmt.Sprintf(
		DefaultEmailOTPPayload,
		callbackURL,
		expiry,
	)
}

func smsBody(secret, expiry string) string {
	return fmt.Sprintf(
		DefaultSMSOTPPayload,
		secret,
		expiry,
	)
}
