package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDiscoverTargetsStrictReportsImplementedSpecWithoutManifest(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "01-example")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("**Status**: Draft - implemented, functional verification needed\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}

	targets, err := discoverTargets(root, true)
	if err != nil {
		t.Fatalf("discoverTargets: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("targets length = %d, want 1", len(targets))
	}
	if !targets[0].missing {
		t.Fatal("target missing = false, want true")
	}
}

func TestRunDryRunWritesPassingReport(t *testing.T) {
	root := t.TempDir()
	specDir := filepath.Join(root, "03-models-list-endpoint")
	contractDir := filepath.Join(specDir, "contracts")
	if err := os.MkdirAll(contractDir, 0o755); err != nil {
		t.Fatalf("mkdir spec dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(specDir, "spec.md"), []byte("**Status**: Draft - implemented, functional verification needed\n"), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	if err := os.WriteFile(filepath.Join(contractDir, "models.md"), []byte("# Contract\n"), 0o644); err != nil {
		t.Fatalf("write contract: %v", err)
	}
	manifest := []byte(`feature_id: 03-models-list-endpoint
status: implemented
spec: spec.md
contracts:
  - contracts/models.md
required_evidence:
  - FR-001
test_commands:
  - name: sample automated check
    command: ["go", "version"]
smoke_commands:
  - name: sample smoke check
    command: ["go", "version"]
`)
	if err := os.WriteFile(filepath.Join(specDir, "validation.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	reportPath := filepath.Join(root, "results.json")
	err := run(options{
		specsDir: root,
		report:   reportPath,
		dryRun:   true,
		timeout:  time.Minute,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var rep report
	if err := json.Unmarshal(data, &rep); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if rep.Summary.Passed != 1 || rep.Summary.CheckSkipped != 2 {
		t.Fatalf("summary = %+v, want one passed spec with two skipped checks", rep.Summary)
	}
}

func TestLoadManifestRejectsMissingCommandArgv(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "validation.yaml")
	manifest := []byte(`feature_id: 01-example
status: implemented
spec: spec.md
test_commands:
  - name: missing argv
`)
	if err := os.WriteFile(path, manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	if _, err := loadManifest(path); err == nil {
		t.Fatal("loadManifest error = nil, want error")
	}
}
