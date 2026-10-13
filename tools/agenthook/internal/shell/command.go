package shell

// Command is one simple command a script runs. An argument that cannot be known without running the script, such
// as a command substitution, is empty. Dir is empty once a cd moved to such an unknown directory.
type Command struct {
	Args   []string
	Dir    string
	Writes []string
	Input  string
}

func (c Command) Name() string {
	if len(c.Args) == 0 {
		return ""
	}

	return c.Args[0]
}

func (c Command) Resolve(path string) string {
	return Resolve(c.Dir, path)
}
