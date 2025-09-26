# Quickstart Guide: Development Environment Checker

This guide walks through the primary user scenarios to validate the tool works correctly.

## Prerequisites

- macOS or Linux development machine
- Go 1.22+ installed for building (if testing from source)
- Basic development tools for testing scenarios

## Scenario 1: Check Development Environment

**Goal**: Verify that the tool can detect installed and missing tools with clear status reporting.

### Test Steps:
1. **Run basic environment check**:
   ```bash
   goctor
   ```

2. **Expected Output** (example):
   ```
   [OK] go (require >=1.22)
     Installed: 1.22.1

   [Missing] docker (require >=24)
     Installed: not found
     Why: ローカル実行・CI互換のため
     Links:
       - Download  https://docs.docker.com/get-docker/
       - Docs      https://docs.docker.com/

   [Outdated] gh (require >=2.50)
     Installed: 2.30.0
     Links:
       - Download  https://cli.github.com/
       - Docs      https://cli.github.com/manual/
   ```

3. **Verify exit code**:
   ```bash
   echo $?  # Should be 1 (issues found)
   ```

### Success Criteria:
- ✅ Shows clear status for each tool (OK/Missing/Outdated)
- ✅ Displays installed versions when found
- ✅ Provides rationale for required tools
- ✅ Shows helpful links for problematic tools
- ✅ Returns exit code 1 when issues exist

## Scenario 2: Machine-Readable Output

**Goal**: Validate JSON output format for CI integration.

### Test Steps:
1. **Run with JSON flag**:
   ```bash
   goctor --json
   ```

2. **Expected Output Structure**:
   ```json
   {
     "schema_version": 1,
     "platform": {
       "os": "darwin",
       "arch": "arm64",
       "hostname": "dev-machine.local"
     },
     "summary": {
       "total": 3,
       "ok": 1,
       "missing": 1,
       "outdated": 1,
       "errors": 0
     },
     "manifest_source": "embedded",
     "items": [
       {
         "id": "go",
         "name": "Go",
         "status": "ok",
         "required": ">=1.22",
         "installed": "1.22.1",
         "rationale": "Go製リポのビルドに必須",
         "links": {
           "homepage": "https://go.dev/",
           "download": "https://go.dev/dl/"
         },
         "errors": []
       }
     ],
     "generated_at": "2025-09-22T10:30:00Z"
   }
   ```

3. **Validate JSON parsing**:
   ```bash
   goctor --json | jq '.summary.total'  # Should output number
   ```

### Success Criteria:
- ✅ Valid JSON output format
- ✅ All required fields present
- ✅ Platform detection works correctly
- ✅ Summary statistics match individual items
- ✅ Parseable by standard JSON tools

## Scenario 3: Custom Manifest Support

**Goal**: Verify tool can load and use custom tool requirements.

### Test Steps:
1. **Create test manifest** (`test-manifest.yaml`):
   ```yaml
   meta:
     version: 1
     name: "Test Manifest"
     language: "en"

   tools:
     - id: git
       name: "Git"
       rationale: "Version control system"
       require: ">=2.30"
       check:
         cmd: ["git", "--version"]
         regex: "git version (?P<ver>\\d+\\.\\d+\\.\\d+)"
       links:
         homepage: "https://git-scm.com/"
         download: "https://git-scm.com/downloads"
   ```

2. **Run with custom manifest**:
   ```bash
   goctor -f test-manifest.yaml
   ```

3. **Expected Behavior**:
   - Should check only tools defined in custom manifest
   - Should display git version check result
   - Should ignore default company manifest

### Success Criteria:
- ✅ Loads custom manifest successfully
- ✅ Checks only tools from specified manifest
- ✅ Handles manifest loading errors gracefully
- ✅ Shows appropriate error for malformed YAML

## Scenario 4: Tool Listing

**Goal**: Verify the list command works without performing checks.

### Test Steps:
1. **List default tools**:
   ```bash
   goctor list
   ```

2. **Expected Output**:
   ```
   Tools defined in manifest:
   - go: Go (>=1.22 <1.25) - Go製リポのビルドに必須
   - docker: Docker (>=24) - ローカル実行・CI互換のため
   - gh: GitHub CLI (>=2.0) - GitHub操作用
   ```

3. **List with JSON output**:
   ```bash
   goctor list --json
   ```

### Success Criteria:
- ✅ Shows all tools without executing commands
- ✅ Displays tool requirements and rationale
- ✅ Supports both human and JSON output
- ✅ Returns exit code 0 for successful listing

## Scenario 5: Corporate Deployment

**Goal**: Simulate company-wide deployment scenario.

### Test Steps:
1. **Test with remote manifest** (mock):
   ```bash
   # If remote manifest available
   goctor -f https://company.example.com/dev-manifest.yaml
   ```

2. **Test manifest override**:
   ```bash
   # Create personal override
   goctor -f company-manifest.yaml -f personal-overrides.yaml
   ```

3. **CI Integration test**:
   ```bash
   # Simulate CI environment
   export CI=true
   goctor --json > environment-report.json
   if [ $? -eq 0 ]; then
     echo "Environment validation passed"
   else
     echo "Environment validation failed"
     cat environment-report.json
   fi
   ```

### Success Criteria:
- ✅ Handles remote manifest loading
- ✅ Supports manifest merging/overrides
- ✅ Appropriate for CI integration
- ✅ Clear exit codes for automation

## Error Scenarios

### Network Issues
Test behavior when remote manifest is unavailable:
```bash
goctor -f https://invalid.example.com/manifest.yaml
```

**Expected**: Clear error message, graceful fallback behavior

### Permission Issues
Test with restricted command execution:
```bash
# Run in environment where some tools require elevated permissions
```

**Expected**: Proper error reporting without crashing

### Malformed Manifest
Test with invalid YAML:
```bash
echo "invalid: yaml: content" > bad-manifest.yaml
goctor -f bad-manifest.yaml
```

**Expected**: Clear YAML parsing error message

## Performance Validation

### Timeout Handling
Verify tool checks don't hang indefinitely:
- Should complete within reasonable time (< 30 seconds total)
- Individual tool checks should timeout appropriately
- Should handle slow/hanging commands gracefully

### Concurrent Execution
With multiple tools, execution should be reasonably fast:
- Tools that can be checked independently should run concurrently
- Total execution time should not be sum of individual check times

## Manual Testing Checklist

Before release, manually verify:

- [ ] Works on macOS (Intel and Apple Silicon)
- [ ] Works on Linux (Ubuntu, CentOS)
- [ ] Handles missing tools correctly
- [ ] Handles outdated tools correctly
- [ ] JSON output is valid and complete
- [ ] Links are accessible and helpful
- [ ] Exit codes are appropriate
- [ ] Error messages are clear and actionable
- [ ] Performance is acceptable for CI use
- [ ] Help text is comprehensive
