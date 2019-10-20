// Package verifier is used for validation & verification of email, sms etc.
package verifier

import (
	"fmt"
	"testing"
	"time"

	"github.com/bnkamalesh/verifier/awsses"
	"github.com/bnkamalesh/verifier/awssns"
)

func TestConfig_init(t *testing.T) {
	type fields struct {
		MailCfg           *awsses.Config
		MobileCfg         *awssns.Config
		MaxVerifyAttempts int
		EmailOTPExpiry    time.Duration
		MobileOTPExpiry   time.Duration
		EmailCallbackURL  string
		DefaultFromEmail  string
		DefaultEmailSub   string
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
		{
			name: "max attempts lesser than 1",
			fields: fields{
				MaxVerifyAttempts: 0,
			},
		},
		{
			name: "max attempts equal to 1",
			fields: fields{
				MaxVerifyAttempts: 1,
			},
		},
		{
			name: "max attempts greater than 1",
			fields: fields{
				MaxVerifyAttempts: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				MaxVerifyAttempts: tt.fields.MaxVerifyAttempts,
				EmailOTPExpiry:    tt.fields.EmailOTPExpiry,
				MobileOTPExpiry:   tt.fields.MobileOTPExpiry,
				EmailCallbackURL:  tt.fields.EmailCallbackURL,
				DefaultFromEmail:  tt.fields.DefaultFromEmail,
				DefaultEmailSub:   tt.fields.DefaultEmailSub,
			}
			cfg.init()
			fmt.Println(
				"tt.fields.MaxVerifyAttempts < 1 && cfg.MaxVerifyAttempts != 3",
				tt.fields.MaxVerifyAttempts,
				tt.fields.MaxVerifyAttempts < 1,
				cfg.MaxVerifyAttempts != 3,
			)
			if tt.fields.MaxVerifyAttempts < 1 && cfg.MaxVerifyAttempts != 3 {
				t.Fatalf("Expected max attempts 3, got %d", cfg.MaxVerifyAttempts)
			} else if tt.fields.MaxVerifyAttempts >= 1 {
				if tt.fields.MaxVerifyAttempts != cfg.MaxVerifyAttempts {
					t.Fatalf(
						"Expected max attempts %d, got %d",
						tt.fields.MaxVerifyAttempts,
						cfg.MaxVerifyAttempts,
					)
				}
			}
		})
	}
}

func TestVerifier_validate(t *testing.T) {
	now := time.Now()
	validExpiry := now.Add(time.Minute * 30)
	invalidExpiry := now.Add(-(time.Minute * 30))

	type fields struct {
		cfg           *Config
		emailHandler  emailService
		mobileHandler mobileService
		store         store
	}
	type args struct {
		secret string
		verreq *Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "valid",
			fields: fields{
				cfg: &Config{
					MaxVerifyAttempts: 3,
				},
			},
			args: args{
				secret: "helloworld",
				verreq: &Request{
					Secret:       "helloworld",
					SecretExpiry: &validExpiry,
				},
			},
			wantErr: false,
		},
		{
			name: "exceeded attempts",
			fields: fields{
				cfg: &Config{
					MaxVerifyAttempts: 1,
				},
			},
			args: args{
				secret: "helloworld",
				verreq: &Request{
					Secret:       "helloworld",
					SecretExpiry: &validExpiry,
					Attempts:     2,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid expiry",
			fields: fields{
				cfg: &Config{
					MaxVerifyAttempts: 3,
				},
			},
			args: args{
				secret: "helloworld",
				verreq: &Request{
					Secret:       "helloworld",
					SecretExpiry: &invalidExpiry,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid secret",
			fields: fields{
				cfg: &Config{
					MaxVerifyAttempts: 3,
				},
			},
			args: args{
				secret: "helloworld",
				verreq: &Request{
					Secret:       "helloworld-2",
					SecretExpiry: &validExpiry,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver := &Verifier{
				cfg:           tt.fields.cfg,
				emailHandler:  tt.fields.emailHandler,
				mobileHandler: tt.fields.mobileHandler,
				store:         tt.fields.store,
			}
			if err := ver.validate(tt.args.secret, tt.args.verreq); (err != nil) != tt.wantErr {
				t.Errorf("Verifier.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
