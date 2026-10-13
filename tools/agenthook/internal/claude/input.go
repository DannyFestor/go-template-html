package claude

// input is the JSON Claude Code writes to a hook's standard input. Only the fields the hooks read are listed.
//
//nolint:tagliatelle // Claude Code names its hook input fields in snake_case.
type input struct {
	SessionID string    `json:"session_id"`
	Cwd       string    `json:"cwd"`
	ToolName  string    `json:"tool_name"`
	ToolInput toolInput `json:"tool_input"`
}

//nolint:tagliatelle // Claude Code names its tool input fields in snake_case.
type toolInput struct {
	FilePath     string `json:"file_path"`
	OldString    string `json:"old_string"`
	NewString    string `json:"new_string"`
	Content      string `json:"content"`
	NotebookPath string `json:"notebook_path"`
	NewSource    string `json:"new_source"`
	Command      string `json:"command"`
}
