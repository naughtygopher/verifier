// Package verifier is used for validation & verification of email, sms etc.
package verifier

import (
	"errors"
	"time"

	"github.com/bnkamalesh/verifier/awsses"
	"github.com/bnkamalesh/verifier/awssns"
)

const (
	// CommTypeMobile communication type mobile
	CommTypeMobile = commType("mobile")
	// CommTypeEmail communication type email
	CommTypeEmail = commType("email")

	// VerStatusPending verification status pending
	VerStatusPending = verificationStatus("pending")
	// VerStatusExpired verification status expired
	VerStatusExpired = verificationStatus("expired")
	// VerStatusVerified verification status verified
	VerStatusVerified = verificationStatus("verified")
	// VerStatusRejected verification status rejected
	VerStatusRejected = verificationStatus("rejected")
	// VerStatusExceededAttempts verification status when attempts are exceeded
	VerStatusExceededAttempts = verificationStatus("exceeded-attempts")
)

var (
	// ErrMaximumAttemptsExceeded is the error returned when maximum verification attempts have exceeded
	ErrMaximumAttemptsExceeded = errors.New("maximum attempts exceeded")
	// ErrSecretExpired is the error returned when the verification secret has expired
	ErrSecretExpired = errors.New("verification secret expired")
	// ErrInvalidSecret is the error returned upon receiving invalid secret
	ErrInvalidSecret = errors.New("invalid verification secret")
	// ErrInvalidMobileNumber is the error returned upon receiving invalid mobile number
	ErrInvalidMobileNumber = errors.New("invalid mobile number provided")
	// ErrInvalidEmail is the error returned upon receiving invalid email address
	ErrInvalidEmail = errors.New("invalid email address provided")
	// ErrEmptyEmailBody is the error returned when using custom email body and it's empty
	ErrEmptyEmailBody = errors.New("empty email body")
	// ErrEmptyMobileMessageBody is the error returned when using custom mobile message body and it's empty
	ErrEmptyMobileMessageBody = errors.New("empty mobile message body")
)

// commType defines the communication type (SMS, Email)
type commType string

// verificationStatus defines the status of a verification request (e.g. pending, verified, rejected)
type verificationStatus string

type emailService interface {
	// the interface returned is expected to be a reference ID for the communication sent
	// This might be a single ref ID or more info based on the service we're using
	Send(sender, recipient, subject, body string) (interface{}, error)
}

type mobileService interface {
	// the interface returned is expected to be a reference ID for the communication sent
	// This might be a single ref ID or more info based on the service we're using
	Send(recipient, body string) (interface{}, error)
}

func newID() string {
	return randomString(32)
}

// Config has all the configurations required for verifier package to function
type Config struct {
	// MailCfg is used to configure AWS SES handler
	MailCfg *awsses.Config `json:"mailCfg,omitempty"`
	// MobileCfg is used to configure AWS SNS handler
	MobileCfg *awssns.Config `json:"mobileCfg,omitempty"`

	// MaxVerifyAttempts is used to set the maximum number of times verification attempts can be made
	MaxVerifyAttempts int `json:"maxVerifyAttempts,omitempty"`
	// EmailOTPExpiry is used to define the expiry of a secret generated to verify email
	EmailOTPExpiry time.Duration `json:"emailOTPExpiry,omitempty"`
	// MobileOTPExpiry is used to define the expiry of a secret generated to verify mobile phone number
	MobileOTPExpiry time.Duration `json:"mobileOTPExpiry,omitempty"`

	// EmailCallbackURL is used to generate the callback link which is sent in the verification email
	/*
	   Two query string attributes are added to the callback URL while sending the verification link.
	   1) 'email' - The email address to which verification mail was sent
	   2) 'secret' - The secret generated for the email, this is required while verification
	*/
	EmailCallbackURL string `json:"emailCallbackURL,omitempty"`
	// DefaultFromEmail is used to set the "from" email while sending verification emails
	DefaultFromEmail string `json:"defaultFromEmail,omitempty"`
	// DefaultEmailSub is the email subject set while sending verification emails
	/*
	   If not set, a hardcoded string "Email verification request" is set as the subject.
	   The default subject is used if no subject is sent while calling the Send function
	*/
	DefaultEmailSub string `json:"defaultEmailSub,omitempty"`
}

func (cfg *Config) init() {
	if cfg.MaxVerifyAttempts < 1 {
		cfg.MaxVerifyAttempts = 3
	}
}

