package frmclient

import (
	"bitbucket.org/itskovich/goava/pkg/goava/errs"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ExecuteWidthFrmAPI(resp *http.Response, resultStubGetter func() interface{}) (interface{}, error) {
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respModel Result
	if resultStubGetter == nil {
		respModel.Res = map[string]interface{}{}
	} else {
		respModel.Res = resultStubGetter()
	}
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
