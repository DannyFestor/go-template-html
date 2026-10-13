package claude

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

// Every decision goes out as JSON with exit status 0, so a non-zero status only ever means the hook itself broke,
// which Claude Code reports without blocking the tool.
type permissionOutput struct {
	HookSpecificOutput permissionDecision `json:"hookSpecificOutput"`
}

type permissionDecision struct {
	HookEventName            string `json:"hookEventName"`
	PermissionDecision       string `json:"permissionDecision"`
	PermissionDecisionReason string `json:"permissionDecisionReason"`
}

type blockOutput struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

func writePermission(stdout io.Writer, judged decision.Decision) error {
	verdicts := map[decision.Verdict]string{decision.Ask: "ask", decision.Deny: "deny"}

	verdict, decided := verdicts[judged.Verdict]
	if !decided {
		return nil
	}

	return writeJSON(stdout, permissionOutput{HookSpecificOutput: permissionDecision{
		HookEventName:            "PreToolUse",
		PermissionDecision:       verdict,
		PermissionDecisionReason: judged.Reason,
	}})
}

func writeBlock(stdout io.Writer, reason string) error {
	if reason == "" {
		return nil
	}

	return writeJSON(stdout, blockOutput{Decision: "block", Reason: reason})
}

func writeJSON(stdout io.Writer, value any) error {
	err := json.NewEncoder(stdout).Encode(value)
	if err != nil {
		return fmt.Errorf("write hook output: %w", err)
	}

	return nil
}
