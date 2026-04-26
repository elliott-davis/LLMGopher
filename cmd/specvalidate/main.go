package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

const implementedNeedsVerification = "implemented, functional verification needed"

type manifest struct {
	FeatureID        string        `yaml:"feature_id"`
	Status           string        `yaml:"status"`
	Spec             string        `yaml:"spec"`
	Contracts        []string      `yaml:"contracts"`
	RequiredEvidence []string      `yaml:"required_evidence"`
	TestCommands     []commandSpec `yaml:"test_commands"`
	SmokeCommands    []commandSpec `yaml:"smoke_commands"`
}

type commandSpec struct {
	Name     string   `yaml:"name"`
	Command  []string `yaml:"command"`
	Timeout  string   `yaml:"timeout"`
	Evidence []string `yaml:"evidence"`
}

type report struct {
	GeneratedAt time.Time    `json:"generated_at"`
	SpecsDir    string       `json:"specs_dir"`
	Strict      bool         `json:"strict"`
	DryRun      bool         `json:"dry_run"`
	Summary     reportTotals `json:"summary"`
	Specs       []specReport `json:"specs"`
}

type reportTotals struct {
	Specs        int `json:"specs"`
	Passed       int `json:"passed"`
	Failed       int `json:"failed"`
	Skipped      int `json:"skipped"`
	Checks       int `json:"checks"`
	CheckPassed  int `json:"check_passed"`
	CheckFailed  int `json:"check_failed"`
	CheckSkipped int `json:"check_skipped"`
}

type specReport struct {
	FeatureID    string        `json:"feature_id"`
	ManifestPath string        `json:"manifest_path,omitempty"`
	SpecPath     string        `json:"spec_path"`
	SpecStatus   string        `json:"spec_status,omitempty"`
	Status       string        `json:"status"`
	Contracts    []string      `json:"contracts,omitempty"`
	Evidence     []string      `json:"required_evidence,omitempty"`
	Checks       []checkReport `json:"checks,omitempty"`
	Errors       []string      `json:"errors,omitempty"`
	DurationMS   int64         `json:"duration_ms"`
}

type checkReport struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Command    []string `json:"command"`
	Evidence   []string `json:"evidence,omitempty"`
	Status     string   `json:"status"`
	ExitCode   int      `json:"exit_code,omitempty"`
	DurationMS int64    `json:"duration_ms"`
	Output     string   `json:"output,omitempty"`
	Error      string   `json:"error,omitempty"`
}

type options struct {
	specsDir string
	report   string
	strict   bool
	dryRun   bool
	timeout  time.Duration
}

