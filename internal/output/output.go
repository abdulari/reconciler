package output

import (
	"fmt"
	"io"
	"time"
)

type Logger interface {
	StepStart(id, name string)
	StepSkip(id string)
	StepExec(id string)
	StepVerify(id string)
	StepDone(id string)
	StepFail(id, phase string, err error)
	ScriptOutput(line string)
}

type Format string

const (
	FormatHuman  Format = "human"
	FormatLogfmt Format = "logfmt"
)

func New(w io.Writer, format Format) Logger {
	switch format {
	case FormatLogfmt:
		return &logfmtLogger{w: w}
	default:
		return &humanLogger{w: w}
	}
}

type humanLogger struct {
	w io.Writer
}

func (l *humanLogger) StepStart(id, name string) {
	fmt.Fprintf(l.w, "── %s (%s) ──\n", id, name)
}

func (l *humanLogger) StepSkip(id string) {
	fmt.Fprintf(l.w, "  ✓ %s already fulfilled, skipping\n", id)
}

func (l *humanLogger) StepExec(id string) {
	fmt.Fprintf(l.w, "  ▶ %s not fulfilled, running execScript\n", id)
}

func (l *humanLogger) StepVerify(id string) {
	fmt.Fprintf(l.w, "  ◇ %s verifying state after execScript\n", id)
}

func (l *humanLogger) StepDone(id string) {
	fmt.Fprintf(l.w, "  ✔ %s done\n", id)
}

func (l *humanLogger) StepFail(id, phase string, err error) {
	fmt.Fprintf(l.w, "  ✘ %s %s: %v\n", id, phase, err)
}

func (l *humanLogger) ScriptOutput(line string) {
	fmt.Fprintf(l.w, "    | %s\n", line)
}

type logfmtLogger struct {
	w io.Writer
}

func (l *logfmtLogger) emit(kv ...string) {
	now := time.Now().Format(time.RFC3339)
	fmt.Fprintf(l.w, "time=%q", now)
	for i := 0; i < len(kv); i += 2 {
		fmt.Fprintf(l.w, " %s=%q", kv[i], kv[i+1])
	}
	fmt.Fprintln(l.w)
}

func (l *logfmtLogger) StepStart(id, name string) {
	l.emit("id", id, "event", "step_start", "name", name)
}

func (l *logfmtLogger) StepSkip(id string) {
	l.emit("id", id, "event", "step_skip")
}

func (l *logfmtLogger) StepExec(id string) {
	l.emit("id", id, "event", "step_exec")
}

func (l *logfmtLogger) StepVerify(id string) {
	l.emit("id", id, "event", "step_verify")
}

func (l *logfmtLogger) StepDone(id string) {
	l.emit("id", id, "event", "step_done")
}

func (l *logfmtLogger) StepFail(id, phase string, err error) {
	l.emit("id", id, "event", "step_fail", "phase", phase, "error", err.Error())
}

func (l *logfmtLogger) ScriptOutput(line string) {
	l.emit("event", "script_output", "line", line)
}