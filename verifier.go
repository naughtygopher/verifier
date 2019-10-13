// Package verifier is used for validation & verification of email, sms etc.
package verifier

import (
	"database/sql"
	"fmt"
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

	maxVerifyAttempts = 3
)

// commType defines the communication type (SMS, Email)
type commType string

// verificationStatus defines the status of a verification request (e.g. pending, verified, rejected)
type verificationStatus string

type emailHandler interface {
	// the interface returned is expected to be a reference ID for the communication sent
	// This might be a single ref ID or more info based on the service we're using
	Send(sender, recipient, subject, body string) (interface{}, error)
}

type mobileHandler interface {
	// the interface returned is expected to be a reference ID for the communication sent
	// This might be a single ref ID or more info based on the service we're using
	SendSMS(recipient, body string) (interface{}, error)
}

func newID() string {
	return fmt.Sprintf("verreq-%s", randomString(32))
}

// Config has all the configurations required for verifier package to function
type Config struct {
	MailCfg                *awsses.Config `json:"mailCfg,omitempty"`
	MobileCfg              *awssns.Config `json:"mobileCfg,omitempty"`
	DefaultEmailSenderAddr string         `json:"defaultEmailSenderAddr,omitempty"`
	EmailCallbackURL       string         `json:"emailCallbackURL,omitempty"`
	DefaultFromEmail       string         `json:"defaultFromEmail,omitempty"`
	EmailCallBackHost      string         `json:"emailCallBackHost,omitempty"`
	DefaultEmailSub        string         `json:"defaultEmailSub,omitempty"`
	EmailOTPExpiry         time.Duration  `json:"emailOTPExpiry,omitempty"`
	SMSOTPExpiry           time.Duration  `json:"smsotpExpiry,omitempty"`
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
	Attempts   int                `json:"attempts,omitempty"`
	CommStatus []CommStatus       `json:"commStatus,omitempty"`
	Status     verificationStatus `json:"status,omitempty"`
	CreatedAt  *time.Time         `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time         `json:"updatedAt,omitempty"`
}

// Verifier struct exposes all services provided by verify package
type Verifier struct {
	cfg           *Config
	emailHandler  emailHandler
	mobileHandler mobileHandler
	store         store
}

// New returns a new isntance of Verifier with all the dependencies initialized
func New(cfg *Config, storeDriver *sql.DB) (*Verifier, error) {
	vs, err := newstore(storeDriver)
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

	v := &Verifier{
		cfg:           cfg,
		emailHandler:  email,
		mobileHandler: mobile,
		store:         vs,
	}

	return v, nil
}
