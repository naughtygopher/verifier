package awsses

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

const (
	charset = "UTF-8"
)

// Config holds all the configurations required for AWS SES to function
type Config struct {
	Region     string
	AccessKey  string
	Secret     string
	HTTPClient *http.Client
}

// AWSSES struct exposes all the services provided by this package
type AWSSES struct {
	cfg *Config
	ses *ses.SES
}

func (awsses *AWSSES) emailInput(sender, recipient, subject, htmlbody, textbody string) *ses.SendEmailInput {
	return &ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: []*string{},
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String(charset),
					Data:    aws.String(htmlbody),
				},
				// Text: &ses.Content{
				// 	Charset: aws.String(charset),
				// 	Data:    aws.String(textbody),
				// },
			},
			Subject: &ses.Content{
				Charset: aws.String(charset),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}
}

// Send sends an email
func (awsses *AWSSES) Send(sender, recipient, subject, body string) (interface{}, error) {
	inp := awsses.emailInput(sender, recipient, subject, body, "")
	result, err := awsses.ses.SendEmail(inp)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// NewService returns an instance of AWSES after initializing all required dependencies
func NewService(cfg *Config) (*AWSSES, error) {
	sess, err := session.NewSession(
		&aws.Config{
			Region:      &cfg.Region,
			Credentials: credentials.NewStaticCredentials(cfg.AccessKey, cfg.Secret, ""),
			HTTPClient:  cfg.HTTPClient,
		},
	)

	if err != nil {
		return nil, err
	}

	awsses := &AWSSES{
		cfg: cfg,
		ses: ses.New(sess),
	}

	return awsses, nil
}
