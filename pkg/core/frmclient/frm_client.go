package frmclient

import (
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"bitbucket.org/itskovich/goava/pkg/goava/utils"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ExecuteWidthFrmAPI(resp *http.Response, resultGetter interface{}) (interface{}, error) {

	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resultGetter != nil {
		println(utils.GetType(resultGetter))
		switch e := resultGetter.(type) {
		case *bytes.Buffer:
			if resp.StatusCode == 200 {
				_, err := e.Write(respBytes)
				if err != nil {
					return nil, &FrmClientError{
						BaseError: *errs.NewBaseErrorFromCause(err),
					}
				} else {
					return resultGetter, nil
				}
			} else {
				resultGetter = nil
			}
			break
		}
	} else {
		resultGetter = map[string]interface{}{}
	}

	var respModel Result
	respModel.Res = resultGetter
	err = json.Unmarshal(respBytes, &respModel)
	if err != nil {
		return nil, err
	}

	if respModel.Err != nil {
		return nil, &FrmClientError{
			BaseError: *errs.NewBaseErrorWithReason(respModel.Err.Message, respModel.Err.Reason),
			Err:       respModel.Err,
		}
	}

	return respModel.Res, nil

}

type FrmClientError struct {
	errs.BaseError
	Err *Err
}

type Err struct {
	Reason  string `json:"reason"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Cause   string `json:"сause,omitempty"`
}

type Result struct {
	Res             interface{} `json:"result,omitempty"`
	Err             *Err        `json:"error,omitempty"`
	ExecutionTimeMs int64       `json:"executionTimeMs"`
}
