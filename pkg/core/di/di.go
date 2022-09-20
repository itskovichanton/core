package di

import (
	"github.com/itskovichanton/core/pkg/core"
	"github.com/itskovichanton/core/pkg/core/app"
	"github.com/itskovichanton/core/pkg/core/cmdservice"
	"github.com/itskovichanton/core/pkg/core/logger"
	"github.com/itskovichanton/goava/pkg/goava"
	"github.com/patrickmn/go-cache"
	"go.uber.org/dig"
	"net/http"
)

type DI struct {
	Container *dig.Container
}

func (c *DI) InitDI(container *dig.Container) {
	c.Container = c.buildContainer(container)
}

func (c *DI) buildContainer(container *dig.Container) *dig.Container {

	container.Provide(c.NewCache)
	container.Provide(c.NewLoggerService)
	container.Provide(c.NewHttpClient)
	container.Provide(c.NewConfigService)
	container.Provide(c.NewConfig)
	container.Provide(c.NewFRService)
	container.Provide(c.NewEmailService)
	container.Provide(c.NewErrorHandler)
	container.Provide(c.NewAlertParamsPostProcessor)
	container.Provide(c.NewGenerator)
	container.Provide(c.NewCmdRunnerService)
	container.Provide(c.NewCmdService)
	container.Provide(c.NewOSFunctionsService)
	container.Provide(c.NewAppRunner)

	return container
}

func (c *DI) NewAppRunner(a app.IApp, config *core.Config) app.IAppRunner {
	return &app.AppRunnerImpl{
		Config: config,
		App:    a,
	}
}

func (c *DI) NewOSFunctionsService(cmdService cmdservice.ICmdService) cmdservice.IOSFunctionsService {
	return &cmdservice.OSFunctionsServiceImpl{
		CmdService: cmdService,
	}
}

func (c *DI) NewCmdService(cmdRunner cmdservice.ICmdRunnerService, config *core.Config) cmdservice.ICmdService {
	r := &cmdservice.CmdServiceImpl{
		Config:           config,
		CmdRunnerService: cmdRunner,
	}
	r.Init()
	return r
}

func (c *DI) NewCmdRunnerService() cmdservice.ICmdRunnerService {
	return &cmdservice.CmdRunnerServiceImpl{}
}

func (c *DI) NewGenerator() goava.IGenerator {
	r := &goava.GeneratorImpl{}
	r.Reset()
	return r
}

func (c *DI) NewConfig(configService core.IConfigService) (*core.Config, error) {
	return configService.LoadConfig()
}

func (c *DI) NewConfigService() core.IConfigService {
	return &core.ConfigServiceImpl{}
}

func (c *DI) NewHttpClient() *http.Client {
	return &http.Client{
		//Timeout: 3 * time.Minute,
	}
}

func (c *DI) NewEmailService(config *core.Config) core.IEmailService {
	return &core.EmailServiceImpl{
		Config: config,
	}
}

func (c *DI) NewFRService(httpClient *http.Client, config *core.Config) core.IFRService {
	return &core.FRServiceImpl{
		HttpClient: httpClient,
		Config:     config,
	}
}

func (c *DI) NewCache() *cache.Cache {
	return cache.New(cache.NoExpiration, cache.NoExpiration)
}

func (c *DI) NewLoggerService(config *core.Config) logger.ILoggerService {
	return &logger.LoggerServiceImpl{
		Config: config,
		Cache:  c.NewCache(),
	}
}

func (c *DI) NewAlertParamsPostProcessor() core.IParamsPostProcessor {
	r := &core.AlertParamsPostProcessorReducerImpl{}
	r.Init()
	return r
}

func (c *DI) NewErrorHandler(paramsPostProcessor core.IParamsPostProcessor, emailService core.IEmailService, config *core.Config, frservice core.IFRService) core.IErrorHandler {
	r := &core.ErrorHandlerImpl{
		ParamsPostProcessor: paramsPostProcessor,
		EmailService:        emailService,
		Config:              config,
		FRService:           frservice,
	}
	r.Init()
	return r
}
