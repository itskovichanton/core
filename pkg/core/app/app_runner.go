package app

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava"
	"github.com/kardianos/service"
)

type IAppRunner interface {
	Run() error
}

type AppRunnerImpl struct {
	IAppRunner

	Config *core.Config
	App    IApp
}

func (c *AppRunnerImpl) Run() error {
	if c.Config.IsServceMode() {
		return c.runAsWindowsService()
	}
	return c.App.Run()
}

func (c *AppRunnerImpl) runAsWindowsService() error {

	options := make(service.KeyValue)
	optsFromYml := c.Config.Get("service", "options")
	if optsFromYml != nil {
		for k, v := range optsFromYml.(map[string]interface{}) {
			options[k] = v
		}
	}
	options["Restart"] = "on-success"
	options["SuccessExitStatus"] = "1 2 8 SIGKILL"

	srvName := c.Config.App.GetFullName() + "__service"
	srv := &goava.Service{
		Config: &service.Config{
			Name:        srvName,
			DisplayName: srvName,
			Description: c.Config.GetStr("service", "description"),
			Dependencies: []string{
				"Requires=network.target",
				"After=network-online.target syslog.target",
			},
			Option: options,
		},
		Action: func(logger service.Logger) {
			c.App.Run()
		},
	}

	return srv.Run()
}
