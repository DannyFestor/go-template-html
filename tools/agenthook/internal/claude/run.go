package claude

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/hook"
)

var errUnknownEvent = errors.New("unknown hook event")

func Run(event string, stdin io.Reader, stdout io.Writer, environ []string) error {
	var hookInput input

	err := json.NewDecoder(stdin).Decode(&hookInput)
	if err != nil {
		return fmt.Errorf("decode hook input: %w", err)
	}

	session := hook.NewSession(hookInput.SessionID, environ)

	switch event {
	case "pre-tool-use":
		return preToolUse(session, hookInput, stdout)
	case "post-tool-use":
		return postToolUse(session, hookInput, stdout)
	case "stop":
		return stop(session, hookInput, stdout)
	default:
		return fmt.Errorf("%w %q", errUnknownEvent, event)
	}
}

func preToolUse(session hook.Session, hookInput input, stdout io.Writer) error {
	if hookInput.ToolName == "Bash" {
		return writePermission(stdout, session.BeforeCommand(hookInput.ToolInput.Command, hookInput.Cwd))
	}

	edit, isEdit := fileEdit(hookInput)
	if !isEdit {
		return nil
	}

	judged, err := session.BeforeEdit(edit)
	if err != nil {
		return fmt.Errorf("judge the edit of %s: %w", edit.Path, err)
	}

	return writePermission(stdout, judged)
}

func postToolUse(session hook.Session, hookInput input, stdout io.Writer) error {
	edit, isEdit := fileEdit(hookInput)
	if !isEdit {
		return nil
	}

	feedback, err := session.AfterEdit(edit.Path)
	if err != nil {
		return fmt.Errorf("check the edit of %s: %w", edit.Path, err)
	}

	return writeBlock(stdout, feedback)
}

func stop(session hook.Session, hookInput input, stdout io.Writer) error {
	failure, err := session.EndOfTurn(hookInput.Cwd)
	if err != nil {
		return fmt.Errorf("gate the end of the turn: %w", err)
	}

	return writeBlock(stdout, failure)
}

func fileEdit(hookInput input) (hook.FileEdit, bool) {
	tool := hookInput.ToolInput

	switch hookInput.ToolName {
	case "Edit":
		return hook.FileEdit{
			Path:      tool.FilePath,
			Removed:   tool.OldString,
			Inserted:  tool.NewString,
			WholeFile: false,
		}, true
	case "Write":
		return hook.FileEdit{Path: tool.FilePath, Removed: "", Inserted: tool.Content, WholeFile: true}, true
	case "NotebookEdit":
		return hook.FileEdit{Path: tool.NotebookPath, Removed: "", Inserted: tool.NewSource, WholeFile: false}, true
	default:
		return hook.FileEdit{Path: "", Removed: "", Inserted: "", WholeFile: false}, false
	}
}
