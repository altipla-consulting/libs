package box_test

import (
	"testing"

	"github.com/logrusorgru/aurora"
	"libs.altipla.consulting/box"
)

func ExampleBox() {
	o := box.Box{}
	o.AddLine("foo", "bar")
	o.AddLine("before", aurora.Red("colored"), "after")
	o.Render()
}

func TestBox(t *testing.T) {
	o := box.Box{}
	o.AddLine("foo", "bar")
	o.AddLine("before", aurora.Red("colored"), "after")
	o.Render()
}
