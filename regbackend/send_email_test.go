package main

import (
	"fmt"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/CCJ16/registration/regbackend/boltorm"
)

type testEmail struct {
	From string
	To   []string
	Msg  []byte
}

type testEmailSender struct {
	Emails []testEmail
}

func (t *testEmailSender) Send(from string, to []string, msg []byte) error {
	t.Emails = append(t.Emails, testEmail{
		From: from,
		To:   to,
		Msg:  msg,
	})
	return nil
}

func TestEmailRequest(t *testing.T) {
	fromAddress := "testsender@examplesending.com"
	Convey("With mock instance for end services", t, func() {
		testEmailSender := &testEmailSender{}
		testPreRegDb, err := NewPreRegBoltDb(boltorm.NewMemoryDB(), &configType{}, nil)
		So(err, ShouldBeNil)

		ces := NewConfirmationEmailService("examplesite.com", fromAddress, "Test Sender Name", "info@infoexample.com", testEmailSender, testPreRegDb)
		Convey("With a valid unconfirmed group preregistration", func() {
			gpr := &GroupPreRegistration{
				PackName:               "Pack A",
				GroupName:              "Test Group",
				Council:                "1st Testingway",
				ContactLeaderEmail:     "test+plus@example.com",
				ContactLeaderFirstName: "MyFirst",
				ContactLeaderLastName:  "MyLast",
				EmailApprovalGivenAt:   time.Now(),
			}
			testPreRegDb.CreateRecord(gpr)

			Convey("Requesting an email send", func() {
				err := ces.RequestEmailConfirmation(gpr)
				Convey("Should return no error", func() {
					So(err, ShouldBeNil)
				})
				Convey("Should leave a single email", func() {
					So(len(testEmailSender.Emails), ShouldEqual, 1)
					email := testEmailSender.Emails[0]
					Convey("From my from address", func() {
						So(email.From, ShouldEqual, fromAddress)
					})
					Convey("Sent only to the contact leader", func() {
						So(email.To, ShouldResemble, []string{"test+plus@example.com"})
					})
					Convey("With the correct body", func() {
						So(string(email.Msg), ShouldResemble, fmt.Sprintf(
							`From: Test Sender Name <testsender@examplesending.com>
To: test+plus@example.com
Subject: Confirm CCJ16 Preregistration
Content-Type: text/plain; charset=UTF-8

Hi Scouter MyFirst MyLast,

Thank you for preregistering for CCJ16!  We hope you are as excited about this amazing camp as we are.  In order to confirm your CCJ16 preregistration, we ask that you confirm your email address by visiting the following page:

https://examplesite.com/confirmpreregistration?email=test%%2Bplus%%40example.com&token=%[2]s

If you are unable to click the link, please copy and paste it into your web browser.

If you wish to review your preregistration, please visit the following page:

https://examplesite.com/registration/%[1]s


Thanks again,
--
The CCJ16 team

If you have any questions, please contact us at info@infoexample.com`, gpr.SecurityKey, gpr.ValidationToken))
					})
					Convey("And the record records the email as having been sent", func() {
						//So(
					})
				})
			})
		})
	})
}
