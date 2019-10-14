package main

import (
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bnkamalesh/verifier"
	"github.com/bnkamalesh/verifier/awsses"
	"github.com/bnkamalesh/verifier/awssns"
)

func config() *verifier.Config {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 3 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Second * 5,
	}
	cfg := &verifier.Config{
		MailCfg: &awsses.Config{
			Region:     "us-west-2",
			AccessKey:  os.Getenv("AWSSES_AK"),
			Secret:     os.Getenv("AWSSES_SEC"),
			HTTPClient: httpClient,
		},
		MobileCfg: &awssns.Config{
			Region:     "us-west-2",
			AccessKey:  os.Getenv("AWSSES_AK"),
			Secret:     os.Getenv("AWSSES_SEC"),
			HTTPClient: httpClient,
		},

		DefaultEmailSub:  "",
		DefaultFromEmail: "noreply@example.com",
		EmailCallbackURL: "https://example.com/verification",
		EmailOTPExpiry:   time.Hour * 12,
		MobileOTPExpiry:  time.Minute * 10,
	}
	return cfg
}

func main() {
	vsvc, err := verifier.New(config())
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.NewMobile("+919876543210")
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.NewEmail("john.doe@example.com", "")
	if err != nil {
		println(err.Error())
		return
	}
}
