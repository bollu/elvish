package cliedit

import (
	"testing"

	"github.com/elves/elvish/eval/vals"
	"github.com/elves/elvish/styled"
	"github.com/elves/elvish/util"
)

func TestCompletionAddon(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"a": "", "b": ""})

	feedInput(f.TTYCtrl, "echo \t")
	wantBuf := bb().
		WriteStyled(styled.MarkLines(
			"~> echo a ", styles,
			"   gggg --",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ")).
		SetDotToCursor().
		Newline().
		WriteStyled(styled.MarkLines(
			"a  b", styles,
			"#   ",
		)).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestCompleteFilename(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"d": util.Dir{"a": "", "b": ""}})

	evals(f.Evaler, `@cands = (edit:complete-filename ls ./d/a)`)
	testGlobal(t, f.Evaler,
		"cands",
		vals.MakeList(
			complexItem{Stem: "./d/a", CodeSuffix: " "},
			complexItem{Stem: "./d/b", CodeSuffix: " "}))
}

func TestComplexCandidate(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`cand  = (edit:complex-candidate a/b/c &code-suffix=' ' &display-suffix='x')`,
		// Identical to $cand.
		`cand2 = (edit:complex-candidate a/b/c &code-suffix=' ' &display-suffix='x')`,
		// Different from $cand.
		`cand3 = (edit:complex-candidate a/b/c)`,
		`kind  = (kind-of $cand)`,
		`@keys = (keys $cand)`,
		`repr  = (repr $cand)`,
		`eq2   = (eq $cand $cand2)`,
		`eq2h  = [&$cand=$true][$cand2]`,
		`eq3   = (eq $cand $cand3)`,
		`stem code-suffix display-suffix = $cand[stem code-suffix display-suffix]`,
	)
	testGlobals(t, f.Evaler, map[string]interface{}{
		"kind": "map",
		"keys": vals.MakeList("stem", "code-suffix", "display-suffix"),
		"repr": "(edit:complex-candidate a/b/c &code-suffix=' ' &display-suffix=x)",
		"eq2":  true,
		"eq2h": true,
		"eq3":  false,

		"stem":           "a/b/c",
		"code-suffix":    " ",
		"display-suffix": "x",
	})
}

func TestCompletionArgCompleter_ArgsAndValueOutput(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`foo-args = []`,
		`fn foo { }`,
		`edit:completion:arg-completer[foo] = [@args]{
		   foo-args = $args
		   put val1
		   edit:complex-candidate val2 &display-suffix=_
		 }`)

	feedInput(f.TTYCtrl, "foo foo1 foo2 \t")
	wantBuf := bb().
		WriteStyled(styled.MarkLines(
			"~> foo foo1 foo2 val1", styles,
			"   ggg           ----",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ")).
		SetDotToCursor().
		Newline().
		WriteStyled(styled.MarkLines(
			"val1  val2_", styles,
			"####       ",
		)).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
	testGlobal(t, f.Evaler,
		"foo-args", vals.MakeList("foo", "foo1", "foo2", ""))
}

func TestCompletionArgCompleter_BytesOutput(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`fn foo { }`,
		`edit:completion:arg-completer[foo] = [@args]{
		   echo val1
		   echo val2
		 }`)

	feedInput(f.TTYCtrl, "foo foo1 foo2 \t")
	wantBuf := bb().
		WriteStyled(styled.MarkLines(
			"~> foo foo1 foo2 val1", styles,
			"   ggg           ----",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ")).
		SetDotToCursor().
		Newline().
		WriteStyled(styled.MarkLines(
			"val1  val2", styles,
			"####      ",
		)).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestCompleteSudo(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`fn foo { }`,
		`edit:completion:arg-completer[foo] = [@args]{
		   echo val1
		   echo val2
		 }`,
		`@cands = (edit:complete-sudo sudo foo '')`)
	testGlobal(t, f.Evaler, "cands", vals.MakeList("val1", "val2"))
}

func TestCompletionMatcher(t *testing.T) {
	f := setup()
	defer f.Cleanup()
	util.ApplyDir(util.Dir{"foo": "", "oof": ""})

	evals(f.Evaler, `edit:completion:matcher[''] = $edit:match-substr~`)
	feedInput(f.TTYCtrl, "echo f\t")
	wantBuf := bb().
		WriteStyled(styled.MarkLines(
			"~> echo foo ", styles,
			"   gggg ----",
			"COMPLETING argument ", styles,
			"mmmmmmmmmmmmmmmmmmm ")).
		SetDotToCursor().
		Newline().
		WriteStyled(styled.MarkLines(
			"foo  oof", styles,
			"###     ",
		)).
		Buffer()
	f.TTYCtrl.TestBuffer(t, wantBuf)
}

func TestBuiltinMatchers(t *testing.T) {
	f := setup()
	defer f.Cleanup()

	evals(f.Evaler,
		`@prefix = (edit:match-prefix ab [ab abc cab acb ba [ab] [a b] [b a]])`,
		`@substr = (edit:match-substr ab [ab abc cab acb ba [ab] [a b] [b a]])`,
		`@subseq = (edit:match-subseq ab [ab abc cab acb ba [ab] [a b] [b a]])`,
	)
	testGlobals(t, f.Evaler, map[string]interface{}{
		"prefix": vals.MakeList(true, true, false, false, false, false, false, false),
		"substr": vals.MakeList(true, true, true, false, false, true, false, false),
		"subseq": vals.MakeList(true, true, true, true, false, true, true, false),
	})
}
