# Spec Functional Validation

Functional validation is recorded with a small `validation.yaml` file in each
spec directory that is ready for automated proof. The spec remains the source
of expected behavior; the manifest only maps that behavior to executable checks.

## Manifest Fields

- `feature_id`: Spec directory name, for example `03-models-list-endpoint`.
- `status`: Validation lifecycle status. Use `implemented` while the spec is
  still marked `implemented, functional verification needed`; use `verified`
  only after validation evidence has been accepted.
- `spec`: Relative path to `spec.md`.
- `contracts`: Relative paths to contract files that define the public surface.
- `required_evidence`: Requirement, success-criteria, or compatibility IDs that
  must be proven by the declared checks.
- `test_commands`: Deterministic automated checks, normally focused `go test`
  commands.
- `smoke_commands`: Higher-level checks that exercise gateway behavior through
  a route, provider fake, or live gateway equivalent.

Commands are stored as argv arrays so the validator does not need a shell:

```yaml
test_commands:
  - name: models endpoint contract
    command: ["go", "test", "./internal/api", "-run", "TestModelsListRoute", "-count=1"]
    evidence: ["FR-001", "FR-002"]
```

## Running Validation

```bash
make spec-validate
```

The command writes JSON evidence to `.spec-validation/results.json`. Use strict
mode when the goal is to fail for every spec marked `implemented, functional
verification needed` that does not yet have a manifest:

```bash
go run ./cmd/specvalidate --strict
```