func main() {
	var opts options
	flag.StringVar(&opts.specsDir, "specs-dir", "specs", "directory containing numbered spec folders")
	flag.StringVar(&opts.report, "report", ".spec-validation/results.json", "path for JSON validation evidence")
	flag.BoolVar(&opts.strict, "strict", false, "fail when an implemented spec lacks a validation manifest")
	flag.BoolVar(&opts.dryRun, "dry-run", false, "validate manifests without executing commands")
	flag.DurationVar(&opts.timeout, "timeout", 5*time.Minute, "default timeout per validation command")
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(opts options) error {
	if opts.specsDir == "" {
		return errors.New("specs-dir is required")
	}
	if opts.timeout <= 0 {
		return errors.New("timeout must be greater than zero")
	}

	start := time.Now()
	rep := report{
		GeneratedAt: time.Now().UTC(),
		SpecsDir:    opts.specsDir,
		Strict:      opts.strict,
		DryRun:      opts.dryRun,
	}

	targets, err := discoverTargets(opts.specsDir, opts.strict)
	if err != nil {
		return err
	}
	for _, target := range targets {
		specStart := time.Now()
		sr := validateTarget(target, opts)
		sr.DurationMS = time.Since(specStart).Milliseconds()
		rep.Specs = append(rep.Specs, sr)
	}

	for _, sr := range rep.Specs {
		rep.Summary.Specs++
		switch sr.Status {
		case "passed":
			rep.Summary.Passed++
		case "skipped":
			rep.Summary.Skipped++
		default:
			rep.Summary.Failed++
		}
		for _, cr := range sr.Checks {
			rep.Summary.Checks++
			switch cr.Status {
			case "passed":
				rep.Summary.CheckPassed++
			case "skipped":
				rep.Summary.CheckSkipped++
			default:
				rep.Summary.CheckFailed++
			}
		}
	}

	if err := writeReport(opts.report, rep); err != nil {
		return err
	}

	fmt.Printf("spec validation: %d passed, %d failed, %d skipped in %s\n",
		rep.Summary.Passed, rep.Summary.Failed, rep.Summary.Skipped, time.Since(start).Round(time.Millisecond))
	fmt.Printf("evidence written to %s\n", opts.report)

	if rep.Summary.Failed > 0 {
		return fmt.Errorf("spec validation failed: %d spec(s) failed", rep.Summary.Failed)
	}
	return nil
}

type validationTarget struct {
	featureID    string
	specPath     string
	manifestPath string
	missing      bool
}

func discoverTargets(specsDir string, strict bool) ([]validationTarget, error) {
	manifestPaths, err := filepath.Glob(filepath.Join(specsDir, "*", "validation.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(manifestPaths)

	targetsBySpec := make(map[string]validationTarget, len(manifestPaths))
	for _, manifestPath := range manifestPaths {
		specPath := filepath.Join(filepath.Dir(manifestPath), "spec.md")
		featureID := filepath.Base(filepath.Dir(manifestPath))
		targetsBySpec[specPath] = validationTarget{
			featureID:    featureID,
			specPath:     specPath,
			manifestPath: manifestPath,
		}
	}

	if strict {
		specPaths, err := filepath.Glob(filepath.Join(specsDir, "*", "spec.md"))
		if err != nil {
			return nil, err
		}
		sort.Strings(specPaths)
		for _, specPath := range specPaths {
			status, err := readSpecStatus(specPath)
			if err != nil || !strings.Contains(status, implementedNeedsVerification) {
				continue
			}
			if _, ok := targetsBySpec[specPath]; !ok {
				featureID := filepath.Base(filepath.Dir(specPath))
				targetsBySpec[specPath] = validationTarget{
					featureID: featureID,
					specPath:  specPath,
					missing:   true,
				}
			}
		}
	}

	targets := make([]validationTarget, 0, len(targetsBySpec))
	for _, target := range targetsBySpec {
		targets = append(targets, target)
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].specPath < targets[j].specPath
	})
	return targets, nil
}

func validateTarget(target validationTarget, opts options) specReport {
	sr := specReport{
		FeatureID:    target.featureID,
		ManifestPath: target.manifestPath,
		SpecPath:     target.specPath,
		Status:       "passed",
	}

	specStatus, err := readSpecStatus(target.specPath)
	if err != nil {
		sr.Errors = append(sr.Errors, err.Error())
		sr.Status = "failed"
	} else {
		sr.SpecStatus = specStatus
	}

	if target.missing {
		sr.Errors = append(sr.Errors, "implemented spec is missing validation.yaml")
		sr.Status = "failed"
		return sr
	}

	manifest, err := loadManifest(target.manifestPath)
	if err != nil {
		sr.Errors = append(sr.Errors, err.Error())
		sr.Status = "failed"
		return sr
	}
	if manifest.FeatureID != "" {
		sr.FeatureID = manifest.FeatureID
	}
	sr.Contracts = manifest.Contracts
	sr.Evidence = manifest.RequiredEvidence

	if err := validateManifestPaths(target, manifest); err != nil {
		sr.Errors = append(sr.Errors, err.Error())
		sr.Status = "failed"
	}

	checks := appendCommands("automated", manifest.TestCommands)
	checks = append(checks, appendCommands("smoke", manifest.SmokeCommands)...)
	if len(checks) == 0 {
		sr.Errors = append(sr.Errors, "validation manifest must declare at least one test or smoke command")
		sr.Status = "failed"
		return sr
	}

	for _, check := range checks {
		cr := runCheck(check, opts)
		sr.Checks = append(sr.Checks, cr)
		if cr.Status == "failed" {
			sr.Status = "failed"
		}
	}
	return sr
}

type typedCommand struct {
	kind string
	spec commandSpec
}

func appendCommands(kind string, commands []commandSpec) []typedCommand {
	out := make([]typedCommand, 0, len(commands))
	for _, command := range commands {
		out = append(out, typedCommand{kind: kind, spec: command})
	}
	return out
}

func runCheck(check typedCommand, opts options) checkReport {
	cr := checkReport{
		Name:     check.spec.Name,
		Type:     check.kind,
		Command:  check.spec.Command,
		Evidence: check.spec.Evidence,
		Status:   "passed",
	}
	if opts.dryRun {
		cr.Status = "skipped"
		return cr
	}

	timeout := opts.timeout
	if check.spec.Timeout != "" {
		parsed, err := time.ParseDuration(check.spec.Timeout)
		if err != nil {
			cr.Status = "failed"
			cr.Error = fmt.Sprintf("invalid timeout %q: %v", check.spec.Timeout, err)
			return cr
		}
		timeout = parsed
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, check.spec.Command[0], check.spec.Command[1:]...)
	output, err := cmd.CombinedOutput()
	cr.DurationMS = time.Since(start).Milliseconds()
	cr.Output = strings.TrimSpace(string(output))
	if ctx.Err() == context.DeadlineExceeded {
		cr.Status = "failed"
		cr.Error = fmt.Sprintf("command timed out after %s", timeout)
		return cr
	}
	if err != nil {
		cr.Status = "failed"
		cr.Error = err.Error()
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) {
			cr.ExitCode = exitErr.ExitCode()
		}
	}
	return cr
}

func loadManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, err
	}
	var m manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return manifest{}, fmt.Errorf("%s: %w", path, err)
	}
	if strings.TrimSpace(m.FeatureID) == "" {
		return manifest{}, fmt.Errorf("%s: feature_id is required", path)
	}
	if strings.TrimSpace(m.Status) == "" {
		return manifest{}, fmt.Errorf("%s: status is required", path)
	}
	if strings.TrimSpace(m.Spec) == "" {
		return manifest{}, fmt.Errorf("%s: spec is required", path)
	}
	for _, command := range append(m.TestCommands, m.SmokeCommands...) {
		if strings.TrimSpace(command.Name) == "" {
			return manifest{}, fmt.Errorf("%s: command name is required", path)
		}
		if len(command.Command) == 0 {
			return manifest{}, fmt.Errorf("%s: command %q must include an argv command", path, command.Name)
		}
	}
	return m, nil
}

func validateManifestPaths(target validationTarget, m manifest) error {
	base := filepath.Dir(target.manifestPath)
	specPath := filepath.Clean(filepath.Join(base, m.Spec))
	if specPath != filepath.Clean(target.specPath) {
		return fmt.Errorf("%s: spec points to %s, want %s", target.manifestPath, specPath, target.specPath)
	}
	for _, contract := range m.Contracts {
		contractPath := filepath.Join(base, contract)
		if _, err := os.Stat(contractPath); err != nil {
			return fmt.Errorf("%s: contract %s is not readable: %w", target.manifestPath, contract, err)
		}
	}
	return nil
}

var statusPattern = regexp.MustCompile(`(?m)^\*\*Status\*\*:\s*(.+?)\s*$`)

func readSpecStatus(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	match := statusPattern.FindSubmatch(data)
	if match == nil {
		return "", fmt.Errorf("%s: status line not found", path)
	}
	return strings.TrimSpace(string(match[1])), nil
}

func writeReport(path string, rep report) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
