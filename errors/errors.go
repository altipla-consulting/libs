package errors

import (
	"bytes"
	"errors" // revive:disable-line:imports-blacklist
	"fmt"
	"io"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

// stack is a comparable []uintptr slice.
type stack struct {
	frames []uintptr
}

// An altiplaError annotates a cause error with a stacktrace and an explanatory
// message.
type altiplaError struct {
	// The cause error.
	cause error
	// The previous altiplaError, if any.
	previous *altiplaError
	// The current stacktrace. Might be the same as previous' stacktrace if that
	// is another altiplaError.
	stack *stack
	// A small explanatory message what went wrong at this level in the stack.
	reason string
	// The index of the stack frame where this altiplaError was added.
	index int
}

// Error implements error, and outputs a full backtrace.
func (e *altiplaError) Error() string {
	return e.cause.Error()
}

// Cause implements important interfaces in other libs to detect the cause of an error.
func (e *altiplaError) Cause() error {
	return e.cause
}

// GRPCStatus allows us to wrap GRPC errors without losing the original error
// code and retaining the ability to compare it.
func (e *altiplaError) GRPCStatus() *status.Status {
	return status.Convert(e.cause)
}

type errorClasser interface {
	ErrorClass() string
}

// ErrorClass implements NewRelic helpers.
func (e *altiplaError) ErrorClass() string {
	if impl, ok := e.cause.(errorClasser); ok {
		return impl.ErrorClass()
	}
	return e.cause.Error()
}

type errorAttributer interface {
	ErrorAttributes() map[string]interface{}
}

// ErrorAttributes implements NewRelic helpers.
func (e *altiplaError) ErrorAttributes() map[string]interface{} {
	if impl, ok := e.cause.(errorAttributer); ok {
		return impl.ErrorAttributes()
	}
	return nil
}

// StackTrace implements NewRelic helpers.
func (e *altiplaError) StackTrace() []uintptr {
	if e.previous != nil {
		return e.previous.StackTrace()
	}
	return e.stack.frames
}

// A Frame represents a Frame in an altipla callstack. The Reason is the manual
// annotation passed to altipla.Wrapf.
type Frame struct {
	File     string
	Function string
	Line     int
	Reason   string
}

type stackWithReasons struct {
	stack   *stack
	reasons []string
}

// Frames extracts all frames from an altipla error. If err is not an altipla error,
// nil is returned.
func Frames(err error) [][]Frame {
	e, ok := err.(*altiplaError)
	if !ok {
		return nil
	}

	// Walk the chain of altiplaErrors backwards, collecting a set of stacks and
	// reasons.
	stacks := make([]stackWithReasons, 0, 8)
	for ; e != nil; e = e.previous {
		// If the current error's stack is different from the previous, add it to
		// the set of stacks.
		if len(stacks) == 0 || stacks[len(stacks)-1].stack != e.stack {
			stacks = append(stacks, stackWithReasons{
				stack:   e.stack,
				reasons: make([]string, len(e.stack.frames)),
			})
		}
		// Store the reason with its stack frame.
		stacks[len(stacks)-1].reasons[e.index] = e.reason
	}

	parsedStacks := make([][]Frame, 0, len(stacks))

	// Walk the set of stacks backwards, starting with stack most closest to the
	// cause error.
	for i := len(stacks) - 1; i >= 0; i-- {
		frames := stacks[i].stack.frames
		reasons := stacks[i].reasons

		parsedFrames := make([]Frame, 0, 8)

		// Iterate over the stack frames.
		iter := runtime.CallersFrames(frames)
		// j tracks the index in the combined frames / reasons array of iter' stack
		// frame. Each frame in frames / reasons array appears at least once in the
		// iterator's frames, but the iterator's frame might have more frames (for
		// example, cgo frames, or inlined frames.)
		j := 0
		for {
			frame, ok := iter.Next()
			if !ok {
				break
			}

			// Advance j and load the reason whenever the current iterator's frame
			// matches. The iterator's frame's PC might differ by 1 because the
			// iterator adjusts for the difference between callsite and return
			// address.
			var reason string
			if j < len(frames) && (frame.PC == frames[j] || frame.PC+1 == frames[j]) {
				reason = reasons[j]
				j++
			}

			file := frame.File
			i := strings.LastIndex(file, "/src/")
			if i >= 0 {
				file = file[i+len("/src/"):]
			}

			parsedFrames = append(parsedFrames, Frame{
				File:     file,
				Function: frame.Function,
				Line:     frame.Line,
				Reason:   reason,
			})
		}

		parsedStacks = append(parsedStacks, parsedFrames)
	}
	return parsedStacks
}

// writeStackTrace unwinds a chain of altiplaErrors and prints the stacktrace
// annotated with explanatory messages.
func (e *altiplaError) writeStackTrace(w io.Writer) {
	fmt.Fprintf(w, "%s\n\n", e.cause.Error())

	for i, stack := range Frames(e) {
		// Include a newline between stacks.
		if i > 0 {
			fmt.Fprintf(w, "\n")
		}

		for _, frame := range stack {
			// Print the current function.
			if frame.Reason != "" {
				fmt.Fprintf(w, "%s: %s\n", frame.Function, frame.Reason)
			} else {
				fmt.Fprintf(w, "%s\n", frame.Function)
			}
			fmt.Fprintf(w, "\t%s:%d\n", frame.File, frame.Line)
		}
	}
}

// Unwrap implements the standard for unwrappable errors.
func (e *altiplaError) Unwrap() error {
	return e.cause
}

func Details(err error) string {
	e, ok := err.(*altiplaError)
	if !ok {
		return "{" + err.Error() + "}"
	}

	result := []string{
		"{" + e.cause.Error() + "}",
	}
	for i, stack := range Frames(e) {
		if i > 0 {
			result = append(result, "{-----}")
		}

		for _, frame := range stack {
			if frame.Reason != "" {
				result = append(result, fmt.Sprintf("{%s:%d:%s: %s}", frame.File, frame.Line, frame.Function, frame.Reason))
			} else {
				result = append(result, fmt.Sprintf("{%s:%d:%s}", frame.File, frame.Line, frame.Function))
			}
		}
	}
	return strings.Join(result, " ")
}

// isPrefix checks if a is a prefix of b.
func isPrefix(a []uintptr, b []uintptr) bool {
	if len(a) > len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func internalWrapf(err error, reason string) error {
	var cause error
	var previous *altiplaError

	var st *stack
	var index int
	found := false

	if e, ok := err.(*altiplaError); ok {
		cause = e.cause
		previous = e

		// Figure out where we are in the existing callstack. Since Wrapf isn't
		// guaranteed to be called at every stack frame, we need to search to find
		// the current callsite. We start searching one level past the previous
		// level (and assume that Wrapf is called at most once per stack).
		st = e.stack
		index = e.index + 1

		// To check where we are, match a number of return frames in the stack. We
		// check one level deeper than the level we are annotating, because the
		// frame in the calling function likely doesn't match:
		//
		// - parent() calls Wrapf and return an error with a stacktrace - child()
		// check cause's return value and then calls Wrapf on parent's error -
		// compare() is the frame that gets compared
		//
		// When parent calls Wrapf and captures the stack frame, the program
		// counter in child will point the if statement that checks the parent's
		// return value. When the child then calls Wrapf, it's program counter
		// will have advanced to the Wrapf call, and will no longer match the
		// program originally captured by the parent. However, the program counter
		// in compare will still match, and so we compare against that.
		//
		// To paper over small numbers of dupliate frames (eg. when using
		// recursion), we compare not just 1 frame, but several. We compare only
		// some frames (instead of all) to keep the runtime of Wrapf efficient.

		var buffer [8]uintptr
		// 0 is the frame of Callers, 1 is us, 2 is the public wrapper, 3 is its
		// caller (child), 4 is the caller's caller (compare).
		compare := buffer[:runtime.Callers(4, buffer[:])]

		for index+1 < len(st.frames) {
			if isPrefix(compare, st.frames[index+1:]) {
				found = true
				break
			}
			index++
		}
	} else {
		cause = err
	}

	if !found {
		var buffer [256]uintptr
		// 0 is the frame of Callers, 1 is us, 2 is the public wrapper, 3 is its
		// caller.
		n := runtime.Callers(3, buffer[:])
		frames := make([]uintptr, n)
		copy(frames, buffer[:n])

		index = 0
		st = &stack{frames: frames}
	}

	return &altiplaError{
		cause:    cause,
		previous: previous,
		stack:    st,
		reason:   reason,
		index:    index,
	}
}

// Errorf creates a new error with a reason and a stacktrace.
//
// Use Errorf in places where you would otherwise return an error using
// fmt.Errorf or errors.New.
//
// Note that the result of Errorf includes a stacktrace. This means
// that Errorf is not suitable for storing in global variables. For
// such errors, keep using errors.New.
func Errorf(format string, a ...interface{}) error {
	return internalWrapf(fmt.Errorf(format, a...), "")
}

// New creates a new error without stacktrace.
func New(err string) error {
	return errors.New(err)
}

// Wrapf annotates an error with a reason and a stacktrace. If err is nil,
// Wrapf returns nil.
//
// Use Wrapf in places where you would otherwise return an error directly. If
// the error passed to Wrapf is nil, Wrapf will also return nil. This makes it
// safe to use in one-line return statements.
//
// To check if a wrapped error is a specific error, such as io.EOF, you can
// extract the error passed in to Wrapf using Cause.
func Wrapf(err error, format string, a ...interface{}) error {
	if err == nil {
		return nil
	}
	return internalWrapf(err, fmt.Sprintf(format, a...))
}

// Trace annotates an error with a stacktrace.
//
// Use Trace in places where you would otherwise return an error directly. If
// the error passed to Trace is nil, Trace will also return nil. This makes it
// safe to use in one-line return statements.
//
// To check if a wrapped error is a specific error, such as io.EOF, you can
// extract the error passed in to Trace using Cause.
func Trace(err error) error {
	if err == nil {
		return nil
	}
	return internalWrapf(err, "")
}

// Cause extracts the cause error of an altipla error. If err is not an altipla
// error, err itself is returned.
//
// You can use Cause to check if an error is an expected error. For example, if
// you know than EOF error is fine, you can handle it with Cause.
func Cause(err error) error {
	if e, ok := err.(*altiplaError); ok {
		return e.cause
	}
	return err
}

// Is reports whether err or its cause matches target.
func Is(err, target error) bool {
	if target == nil {
		return err == target
	}
	if err == target {
		return true
	}
	if e, ok := err.(*altiplaError); ok {
		return e.cause == target
	}
	return false
}

// Recover recovers from a panic in a defer. If there is no panic, Recover()
// returns nil. To use, call error.Recover(recover()) and compare the result to nil.
func Recover(p interface{}) error {
	if p == nil {
		return nil
	}
	if err, ok := p.(error); ok {
		return internalWrapf(err, "panic")
	}
	return internalWrapf(fmt.Errorf("panic: %v", p), "")
}

func LogFields(err error) log.Fields {
	return log.Fields{
		"error":   err.Error(),
		"details": Details(err),
	}
}

func Stack(err error) string {
	e, ok := err.(*altiplaError)
	if !ok {
		return err.Error()
	}

	var buffer bytes.Buffer
	e.writeStackTrace(&buffer)
	return buffer.String()
}
