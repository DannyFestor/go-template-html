package decision

import (
	"slices"
	"strings"
)

// Verdict values are ordered from weakest to strongest, so the strongest of several decisions wins.
type Verdict int

const (
	Pass Verdict = iota
	Ask
	Deny
)

// Decision tells the agent whether a tool call may run. Reason is the way forward when it may not.
type Decision struct {
	Verdict Verdict
	Reason  string
}

func Passed() Decision {
	return Decision{Verdict: Pass, Reason: ""}
}

func Asked(reason string) Decision {
	return Decision{Verdict: Ask, Reason: reason}
}

func Denied(reason string) Decision {
	return Decision{Verdict: Deny, Reason: reason}
}

func Strongest(decisions ...Decision) Decision {
	verdict := Pass
	for _, decision := range decisions {
		verdict = max(verdict, decision.Verdict)
	}

	var reasons []string

	for _, decision := range decisions {
		if decision.Verdict == verdict && decision.Reason != "" && !slices.Contains(reasons, decision.Reason) {
			reasons = append(reasons, decision.Reason)
		}
	}

	return Decision{Verdict: verdict, Reason: strings.Join(reasons, "\n\n")}
}