// CommStatus stores the status of the communication sent
type CommStatus struct {
	Status string                 `json:"status,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// Verification struct holds all data related to a single verification request
type Verification struct {
	ID           string            `json:"id,omitempty"`
	Type         commType          `json:"type,omitempty"`
	Sender       string            `json:"sender,omitempty"`
	Recipient    string            `json:"recipient,omitempty"`
	Data         map[string]string `json:"data,omitempty"`
	Secret       string            `json:"secret,omitempty"`
	SecretExpiry *time.Time        `json:"secretExpiry,omitempty"`
	// Attempts has the number of times verification has been attempted
	Attempts int `json:"attempts,omitempty"`
	// CommStatus is the communication status, and is maintained as a list to later store
	// statuses of retries
	CommStatus []CommStatus       `json:"commStatus,omitempty"`
	Status     verificationStatus `json:"status,omitempty"`
	CreatedAt  *time.Time         `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time         `json:"updatedAt,omitempty"`
}

func (v *Verification) setStatus(status interface{}, err error) {
	if status != nil {
		if len(v.CommStatus) == 0 {
			v.CommStatus = make([]CommStatus, 0, 1)
		}
		v.CommStatus = append(
			v.CommStatus,
			CommStatus{
				Status: "queued",
				Data: map[string]interface{}{
					"status": status,
				},
			},
		)
		return
	}

	if err != nil {
		if len(v.CommStatus) == 0 {
			v.CommStatus = make([]CommStatus, 0, 1)
		}

		v.CommStatus = append(
			v.CommStatus,
			CommStatus{
				Status: "failed",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			},
		)
		return
	}
}

// Verifier struct exposes all services provided by verify package
type Verifier struct {
	cfg           *Config
	emailHandler  emailService
	mobileHandler mobileService
	store         store
}

