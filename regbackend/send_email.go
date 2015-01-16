package main

import (
	"bytes"
	"net/smtp"
	"text/template"
)

type EmailSender interface {
	Send(from string, to []string, msg []byte) error
}

type localMailer struct {
	serverAddr string
}

func NewLocalMailder(serverAddr string) EmailSender {
	return localMailer{
		serverAddr: serverAddr,
	}
}

// Shamelessly copied from net/smtp's SendMail function, to avoid the TLS connection.
func (l localMailer) Send(from string, to []string, msg []byte) error {
	c, err := smtp.Dial(l.serverAddr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Hello("registration.cubjamboree.ca"); err != nil {
		return err
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

type ConfirmationEmailService struct {
	domain         string
	emailSender    EmailSender
	fromAddress    string
	contactAddress string
	preRegDb       PreRegDb
}

func NewConfirmationEmailService(domain, fromAddress, contactAddress string, emailSender EmailSender, preRegDb PreRegDb) *ConfirmationEmailService {
	ret := &ConfirmationEmailService{
		domain:         domain,
		fromAddress:    fromAddress,
		emailSender:    emailSender,
		contactAddress: contactAddress,
		preRegDb:       preRegDb,
	}

	return ret
}

func (c *ConfirmationEmailService) RequestEmailConfirmation(gpr *GroupPreRegistration) error {
	buf := &bytes.Buffer{}
	type confirmationData struct {
		ToAddress, FirstName, LastName, SecurityKey, ValidationToken, Domain, FromAddress, ContactAddress string
	}
	if err := emailTemplate.Execute(buf, confirmationData{
		ToAddress:       gpr.ContactLeaderEmail,
		FirstName:       gpr.ContactLeaderFirstName,
		LastName:        gpr.ContactLeaderLastName,
		SecurityKey:     gpr.SecurityKey,
		ValidationToken: gpr.ValidationToken,
		Domain:          c.domain,
		FromAddress:     c.fromAddress,
		ContactAddress:  c.contactAddress,
	}); err != nil {
		return err
	}
	if err := c.emailSender.Send(c.fromAddress, []string{gpr.ContactLeaderEmail}, buf.Bytes()); err != nil {
		return err
	}
	return c.preRegDb.NoteConfirmationEmailSent(gpr)
}

const emailTemplateString = `From: {{.FromAddress}}
To: {{.ToAddress}}
Subject: Confirm CCJ16 Preregistration
Content-Type: text/plain; charset=UTF-8

Hi Scouter {{.FirstName}} {{.LastName}},

Thank you for preregistering for CCJ16!  We hope you are as excited about this amazing camp as we are.  In order to confirm your CCJ16 preregistration, we ask that you confirm your email address by visiting the following page:

https://{{.Domain}}/confirmpreregistration?email={{.ToAddress}}&token={{.ValidationToken}}

If you are unable to click the link, please copy and paste it into your web browser.

If you wish to review your preregistration, please visit the following page:

https://{{.Domain}}/registration/{{.SecurityKey}}


Thanks again,
--
The CCJ16 team

If you have any questions, please contact us at {{.ContactAddress}}`

var emailTemplate = template.Must(template.New("email").Parse(emailTemplateString))
