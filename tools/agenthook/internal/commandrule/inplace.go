package commandrule

// An inPlaceEditor rewrites the files it names when given its in-place option. Without a script option its first
// operand is the script, not a file.
type inPlaceEditor struct {
	spec          argSpec
	inPlace       []string
	scriptOptions []string
}

func sedEditor() inPlaceEditor {
	return inPlaceEditor{
		spec: argSpec{
			valueShorts:    "efl",
			attachedShorts: "i",
			valueLongs:     []string{"--expression", "--file", "--line-length"},
		},
		inPlace:       []string{"-i", "--in-place"},
		scriptOptions: []string{"-e", "-f", "--expression", "--file"},
	}
}

func perlEditor() inPlaceEditor {
	return inPlaceEditor{
		spec:          argSpec{valueShorts: "eE", attachedShorts: "iIMmxdDlFC0", valueLongs: nil},
		inPlace:       []string{"-i"},
		scriptOptions: []string{"-e", "-E"},
	}
}

func (e inPlaceEditor) targets(command invocation) []string {
	scanned := e.spec.scan(command.Args[1:])
	if !scanned.has(e.inPlace...) {
		return nil
	}

	if scanned.has(e.scriptOptions...) || len(scanned.operands) == 0 {
		return scanned.operands
	}

	return scanned.operands[1:]
}