// NewRequest is used to create a new verification request
func (ver *Verifier) NewRequest(ctype commType, recipient string) (*Verification, error) {
	now := time.Now()
	secExpiry := now.Add(ver.cfg.EmailOTPExpiry)
	secret := randomString(256)

	switch ctype {
	case CommTypeMobile:
		{
			secExpiry = now.Add(ver.cfg.MobileOTPExpiry)
			secret = randomNumericString(6)
		}
	}

	verReq := &Verification{
		ID:           newID(),
		Type:         ctype,
		Recipient:    recipient,
		Data:         nil,
		Secret:       secret,
		SecretExpiry: &secExpiry,
		Status:       VerStatusPending,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
	verReq, err := ver.store.Create(verReq)
	if err != nil {
		return nil, err
	}
	return verReq, nil
}

func (ver *Verifier) verifySecret(ctype commType, recipient, secret string) error {
	verreq, err := ver.store.ReadLastPending(ctype, recipient)
	if err != nil {
		return err
	}

	return ver.verifyAndUpdate(secret, verreq)
}

func (ver *Verifier) validate(secret string, verreq *Verification) error {
	if verreq.Attempts > ver.cfg.MaxVerifyAttempts {
		return ErrMaximumAttemptsExceeded
	}

	now := time.Now()
	if verreq.SecretExpiry.Before(now) {
		return ErrSecretExpired
	}

	if secret != verreq.Secret {
		return ErrInvalidSecret
	}

	return nil
}

// verifyAndUpdate verifies all conditions required to verify a secret. And then update
// the status of verification in the store
func (ver *Verifier) verifyAndUpdate(secret string, verreq *Verification) error {
	var err error
	now := time.Now()
	verreq.UpdatedAt = &now
	verreq.Attempts++

	validationErr := ver.validate(secret, verreq)
	switch validationErr {
	case ErrMaximumAttemptsExceeded:
		{
			verreq.Status = VerStatusExceededAttempts
			verreq, err = ver.store.Update(verreq.ID, verreq)
			if err != nil {
				return err
			}
		}

	case ErrSecretExpired:
		{
			verreq.Status = VerStatusExpired
			verreq, err = ver.store.Update(verreq.ID, verreq)
			if err != nil {
				return err
			}
		}

	case ErrInvalidSecret:
		{
			verreq.Status = VerStatusRejected
			verreq, err = ver.store.Update(verreq.ID, verreq)
			if err != nil {
				return err
			}
		}
	}

	if validationErr != nil {
		return validationErr
	}

	verreq.Status = VerStatusVerified
	verreq, err = ver.store.Update(verreq.ID, verreq)
	if err != nil {
		return err
	}

	return nil
}

// VerifyEmailSecret validates an email and its verification secret
func (ver *Verifier) VerifyEmailSecret(recipient, secret string) error {
	return ver.verifySecret(CommTypeEmail, recipient, secret)
}

// NewEmailWithReq is used to send a mail with a custom verification request
func (ver *Verifier) NewEmailWithReq(verreq *Verification, subject, body string) error {
	err := validateEmailAddress(verreq.Recipient)
	if err != nil {
		return err
	}

	if body == "" {
		return ErrEmptyEmailBody
	}

	if subject == "" {
		subject = ver.cfg.DefaultEmailSub
		if subject == "" {
			subject = "Email verification request"
		}
	}

	status, sendErr := ver.emailHandler.Send(
		ver.cfg.DefaultFromEmail,
		verreq.Recipient,
		subject,
		body,
	)

	verreq.setStatus(status, sendErr)

	verreq, err = ver.store.Update(verreq.ID, verreq)
	if err != nil {
		return err
	}

	if sendErr != nil {
		return sendErr
	}

	return nil
}

// NewEmail creates a new request for email verification
func (ver *Verifier) NewEmail(recipient, subject string) error {
	err := validateEmailAddress(recipient)
	if err != nil {
		return err
	}

	verreq, err := ver.NewRequest(CommTypeEmail, recipient)
	if err != nil {
		return err
	}

	callbackURL, err := EmailCallbackURL(ver.cfg.EmailCallbackURL, verreq.Recipient, verreq.Secret)
	if err != nil {
		return err
	}

	if subject == "" {
		subject = ver.cfg.DefaultEmailSub
		if subject == "" {
			subject = "Email verification request"
		}
	}

	return ver.NewEmailWithReq(
		verreq,
		subject,
		emailBody(callbackURL, ver.cfg.EmailOTPExpiry.String()),
	)
}

// NewMobileWithReq creates a new request for mobile number verification
func (ver *Verifier) NewMobileWithReq(verreq *Verification, body string) error {
	err := validateMobile(verreq.Recipient)
	if err != nil {
		return err
	}

	if body == "" {
		return ErrEmptyMobileMessageBody
	}

	status, sendErr := ver.mobileHandler.Send(
		verreq.Recipient,
		body,
	)
	verreq.setStatus(status, sendErr)

	verreq, err = ver.store.Update(verreq.ID, verreq)
	if err != nil {
		return nil
	}

	if sendErr != nil {
		return sendErr
	}

	return nil
}

// NewMobile creates a new request for mobile number verification with default setting
func (ver *Verifier) NewMobile(recipient string) error {
	err := validateMobile(recipient)
	if err != nil {
		return err
	}

	verreq, err := ver.NewRequest(CommTypeMobile, recipient)
	if err != nil {
		return err
	}

	return ver.NewMobileWithReq(
		verreq,
		smsBody(verreq.Secret, ver.cfg.MobileOTPExpiry.String()),
	)
}

// VerifyMobileSecret validates a mobile number and its verification secret (OTP)
func (ver *Verifier) VerifyMobileSecret(recipient, secret string) error {
	return ver.verifySecret(CommTypeMobile, recipient, secret)
}

// CustomEmailHandler is used to set a custom email sending service
func (ver *Verifier) CustomEmailHandler(email emailService) error {
	ver.emailHandler = email
	// TODO: implement a validation method later, by implementing Ping
	return nil
}

// CustomStore is used to set a custom persistent store
func (ver *Verifier) CustomStore(verStore store) error {
	ver.store = verStore
	// TODO: implement a validation method later, by implementing Ping
	return nil
}

// CustomMobileHandler is used to set a custom mobile message sending service
func (ver *Verifier) CustomMobileHandler(mobile mobileService) error {
	ver.mobileHandler = mobile
	// TODO: implement a validation method later, by implementing Ping
	return nil
}

// New returns a new isntance of Verifier with all the dependencies initialized
func New(cfg *Config) (*Verifier, error) {
	verstore, err := newstore()
	if err != nil {
		return nil, err
	}

	email, err := awsses.NewService(cfg.MailCfg)
	if err != nil {
		return nil, err
	}

	mobile, err := awssns.NewService(cfg.MobileCfg)
	if err != nil {
		return nil, err
	}

	if cfg.MaxVerifyAttempts < 1 {
		cfg.MaxVerifyAttempts = 3
	}

	v := &Verifier{
		cfg:           cfg,
		emailHandler:  email,
		mobileHandler: mobile,
		store:         verstore,
	}

	return v, nil
}

// NewCustom lets you customize various components
func NewCustom(cfg *Config, verStore store, email emailService, mobile mobileService) (*Verifier, error) {
	cfg.init()

	v := &Verifier{
		cfg: cfg,
	}

	err := v.CustomEmailHandler(email)
	if err != nil {
		return nil, err
	}

	err = v.CustomMobileHandler(mobile)
	if err != nil {
		return nil, err
	}

	err = v.CustomStore(verStore)
	if err != nil {
		return nil, err
	}

	return v, nil
}
