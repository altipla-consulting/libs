package customrules

import (
	"go/ast"

	"github.com/mgechev/revive/lint"
)

type MultilineIfRule struct{}

func (r *MultilineIfRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	walker := &multilineIf{
		file: file,
	}
	ast.Walk(walker, file.AST)

	return walker.failures
}

func (r *MultilineIfRule) Name() string {
	return "multiline-if"
}

type multilineIf struct {
	file     *lint.File
	failures []lint.Failure
}

func (w *multilineIf) Visit(n ast.Node) ast.Visitor {
	if n, ok := n.(*ast.IfStmt); ok {
		expected := w.file.ToPosition(n.Cond.Pos()).Line
		if n.Init != nil {
			expected = w.file.ToPosition(n.Init.Pos()).Line
			if w.file.ToPosition(n.Init.Pos()).Line != expected || w.file.ToPosition(n.Init.End()).Line != expected {
				w.failures = append(w.failures, lint.Failure{
					Failure:    "if statement cannot start and end in different lines, use variables",
					Node:       n.Init,
					Confidence: 1,
				})
				return nil
			}
		}
		if w.file.ToPosition(n.Cond.End()).Line != expected {
			w.failures = append(w.failures, lint.Failure{
				Failure:    "if statement cannot start and end in different lines, use variables",
				Node:       n.Cond,
				Confidence: 1,
			})
		}
	}

	return w
}
