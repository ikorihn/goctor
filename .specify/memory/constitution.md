<!--
Sync Impact Report:
- Version change: [Initial] → 1.0.0
- List of modified principles:
  - Added: Test-Driven Development (NON-NEGOTIABLE)
  - Added: Code Quality Standards
  - Added: Contract-First Design
  - Added: Continuous Integration
  - Added: Documentation as Code
- Added sections: Code Standards, Development Workflow
- Removed sections: None
- Templates requiring updates:
  ✅ plan-template.md (already aligned with TDD principles)
  ✅ tasks-template.md (already aligned with test-first approach)
  ✅ spec-template.md (no changes needed)
- Follow-up TODOs: None
-->

# Doctor-Cmd Constitution

## Core Principles

### I. Test-Driven Development (NON-NEGOTIABLE)
Tests MUST be written before implementation code. Red-Green-Refactor cycle strictly enforced: Write failing test → Make test pass → Refactor for quality. No implementation without corresponding tests. Contract tests for all APIs, integration tests for user workflows, unit tests for business logic.

**Rationale**: TDD ensures code correctness, drives better design, and provides safety net for refactoring. Prevents regression bugs and enforces requirements validation.

### II. Code Quality Standards
All code MUST pass automated quality gates: linting, type checking, security scanning, and code coverage >80%. Follow language-specific conventions and maintain consistent formatting. No code smells: duplicated logic, overly complex functions, or unclear naming.

**Rationale**: High code quality reduces maintenance burden, improves readability, and prevents bugs. Consistent standards enable team collaboration and code review efficiency.

### III. Contract-First Design
APIs and interfaces MUST be designed and documented before implementation. Use schemas (OpenAPI, GraphQL) to define contracts. Generate tests from contracts to ensure compliance. All external integrations require explicit contracts.

**Rationale**: Contract-first approach prevents integration issues, enables parallel development, and serves as living documentation. Reduces communication overhead and API breaking changes.

### IV. Continuous Integration
Every code change MUST pass CI pipeline: tests, quality checks, security scans, and build verification. Failed pipelines block merging. Builds MUST be reproducible and fast (<10 minutes). Automate deployment to staging environments.

**Rationale**: CI catches issues early, ensures consistent quality, and enables rapid feedback. Automated processes reduce human error and deployment risks.

### V. Documentation as Code
Documentation MUST be maintained alongside code in version control. Use markdown for specifications, inline comments for complex logic, and automated API docs generation. Examples and quickstart guides are mandatory for all features.

**Rationale**: Co-located documentation stays current, reduces onboarding time, and improves maintainability. Living documentation prevents knowledge silos and enables self-service.

## Code Standards

Code MUST be self-documenting through clear naming and structure. Use descriptive variable and function names. Limit function complexity (max 15 lines, single responsibility). Organize code in logical modules with clear boundaries. Follow established architectural patterns and avoid premature optimization.

Error handling MUST be explicit and comprehensive. Use typed exceptions where possible. Log errors with sufficient context for debugging. Implement graceful degradation for non-critical failures.

## Development Workflow

All changes MUST follow feature branch workflow with pull request reviews. Minimum two approvals required for merging. Reviewers MUST verify test coverage, code quality, and documentation updates. Commits MUST be atomic with descriptive messages.

Production deployments MUST be approved by designated maintainers. Use blue-green or canary deployment strategies. Implement monitoring and alerting for all critical paths. Maintain rollback procedures for all releases.

## Governance

Constitution supersedes all other development practices. Amendments require documented justification, team approval, and migration plan. All code reviews MUST verify constitutional compliance. Complexity that violates principles MUST be justified in writing with alternative approaches considered.

**Version**: 1.0.0 | **Ratified**: 2025-09-22 | **Last Amended**: 2025-09-22