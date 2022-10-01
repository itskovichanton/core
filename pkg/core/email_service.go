package core

import (
	"bytes"
	"github.com/itskovichanton/core/pkg/core/email"
	"html/template"
	"net/smtp"
	"strings"
)

type IEmailService interface {
	Send(p *Params) error
}

type EmailServiceImpl struct {
	IEmailService

	Config *Config
}

type Params struct {
	From                string
	To                  []string
	Subject             string
	Body                string
	Template            *Template
	AttachmentFileNames []string
}

type Template struct {
	TemplateFileName string
	Data             interface{}
}

func (r *Params) parseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.Body = buf.String()
	return nil
}

func (r *EmailServiceImpl) Send(p *Params) error {

	var tmpl *template.Template
	var err error
	var wr *bytes.Buffer

	if p.Template != nil {
		tmpl, err = template.ParseFiles(p.Template.TemplateFileName)
		if err != nil {
			tmpl = nil
			//p.Body = p.
		} else {
			wr = bytes.NewBuffer(make([]byte, 256))
			err = tmpl.Execute(wr, p.Template.Data)
			if err != nil {
				wr = nil
				tmpl = nil
			}
		}
	} else {
		tmpl = nil
	}

	emailsvc := email.New(r.Config.GetStr("email", "address"))
	emailsvc.Auth = smtp.PlainAuth("",
		r.Config.GetStr("email", "username"),
		r.Config.GetStr("email", "password"),
		r.Config.GetStr("email", "host"),
	)
	emailsvc.Header = map[string]string{
		"Content-Type": "multipart/mixed; charset=UTF-8",
	}
	emailsvc.Template = tmpl

	msg := email.Message{
		From:        p.From,
		To:          strings.Join(p.To, ","),
		CC:          "",
		Subject:     p.Subject,
		Inlines:     nil,
		Attachments: nil,
	}

	if wr != nil {
		msg.BodyHTML = wr.String()
	} else {
		msg.BodyText = p.Body
	}

	return emailsvc.Send(&msg)
}
