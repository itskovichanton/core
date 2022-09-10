package core

import (
	"bitbucket.org/itskovich/goava/pkg/goava/httputils"
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

type Post struct {
	project, msg string
	level        int
	attachment   *os.File
}

type IFRService interface {
	PostMsg(post *Post)
}

type FRServiceImpl struct {
	IFRService

	Config     *Config
	HttpClient *http.Client
}

func (c *FRServiceImpl) PostMsg(a *Post) {
	if c.Config.FR != nil {
		go func() { c.postMsg(a, c.Config.FR) }()
	}
	//if a.level > 2 && c.Config.FR2 != nil {
	//	go func() { c.postMsg(a, c.Config.FR2) }()
	//}
}

func (c *FRServiceImpl) postMsg(a *Post, fr *FR) (string, error) {
	req, err := c.getPostHttpRequest(a, fr)
	if err != nil {
		return "", err
	}
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

func (c *FRServiceImpl) getPostHttpRequest(a *Post, fr *FR) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("msg", a.msg)
	writer.WriteField("project", a.project)
	writer.WriteField("level", strconv.Itoa(a.level))
	//writer.WriteField("ids", strings.Join(utils.ToStringSliceInts([]int{fr.DeveloperId}), ","))
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

	req, err := http.NewRequest("POST", fr.Url+"/postMsg", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}
