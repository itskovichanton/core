package core

import (
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cast"
	"os"
	"time"
)

type AlertParams struct {
	Message, Subject string
	Attachments      []*os.File
	ByEmail, ByFR    bool
	Level            int
	Send             bool
}

type IErrorHandler interface {
	HandleWithMessage(err error, message interface{}, byFR bool) *AlertParams
	HandleWithCustomParams(err error, alertParamsPreprocessor func(alertParams *AlertParams)) *AlertParams
	Handle(err error, byFR bool) *AlertParams
	SendAlert(a *AlertParams)
}

type IParamsPostProcessor interface {
	Process(params *AlertParams)
}

type AlertParamsPostProcessorReducerImpl struct {
	IParamsPostProcessor

	lastSentMsgs *cache.Cache
	GetEntry     func(params *AlertParams) (string, time.Duration)
}

func (c *AlertParamsPostProcessorReducerImpl) Init() {
	c.lastSentMsgs = cache.New(cache.NoExpiration, cache.NoExpiration)
}

func (c *AlertParamsPostProcessorReducerImpl) Process(params *AlertParams) {

	key, sameKeysDelayInterval := c.getEntryParams(params)

	_, found := c.lastSentMsgs.Get(key)

	if found {
		params.Send = false
	} else {
		c.lastSentMsgs.Set(key, 1, sameKeysDelayInterval)
	}
}

func (c *AlertParamsPostProcessorReducerImpl) getEntryParams(params *AlertParams) (string, time.Duration) {
	if c.GetEntry != nil {
		return c.GetEntry(params)
	}
	return params.Subject + utils.MD5(params.Message), 5 * time.Minute
}

type ErrorHandlerImpl struct {
	IErrorHandler

	EmailService        IEmailService
	Config              *Config
	FRService           IFRService
	ParamsPostProcessor IParamsPostProcessor
	alertEmails         []string
}

func (c *ErrorHandlerImpl) Init() {
	c.alertEmails = cast.ToStringSlice(c.Config.Get("alerts", "emails"))
	if len(c.alertEmails) == 0 {
		c.alertEmails = []string{"a.itskovich@molbulak.com"}
	}
}

func (c *ErrorHandlerImpl) HandleWithMessage(err error, message interface{}, byFR bool) *AlertParams {
	return c.HandleWithCustomParams(err, func(p *AlertParams) {
		p.ByFR = byFR
		p.Message += "\n" + utils.ToJson(message)
	})
}

func (c *ErrorHandlerImpl) HandleWithCustomParams(err error, alertParamsPreprocessor func(alertParams *AlertParams)) *AlertParams {

	alertParams := &AlertParams{
		Message: utils.GetErrorFullInfo(err),
		Subject: c.Config.App.GetFullName() + "-[" + c.Config.Profile + "]",
		ByEmail: true,
		Level:   1,
		ByFR:    true,
		Send:    true,
	}
	alertParamsPreprocessor(alertParams)

	c.SendAlert(alertParams)

	return alertParams
}

func (c *ErrorHandlerImpl) Handle(err error, byFR bool) *AlertParams {
	return c.HandleWithCustomParams(err, func(ap *AlertParams) {
		ap.ByFR = byFR
	})
}

func (c *ErrorHandlerImpl) SendAlert(a *AlertParams) {

	if !a.Send {
		return
	}

	if len(a.Subject) == 0 {
		a.Subject = c.Config.App.Name
	}

	if c.ParamsPostProcessor != nil {
		c.ParamsPostProcessor.Process(a)
	}

	if !a.Send {
		return
	}

	var pr *Params
	if a.ByEmail {
		pr = &Params{
			From:    "finstart.mailer@molbulak.com",
			To:      c.alertEmails,
			Subject: a.Subject,
			Body:    a.Message,
			Template: &Template{
				TemplateFileName: c.Config.GetResourceFilePath("developer_email.html"),
				Data: struct {
					Msg string
				}{
					Msg: a.Message,
				},
			},
			AttachmentFileNames: []string{},
		}
	}

	if a.ByFR {
		p := Post{
			project: a.Subject,
			msg:     utils.ChopOffString(a.Message, 4000),
			level:   a.Level,
		}
		if len(a.Attachments) > 0 {
			p.attachment = a.Attachments[0]
		}
		c.FRService.PostMsg(&p)

	}

	if pr != nil {
		go c.EmailService.Send(pr)
	}

}
