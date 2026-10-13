package check

import "strings"

const (
	goModuleVariable = "GO_MODULE"
	filesVariable    = "FILES"
)

// Target is one Make target the hooks run. GoModule is the module, relative to the checkout root, that a Go
// target works in, and Files are relative to it. Targets that are not Go's leave GoModule empty and take Files
// relative to the checkout root.
type Target struct {
	Name     string
	GoModule string
	Files    []string
}

func goTarget(name, module string, files []string) Target {
	return Target{Name: name, GoModule: module, Files: files}
}

func checkoutTarget(name string, files []string) Target {
	return Target{Name: name, GoModule: "", Files: files}
}

func (t Target) String() string {
	return "make " + t.Name
}

func (t Target) makeArgs() []string {
	args := []string{t.Name}
	if t.GoModule != "" {
		args = append(args, goModuleVariable+"="+t.GoModule)
	}

	if len(t.Files) > 0 {
		args = append(args, filesVariable+"="+strings.Join(t.Files, " "))
	}

	return args
}
