package app

import (
	"github.com/kardianos/service"
	"log"
)

type IService interface {
	Run() error
}

type Service struct {
	Config *service.Config
	Action func()
}

type program struct {
	exit   chan struct{}
	logger service.Logger
	action func()
}

func (p *program) Start(s service.Service) error {
	if service.Interactive() {
		p.logger.Info("Running in terminal.")
	} else {
		p.logger.Info("Running under service manager.")
	}
	p.exit = make(chan struct{})

	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) run() {
	p.action()
}

func (p *program) Stop(s service.Service) error {
	// Any work in Stop should be quick, usually a few seconds at most.
	p.logger.Info("I'm Stopping!")
	close(p.exit)
	return nil
}

// Service setup.
//   Define service config.
//   Create the service.
//   Setup the logger.
//   Handle service controls (optional).
//   Run the service.
func (c *Service) Run() error {

	prg := &program{action: c.Action}
	srv, err := service.New(prg, c.Config)
	if err != nil {
		log.Fatal(err)
	}
	errs := make(chan error, 5)
	prg.logger, err = srv.Logger(errs)
	if err != nil {
		return err
	}

	go func() {
		for {
			err := <-errs
			if err != nil {
				log.Print(err)
			}
		}
	}()

	err = service.Control(srv, "")
	if err != nil {
		log.Printf("Valid actions: %q\n", service.ControlAction)
		log.Fatal(err)
		return err
	}

	err = srv.Run()
	if err != nil {
		prg.logger.Error(err)
		return err
	}

	return nil
}
