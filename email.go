package verifier

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	defaultEmailOTPPayload = `
    <html style="background: #fefefe; font-size: 14px; font-family: sans-serif; color: #333;">
    <body style="max-width: 780px; margin: 0 auto; padding: 2rem;">
        <div>Hello,</div>
        <p>
        Please click
        <a href="%s" style="font-weight: 700; text-decoration: underline">here</a>
        to verify your email.
        </p>

        <h5>Note: This link is valid only for %s.</h5>
        <p style="margin-top: 3rem; color: #999; font-size: 0.75rem;">
            <em>
                Disclaimer: This is a system generated email, please do not reply to
                this address.
            </em>
        </p>
    </body></html>
    `
)

// EmailCallbackURL adds the relevant query string parameters to the email callback URL
func (ver *Verifier) EmailCallbackURL(email, secret string) (string, error) {
	callbackURL, err := url.Parse(ver.cfg.EmailCallbackURL)
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
func (ver *Verifier) emailBody(body, callbackURL string) string {
	if body != "" {
		return body
	}

	body = fmt.Sprintf(
		defaultEmailOTPPayload,
		callbackURL,
		ver.cfg.EmailOTPExpiry.String(),
	)
	return body
}

func (ver *Verifier) emailSubject(subject string) string {
	if subject == "" {
		return subject
	}
	return ver.cfg.DefaultEmailSub
}

// Email validates an email offline (regex) and then send an email with unique secret
func (ver *Verifier) Email(ctx context.Context, recipient, subject, body string) error {
	err := validateEmailAddress(recipient)
	if err != nil {
		return err
	}

	body = strings.TrimSpace(body)
	now := time.Now()
	secExpiry := now.Add(ver.cfg.EmailOTPExpiry)
	secret := randomString(256)

	verReq := &Verification{
		ID:           newID(),
		Type:         CommTypeEmail,
		Recipient:    recipient,
		Data:         nil,
		Secret:       secret,
		SecretExpiry: &secExpiry,
		Status:       VerStatusPending,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	callbackURL, err := ver.EmailCallbackURL(verReq.Recipient, verReq.Secret)
	if err != nil {
		return err
	}

	body = ver.emailBody(body, callbackURL)
	subject = ver.emailSubject(subject)

	status, emailerr := ver.emailHandler.Send(ver.cfg.DefaultEmailSenderAddr, recipient, subject, body)
	if emailerr != nil {
		verReq.CommStatus = []CommStatus{
			CommStatus{
				Status: "failed",
				Data: map[string]interface{}{
					"error": emailerr.Error(),
				},
			},
		}
	} else {
		verReq.CommStatus = []CommStatus{
			CommStatus{
				Status: "queued",
				Data: map[string]interface{}{
					"status": status,
				},
			},
		}
	}

	verReq, err = ver.store.Create(verReq)
	if err != nil {
		return err
	}

	if emailerr != nil {
		return emailerr
	}

	return nil
}

// VerifyEmailSecret validates an email and its verification secret
func (ver *Verifier) VerifyEmailSecret(ctx context.Context, recipient, secret string) error {
	verreq, err := ver.store.ReadLastPending(CommTypeEmail, recipient)
	if err != nil {
		return err
	}

	now := time.Now()
	verreq.UpdatedAt = &now
	verreq.Attempts++

	if verreq.Attempts > maxVerifyAttempts {
		verreq.Status = VerStatusRejected
		verreq, err = ver.store.Update(verreq.ID, verreq)
		if err != nil {
			return fmt.Errorf("unknown error occurred %w", err)
		}
		return errors.New("maximum attempts exceeded")
	}

	if verreq.SecretExpiry.Before(now) {
		verreq.Status = VerStatusExpired
		verreq, err = ver.store.Update(verreq.ID, verreq)
		if err != nil {
			return fmt.Errorf("unknown error occurred. %w", err)
		}
		return errors.New("verification secret expired")
	}

	if secret != verreq.Secret {
		verreq, err = ver.store.Update(verreq.ID, verreq)
		if err != nil {
			return fmt.Errorf("unknown error occurred. %w", err)
		}
		return errors.New("invalid verification secret")
	}

	verreq.Status = VerStatusVerified
	verreq, err = ver.store.Update(verreq.ID, verreq)
	if err != nil {
		return fmt.Errorf("unknown error occurred. %w", err)
	}

	return nil
}
