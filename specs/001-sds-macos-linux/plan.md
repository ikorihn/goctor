
# Implementation Plan: Development Environment Checker

**Branch**: `001-sds-macos-linux` | **Date**: 2025-09-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/Users/hiroki/work/playground-environment-checker/doctor-cmd/specs/001-sds-macos-linux/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Development environment checker CLI tool that verifies required development tools are installed with correct versions, provides guidance links for missing/outdated tools, and supports corporate manifest management. Implementation in Go using only standard library for minimal dependencies and maximum portability.

## Technical Context
**Language/Version**: Go 1.22+ (using standard library only per user requirement)
**Primary Dependencies**: None (standard library only), YAML parsing via encoding/yaml, semver via custom implementation
**Storage**: YAML manifest files (local and remote), no persistent data storage required
**Testing**: Go testing package (go test), table-driven tests for CLI scenarios
**Target Platform**: macOS and Linux (darwin/amd64, darwin/arm64, linux/amd64, linux/arm64)
**Project Type**: single (CLI tool)
**Performance Goals**: Tool version checks <5 seconds timeout, manifest loading <10 seconds
**Constraints**: No external dependencies, no shell execution, regex-based version parsing, read-only operations only
**Scale/Scope**: Corporate deployment (100s of developers), 20-50 tools per manifest, CI integration

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**I. Test-Driven Development (NON-NEGOTIABLE)**: ✅ PASS
- Will implement comprehensive unit tests for semver parsing, tool detection, manifest loading
- Contract tests for CLI command interfaces (doctor, list commands)
- Integration tests for end-to-end user scenarios from spec

**II. Code Quality Standards**: ✅ PASS
- Go fmt, go vet, go lint for code quality
- Target >90% test coverage given critical nature of development toolchain verification
- Clear naming conventions, max 15-line functions, single responsibility

**III. Contract-First Design**: ✅ PASS
- CLI interface contracts defined (commands, flags, exit codes, output formats)
- YAML manifest schema specification
- JSON output schema for machine consumption

**IV. Continuous Integration**: ✅ PASS
- Multi-platform build verification (macOS/Linux, multiple architectures)
- Test execution on both target platforms
- Binary distribution with checksums

**V. Documentation as Code**: ✅ PASS
- Embedded usage help, manifest examples
- Quickstart guide for onboarding
- README with clear installation and usage instructions

## Project Structure

### Documentation (this feature)
```
specs/001-sds-macos-linux/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
# Go CLI Project Structure
cmd/
└── doctor/              # Main CLI application
    └── main.go

internal/
├── manifest/            # Manifest loading and parsing
├── checker/             # Tool detection and version checking
├── semver/              # Semantic version parsing and comparison
├── output/              # Human and JSON output formatting
└── platform/            # Platform-specific operations

testdata/
├── manifests/           # Sample manifest files for testing
└── fixtures/            # Test fixture data

tests/
├── integration/         # End-to-end CLI tests
└── contract/            # CLI interface contract tests

# Generated files
go.mod                   # Go module definition
go.sum                   # Dependency checksums
README.md                # Usage documentation
```

**Structure Decision**: Go CLI single project structure following Go conventions

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh claude`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate tasks from Phase 1 design docs (contracts, data model, quickstart)
- CLI interface contracts → command handler test tasks [P]
- Data model entities → struct and validation test tasks [P]
- Quickstart scenarios → integration test tasks [P]
- Implementation tasks following TDD approach

**Go-Specific Task Categories**:
1. **Project Setup**: go.mod, directory structure, basic CLI framework
2. **Core Types**: manifest, tool definition, check result structs [P]
3. **Contract Tests**: CLI command interface validation [P]
4. **Business Logic**: semver parsing, tool detection, manifest loading [P]
5. **Integration**: end-to-end scenarios from quickstart [P]
6. **Output Formatting**: human-readable and JSON formatters [P]
7. **Build & Distribution**: cross-platform builds, binary packaging

**Ordering Strategy**:
- TDD order: Tests before implementation (Go testing package)
- Dependency order: Types → Business Logic → CLI Interface → Integration
- Mark [P] for parallel execution (independent packages/files)
- Go package structure enables parallel development

**Estimated Output**: 20-25 numbered, ordered tasks in tasks.md
- Fewer tasks due to Go's simplicity and standard library usage
- Focus on core functionality with minimal external dependencies

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command)
- [x] Phase 1: Design complete (/plan command)
- [x] Phase 2: Task planning complete (/plan command - describe approach only)
- [ ] Phase 3: Tasks generated (/tasks command)
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved
- [x] Complexity deviations documented

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
