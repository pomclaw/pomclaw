package tools

// Result is kept for potential future use with async tools or other features.
// Currently not used by InvokableTool implementations.
type Result struct {
	ForLLM  string `json:"for_llm"`
	ForUser string `json:"for_user,omitempty"`
	Silent  bool   `json:"silent"`
	IsError bool   `json:"is_error"`
	Async   bool   `json:"async"`
	Err     error  `json:"-"`
}
