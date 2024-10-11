<p align="center"><img src="https://repository-images.githubusercontent.com/214951539/5b1d4880-be23-11ea-956f-13b099260266" alt="verifier gopher" width="256px"/></p>

[![](https://github.com/naughtygopher/verifier/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/naughtygopher/verifier/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/naughtygopher/verifier.svg)](https://pkg.go.dev/github.com/naughtygopher/verifier)
[![Go Report Card](https://goreportcard.com/badge/github.com/naughtygopher/verifier)](https://goreportcard.com/report/github.com/naughtygopher/verifier)
[![Coverage Status](https://coveralls.io/repos/github/naughtygopher/verifier/badge.svg?branch=master)](https://coveralls.io/github/naughtygopher/verifier?branch=master)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/creativecreature/sturdyc/blob/master/LICENSE)

# Verifier

Verifier package lets you verify emails & phone numbers, with customization available at different components. There's a functional (if provided with valid configurations) sample app provided [here](https://github.com/naughtygopher/verifier/blob/master/cmd/main.go).

## How does it work?

Verifier generates secrets with an expiry, appropriate for emails & mobile phones. In case of emails,
it generates a 256 character long random alpha-numeric string, and a 6 character long numeric string
for mobile phones.

By default, it uses [AWS SES](https://aws.amazon.com/ses/) for sending e-mails & [AWS SNS](https://aws.amazon.com/sns/) for sending SMS/text messages.

## How to customize?

You can customize the following components of verifier.

```golang

    // Customize the default templates
    // there should be 2 string placeholders for email body. First is the 'callback URL' and 2nd is the expiry
    verifier.DefaultEmailOTPPayload = ``
    // there should be 1 string placeholder for SMS body. It will be the secret itself
    verifier.DefaultSMSOTPPayload = ``
    // ==

    vsvc, err := verifier.NewCustom(&Config{}, nil,nil,nil)
	if err != nil {
		log.Println(err)
		return
    }

    // Service provider for sending emails
    err := v.CustomEmailHandler(email)
	if err != nil {
        log.Println(err)
		return
    }
    // ==

    // Service provider for sending messages to mobile
	err = v.CustomMobileHandler(mobile)
	if err != nil {
        log.Println(err)
		return err
    }
    // ==

    // Persistent store used by verifier for storing secrets and all the requests
	err = v.CustomStore(verStore)
	if err != nil {
        log.Println(err)
		return
    }
    // ==

    // Using custom email & text message body
    verreq, err := vsvc.NewRequest(verifier.CommTypeEmail, recipient)
    if err != nil {
        log.Println(err)
        return
    }

    // callbackURL can be used inside the custom email body
    callbackURL, err := verifier.EmailCallbackURL("https://example.com", verreq.Recipient, verreq.Secret)
    if err != nil {
        log.Println(err)
        return
    }

    err = vsvc.NewEmailWithReq(verreq, "subject", "body")
    if err != nil {
        log.Println(err)
        return
    }

    err = vsvc.NewMobileWithReq(verreq, fmt.Sprintf("%s is your OTP", verreq.Secret))
    if err != nil {
        log.Println(err)
        return
    }
    // ==
```

## TODO

1. Unit tests
2. Setup a web service, which can be independently run, and consumed via APIs

## The gopher

The gopher used here was created using [Gopherize.me](https://gopherize.me/). Verifier helps you keep those scammers and bots away just like our hacker gopher!
