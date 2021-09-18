package core

import (
	"bitbucket.org/itskovich/goava/pkg/goava/httputils"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Post struct {
	ids          []int
	project, msg string
	level        int
	attachment   *os.File
}

type IFRService interface {
	PostMsg(post *Post) (string, error)
}

type FRServiceImpl struct {
	IFRService
	Config     *Config
	HttpClient *http.Client
}

func (c *FRServiceImpl) PostMsg(a *Post) (string, error) {

	req, err := c.getPostHttpRequest(a)
	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return "", err
	}

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	//resp.Body.Close()
	//fmt.Println(resp.StatusCode)
	//fmt.Println(resp.Header)
	b := fmt.Sprint(body)
	return b, nil
}

func (c *FRServiceImpl) getPostHttpRequest(a *Post) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("msg", a.msg)
	writer.WriteField("project", a.project)
	writer.WriteField("level", strconv.Itoa(a.level))
	writer.WriteField("ids", strings.Join(utils.ToStringSliceInts(a.ids), ","))
	if a.attachment != nil {
		err := httputils.AddFile("attachment", a.attachment.Name(), writer)
		if err != nil {
			return nil, err
		}
	}

	err := writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.Config.FR.Url+"/postMsg", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
