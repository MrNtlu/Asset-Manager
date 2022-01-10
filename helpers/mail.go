package helpers

import (
	"net/smtp"
	"os"

	"github.com/jordan-wright/email"
)

func SendForgotPasswordEmail(token, mail string) error {
	url := (os.Getenv("BASE_URL") + "/auth/confirm-password-reset?token=" + token + "&mail=" + mail)

	e := email.NewEmail()
	e.From = "Asset Manager <" + os.Getenv("FROM_MAIL") + ">"
	e.To = []string{mail}
	e.Subject = "Forgot Password"
	e.HTML = []byte(
		`<p>If you click the link your password will reset.</p>
		<b><a href="` + url + `">Reset Password</a></b>
		<br>
		<br>
		<small>If it wasn't you, please change your password.</small>`,
	)
	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", os.Getenv("FROM_MAIL"), os.Getenv("FROM_MAIL_PASSWORD"), "smtp.gmail.com"))
	if err != nil {
		return err
	}

	return nil
}

func SendPasswordChangedEmail(content, mail string) error {
	e := email.NewEmail()
	e.From = "Asset Manager <" + os.Getenv("FROM_MAIL") + ">"
	e.To = []string{mail}
	e.Subject = "Password Reset"
	e.Text = []byte("Your password changed. New password: " + content + " \nDon't forget to change your password!")
	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", os.Getenv("FROM_MAIL"), os.Getenv("FROM_MAIL_PASSWORD"), "smtp.gmail.com"))
	if err != nil {
		return err
	}

	return nil
}
