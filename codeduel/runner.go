package codeduel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Runner struct {
	url string
}

type ApiResult struct {
	Result []ExecutionResult `json:"result"`
}

type ExecutionResult struct {
	Output string `json:"output"`
	Error  string `json:"errors"`
	Status int64  `json:"status"`
}

func NewRunner(url string) Runner {
	return Runner{url}
}

func (r *Runner) AvailableLanguages() ([]string, error) {
	response, err := http.Get(r.url + "/api/v1/languages")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	errorResult := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "",
	}
	err = json.Unmarshal(bytes, &errorResult)
	if err != nil {
		return nil, err
	}
	if errorResult.Error {
		return nil, fmt.Errorf(errorResult.Message)
	}
	var result struct {
		Result struct {
			Languages []string `json:"languages"`
		} `json:"result"`
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result.Result.Languages, nil
}

func (r *Runner) Run(language, code string, input []string) ([]ExecutionResult, error) {
	raw, _ := json.Marshal(struct {
		Language string   `json:"language"`
		Code     string   `json:"code"`
		Input    []string `json:"input"`
	}{
		Language: language,
		Code:     code,
		Input:    input,
	})
	body := bytes.NewBuffer(raw)
	response, err := http.Post(r.url+"/api/v1/run", "application/json", body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	errorResult := struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}{
		Error:   false,
		Message: "",
	}
	err = json.Unmarshal(bytes, &errorResult)
	if err != nil {
		return nil, err
	}
	if errorResult.Error {
		return nil, fmt.Errorf(errorResult.Message)
	}
	var result ApiResult
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result.Result, nil
}
