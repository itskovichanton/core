package core

import (
	"git.molbulak.ru/a.itskovich/molbulak-services-golang/pkg/core/validation"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
)

type Caller struct {
	IP       string
	Version  *Version
	Type     string
	Language string
	AuthArgs *AuthArgs
	Session  *Session
}

type AuthArgs struct {
	Username, Password string
	SessionToken       string
}

type Version struct {
	Code int
	Name string
}

type CallParams struct {
	Request    interface{} `json:"-"`
	Parameters map[string][]interface{}
	URL        string
	Caller     *Caller
	Raw        string
}

func (c *CallParams) GetParamFloat32(key string, failValue float32) float32 {
	v, err := validation.CheckFloat32(key, c.GetParamStr(key))
	if err == nil {
		return v
	}
	return failValue
}

func (c *CallParams) GetParamFloat64(key string, failValue float64) float64 {
	v, err := validation.CheckFloat64(key, c.GetParamStr(key))
	if err == nil {
		return v
	}
	return failValue
}

func (c *CallParams) GetParamInt(key string, failValue int) int {
	v, err := validation.CheckInt(key, c.GetParamStr(key))
	if err == nil {
		return v
	}
	return failValue
}

func (c *CallParams) GetParamBool(key string, failValue bool) bool {
	v, err := validation.CheckBool(key, c.GetParamStr(key))
	if err == nil {
		return v
	}
	return failValue
}

func (c *CallParams) GetParamInt64(key string, failValue int64) int64 {
	v, err := validation.CheckInt64(key, c.GetParamStr(key))
	if err == nil {
		return v
	}
	return failValue
}

func (c *CallParams) GetParamStr(key string) string {
	p := c.GetParam(key)
	if len(p) == 0 {
		return ""
	}
	return cast.ToString(p[0])
}

func (c *CallParams) GetParam(key string) []interface{} {
	return c.Parameters[key]
}

func (c *CallParams) SetParam(key string, value interface{}) {
	c.Parameters[key] = []interface{}{value}
}

func (c *CallParams) GetStrParams() map[string]string {
	r := map[string]string{}
	for k, v := range c.Parameters {
		r[k] = cast.ToString(v[0])
	}
	return r
}

func (c *CallParams) GetParamsUsingFirstValue() map[string]interface{} {
	r := map[string]interface{}{}
	for k, v := range c.Parameters {
		r[k] = v[0]
	}
	return r
}

func (c *CallParams) Context() echo.Context {
	return c.Request.(echo.Context)
}

type Session struct {
	Token   string   `json:"token"`
	Account *Account `json:"account"`
}

const RoleAdmin = "admin"
const RoleUser = "user"

type Account struct {
	ID           int64  `json:"id"`
	CID          int64  `json:"cid"`
	MCLID        int64  `json:"mclId"`
	Username     string `json:"username"`
	Lang         string `json:"lang"`
	FullName     string `json:"fullName"`
	SessionToken string `json:"sessionToken,omitempty"`
	Role         string `json:"role"`
	Password     string `json:"password"`
	IP           string `json:"ip"`
}
