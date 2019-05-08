package parse

import (
	"bytes"
	"fmt"
	"strings"
)

type Node interface {
	Type() NodeType
	String() string
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeText NodeType = iota
	NodeList
	NodeVariable
	NodePlural
	NodePluralCase
)

type TextNode struct {
	NodeType
	Text string
}

func (node *TextNode) String() string {
	return node.Text
}

type ListNode struct {
	NodeType
	Nodes []Node
}

func (l *ListNode) String() string {
	var children bytes.Buffer
	for _, n := range l.Nodes {
		fmt.Fprint(&children, n)
	}
	return children.String()
}

type VariableNode struct {
	NodeType
	Name string
}

func (v *VariableNode) String() string {
	return "{" + v.Name + "}"
}

type PluralNode struct {
	NodeType
	Variable string
	Cases    []*PluralCase
}

func (p *PluralNode) String() string {
	cases := make([]string, len(p.Cases))
	for i, c := range p.Cases {
		cases[i] = c.String()
	}
	return "{" + p.Variable + ", plural, " + strings.Join(cases, " ") + "}"
}

type PluralCase struct {
	NodeType
	Value    int
	Category PluralCategory
	Content  *ListNode
}

func (p *PluralCase) String() string {
	switch p.Category {
	case PluralOne, PluralOther:
		return string(p.Category) + " {" + p.Content.String() + "}"
	case PluralValue:
		return fmt.Sprintf("=%d {%s}", p.Value, p.Content.String())
	}
	panic("should not reach here")
}

type PluralCategory string

const (
	PluralOne   = PluralCategory("one")
	PluralOther = PluralCategory("other")
	PluralValue = PluralCategory("value")
)
