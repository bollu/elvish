package eval_test

import (
	"bytes"
	"strings"
	"testing"

	. "github.com/elves/elvish/pkg/eval"
	"github.com/elves/elvish/pkg/parse"
	"github.com/elves/elvish/pkg/prog"
)

func TestDeprecatedBuiltin(t *testing.T) {
	testCompileTimeDeprecation(t, "ord a", `the "ord" command is deprecated`, 15)
	// Deprecations of other builtins are implemented in the same way, so we
	// don't test them repeatedly
}

func testCompileTimeDeprecation(t *testing.T, code, wantWarning string, level int) {
	restore := prog.SetDeprecationLevel(level)
	defer restore()

	ev := NewEvaler()
	errOutput := new(bytes.Buffer)

	parseErr, compileErr := ev.Check(parse.Source{Code: code}, errOutput)
	if parseErr != nil {
		t.Errorf("got parse err %v", parseErr)
	}
	if compileErr != nil {
		t.Errorf("got compile err %v", compileErr)
	}

	warning := errOutput.String()
	if !strings.Contains(warning, wantWarning) {
		t.Errorf("got warning %q, want warning containing %q", warning, wantWarning)
	}
}
