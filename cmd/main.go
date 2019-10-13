package main

import (
	"net"
	"net/http"
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
			AccessKey:  "",
			Secret:     "",
			HTTPClient: httpClient,
		},
		MobileCfg: &awssns.Config{
			Region:     "us-west-2",
			AccessKey:  "",
			Secret:     "",
			HTTPClient: httpClient,
		},

		DefaultEmailSub:   "",
		DefaultFromEmail:  "noreply@example.com",
		EmailCallBackHost: "https://example.com/verification",
	}
	return cfg
}

func main() {
	_, err := verifier.New(config(), nil)
	if err != nil {
		println(err.Error())
	}
}
