package edit

import (
	"fmt"
	"os"
	"strings"

	"github.com/elves/elvish/pkg/cli"
	"github.com/elves/elvish/pkg/cli/histutil"
	"github.com/elves/elvish/pkg/diag"
	"github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/eval/vals"
	"github.com/elves/elvish/pkg/eval/vars"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/store"
)

//elvdoc:var max-height
//
// Maximum height the editor is allowed to use, defaults to `+Inf`.
//
// By default, the height of the editor is only restricted by the terminal
// height. Some modes like location mode can use a lot of lines; as a result,
// it can often occupy the entire terminal, and push up your scrollback buffer.
// Change this variable to a finite number to restrict the height of the editor.

func initMaxHeight(appSpec *cli.AppSpec, nb eval.NsBuilder) {
	maxHeight := newIntVar(-1)
	appSpec.MaxHeight = func() int { return maxHeight.GetRaw().(int) }
	nb.Add("max-height", maxHeight)
}

func initReadlineHooks(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	initBeforeReadline(appSpec, ev, nb)
	initAfterReadline(appSpec, ev, nb)
}

//elvdoc:var before-readline
//
// A list of functions to call before each readline cycle. Each function is
// called without any arguments.

func initBeforeReadline(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	hook := newListVar(vals.EmptyList)
	nb["before-readline"] = hook
	appSpec.BeforeReadline = append(appSpec.BeforeReadline, func() {
		callHooks(ev, "$<edit>:before-readline", hook.Get().(vals.List))
	})
}

//elvdoc:var after-readline
//
// A list of functions to call after each readline cycle. Each function is
// called with a single string argument containing the code that has been read.

func initAfterReadline(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder) {
	hook := newListVar(vals.EmptyList)
	nb["after-readline"] = hook
	appSpec.AfterReadline = append(appSpec.AfterReadline, func(code string) {
		callHooks(ev, "$<edit>:after-readline", hook.Get().(vals.List), code)
	})
}

//elvdoc:var add-cmd-filters
//
// List of filters to run before adding a command to history.
//
// A filter is a function that takes a command as argument and outputs
// a boolean value. If any of the filters outputs `$false`, the
// command is not saved to history, and the rest of the filters are
// not run. The default value of this list contains a filter which
// ignores command starts with space.

func initAddCmdFilters(appSpec *cli.AppSpec, ev *eval.Evaler, nb eval.NsBuilder, s histutil.Store) {
	ignoreLeadingSpace := eval.NewGoFn("<ignore-cmd-with-leading-space>",
		func(s string) bool { return !strings.HasPrefix(s, " ") })
	filters := newListVar(vals.MakeList(ignoreLeadingSpace))
	nb["add-cmd-filters"] = filters

	appSpec.AfterReadline = append(appSpec.AfterReadline, func(code string) {
		if code != "" &&
			callFilters(ev, "$<edit>:add-cmd-filters",
				filters.Get().(vals.List), code) {
			s.AddCmd(store.Cmd{Text: code, Seq: -1})
		}
		// TODO(xiaq): Handle the error.
	})
}

func callHooks(ev *eval.Evaler, name string, hook vals.List, args ...interface{}) {
	i := -1
	for it := hook.Iterator(); it.HasElem(); it.Next() {
		i++
		name := fmt.Sprintf("%s[%d]", name, i)
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			// TODO(xiaq): This is not testable as it depends on stderr.
			// Make it testable.
			diag.Complainf(os.Stderr, "%s not function", name)
			continue
		}
		// TODO(xiaq): This should use stdPorts, but stdPorts is currently
		// unexported from eval.
		ports := []*eval.Port{
			{File: os.Stdin}, {File: os.Stdout}, {File: os.Stderr}}
		fm := eval.NewTopFrame(ev, parse.Source{Name: name}, ports)
		fn.Call(fm, args, eval.NoOpts)
	}
}

func callFilters(ev *eval.Evaler, name string, filters vals.List, args ...interface{}) bool {
	i := -1
	for it := filters.Iterator(); it.HasElem(); it.Next() {
		i++
		name := fmt.Sprintf("%s[%d]", name, i)
		fn, ok := it.Elem().(eval.Callable)
		if !ok {
			// TODO(xiaq): This is not testable as it depends on stderr.
			// Make it testable.
			diag.Complainf(os.Stderr, "%s not function", name)
			continue
		}
		// TODO(xiaq): This should use stdPorts, but stdPorts is currently
		// unexported from eval.
		ports := []*eval.Port{
			eval.DevNullClosedChan, {File: os.Stdout}, {File: os.Stderr}}
		fm := eval.NewTopFrame(ev, parse.Source{Name: name}, ports)
		out, err := fm.CaptureOutput(func(fm *eval.Frame) error { return fn.Call(fm, args, eval.NoOpts) })
		if err != nil {
			diag.Complainf(os.Stderr, "%s return error", name)
			continue
		}
		if len(out) != 1 {
			diag.Complainf(os.Stderr, "filter %s should only return $true or $false", name)
			continue
		}
		p, ok := out[0].(bool)
		if !ok {
			diag.Complainf(os.Stderr, "filter %s should return bool", name)
			continue
		}
		if !p {
			return false
		}
	}
	return true
}

func newIntVar(i int) vars.PtrVar            { return vars.FromPtr(&i) }
func newFloatVar(f float64) vars.PtrVar      { return vars.FromPtr(&f) }
func newBoolVar(b bool) vars.PtrVar          { return vars.FromPtr(&b) }
func newListVar(l vals.List) vars.PtrVar     { return vars.FromPtr(&l) }
func newMapVar(m vals.Map) vars.PtrVar       { return vars.FromPtr(&m) }
func newFnVar(c eval.Callable) vars.PtrVar   { return vars.FromPtr(&c) }
func newBindingVar(b BindingMap) vars.PtrVar { return vars.FromPtr(&b) }
