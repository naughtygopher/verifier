package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/bnkamalesh/verifier"
	"github.com/bnkamalesh/verifier/awsses"
	"github.com/bnkamalesh/verifier/awssns"
	"github.com/bnkamalesh/verifier/stores"
)

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 3 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 3 * time.Second,
		},
		Timeout: time.Second * 5,
	}
}

func mailmobileConfig() (*awsses.Config, *awssns.Config) {
	httpClient := newHTTPClient()
	return &awsses.Config{
			Region:     "us-west-2",
			AccessKey:  os.Getenv("AWSSES_AK"),
			Secret:     os.Getenv("AWSSES_SEC"),
			HTTPClient: httpClient,
		},
		&awssns.Config{
			Region:     "us-west-2",
			AccessKey:  os.Getenv("AWSSES_AK"),
			Secret:     os.Getenv("AWSSES_SEC"),
			HTTPClient: httpClient,
		}
}

func redisConfig() *stores.RedisConfig {
	return &stores.RedisConfig{
		Hosts: []string{
			"localhost:6379",
		},
		DialTimeoutSecs:  time.Second * 3,
		ReadTimeoutSecs:  time.Second * 1,
		WriteTimeoutSecs: time.Second * 2,
	}
}

func postgresConfig() *stores.PostgresConfig {
	return &stores.PostgresConfig{
		Host:      "localhost",
		Port:      "5432",
		Username:  "user1",
		Password:  "password",
		StoreName: "mydb",
		PoolSize:  100,

		DialTimeoutSecs:  time.Second * 10,
		ReadTimeoutSecs:  time.Second * 10,
		WriteTimeoutSecs: time.Second * 20,
		IdleTimeoutSecs:  time.Minute * 5,

		TableName: "VerificationRequests",
	}
}

func config() *verifier.Config {
	cfg := &verifier.Config{
		DefaultEmailSub:  "",
		DefaultFromEmail: "noreply@example.com",
		EmailCallbackURL: "https://example.com/verify-email",
		EmailOTPExpiry:   time.Hour * 12,
		MobileOTPExpiry:  time.Minute * 10,
	}
	return cfg
}

func notifyWithoutCustomRequest(vsvc *verifier.Verifier) {
	const mobile = "+919876543210"
	err := vsvc.NewMobile(mobile)
	if err != nil {
		println(err.Error())
		return
	}

	const email = "john.doe@example.com"
	err = vsvc.NewEmail(email, "")
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.VerifyMobileSecret(mobile, "secret")
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.VerifyEmailSecret(email, "secret")
	if err != nil {
		println(err.Error())
		return
	}
}

func notifyWithCustomRequest(vsvc *verifier.Verifier) {
	req, err := vsvc.NewRequest(
		verifier.CommTypeMobile,
		"+919876543210",
	)
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.NewMobileWithReq(
		req,
		fmt.Sprintf(
			verifier.DefaultSMSOTPPayload,
			req.Secret,
			req.SecretExpiry.String(),
		),
	)
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.VerifyMobileSecret(req.Recipient, req.Secret)
	if err != nil {
		println(err.Error())
		return
	}

	req, err = vsvc.NewRequest(
		verifier.CommTypeEmail,
		"john.doe@example.com",
	)
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.NewEmailWithReq(
		req,
		"verify your email",
		fmt.Sprintf(
			verifier.DefaultEmailOTPPayload,
			req.Secret,
			req.SecretExpiry.String(),
		),
	)
	if err != nil {
		println(err.Error())
		return
	}

	err = vsvc.VerifyEmailSecret(req.Recipient, req.Secret)
	if err != nil {
		println(err.Error())
		return
	}
}

func main() {

	mailCfg, mobCfg := mailmobileConfig()

	mailservice, err := awsses.NewService(mailCfg)
	if err != nil {
		println(err.Error())
		return
	}

	mobService, err := awssns.NewService(mobCfg)
	if err != nil {
		println(err.Error())
		return
	}

	// redisstore, err := stores.NewRedis(redisConfig())
	// if err != nil {
	// 	println(err.Error())
	// 	return
	// }
	// vsvc, err := verifier.New(
	// 	config(),
	// 	redisstore,
	// 	mailservice,
	// 	mobService,
	// )

	postgrestore, err := stores.NewPostgres(postgresConfig())
	if err != nil {
		println(err.Error())
		return
	}
	vsvc, err := verifier.New(
		config(),
		postgrestore,
		mailservice,
		mobService,
	)
	if err != nil {
		println(err.Error())
		return
	}

	notifyWithCustomRequest(vsvc)
	// notifyWithoutCustomRequest(vsvc)
}
