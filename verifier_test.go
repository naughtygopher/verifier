// Package verifier is used for validation & verification of email, sms etc.
package verifier

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/naughtygopher/verifier/awsses"
	"github.com/naughtygopher/verifier/awssns"
)

type mockstore struct {
	data map[string]*Request
}

func (ms *mockstore) Create(ver *Request) (*Request, error) {
	key := fmt.Sprintf(
		"%s-%s",
		ver.Type,
		ver.Recipient,
	)
	ms.data[key] = ver
	return ver, nil
}
func (ms *mockstore) ReadLastPending(ctype CommType, recipient string) (*Request, error) {
	key := fmt.Sprintf(
		"%s-%s",
		ctype,
		recipient,
	)
	req, ok := ms.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return req, nil
}

func (ms *mockstore) Update(verID string, ver *Request) (*Request, error) {
	key := fmt.Sprintf(
		"%s-%s",
		ver.Type,
		ver.Recipient,
	)
	ms.data[key] = ver
	return ver, nil
}

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

func Test_newID(t *testing.T) {
	regex := regexp.MustCompile("^[0-9a-zA-Z]{32}$")
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
		{
			name: "alpha numeric 32 char long random string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newID()
			if !regex.MatchString(got) {
				t.Fatalf("Expected 32 chr long alpha numeric random string, got '%s'", got)
			}
		})
	}
}

func TestRequest_setStatus(t *testing.T) {
	type fields struct {
		ID           string
		Type         CommType
		Sender       string
		Recipient    string
		Data         map[string]string
		Secret       string
		SecretExpiry *time.Time
		Attempts     int
		CommStatus   []CommStatus
		Status       verificationStatus
		CreatedAt    *time.Time
		UpdatedAt    *time.Time
	}
	type args struct {
		status interface{}
		err    error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "with status",
			args: args{
				status: "hello world",
			},
			fields: fields{},
		},
		{
			name: "with error",
			args: args{
				status: nil,
				err:    errors.New("some error"),
			},
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Request{
				ID:           tt.fields.ID,
				Type:         tt.fields.Type,
				Sender:       tt.fields.Sender,
				Recipient:    tt.fields.Recipient,
				Data:         tt.fields.Data,
				Secret:       tt.fields.Secret,
				SecretExpiry: tt.fields.SecretExpiry,
				Attempts:     tt.fields.Attempts,
				CommStatus:   tt.fields.CommStatus,
				Status:       tt.fields.Status,
				CreatedAt:    tt.fields.CreatedAt,
				UpdatedAt:    tt.fields.UpdatedAt,
			}
			v.setStatus(tt.args.status, tt.args.err)

			if tt.args.err != nil {
				got := v.CommStatus[0]
				if got.Status != "failed" {
					t.Fatalf("expected status 'failed', got '%s'", got.Status)
				}
				wantData := map[string]interface{}{
					"error": tt.args.err.Error(),
				}
				if !reflect.DeepEqual(wantData, got.Data) {
					t.Fatalf("expected '%v', got '%v'", wantData, got.Data)
				}
				return
			}

			if tt.args.status != nil {
				got := v.CommStatus[0]
				if got.Status != "queued" {
					t.Fatalf("expected status 'queued', got '%s'", got.Status)
				}

				wantData := map[string]interface{}{
					"status": tt.args.status,
				}
				if !reflect.DeepEqual(wantData, got.Data) {
					t.Fatalf("expected '%v', got '%v'", wantData, got.Data)
				}
			}
		})
	}
}
