package awssns

import (
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// Config holds all the configurations required for AWS SES to function
type Config struct {
	Region     string
	AccessKey  string
	Secret     string
	HTTPClient *http.Client
}

// AWSSNS struct exposes all the services provided by this package
type AWSSNS struct {
	cfg *Config
	sns *sns.SNS
}

// Send sends a transactional SMS using AWS SNS service
func (awssns *AWSSNS) Send(recipient string, body string) (interface{}, error) {
	params := &sns.PublishInput{
		Message:     aws.String(body),
		PhoneNumber: aws.String(recipient),
	}

	resp, err := awssns.sns.Publish(params)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// NewService returns a new instance of this package with all the required initialization
func NewService(cfg *Config) (*AWSSNS, error) {
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

	svc := sns.New(sess)

	awssns := &AWSSNS{
		cfg: cfg,
		sns: svc,
	}

	return awssns, nil
}
