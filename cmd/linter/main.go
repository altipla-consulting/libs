package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mgechev/dots"
	"github.com/mgechev/revive/formatter"
	"github.com/mgechev/revive/lint"
	"github.com/mgechev/revive/rule"
	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	"libs.altipla.consulting/cmd/linter/customrules"
	"libs.altipla.consulting/errors"
)

type ruleConfig struct {
	run  lint.Rule
	args []interface{}
}

var allRules = []ruleConfig{
	// {new(rule.AddConstantRule)},
	// {new(rule.ArgumentsLimitRule)},
	{new(rule.AtomicRule), nil},
	{new(rule.BareReturnRule), nil},
	{new(rule.BlankImportsRule), nil},
	{new(rule.BoolLiteralRule), nil},
	{new(rule.CallToGCRule), nil},
	// {new(rule.CognitiveComplexityRule)},
	{new(rule.ConfusingNamingRule), nil},
	{new(rule.ConfusingResultsRule), nil},
	{new(rule.ConstantLogicalExprRule), nil},
	{new(rule.ContextAsArgumentRule), nil},
	{new(rule.ContextKeysType), nil},
	// {new(rule.CyclomaticRule)},
	{new(rule.DeepExitRule), nil},
	{new(rule.DotImportsRule), nil},
	{new(rule.DuplicatedImportsRule), nil},
	{new(rule.EmptyBlockRule), nil},
	{new(rule.EmptyLinesRule), nil},
	{new(rule.ErrorNamingRule), nil},
	{new(rule.ErrorReturnRule), nil},
	{new(rule.ErrorStringsRule), nil},
	{new(rule.ErrorfRule), nil},
	// {new(rule.ExportedRule), nil},
	// {new(rule.FileHeaderRule)},
	// {new(rule.FlagParamRule)},
	{
		new(rule.FunctionResultsLimitRule),
		[]interface{}{
			int64(3),
		},
	},
	{new(rule.GetReturnRule), nil},
	{new(rule.IfReturnRule), nil},
	// {new(rule.ImportShadowingRule), nil},
	{
		new(rule.ImportsBlacklistRule),
		[]interface{}{
			"errors",
			"github.com/juju/errors",
			"github.com/altipla-consulting/errors",
		},
	},
	{new(rule.IncrementDecrementRule), nil},
	{new(rule.IndentErrorFlowRule), nil},
	// {new(rule.LineLengthLimitRule)},
	// {new(rule.MaxPublicStructsRule)},
	{new(rule.ModifiesParamRule), nil},
	{new(rule.ModifiesValRecRule), nil},
	{new(rule.PackageCommentsRule), nil},
	{new(rule.RangeRule), nil},
	{new(rule.RangeValInClosureRule), nil},
	{new(rule.ReceiverNamingRule), nil},
	{new(rule.RedefinesBuiltinIDRule), nil},
	{new(rule.StructTagRule), nil},
	{new(rule.SuperfluousElseRule), nil},
	{new(rule.TimeNamingRule), nil},
	{new(rule.UnexportedReturnRule), nil},
	{
		new(rule.UnhandledErrorRule),
		[]interface{}{
			"fmt.Fprint",
			"fmt.Fprintf",
			"fmt.Fprintln",
			"fmt.Println",
		},
	},
	{new(rule.UnnecessaryStmtRule), nil},
	{new(rule.UnreachableCodeRule), nil},
	// {new(rule.UnusedParamRule)},
	// {new(rule.UnusedReceiverRule)},
	{new(rule.VarDeclarationsRule), nil},
	{new(rule.VarNamingRule), nil},
	{new(rule.WaitGroupByValueRule), nil},
	{new(customrules.ImportShadowingRule), nil},
	{new(customrules.MultilineIfRule), nil},
}

func main() {
	if err := run(); err != nil {
		log.Fatal(errors.Stack(err))
	}
}

func run() error {
	flag.Parse()

	packages, err := dots.ResolvePackages(flag.Args(), nil)
	if err != nil {
		return errors.Trace(err)
	}

	reader := func(file string) ([]byte, error) {
		return ioutil.ReadFile(file)
	}
	cnf := lint.Config{
		Severity:   "error",
		Confidence: 0.8,
		Rules:      make(lint.RulesConfig),
	}

	var rules []lint.Rule
	for _, r := range allRules {
		rules = append(rules, r.run)
		cnf.Rules[r.run.Name()] = lint.RuleConfig{
			Arguments: r.args,
			Severity:  cnf.Severity,
		}
	}

	revive := lint.New(reader)
	failures, err := revive.Lint(packages, rules, cnf)
	if err != nil {
		return errors.Trace(err)
	}

	failureCh := make(chan lint.Failure)
	doneCh := make(chan error)
	formatter := new(formatter.Stylish)
	go func() {
		output, err := formatter.Format(failureCh, cnf)
		if output != "" {
			fmt.Println(output)
		}
		doneCh <- errors.Trace(err)
	}()
	for failure := range failures {
		if failure.Confidence >= cnf.Confidence {
			failureCh <- failure
		}
	}

	close(failureCh)
	return errors.Trace(<-doneCh)
}
