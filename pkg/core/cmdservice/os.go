package cmdservice

import (
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"strconv"
	"strings"
	"time"
)

type IOSFunctionsService interface {
	IsPortBusy(port int) bool
	GetCmdService() ICmdService
	GetNslookupIPs(url string) ([]string, error)
	KillByPort(port int, hard bool)
	RestartByPort(port int, killHard bool, starterArgs ...string) ([]byte, error)
}

type OSFunctionsServiceImpl struct {
	IOSFunctionsService

	CmdService ICmdService
}

func (c *OSFunctionsServiceImpl) GetCmdService() ICmdService {
	return c.CmdService
}

func (c *OSFunctionsServiceImpl) IsPortBusy(port int) bool {
	r, _ := c.CmdService.Run(&NetStatCmd{}, strconv.Itoa(port))
	return len(r) > 0
}

func (c *OSFunctionsServiceImpl) RestartByPort(port int, killHard bool, starterArgs ...string) ([]byte, error) {
	c.KillByPort(port, killHard)
	return c.CmdService.GetCmdRunnerService().StartE(starterArgs...)
}

func (c *OSFunctionsServiceImpl) KillByPort(port int, hard bool) {
	for {
		var cmd ICmd
		if hard {
			cmd = &KillHardByPortCmd{}
		} else {
			cmd = &KillByPortCmd{}
		}
		c.CmdService.Run(cmd, strconv.Itoa(port))
		time.Sleep(1 * time.Second)
		if !c.IsPortBusy(port) {
			return
		}
	}
}

func (c *OSFunctionsServiceImpl) GetNslookupIPs(url string) ([]string, error) {
	r, err := c.CmdService.Run(&NslookupCmd{}, url)
	if err != nil {
		return nil, err
	}
	return utils.RetrieveIPs(r[strings.Index(r, url)+len(url):]), nil
}
