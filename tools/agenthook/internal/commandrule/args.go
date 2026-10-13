package commandrule

import (
	"slices"
	"strings"
)

type argSpec struct {
	// valueShorts take the next word as their value when they end a cluster, or the rest of the cluster.
	valueShorts string
	// attachedShorts take the rest of their cluster as their value, never the next word.
	attachedShorts string
	// valueLongs take the next word as their value unless it is attached with =.
	valueLongs []string
}

// scannedArgs maps each option given, as "-x" or "--long", to its value.
type scannedArgs struct {
	operands []string
	options  map[string]string
}

func plainSpec() argSpec {
	return argSpec{valueShorts: "", attachedShorts: "", valueLongs: nil}
}

func (a scannedArgs) has(options ...string) bool {
	for _, option := range options {
		if _, given := a.options[option]; given {
			return true
		}
	}

	return false
}

func (a scannedArgs) value(options ...string) string {
	for _, option := range options {
		if value, given := a.options[option]; given {
			return value
		}
	}

	return ""
}

func (s argSpec) scan(args []string) scannedArgs {
	scanned := scannedArgs{operands: nil, options: map[string]string{}}

	for len(args) > 0 {
		arg, rest := args[0], args[1:]

		switch {
		case arg == "--":
			scanned.operands = append(scanned.operands, rest...)

			return scanned
		case strings.HasPrefix(arg, "--"):
			args = s.scanLong(scanned, arg, rest)
		case strings.HasPrefix(arg, "-") && arg != "-":
			args = s.scanShorts(scanned, arg[1:], rest)
		default:
			scanned.operands = append(scanned.operands, arg)
			args = rest
		}
	}

	return scanned
}

func (s argSpec) scanLong(scanned scannedArgs, arg string, rest []string) []string {
	name, value, attached := strings.Cut(arg, "=")
	if attached || !slices.Contains(s.valueLongs, name) || len(rest) == 0 {
		scanned.options[name] = value

		return rest
	}

	scanned.options[name] = rest[0]

	return rest[1:]
}

func (s argSpec) scanShorts(scanned scannedArgs, cluster string, rest []string) []string {
	for index, option := range cluster {
		name := "-" + string(option)
		value := cluster[index+1:]

		switch {
		case strings.ContainsRune(s.attachedShorts, option):
			scanned.options[name] = value

			return rest
		case strings.ContainsRune(s.valueShorts, option) && value == "" && len(rest) > 0:
			scanned.options[name] = rest[0]

			return rest[1:]
		case strings.ContainsRune(s.valueShorts, option):
			scanned.options[name] = value

			return rest
		default:
			scanned.options[name] = ""
		}
	}

	return rest
}
