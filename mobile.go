package verifier

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	defaultSMSOTPPayload = "%s is the OTP to verify your mobile number. It is valid only for %s."
)

func (ver *Verifier) smsBody(body, secret string) string {
	if body != "" {
		return body
	}

	return fmt.Sprintf(
		defaultSMSOTPPayload,
		secret,
		ver.cfg.SMSOTPExpiry.String(),
	)
}

// Mobile validates a mobile number offline (regex) and then send SMS with unique secret.
// body if provided will be the body of the SMS being sent
func (ver *Verifier) Mobile(ctx context.Context, recipient, body string) error {
	err := validateMobile(recipient)
	if err != nil {
		return err
	}

	secret := randomNumericString(6)
	body = ver.smsBody(body, secret)

	now := time.Now()
	secExpiry := now.Add(ver.cfg.SMSOTPExpiry)

	verReq := &Verification{
		ID:           newID(),
		Type:         CommTypeMobile,
		Recipient:    recipient,
		Data:         nil,
		Secret:       secret,
		SecretExpiry: &secExpiry,
		Status:       VerStatusPending,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	status, smsErr := ver.mobileHandler.SendSMS(recipient, body)
	if smsErr != nil {
		verReq.CommStatus = []CommStatus{
			CommStatus{
				Status: "failed",
				Data: map[string]interface{}{
					"error": smsErr.Error(),
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

	if smsErr != nil {
		return smsErr
	}

	return nil
}

// VerifyMobileSecret validates a mobile number and its verification secret (OTP)
func (ver *Verifier) VerifyMobileSecret(ctx context.Context, recipient, secret string) error {
	verreq, err := ver.store.ReadLastPending(CommTypeMobile, recipient)
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
			return err
		}
		return errors.New("maximum attempts exceeded")
	}

	if verreq.SecretExpiry.Before(now) {
		verreq.Status = VerStatusExpired
		verreq, err = ver.store.Update(verreq.ID, verreq)
		if err != nil {
			return err
		}
		return errors.New("verification secret expired")
	}

	if secret != verreq.Secret {
		verreq, err = ver.store.Update(verreq.ID, verreq)
		if err != nil {
			return err
		}
		return errors.New("invalid verification secret")
	}

	verreq.Status = VerStatusVerified
	verreq, err = ver.store.Update(verreq.ID, verreq)
	if err != nil {
		return err
	}

	return nil
}
