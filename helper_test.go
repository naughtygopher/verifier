package verifier

import (
	"fmt"
	"regexp"
	"testing"
)

func Test_randomNumericString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "10 digits",
			args: args{
				n: 10,
			},
		},
		{
			name: "5 digits",
			args: args{
				n: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexNumeric := regexp.MustCompile(fmt.Sprintf("^([0-9]+){%d}", tt.args.n))
			got := randomNumericString(tt.args.n)
			if !regexNumeric.MatchString(got) {
				t.Fatalf("Expected %d character numeric string, got '%s'", tt.args.n, got)
			}
		})
	}
}

func Test_randomString(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "10 characters",
			args: args{
				n: 10,
			},
		},
		{
			name: "5 characters",
			args: args{
				n: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexAlphaNumeric := regexp.MustCompile(fmt.Sprintf("^([0-9a-zA-Z]+){%d}", tt.args.n))
			got := randomString(tt.args.n)
			if !regexAlphaNumeric.MatchString(got) {
				t.Fatalf("Expected %d character alpha numeric string, got '%s'", tt.args.n, got)
			}
		})
	}
}

func Test_validateEmailAddress(t *testing.T) {
	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Valid email",
			args: args{
				email: "hello@example.com",
			},
			wantErr: false,
		},
		{
			name: "Invalid email - no '@'",
			args: args{
				email: "example.com",
			},
			wantErr: true,
		},
		{
			name: "Invalid email - no '.'",
			args: args{
				email: "hello@examplecom",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateEmailAddress(tt.args.email); (err != nil) != tt.wantErr {
				t.Errorf("validateEmailAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEmailCallbackURL(t *testing.T) {
	type args struct {
		baseurl string
		email   string
		secret  string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Valid callback",
			args: args{
				baseurl: "https://example.com",
				email:   "hello@example.com",
				secret:  "secret with special chars!@to be encoded",
			},
			want:    "https://example.com?email=hello%40example.com&secret=secret+with+special+chars%21%40to+be+encoded",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EmailCallbackURL(tt.args.baseurl, tt.args.email, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("EmailCallbackURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EmailCallbackURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateMobile(t *testing.T) {
	type args struct {
		mobile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "valid mobile",
			args: args{
				mobile: "+919876543210",
			},
			wantErr: false,
		},
		{
			name: "invalid mobile (special symbols)",
			args: args{
				mobile: "#919876543210",
			},
			wantErr: true,
		},
		{
			name: "invalid mobile (lesser than 7)",
			args: args{
				mobile: "919875",
			},
			wantErr: true,
		},
		{
			name: "invalid mobile (greater than 24)",
			args: args{
				mobile: "919876543919876543919876543",
			},
			wantErr: true,
		},
		{
			name: "invalid mobile (alpha numeric)",
			args: args{
				mobile: "91a87654321",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMobile(tt.args.mobile); (err != nil) != tt.wantErr {
				t.Errorf("validateMobile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
