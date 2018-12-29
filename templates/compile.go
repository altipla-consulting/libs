package templates

import (
	"html/template"
	"text/template/parse"

	"github.com/altipla-consulting/errors"
	"github.com/altipla-consulting/langs"

	"github.com/altipla-consulting/apps/pkg/templates/funcs"
)

var (
	viewsFuncs = template.FuncMap{
		"safehtml": funcs.SafeHTML,
		"safejs":   funcs.SafeJavascript,
		"safeurl":  funcs.SafeURL,
		"safecss":  funcs.SafeCSS,

		"thumbnail": funcs.Thumbnail,

		"nospaces":  funcs.NoSpaces,
		"camelcase": funcs.CamelCase,
		"nl2br":     funcs.Nl2Br,
		"hasprefix": funcs.HasPrefix,
		"join":      funcs.Join,

		"newvar": funcs.NewVar,
		"setvar": funcs.SetVar,
		"getvar": funcs.GetVar,

		"dict":   funcs.Dict,
		"json":   funcs.JSON,
		"vue":    funcs.Vue,
		"randid": funcs.RandID,

		"development": funcs.Development,

		"nativename": langs.NativeName,
		"msgformat":  funcs.MsgFormat,
		"__":         funcs.Translate,

		"even":       funcs.Even,
		"odd":        funcs.Odd,
		"trio":       funcs.Trio,
		"mod":        funcs.Mod,
		"div":        funcs.Div,
		"times":      funcs.Times,
		"add":        funcs.Add,
		"percentage": funcs.Percentage,

		"genrange": funcs.GenRange,
		"shuffle":  funcs.Shuffle,
		"limit":    funcs.Limit,
		"slice":    funcs.Slice,
		"randitem": funcs.RandItem,
		"last":     funcs.Last,

		"price": funcs.Price,

		"dateformat":      funcs.DateFormat,
		"protodateformat": funcs.ProtoDateFormat,
		"timestampformat": funcs.TimestampFormat,
		"now":             funcs.Now,
		"timezone":        funcs.Timezone,
		"madrid":          funcs.Madrid,
		"timeformat":      funcs.TimeFormat,

		"component":    funcs.Component,
		"endcomponent": funcs.EndComponent,
		"param":        funcs.ComponentParam,

		"asset": funcs.Asset,
		"rev":   funcs.Rev,

		"include": funcs.Include,
	}
)

// Load receives a list of globs and load every template that matches those folders.
// All folders must have at least one template or it will fail the loading.
func Load(folders ...string) (*template.Template, error) {
	tmpl := template.New("root").Funcs(viewsFuncs)

	for _, folder := range folders {
		if _, err := tmpl.ParseGlob(folder); err != nil {
			return nil, errors.Errorf("cannot parse folder %s: %s", folder, err)
		}
	}

	for _, template := range tmpl.Templates() {
		if template.Tree == nil {
			continue
		}

		template.Tree.Root.Nodes = applyTreeChanges(template.Tree.Root.Nodes)
	}

	return tmpl, nil
}

