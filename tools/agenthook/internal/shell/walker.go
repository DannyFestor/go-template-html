package shell

import (
	"fmt"
	"io"

	"mvdan.cc/sh/v3/syntax"
)

// A scope holds the directory to continue in once syntax.Walk leaves its node: the one it was entered in for a
// subshell, or a cd's target, which only applies after the substitutions in the cd's own words ran.
type scope struct {
	dir       string
	variables int
	restores  bool
	cdTarget  *string
}

type walker struct {
	expander   *expander
	substitute func(Command) (string, bool)
	dir        string
	scopes     []scope
	commands   []Command
	err        error
}

func newWalker(start Start) *walker {
	walker := &walker{
		expander:   newExpander(start.Environ),
		substitute: start.Substitute,
		dir:        start.Dir,
		scopes:     nil,
		commands:   nil,
		err:        nil,
	}
	walker.expander.config.CmdSubst = walker.substitution

	return walker
}

// syntax.Walk calls visit with nil after the children of every node it entered.
func (w *walker) visit(node syntax.Node) bool {
	if node == nil {
		w.leave()

		return true
	}

	w.scopes = append(w.scopes, scope{
		dir:       w.dir,
		variables: len(w.expander.variables),
		restores:  startsSubshell(node),
		cdTarget:  nil,
	})

	if stmt, isStmt := node.(*syntax.Stmt); isStmt {
		w.stmt(stmt)
	}

	return w.err == nil
}

func (w *walker) leave() {
	last := w.scopes[len(w.scopes)-1]
	w.scopes = w.scopes[:len(w.scopes)-1]

	switch {
	case last.restores:
		w.dir = last.dir
		w.expander.keepFirstVariables(last.variables)
	case last.cdTarget != nil:
		w.dir = *last.cdTarget
	}
}

func startsSubshell(node syntax.Node) bool {
	switch node.(type) {
	case *syntax.Subshell, *syntax.CmdSubst, *syntax.ProcSubst:
		return true
	default:
		return false
	}
}

func (w *walker) stmt(stmt *syntax.Stmt) {
	redirects := w.expander.redirects(stmt.Redirs, w.dir)

	switch command := stmt.Cmd.(type) {
	case *syntax.CallExpr:
		w.call(command, redirects)
	case *syntax.DeclClause:
		w.expander.assign(command.Args)
		w.recordRedirects(redirects)
	default:
		w.recordRedirects(redirects)
	}
}

// Assignments without a command last for the rest of the script, while those before a command only reach it.
func (w *walker) call(call *syntax.CallExpr, redirects redirects) {
	if len(call.Args) == 0 {
		w.expander.assign(call.Assigns)
	}

	command := Command{
		Args:   unwrap(w.expander.words(call.Args)),
		Dir:    w.dir,
		Writes: redirects.writes,
		Input:  redirects.input,
	}
	w.commands = append(w.commands, command)
	w.changeDirOnLeave(command)
	w.parseNestedScript(command)
}

func (w *walker) recordRedirects(redirects redirects) {
	if len(redirects.writes) > 0 {
		w.commands = append(w.commands, Command{Args: nil, Dir: w.dir, Writes: redirects.writes, Input: ""})
	}
}

func (w *walker) changeDirOnLeave(command Command) {
	if command.Name() == "cd" {
		target := w.expander.cdTarget(command)
		w.scopes[len(w.scopes)-1].cdTarget = &target
	}
}

func (w *walker) parseNestedScript(command Command) {
	script, isShell := nestedScript(command)
	if !isShell {
		return
	}

	nested, err := Parse(script, Start{Dir: command.Dir, Environ: w.expander.variables, Substitute: w.substitute})
	w.commands = append(w.commands, nested...)
	w.err = err
}

func (w *walker) substitution(output io.Writer, substitution *syntax.CmdSubst) error {
	command, isSimple := w.simpleCommand(substitution.Stmts)
	if !isSimple || w.substitute == nil {
		return errUnresolved
	}

	value, resolved := w.substitute(command)
	if !resolved {
		return errUnresolved
	}

	_, err := io.WriteString(output, value)
	if err != nil {
		return fmt.Errorf("write the substituted value: %w", err)
	}

	return nil
}

func (w *walker) simpleCommand(stmts []*syntax.Stmt) (Command, bool) {
	if len(stmts) != 1 {
		return Command{Args: nil, Dir: "", Writes: nil, Input: ""}, false
	}

	call, isCall := stmts[0].Cmd.(*syntax.CallExpr)
	if !isCall || len(call.Args) == 0 {
		return Command{Args: nil, Dir: "", Writes: nil, Input: ""}, false
	}

	return Command{Args: unwrap(w.expander.words(call.Args)), Dir: w.dir, Writes: nil, Input: ""}, true
}
