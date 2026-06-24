package reconciler

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/abdulari/reconciler/internal/output"
	"github.com/abdulari/reconciler/parser"
)

type Reconciler struct {
	parser parser.Parser
	logger output.Logger
}

func New(p parser.Parser, l output.Logger) *Reconciler {
	return &Reconciler{parser: p, logger: l}
}

func (r *Reconciler) Run() error {
	steps, err := r.parser.Parse()
	if err != nil {
		return fmt.Errorf("parse configuration: %w", err)
	}

	for _, step := range steps {
		if err := r.reconcileStep(step); err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) reconcileStep(step parser.Step) error {
	r.logger.StepStart(step.ID, step.Name)

	if r.runScript(step.VerifyScript) {
		r.logger.StepSkip(step.ID)
		return nil
	}

	r.logger.StepExec(step.ID)
	if !r.runScript(step.ExecScript) {
		r.logger.StepFail(step.ID, "exec", fmt.Errorf("execScript failed"))
		return fmt.Errorf("step %q execScript failed", step.ID)
	}

	r.logger.StepVerify(step.ID)
	if !r.runScript(step.VerifyScript) {
		r.logger.StepFail(step.ID, "verify", fmt.Errorf("state not fulfilled after execScript"))
		return fmt.Errorf("step %q still not fulfilled after execScript", step.ID)
	}

	r.logger.StepDone(step.ID)
	return nil
}

func (r *Reconciler) runScript(script string) bool {
	if script == "" {
		return true
	}

	cmd := exec.Command("sh", "-c", script)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		r.logger.ScriptOutput(fmt.Sprintf("stdout pipe error: %v", err))
		return false
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		r.logger.ScriptOutput(fmt.Sprintf("stderr pipe error: %v", err))
		return false
	}

	if err := cmd.Start(); err != nil {
		r.logger.ScriptOutput(fmt.Sprintf("start error: %v", err))
		return false
	}

	r.captureOutput(stdout, stderr)

	return cmd.Wait() == nil
}

func (r *Reconciler) captureOutput(readers ...io.Reader) {
	lines := make(chan string)
	done := make(chan struct{})

	go func() {
		for line := range lines {
			r.logger.ScriptOutput(line)
		}
		close(done)
	}()

	var wg sync.WaitGroup
	for _, rd := range readers {
		wg.Add(1)
		go func(src io.Reader) {
			defer wg.Done()
			sc := bufio.NewScanner(src)
			for sc.Scan() {
				lines <- sc.Text()
			}
		}(rd)
	}
	wg.Wait()
	close(lines)
	<-done
}