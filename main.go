package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/abdulari/reconciler/internal/filelog"
	"github.com/abdulari/reconciler/internal/output"
	"github.com/abdulari/reconciler/parser"
	"github.com/abdulari/reconciler/reconciler"
)

type config struct {
	Config string
	Format output.Format
	Every  time.Duration
	LogDir string
}

func main() {
	cfg := parseFlags()
	out := openOutput(cfg)
	logger := output.New(out, cfg.Format)

	if cfg.Every > 0 {
		runForever(cfg.Config, logger, cfg.Every)
		return
	}

	if err := runOnce(cfg.Config, logger); err != nil {
		logger.ScriptOutput(fmt.Sprintf("fatal: %v", err))
		os.Exit(1)
	}
}

func parseFlags() config {
	var (
		configPath   string
		outputFormat string
		everySec     int
		logDir       string
	)

	fs := flagSet()
	fs.StringVar(&configPath, "config", "", "path to configuration file (required)")
	fs.StringVar(&outputFormat, "output", "human", "output format: human or logfmt")
	fs.IntVar(&everySec, "every", 0, "run periodically every N seconds")
	fs.StringVar(&logDir, "log-dir", "", "write logs to dated files in this directory")
	fs.Parse(os.Args[1:])

	if configPath == "" {
		fmt.Fprintln(os.Stderr, "error: -config flag is required")
		fs.Usage()
		os.Exit(1)
	}

	return config{
		Config: configPath,
		Format: output.Format(outputFormat),
		Every:  time.Duration(everySec) * time.Second,
		LogDir: logDir,
	}
}

func flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("reconciler", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Println("\nReconciler - A CLI reconciliation loop engine written in Go.")
		fmt.Fprintf(os.Stderr, "Usage: reconciler -config <file> [options]\n\nOptions:\n")
		fs.PrintDefaults()
	}
	return fs
}

func openOutput(cfg config) io.Writer {
	if cfg.LogDir == "" {
		return os.Stdout
	}
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error: create log directory: %v\n", err)
		os.Exit(1)
	}
	return filelog.New(cfg.LogDir, "reconciler")
}

func runForever(path string, logger output.Logger, interval time.Duration) {
	cycle := 0
	for {
		cycle++

		err := runOnce(path, logger)
		status := "ok"
		if err != nil {
			status = fmt.Sprintf("failed: %v", err)
		}
		fmt.Fprintf(os.Stdout, "cycle=%d status=%q next_sleep=%s\n", cycle, status, interval)

		time.Sleep(interval)
	}
}

func runOnce(path string, logger output.Logger) error {
	p, err := buildParser(path)
	if err != nil {
		return err
	}
	return reconciler.New(p, logger).Run()
}

func buildParser(path string) (parser.Parser, error) {
	parserName, err := readParserName(path)
	if err != nil {
		return nil, fmt.Errorf("read parser type: %w", err)
	}

	switch parserName {
	case "DefaultV1":
		return parser.NewDefaultV1(path), nil
	default:
		return nil, fmt.Errorf("unknown parser type %q", parserName)
	}
}
