package logger

import (
	"bitbucket.org/itskovich/core/pkg/core"
	"bitbucket.org/itskovich/core/pkg/core/validation"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"encoding/json"
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/patrickmn/go-cache"
	"log"
	"path/filepath"
	"time"
)

type ILoggerService interface {
	GetLogger(name string, profile string, maxHistory int, loggerProvider func(string) *log.Logger) *log.Logger
	GetFileLogger(name string, profile string, maxHistory int) *log.Logger
	GetDefaultFileOpsLogger() *log.Logger
	GetDefaultActionsLogger() *log.Logger
	GetLogFileName(name string, profile string) string
}

type LoggerServiceImpl struct {
	ILoggerService

	Config *core.Config
	Cache  *cache.Cache
}

func (c *LoggerServiceImpl) GetDefaultActionsLogger() *log.Logger {
	logMaxDays, err := validation.CheckInt("actions-logMaxDays", c.Config.GetStr("actions", "logmaxdays"))
	if err != nil {
		logMaxDays = 1
	}
	return c.GetFileLogger("actions", "", logMaxDays)
}

func (c *LoggerServiceImpl) GetDefaultFileOpsLogger() *log.Logger {
	return c.GetFileLogger("ops", "", 90)
}

func (c *LoggerServiceImpl) GetFileLogger(name string, profile string, maxHistory int) *log.Logger {

	return c.GetLogger(name, profile, maxHistory, func(profile string) *log.Logger {

		filename := c.GetLogFileName(name, profile)
		var err error
		var l *rotatelogs.RotateLogs
		if maxHistory == 0 {
			l, err = rotatelogs.New(
				filename,
				rotatelogs.WithMaxAge(-1),
				rotatelogs.WithRotationTime(time.Hour*24),
			)
		} else {
			l, err = rotatelogs.New(
				filename,
				rotatelogs.WithMaxAge(-1),
				rotatelogs.WithRotationTime(time.Hour*24),
				rotatelogs.WithRotationCount(maxHistory),
			)
		}
		if err != nil {
			log.Fatalf("Failed to Initialize Log File %s", err)
		}
		return log.New(l, "", 0)
	})
}

func Print(logger *log.Logger, ld map[string]interface{}) error {
	_, ignoreExists := ld["ignore"]
	_, rExists := ld["r"]
	_, errExists := ld["err"]
	_, rspExists := ld["rsp"]
	if ignoreExists || (!rExists && !errExists && !rspExists) {
		return nil
	}
	delete(ld, "chopoff-disabled")
	Field(ld, "tm", utils.CurrentTimeMillis())
	jsonBytes, err := json.Marshal(ld)
	if err != nil {
		return err
	}
	logger.Println(string(jsonBytes))
	println(string(jsonBytes))
	delete(ld, "l")
	return nil
}

func Ignore(ld map[string]interface{}) map[string]interface{} {
	return Field(ld, "ignore", true)
}

func Args(ld map[string]interface{}, args interface{}) map[string]interface{} {
	switch ex := args.(type) {
	case *core.CallParams:
		argsX := core.CallParams{
			Request:    nil,
			Parameters: ex.Parameters,
			URL:        ex.URL,
			Caller:     ex.Caller,
			Raw:        ex.Raw,
		}
		if _, callerExists := ld["c"]; callerExists {
			argsX.Caller = nil
		}
		args = argsX
		break
	}

	return Field(ld, "p", args)
}

func Add(ld map[string]interface{}, field string, args string) map[string]interface{} {
	v := ld[field]
	if v == nil {
		Field(ld, field, args)
		return ld
	}

	vStr := ""
	if v != nil {
		switch e := v.(type) {
		case string:
			vStr = e
		default:
			vStr = utils.ToJson(v)
		}
	}

	Field(ld, field, vStr+args)
	return ld
}

func Field(ld map[string]interface{}, field string, args interface{}) map[string]interface{} {
	_, chopOffDisabled := ld["chopoff-disabled"]
	if args != nil {
		switch v := args.(type) {
		case string:
			if !chopOffDisabled {
				args = utils.ChopOffString(fmt.Sprintf("%v", v), 1000)
			}
			break
		case int64, int, float32, float64:
			break
		default:
			argsJsonBytes, err := json.Marshal(args)
			if err == nil {
				argStr := string(argsJsonBytes)
				if chopOffDisabled {
					args = argStr
				} else {
					args = utils.ChopOffString(argStr, 1000)
				}
			}
		}
	}
	ld[field] = args
	return ld
}

func Result(ld map[string]interface{}, result interface{}) map[string]interface{} {
	return Field(ld, "r", result)
}

func Err(ld map[string]interface{}, e interface{}) map[string]interface{} {
	var x interface{}
	switch v := e.(type) {
	case error:
		x = utils.GetErrorFullInfo(v)
	default:
		x = v
	}
	return Field(ld, "err", x)
}

func Action(ld map[string]interface{}, a interface{}) map[string]interface{} {
	return Field(ld, "a", a)
}

func Subject(ld map[string]interface{}, sbj interface{}) map[string]interface{} {
	return Field(ld, "sbj", sbj)
}

func NewLD() map[string]interface{} {
	return map[string]interface{}{}
}

func (c *LoggerServiceImpl) GetLogger(name string, profile string, maxHistory int, loggerProvider func(string) *log.Logger) *log.Logger {
	if len(profile) == 0 {
		profile = c.Config.Profile
	}
	key := fmt.Sprintf("logger:%v-%v", name, profile)
	cached, found := c.Cache.Get(key)
	if found {
		return cached.(*log.Logger)
	}

	logger := loggerProvider(profile)

	c.Cache.Set(key, logger, cache.NoExpiration)
	return logger
}

func (c *LoggerServiceImpl) GetLogFileName(name string, profile string) string {
	return filepath.Join(c.Config.GetLogsDir(), fmt.Sprintf("%v-%v-%v", c.Config.App.Name, name, profile)+"-%d-%m-%Y.txt")
}

func ErrWithLocation(ld map[string]interface{}, e interface{}, strNumber int) map[string]interface{} {
	ld["l"] = strNumber
	return Err(ld, e)
}

func DisableSetChopOffFields(ld map[string]interface{}) map[string]interface{} {
	ld["chopoff-disabled"] = true
	return ld
}
