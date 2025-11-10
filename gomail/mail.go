package gomail

import (
	"bytes"
	"path/filepath"
	"text/template"

	mail "gopkg.in/gomail.v2"
)

// GoMailConfig is a struct that represents the configuration for the GoMail.
type GoMailConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	From         string `mapstructure:"from"`
	RootTemplate string `mapstructure:"rootTemplate"`
}

// SendMailOption is a struct that represents the options for the SendMail function.
type SendMailOption struct {
	Subject     string
	Body        string
	To          []string
	From        string
	Attachments []string
}

// mailConf is a variable that represents the configuration for the GoMail.
var mailConf GoMailConfig

// Init is a function that initializes the GoMail.
// It takes a GoMailConfig and returns an error.
// This is used to initialize the GoMail.
func Init(conf GoMailConfig) {
	mailConf = conf
}

// NewEmail is a function that creates a new SendMailOption.
// It takes a subject, to, from, and attachments and returns a SendMailOption.
// This is used to create a new SendMailOption.
func NewEmail(
	subject string,
	to []string,
	from string,
	attachments []string,
) *SendMailOption {
	return &SendMailOption{
		Subject:     subject,
		To:          to,
		From:        from,
		Attachments: attachments,
	}
}

// LoadTemplate is a function that loads a template and renders it.
// It takes a template name and data and returns an error.
// This is used to load a template and render it.
func (o *SendMailOption) LoadTemplate(templateName string, data any) error {
	t, err := template.ParseFiles(filepath.Join(mailConf.RootTemplate, templateName))
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return err
	}
	o.Body = buf.String()
	return nil
}

// SendMail is a function that sends an email.
// It takes a body and returns an error.
// This is used to send an email.
func (o *SendMailOption) SendMail(body string) error {
	// Create a new dialer.
	d := mail.NewDialer(
		mailConf.Host,
		mailConf.Port,
		mailConf.Username,
		mailConf.Password,
	)
	// Create a new message.
	m := mail.NewMessage()

	// Set the headers.
	m.SetHeader("From", o.From)
	m.SetHeader("To", o.To...)
	m.SetHeader("Subject", o.Subject)
	m.SetHeader("Content-Type", "text/html; charset=utf-8")

	// Set the body.
	if len(body) > 0 {
		m.SetBody("text/html", body)
	} else {
		m.SetBody("text/html", o.Body)
	}

	// Set the attachments.
	for _, attachment := range o.Attachments {
		m.Attach(attachment)
	}

	// Send the email.
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
