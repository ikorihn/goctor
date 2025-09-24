# Tasks: Development Environment Checker

**Input**: Design documents from `/Users/hiroki/work/playground-environment-checker/doctor-cmd/specs/001-sds-macos-linux/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
Go CLI project structure:
- **Main CLI**: `cmd/doctor/main.go`
- **Internal packages**: `internal/*/`
- **Tests**: `tests/contract/`, `tests/integration/`
- **Test data**: `testdata/manifests/`, `testdata/fixtures/`

## Phase 3.1: Setup
- [X] T001 Initialize Go module in doctor-cmd root with go.mod
- [X] T002 Create complete Go project directory structure per plan.md
- [X] T003 [P] Create testdata/manifests/sample.yaml with example tool definitions
- [X] T004 [P] Create testdata/fixtures/tool-outputs.txt with sample command outputs

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**

### Contract Tests [P]
- [X] T005 [P] Contract test for doctor command CLI interface in tests/contract/test_doctor_cmd.go
- [X] T006 [P] Contract test for list command CLI interface in tests/contract/test_list_cmd.go
- [X] T007 [P] Contract test for JSON output schema validation in tests/contract/test_json_output.go
- [X] T008 [P] Contract test for manifest YAML schema validation in tests/contract/test_manifest_schema.go

### Core Type Tests [P]
- [X] T009 [P] Unit tests for ToolDefinition struct and validation in internal/manifest/tool_test.go
- [X] T010 [P] Unit tests for Manifest struct and validation in internal/manifest/manifest_test.go
- [X] T011 [P] Unit tests for CheckResult and status transitions in internal/checker/result_test.go
- [X] T012 [P] Unit tests for semver parsing and comparison in internal/semver/semver_test.go

### Integration Tests [P]
- [X] T013 [P] Integration test for basic environment check scenario in tests/integration/test_basic_check.go
- [X] T014 [P] Integration test for JSON output validation in tests/integration/test_json_output.go
- [X] T015 [P] Integration test for custom manifest loading in tests/integration/test_custom_manifest.go
- [X] T016 [P] Integration test for tool listing functionality in tests/integration/test_tool_listing.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Data Types and Validation [P]
- [X] T017 [P] Implement ToolDefinition struct with validation in internal/manifest/tool.go
- [X] T018 [P] Implement Manifest, ManifestMeta, ManifestDefaults structs in internal/manifest/manifest.go
- [X] T019 [P] Implement CheckResult, CheckStatus, EnvironmentReport structs in internal/checker/result.go
- [X] T020 [P] Implement PlatformInfo and CheckSummary structs in internal/platform/info.go

### Core Business Logic [P]
- [ ] T021 [P] Implement semantic version parsing and comparison in internal/semver/semver.go
- [ ] T022 [P] Implement YAML manifest loading and parsing in internal/manifest/loader.go
- [ ] T023 [P] Implement tool detection and version extraction in internal/checker/detector.go
- [ ] T024 [P] Implement platform detection logic in internal/platform/detector.go

### Output Formatting [P]
- [ ] T025 [P] Implement human-readable output formatter in internal/output/human.go
- [ ] T026 [P] Implement JSON output formatter in internal/output/json.go

### CLI Interface
- [ ] T027 Main CLI application with flag parsing in cmd/doctor/main.go
- [ ] T028 Doctor command handler implementation in cmd/doctor/main.go
- [ ] T029 List command handler implementation in cmd/doctor/main.go

## Phase 3.4: Integration and Error Handling
- [ ] T030 Implement concurrent tool checking with goroutines in internal/checker/detector.go
- [ ] T031 Add timeout handling for tool command execution in internal/checker/detector.go
- [ ] T032 Implement graceful error handling and reporting in internal/checker/detector.go
- [ ] T033 Add remote manifest loading with HTTP client in internal/manifest/loader.go

## Phase 3.5: Polish and Documentation
- [ ] T034 [P] Add comprehensive unit tests for error scenarios in internal/*/errors_test.go
- [ ] T035 [P] Implement command-line help and usage documentation in cmd/doctor/main.go
- [ ] T036 [P] Create embedded default manifest in cmd/doctor/main.go
- [ ] T037 [P] Add version information and build metadata in cmd/doctor/main.go
- [ ] T038 Performance optimization: regex compilation caching in internal/semver/semver.go
- [ ] T039 [P] Create comprehensive README.md with installation and usage guide
- [ ] T040 Run manual testing scenarios from quickstart.md

