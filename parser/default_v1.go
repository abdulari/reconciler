package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DefaultV1 struct {
	configPath string
	baseDir    string
}

type stateDef struct {
	ID           string `yaml:"id"`
	Name         string `yaml:"name"`
	VerifyScript string `yaml:"verifyScript"`
	ExecScript   string `yaml:"execScript"`
}

type flowRef struct {
	ID string `yaml:"id"`
}

type defaultV1Config struct {
	State []stateDef `yaml:"state"`
	Flow  []flowRef  `yaml:"flow"`
}

func NewDefaultV1(path string) *DefaultV1 {
	return &DefaultV1{
		configPath: path,
		baseDir:    filepath.Dir(path),
	}
}

func (d *DefaultV1) Parse() ([]Step, error) {
	raw, err := os.ReadFile(d.configPath)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg defaultV1Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	return d.resolveFlow(cfg.State, cfg.Flow)
}

func (d *DefaultV1) resolveFlow(states []stateDef, flow []flowRef) ([]Step, error) {
	index := make(map[string]Step, len(states))
	for _, s := range states {
		verify, err := d.resolveScript(s.VerifyScript)
		if err != nil {
			return nil, fmt.Errorf("state %q: resolve verifyScript: %w", s.ID, err)
		}
		exec, err := d.resolveScript(s.ExecScript)
		if err != nil {
			return nil, fmt.Errorf("state %q: resolve execScript: %w", s.ID, err)
		}

		index[s.ID] = Step{
			ID:           s.ID,
			Name:         s.Name,
			VerifyScript: verify,
			ExecScript:   exec,
		}
	}

	var steps []Step
	for _, ref := range flow {
		s, ok := index[ref.ID]
		if !ok {
			return nil, fmt.Errorf("flow references state %q, but no matching state definition exists", ref.ID)
		}
		steps = append(steps, s)
	}

	return steps, nil
}

func (d *DefaultV1) resolveScript(value string) (string, error) {
	if value == "" {
		return "", nil
	}

	scriptPath := filepath.Join(d.baseDir, value)
	if info, err := os.Stat(scriptPath); err == nil && !info.IsDir() {
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			return "", fmt.Errorf("read script file %q: %w", scriptPath, err)
		}
		return string(content), nil
	}

	return value, nil
}