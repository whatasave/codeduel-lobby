package codeduel

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Runner struct {
	url string
}

type ExecutionResult struct {
	Output     string `json:"output"`
	Error      string `json:"error"`
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
	var result []ExecutionResult
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
