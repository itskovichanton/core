package app

import (
	"bitbucket.org/itskovich/core/pkg/core"
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
	if c.Config.GetBool("service", "enabled") {
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

	srvName := c.Config.App.GetFullName() + "__service"
	svcConfig := &service.Config{
		Name:        srvName,
		DisplayName: srvName,
		Description: c.Config.GetStr("service", "description"),
		Dependencies: []string{
			"Requires=network.target",
			"After=network-online.target syslog.target"},
		Option: options,
	}

	srv := &Service{
		Config: svcConfig,
		Action: func() {
			c.App.Run()
		},
	}

	return srv.Run()
}