func applyTreeChanges(nodes []parse.Node) []parse.Node {
	for i, node := range nodes {
		if node.Type() != parse.NodePipe {
			continue
		}

		pipe := node.(*parse.PipeNode)

		call := pipe.Cmds[0]
		if call.Args[0].Type() != parse.NodeIdentifier {
			continue
		}

		switch ident := call.Args[0].(*parse.IdentifierNode).Ident; ident {
		case "msgformat", "__", "protodateformat":
			args := []parse.Node{
				parse.NewIdentifier(ident),
				&parse.VariableNode{Ident: []string{"$", "Lang"}},
			}
			args = append(args, call.Args[1:]...)

			pipe = &parse.PipeNode{
				NodeType: parse.NodePipe,
				Cmds: []*parse.CommandNode{
					{
						NodeType: parse.NodeCommand,
						Args:     args,
					},
				},
			}

		case "component":
			pipe = &parse.PipeNode{
				NodeType: parse.NodePipe,
				Decl: []*parse.VariableNode{
					{
						NodeType: parse.NodeVariable,
						Ident:    []string{"$c"},
					},
				},
				Cmds: []*parse.CommandNode{
					{
						NodeType: parse.NodeCommand,
						Args: []parse.Node{
							parse.NewIdentifier(ident),
							call.Args[1],
						},
					},
				},
			}

		case "endcomponent":
			pipe = &parse.PipeNode{
				NodeType: parse.NodePipe,
				Cmds: []*parse.CommandNode{
					{
						NodeType: parse.NodeCommand,
						Args: []parse.Node{
							parse.NewIdentifier(ident),
							&parse.VariableNode{
								NodeType: parse.NodeVariable,
								Ident:    []string{"$c"},
							},
						},
					},
				},
			}

		case "param":
			pipe = &parse.PipeNode{
				NodeType: parse.NodePipe,
				Cmds: []*parse.CommandNode{
					{
						NodeType: parse.NodeCommand,
						Args: []parse.Node{
							parse.NewIdentifier(ident),
							&parse.VariableNode{
								NodeType: parse.NodeVariable,
								Ident:    []string{"$c"},
							},
							call.Args[1],
							call.Args[2],
						},
					},
				},
			}

		case "asset":
			pipe = &parse.PipeNode{
				NodeType: parse.NodePipe,
				Cmds: []*parse.CommandNode{
					{
						NodeType: parse.NodeCommand,
						Args: []parse.Node{
							parse.NewIdentifier(ident),
							&parse.VariableNode{
								NodeType: parse.NodeVariable,
								Ident:    []string{"$", "AssetsHostname"},
							},
							call.Args[1],
						},
					},
				},
			}

		default:
			continue
		}

		pipe.Cmds = append(pipe.Cmds, node.(*parse.PipeNode).Cmds[1:]...)
		nodes[i] = pipe
	}

	for _, node := range nodes {
		switch n := node.(type) {
		case *parse.ListNode:
			n.Nodes = applyTreeChanges(n.Nodes)

		case *parse.BranchNode:
			n.List.Nodes = applyTreeChanges(n.List.Nodes)
			if n.ElseList != nil {
				n.ElseList.Nodes = applyTreeChanges(n.ElseList.Nodes)
			}

		case *parse.IfNode:
			n.List.Nodes = applyTreeChanges(n.List.Nodes)
			if n.ElseList != nil {
				n.ElseList.Nodes = applyTreeChanges(n.ElseList.Nodes)
			}

		case *parse.RangeNode:
			n.List.Nodes = applyTreeChanges(n.List.Nodes)
			if n.ElseList != nil {
				n.ElseList.Nodes = applyTreeChanges(n.ElseList.Nodes)
			}

		case *parse.WithNode:
			n.List.Nodes = applyTreeChanges(n.List.Nodes)
			if n.ElseList != nil {
				n.ElseList.Nodes = applyTreeChanges(n.ElseList.Nodes)
			}

		case *parse.PipeNode:
			n.Cmds = nodesToCmds(applyTreeChanges(cmdsToNode(n.Cmds)))

		case *parse.ActionNode:
			n.Pipe = applyTreeChanges([]parse.Node{n.Pipe})[0].(*parse.PipeNode)

		case *parse.CommandNode:
			n.Args = applyTreeChanges(n.Args)
		}
	}

	return nodes
}

func cmdsToNode(cmds []*parse.CommandNode) []parse.Node {
	nodes := make([]parse.Node, len(cmds))
	for i, node := range cmds {
		nodes[i] = node
	}
	return nodes
}

func nodesToCmds(nodes []parse.Node) []*parse.CommandNode {
	cmds := make([]*parse.CommandNode, len(nodes))
	for i, node := range nodes {
		cmds[i] = node.(*parse.CommandNode)
	}
	return cmds
}