## Dependencies
**Setup Dependencies:**
- T001 → T002 → all other tasks (project structure required)
- T003, T004 → test tasks (test data required)

**TDD Dependencies:**
- Tests (T005-T016) → Implementation (T017-T029)
- T017-T020 → T021-T024 (types before business logic)
- T021-T026 → T027-T029 (business logic before CLI)

**Implementation Dependencies:**
- T017 → T022 (ToolDefinition before manifest loading)
- T019 → T023 (CheckResult before tool detection)
- T020 → T024 (PlatformInfo before platform detection)
- T025, T026 → T027 (output formatters before CLI)

## Parallel Execution Examples

### Phase 3.2 Contract Tests
```bash
# Launch T005-T008 together:
Task: "Contract test for doctor command CLI interface in tests/contract/test_doctor_cmd.go"
Task: "Contract test for list command CLI interface in tests/contract/test_list_cmd.go"
Task: "Contract test for JSON output schema validation in tests/contract/test_json_output.go"
Task: "Contract test for manifest YAML schema validation in tests/contract/test_manifest_schema.go"
```

### Phase 3.2 Core Type Tests
```bash
# Launch T009-T012 together:
Task: "Unit tests for ToolDefinition struct and validation in internal/manifest/tool_test.go"
Task: "Unit tests for Manifest struct and validation in internal/manifest/manifest_test.go"
Task: "Unit tests for CheckResult and status transitions in internal/checker/result_test.go"
Task: "Unit tests for semver parsing and comparison in internal/semver/semver_test.go"
```

### Phase 3.3 Data Types
```bash
# Launch T017-T020 together:
Task: "Implement ToolDefinition struct with validation in internal/manifest/tool.go"
Task: "Implement Manifest, ManifestMeta, ManifestDefaults structs in internal/manifest/manifest.go"
Task: "Implement CheckResult, CheckStatus, EnvironmentReport structs in internal/checker/result.go"
Task: "Implement PlatformInfo and CheckSummary structs in internal/platform/info.go"
```

### Phase 3.3 Business Logic
```bash
# Launch T021-T024 together:
Task: "Implement semantic version parsing and comparison in internal/semver/semver.go"
Task: "Implement YAML manifest loading and parsing in internal/manifest/loader.go"
Task: "Implement tool detection and version extraction in internal/checker/detector.go"
Task: "Implement platform detection logic in internal/platform/detector.go"
```

## Notes
- [P] tasks target different packages/files with no shared dependencies
- Verify tests fail before implementing (Red-Green-Refactor)
- Follow Go conventions: package names match directory names
- Use table-driven tests for Go unit tests
- Include benchmark tests for performance-critical code (semver, regex)
- Commit after each major task completion
- Use `go fmt`, `go vet`, `go test` for quality checks

## Task Generation Rules Applied

1. **From CLI Contracts**: Each command (doctor, list) → contract test + implementation
2. **From Data Model**: Each entity → struct definition + validation tests + unit tests
3. **From Quickstart Scenarios**: Each scenario → integration test
4. **Go-Specific Patterns**:
   - Package-based parallelization for [P] tasks
   - Test files mirror implementation files
   - Internal packages for implementation details
   - cmd/ package for CLI entry point

## Validation Checklist

- [x] All CLI contracts have corresponding tests (T005-T008)
- [x] All data model entities have implementation tasks (T017-T020)
- [x] All tests come before implementation (Phase 3.2 → 3.3)
- [x] Parallel tasks truly independent (different packages/files)
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] Quickstart scenarios covered by integration tests (T013-T016)