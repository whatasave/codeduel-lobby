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

type ExecutionResult struct {
	Output     string `json:"output"`
	Error      string `json:"errors"`
	Terminated bool   `json:"terminated"`
}

func NewRunner(url string) Runner {
	return Runner{url}
}

func (r *Runner) Run(code string, input []string) ([]ExecutionResult, error) {
	raw, _ := json.Marshal(struct {
		Code  string   `json:"code"`
		Input []string `json:"input"`
	}{
		Code:  code,
		Input: input,
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
	var errorResult struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(bytes, &errorResult)
	if err != nil {
		return nil, err
	}
	if errorResult.Error {
		return nil, fmt.Errorf(errorResult.Message)
	}
	var result []ExecutionResult
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
