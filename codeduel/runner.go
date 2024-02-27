package codeduel

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

func (r *Runner) run(code string, input []string) []ExecutionResult {
	// TODO
	return []ExecutionResult{}
}
