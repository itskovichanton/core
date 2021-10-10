package core

import (
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"flag"
	"fmt"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	ProfilePROD = "prod"
	ProfileDEV  = "dev"
)

type AppInfo struct {
	Version string
	Name    string
}

func (a *AppInfo) GetFullName() string {
	return strings.Join([]string{a.Name, a.Version}, "-")
}

type Props struct {
	MainServiceUrl string
	Apis           *Apis
	ResourcesPath  string
}

type Apis struct {
}

type Actions struct {
	DefaultLang      string
	ApiMethodSystems []string
}

func (c *Actions) IsSystemPermitted(system string) bool {
	if c.ApiMethodSystems[0] == "ANY" {
		return true
	}
	return utils.HasElem(c.ApiMethodSystems, system)
}

type FR struct {
	Url         string
	DeveloperId int
}

type Server struct {
	Port               int
	Http               *Http
	GrpcPort           int
	EnableThrottleMode bool
	EnableCORS         bool
}

type Multipart struct {
	MaxRequestSizeBytes string
}

func (c Multipart) GetMaxRequestSizeBytes() (uint64, error) {
	return utils.ParseMemory(c.MaxRequestSizeBytes)
}

type Ssl struct {
	CertFile string
	KeyFile  string
	Enabled  bool
	Network  string
}

type Http struct {
	Multipart *Multipart
	Ssl       *Ssl
}

type Config struct {
	Profile  string
	Server   *Server
	FR, FR2  *FR
	App      *AppInfo
	Props    *Props
	Actions  *Actions
	Settings map[string]interface{}
}

func (c *Config) GetLogsDir() string {
	return c.GetDir("logs")
}

func (c *Config) GetFileStorageDir() string {
	return c.GetDir("filestorage")
}

func (c *Config) GetSettingsDir() string {
	return c.GetDir("settings")
}

func (c *Config) GetSettingsFile() (*os.File, error) {
	return utils.CreateFileIfNotExists(filepath.Join(c.GetSettingsDir(), "settings.yml"))
}

func (c *Config) GetSecurityFile() (*os.File, error) {
	return utils.CreateFileIfNotExists(filepath.Join(c.GetSettingsDir(), "security.yml"))
}

func (c *Config) GetAppName() string {
	return c.App.Name + "-" + c.App.Version + "-" + "[" + c.Profile + "]"
}

func (c *Config) GetBaseWorkDir() string {
	return filepath.Join(c.App.Name, "workdir")
}

func (c *Config) GetOnBaseWorkDir(s ...string) string {
	s = append([]string{c.GetBaseWorkDir()}, s...)
	r := filepath.Join(s...)
	os.MkdirAll(r, os.ModeDir)
	return r
}

func (c *Config) GetDir(s ...string) string {
	s = append([]string{c.GetBaseWorkDir(), c.Profile}, s...)
	r := filepath.Join(s...)
	os.MkdirAll(r, os.ModeDir)
	return r
}

func (c *Config) IsProfileProd() bool {
	return strings.EqualFold("prod", c.Profile)
}

func (c *Config) GetTempFilesStorageDir() string {
	dir := filepath.Join(c.GetFileStorageDir(), "tmp")
	os.MkdirAll(dir, os.ModeDir)
	return dir
}

func (c *Config) GetTempFile(pattern string) (*os.File, error) {
	return ioutil.TempFile(c.GetTempFilesStorageDir(), pattern)
}

func (c *Config) GetStr(path ...string) string {
	return cast.ToString(c.Get(path...))
}

func (c *Config) GetBool(path ...string) bool {
	return c.GetBoolWithDefaultValue(false, path...)
}

func (c *Config) GetBoolWithDefaultValue(defaultValue bool, path ...string) bool {
	r, err := cast.ToBoolE(c.Get(path...))
	if err != nil {
		return defaultValue
	}
	return r
}

func (c *Config) Get(path ...string) interface{} {
	r := c.Settings
	for _, p := range path {
		switch r[p].(type) {
		case map[string]interface{}:
			r = r[p].(map[string]interface{})
			break
		default:
			return r[p]
		}
	}
	return r
}

func (c *Config) GetResourceFilePath(resourcePath string) string {
	return filepath.Join(c.Props.ResourcesPath, resourcePath)
}

type IConfigService interface {
	LoadConfig() (*Config, error)
}

type ConfigServiceImpl struct {
	IConfigService
}

func (c *ConfigServiceImpl) LoadConfig() (*Config, error) {

	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	if len(os.Args) >= 2 {
		viper.AddConfigPath(fmt.Sprintf("%v/%v", os.Args[1], "config"))
	}

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var C Config
	C.Settings = viper.AllSettings()
	if err := viper.Unmarshal(&C); err != nil {
		return nil, err
	}

	profiledProps := viper.Sub(C.Profile)
	if err := profiledProps.Unmarshal(&C.Props); err != nil {
		return nil, err
	}

	for k, v := range profiledProps.AllSettings() {
		C.Settings[k] = v
	}

	c.initDirs(&C)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	return &C, nil
}

func (c *ConfigServiceImpl) initDirs(cfg *Config) {
	os.MkdirAll(cfg.GetBaseWorkDir(), os.ModePerm)
}
