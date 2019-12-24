package textview

import (
	"reflect"
	"testing"

	"github.com/elves/elvish/pkg/cli/el"
	"github.com/elves/elvish/pkg/cli/term"
	"github.com/elves/elvish/pkg/ui"
)

var bb = term.NewBufferBuilder

var renderTests = []el.RenderTest{
	{
		Name: "text fits entirely",
		Given: NewTextView(TextViewSpec{State: TextViewState{
			Lines: []string{"line 1", "line 2", "line 3"}}}),
		Width: 10, Height: 4,
		Want: bb(10).
			Write("line 1").Newline().
			Write("line 2").Newline().
			Write("line 3").Buffer(),
	},
	{
		Name: "text cropped horizontally",
		Given: NewTextView(TextViewSpec{State: TextViewState{
			Lines: []string{"a very long line"}}}),
		Width: 10, Height: 4,
		Want: bb(10).
			Write("a very lon").Buffer(),
	},
	{
		Name: "text cropped vertically",
		Given: NewTextView(TextViewSpec{State: TextViewState{
			Lines: []string{"line 1", "line 2", "line 3"}}}),
		Width: 10, Height: 2,
		Want: bb(10).
			Write("line 1").Newline().
			Write("line 2").Buffer(),
	},
	{
		Name: "text cropped vertically, with scrollbar",
		Given: NewTextView(TextViewSpec{
			Scrollable: true,
			State: TextViewState{
				Lines: []string{"line 1", "line 2", "line 3", "line 4"}}}),
		Width: 10, Height: 2,
		Want: bb(10).
			Write("line 1   ").
			Write(" ", ui.Inverse, ui.FgMagenta).Newline().
			Write("line 2   ").
			Write("│", ui.FgMagenta).Buffer(),
	},
	{
		Name: "State.First adjusted to fit text",
		Given: NewTextView(TextViewSpec{State: TextViewState{
			First: 2,
			Lines: []string{"line 1", "line 2", "line 3"}}}),
		Width: 10, Height: 3,
		Want: bb(10).
			Write("line 1").Newline().
			Write("line 2").Newline().
			Write("line 3").Buffer(),
	},
}

func TestRender(t *testing.T) {
	el.TestRender(t, renderTests)
}

var handleTests = []el.HandleTest{
	{
		Name: "up doing nothing when not scrollable",
		Given: NewTextView(TextViewSpec{
			State: TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 1}}),
		Event: term.K(ui.Up),

		WantUnhandled: true,
	},
	{
		Name: "up moving window up when scrollable",
		Given: NewTextView(TextViewSpec{
			Scrollable: true,
			State:      TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 1}}),
		Event: term.K(ui.Up),

		WantNewState: TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 0},
	},
	{
		Name: "up doing nothing when already at top",
		Given: NewTextView(TextViewSpec{
			Scrollable: true,
			State:      TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 0}}),
		Event: term.K(ui.Up),

		WantNewState: TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 0},
	},
	{
		Name: "down moving window down when scrollable",
		Given: NewTextView(TextViewSpec{
			Scrollable: true,
			State:      TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 1}}),
		Event: term.K(ui.Down),

		WantNewState: TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 2},
	},
	{
		Name: "down doing nothing when already at bottom",
		Given: NewTextView(TextViewSpec{
			Scrollable: true,
			State:      TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 3}}),
		Event: term.K(ui.Down),

		WantNewState: TextViewState{Lines: []string{"1", "2", "3", "4"}, First: 3},
	},
	{
		Name: "overlay",
		Given: NewTextView(TextViewSpec{
			OverlayHandler: el.MapHandler{term.K('a'): func() {}}}),
		Event: term.K('a'),

		WantNewState: TextViewState{},
	},
}

func TestHandle(t *testing.T) {
	el.TestHandle(t, handleTests)
}

func TestCopyState(t *testing.T) {
	state := TextViewState{Lines: []string{"a", "b", "c"}, First: 1}
	w := NewTextView(TextViewSpec{State: state})
	copied := w.CopyState()
	if !reflect.DeepEqual(copied, state) {
		t.Errorf("Got copied state %v, want %v", copied, state)
	}
}
