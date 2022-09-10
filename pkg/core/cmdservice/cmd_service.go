package cmdservice

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
)

type ICmd interface {
	GetBashScript() string
}

type ICmdService interface {
	Run(cmd ICmd, preprocessor func(cm *exec.Cmd), args ...string) (string, error)
	Init()
	GetCmdRunnerService() ICmdRunnerService
}

type CmdServiceImpl struct {
	ICmdService

	shExecutorFileName string
	cmdsDirName        string

	Config           *core.Config
	CmdRunnerService ICmdRunnerService
}

func (c *CmdServiceImpl) GetCmdRunnerService() ICmdRunnerService {
	return c.CmdRunnerService
}

func (c *CmdServiceImpl) Init() {
	c.shExecutorFileName = c.Config.GetStr("sh", "executor")
	if len(c.shExecutorFileName) == 0 {
		c.shExecutorFileName = "C:\\Program Files\\Git\\bin\\sh.exe"
	}
	c.cmdsDirName = c.Config.GetDir("cmds")
}

func (c *CmdServiceImpl) Run(cmd ICmd, preprocessor func(cm *exec.Cmd), args ...string) (string, error) {
	bashFileName, err := c.prepareShFile(cmd)
	if err != nil {
		return "", err
	}
	cmdName := "bash"
	if c.CmdRunnerService.IsWindows() {
		cmdName = c.shExecutorFileName
	}
	args = append([]string{cmdName, bashFileName}, args...)
	r, err := c.CmdRunnerService.StartE(func(cm *exec.Cmd) {
		cm.Dir = filepath.Dir(bashFileName)
		if preprocessor != nil {
			preprocessor(cm)
		}
	}, args...)

	return string(r), err
}

func (c *CmdServiceImpl) prepareShFile(cmd ICmd) (string, error) {

	cmdFileName := filepath.Join(c.cmdsDirName, utils.GetType(cmd)[1:]+".sh") // исключим * в начале
	f := os.NewFile(3, cmdFileName)
	cmdFileName, err := filepath.Abs(f.Name())
	if err != nil {
		return "", nil
	}

	if !utils.FileExists(cmdFileName) {

		cmdFile, err := utils.CreateFileIfNotExists(cmdFileName)
		if err != nil {
			return "", nil
		}
		w := bufio.NewWriter(cmdFile)
		_, err = w.WriteString(cmd.GetBashScript())
		if err != nil {
			return "", nil
		}
		err = w.Flush()
		if err != nil {
			return "", nil
		}

	}

	return cmdFileName, nil
}
