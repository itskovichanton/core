package cmdservice

import (
	"fmt"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"io"
	"os/exec"
	"runtime"
	"strings"
)

type ICmdRunnerService interface {
	StartE(preprocessor func(cm *exec.Cmd), arg ...string) ([]byte, error)
	Start(arg ...string) *exec.Cmd
	Run(writer func(stdin io.WriteCloser)) (string, error)
	Cmd(input string) (string, error)
	IsWindows() bool
}

type CmdRunnerServiceImpl struct {
	ICmdRunnerService
}

func (c *CmdRunnerServiceImpl) Cmd(input string) (string, error) {
	return c.Run(func(stdin io.WriteCloser) {
		io.WriteString(stdin, input+"\n")
	})
}

func (c *CmdRunnerServiceImpl) Run(writer func(stdin io.WriteCloser)) (string, error) {

	if c.IsWindows() {
		cmd := exec.Command("cmd", "/k")

		stdin, err := cmd.StdinPipe()
		if err != nil {
			return "", err
		}

		go func() {
			defer stdin.Close()
			writer(stdin)
		}()

		r, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}

		rStr := string(r)
		cmd.Process.Kill()
		rStr = rStr[strings.Index(rStr, "\n")+1:]
		end := strings.Index(rStr, "\r\n\r\n")
		if end < 0 {
			end = strings.Index(rStr, "\n")
			if end > 0 {
				rStr = rStr[:end]
			}
		}
		return strings.TrimSpace(rStr), nil
	}

	return "", nil
}

func (c *CmdRunnerServiceImpl) StartE(preprocessor func(cm *exec.Cmd), arg ...string) ([]byte, error) {
	cm := c.Start(arg...)
	if preprocessor != nil {
		preprocessor(cm)
	}
	r, err := cm.Output()
	if err != nil {
		switch ex := err.(type) {
		case *exec.ExitError:
			return nil, errs.NewBaseErrorFromCauseMsg(ex, fmt.Sprintf("%v, Stderr: %v", ex.Error(), string(ex.Stderr)))
		}
		return nil, err
	}
	return r, nil
}

func (c *CmdRunnerServiceImpl) Start(arg ...string) *exec.Cmd {

	if c.IsWindows() {
		arg = append([]string{"cmd", "/k"}, arg...)
		return exec.Command(arg[0], arg[1:]...)
	}

	return exec.Command(arg[0], arg[1:]...)
}

func (c *CmdRunnerServiceImpl) IsWindows() bool {
	return strings.EqualFold(runtime.GOOS, "windows")
}
