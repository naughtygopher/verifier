[![](https://goreportcard.com/badge/github.com/bnkamalesh/verifier)](https://goreportcard.com/report/github.com/bnkamalesh/verifier)
[![](https://api.codeclimate.com/v1/badges/a99a88d28ad37a79dbf6/maintainability)](https://codeclimate.com/github/bnkamalesh/verifier/maintainability)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](http://godoc.org/github.com/bnkamalesh/verifier)

# Verifier

Verifier package lets you verify emails & phone numbers with very flexble implementation at different stages. There's a functional sample app available
in the `cmd` directory.

## How does it work?

It generates secrets with an expiry, appropriate for emails & mobile phones. In case of emails, 
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
1. Complete the Redis store implementation
2. Add a Postgres store implementation
