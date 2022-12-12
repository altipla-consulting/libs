package box

import (
	"fmt"
	"strings"

	"github.com/logrusorgru/aurora"
)

type Box struct {
	// Horizontal padding in spaces between the border of the box and the content.
	//
	// By default it's 3 spaces.
	Px int

	// Vertical padding in lines between the border of the box and the content.
	//
	// By default it's 1 line.
	Py int

	// Color of the border of the box.
	//
	// By default it's bright yellow: aurora.YellowFg | aurora.BrightFg
	BorderColor aurora.Color

	lengths []int
	lines   []string
}

func (box *Box) maxlength() int {
	var max int
	for _, l := range box.lengths {
		if l > max {
			max = l
		}
	}
	return max
}

// AddLine adds a new line to the output content. It accepts multiple strings
// or aurora.Value for colored output.
//
//	o.AddLine("part1", "part2", "joined without spaces")
//
// If you need colored output split the string, this function should be able to read
// every part individually to calculate the length correctly:
//
//	o.AddLine("before", aurora.Red("colored"), "after", aurora.Blue("another colored"))
func (box *Box) AddLine(parts ...interface{}) {
	var length int
	var line string
	for _, part := range parts {
		switch part := part.(type) {
		case string:
			length += len([]rune(part))
			line += part

		case aurora.Value:
			length += len(part.Reset().String())
			line += part.String()

		default:
			panic(fmt.Sprintf("unknown line type: %T", line))
		}
	}
	box.lengths = append(box.lengths, length)
	box.lines = append(box.lines, line)
}

// Render emits the full box to stdout.
func (box *Box) Render() {
	if box.Px == 0 {
		box.Px = 2
	}
	if box.Py == 1 {
		box.Py = 1
	}
	if box.BorderColor == 0 {
		box.BorderColor = aurora.YellowFg | aurora.BrightFg
	}

	fmt.Println()
	fmt.Println(aurora.BrightYellow("   ╭" + strings.Repeat("─", box.maxlength()+box.Px*2) + "╮"))
	for i := 0; i < box.Py; i++ {
		fmt.Println(aurora.BrightYellow("   │" + strings.Repeat(" ", box.Px*2+box.maxlength()) + "│"))
	}
	for i, line := range box.lines {
		line += strings.Repeat(" ", box.maxlength()-box.lengths[i])
		fmt.Println(aurora.BrightYellow("   │").String() + strings.Repeat(" ", box.Px) + line + strings.Repeat(" ", box.Px) + aurora.BrightYellow("│").String())
	}
	for i := 0; i < box.Py; i++ {
		fmt.Println(aurora.BrightYellow("   │" + strings.Repeat(" ", box.Px*2+box.maxlength()) + "│"))
	}
	fmt.Println(aurora.BrightYellow("   ╰" + strings.Repeat("─", box.maxlength()+box.Px*2) + "╯"))
	fmt.Println()
}
