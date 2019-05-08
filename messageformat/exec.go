package messageformat

import (
	"fmt"
	"io"
	"runtime"

	"libs.altipla.consulting/messageformat/parse"
)

type state struct {
	t    *parse.Tree
	wr   io.Writer
	vars map[string]interface{}
	lang string
}

// errorf records an error and terminates processing.
func (s *state) errorf(format string, args ...interface{}) {
	panic(fmt.Errorf(format, args...))
}

// recover is the handler that turns panics into returns from the top
// level of Parse.
func (s *state) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*errp = e.(error)
	}
}

func (s *state) execute() (err error) {
	defer s.recover(&err)
	s.walk(s.t.Root)
	return nil
}

func (s *state) walk(node parse.Node) {
	switch node := node.(type) {
	case *parse.ListNode:
		for _, child := range node.Nodes {
			s.walk(child)
		}

	case *parse.VariableNode:
		value, ok := s.vars[node.Name]
		if !ok {
			s.errorf("unknown variable: %s", node.Name)
		}
		fmt.Fprint(s.wr, value)

	case *parse.TextNode:
		fmt.Fprint(s.wr, node.Text)

	case *parse.PluralNode:
		value, ok := s.vars[node.Variable]
		if !ok {
			s.errorf("unknown variable: %s", node.Variable)
		}

		var n int64
		switch x := value.(type) {
		case int:
			n = int64(x)
		case int64:
			n = x
		case int32:
			n = int64(x)
		default:
			s.errorf("variable must be numeric for plurals: %s: %v", node.Variable, value)
		}

		var bestCase *parse.PluralCase
		for _, c := range node.Cases {
			if c.Category == parse.PluralValue {
				bestCase = c
				break
			}

			if matchesPlural(s.lang, c, n) {
				bestCase = c
			}
		}
		if bestCase == nil {
			s.errorf("plural missing a case for the variable: %s: %v", node.Variable, value)
		}

		s.walk(bestCase.Content)

	default:
		s.errorf("unknown node: %s", node)
	}
}
