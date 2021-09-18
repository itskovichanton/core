package cmdservice

import (
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"strconv"
	"strings"
)

type IOSFunctionsService interface {
	IsPortBusy(port int) bool
	GetCmdService() ICmdService
	GetNslookupIPs(url string) ([]string, error)
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

func (c *OSFunctionsServiceImpl) GetNslookupIPs(url string) ([]string, error) {
	r, err := c.CmdService.Run(&NslookupCmd{}, url)
	if err != nil {
		return nil, err
	}
	return utils.RetrieveIPs(r[strings.Index(r, url)+len(url):]), nil
}
