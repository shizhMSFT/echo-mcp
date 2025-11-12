<!--
  SYNC IMPACT REPORT
  Version Change: Initial → 1.0.0
  Rationale: Initial constitution establishment with comprehensive governance framework
  
  Principles Established:
  - Code Quality Standards (mandatory code review, static analysis, formatting)
  - Test-First Development (TDD with Red-Green-Refactor)
  - User Experience Consistency (MCP protocol compliance, error handling)
  - Performance Requirements (latency, resource usage, scalability)
  
  Templates Status:
  ✅ plan-template.md - Constitution Check section aligns with all principles
  ✅ spec-template.md - User scenarios and requirements sections support UX consistency
  ✅ tasks-template.md - Phase structure supports test-first and quality gates
  
  Follow-up Actions: None - all placeholders resolved
-->

# Echo MCP Server Constitution

## Core Principles

### I. Code Quality Standards (NON-NEGOTIABLE)

All code submitted to this project MUST adhere to the following quality standards:

- **Static Analysis**: Code MUST pass `go vet` and `golangci-lint` with zero errors before merge
- **Formatting**: All Go code MUST be formatted with `gofmt` and adhere to Go standard style guidelines
- **Code Review**: Every change MUST be reviewed by at least one other developer before merge
- **Documentation**: Public functions, types, and packages MUST have meaningful godoc comments
- **Complexity**: Functions with cyclomatic complexity >15 MUST be refactored or explicitly justified
- **Dependencies**: New dependencies MUST be justified, security-scanned, and approved

**Rationale**: Code quality directly impacts maintainability, debuggability, and long-term project
health. MCP servers are integration points that require high reliability and clear interfaces.

### II. Test-First Development (NON-NEGOTIABLE)

This project follows strict Test-Driven Development (TDD) discipline:

- **Red-Green-Refactor Cycle**: Tests MUST be written first, verified to fail, then implementation
  makes them pass
- **Coverage Requirements**: All new code MUST maintain minimum 80% test coverage; critical paths
  require 95%+
- **Test Categories Required**:
  - Unit tests for all business logic
  - Contract tests for all MCP protocol interactions
  - Integration tests for end-to-end message flows
- **Test Independence**: Each test MUST be independently executable and deterministic
- **No Tests = No Merge**: Pull requests without corresponding tests will be rejected

**Rationale**: TDD ensures correctness, provides living documentation, enables fearless refactoring,
and catches regressions early. For protocol servers, contract tests are critical to ensure
compliance with MCP specifications.

### III. User Experience Consistency

All user-facing aspects MUST provide consistent, predictable, and ergonomic experiences:

- **MCP Protocol Compliance**: Server MUST strictly adhere to Model Context Protocol specification
  version compatibility
- **Error Messages**: Errors MUST be actionable, include context, and follow standardized formats
  (JSON for MCP responses)
- **Response Format**: All responses MUST follow consistent schema patterns; breaking changes
  require major version bump
- **Echo Behavior**: Echo responses MUST preserve message structure and metadata; transformations
  MUST be explicitly documented
- **Logging**: User-facing logs MUST be human-readable; structured logs for machine consumption
  MUST be valid JSON
- **Configuration**: Server behavior MUST be configurable via environment variables or config files;
  defaults MUST be sensible

**Rationale**: MCP servers are developer tools. Inconsistent behavior erodes trust and increases
debugging time. Echo servers specifically must be reliable reference implementations.

### IV. Performance Requirements

Performance is a feature. All implementations MUST meet these non-negotiable benchmarks:

- **Latency**: Echo responses MUST complete within 50ms p95 latency for messages <1MB
- **Throughput**: Server MUST handle minimum 1,000 requests/second on standard hardware
  (4 CPU cores, 8GB RAM)
- **Memory**: Server MUST operate within 100MB resident memory under normal load; no memory leaks
- **Concurrency**: Server MUST safely handle concurrent requests; race conditions are critical bugs
- **Graceful Degradation**: Under load, server MUST queue or rate-limit rather than crash
- **Resource Cleanup**: All goroutines, connections, and file handles MUST be properly cleaned up

**Performance Testing Required**:
- Benchmark tests for critical paths
- Load testing before releases
- Memory profiling for resource-intensive operations

**Rationale**: MCP servers run as long-lived processes in developer environments. Poor performance
impacts productivity across all users. Echo servers often serve as testing infrastructure, where
performance bottlenecks can mask issues in client implementations.

## Quality Gates

All code contributions MUST pass these automated and manual gates before merge:

### Automated Gates

- **Build**: `go build` MUST succeed with no warnings
- **Tests**: `go test ./...` MUST pass with >80% coverage
- **Linting**: `golangci-lint run` MUST report zero issues
- **Race Detection**: `go test -race ./...` MUST pass
- **Benchmarks**: Performance regressions >10% MUST be justified

### Manual Gates

- **Code Review**: Minimum one approval from project maintainer
- **Constitution Compliance**: Reviewer MUST verify adherence to all four core principles
- **Documentation**: Changes to public API MUST include updated godoc and README examples
- **Breaking Changes**: Any protocol or API changes MUST include migration guide

## Development Workflow

### Feature Development Process

1. **Specification Phase**:
   - Create feature specification using `.specify/templates/spec-template.md`
   - Define user scenarios with Given-When-Then acceptance criteria
   - Identify required test coverage

2. **Planning Phase**:
   - Generate implementation plan using `.specify/templates/plan-template.md`
   - Verify Constitution Check section passes all principles
   - Document performance impact analysis if applicable

3. **Implementation Phase**:
   - Create feature branch: `###-feature-name`
   - Write failing tests first (Red)
   - Implement minimum code to pass tests (Green)
   - Refactor for quality while keeping tests green (Refactor)
   - Run quality gates continuously

4. **Review Phase**:
   - Submit pull request with test evidence
   - Address review feedback
   - Ensure all automated gates pass
   - Obtain maintainer approval

5. **Integration Phase**:
   - Merge to main branch
   - Verify CI/CD pipeline success
   - Monitor production metrics if applicable

### Complexity Justification Process

When code exceeds complexity thresholds (cyclomatic complexity >15, deeply nested logic,
performance trade-offs), developers MUST:

1. Document the complexity in code comments with rationale
2. Include alternative approaches considered and why they were rejected
3. Add compensating controls (extra tests, monitoring, documentation)
4. Obtain explicit approval in code review

## Versioning and Breaking Changes

This project follows **Semantic Versioning 2.0.0**:

- **MAJOR**: Breaking changes to MCP protocol compliance or public API
- **MINOR**: New features, backward-compatible enhancements
- **PATCH**: Bug fixes, performance improvements, documentation updates

**Breaking Change Policy**:
- MUST be documented in CHANGELOG.md with migration guide
- MUST include deprecation warnings in previous minor version when feasible
- MUST update all examples and documentation
- MUST be approved by project owner

## Governance

### Amendment Process

1. Propose constitution change via pull request to `.specify/memory/constitution.md`
2. Include rationale, impact analysis, and template alignment updates
3. Require consensus approval from all active maintainers
4. Update constitution version following semantic versioning:
   - MAJOR: Principle removal or incompatible redefinition
   - MINOR: New principle or expanded guidance
   - PATCH: Clarifications, wording fixes, non-semantic changes
5. Propagate changes to all dependent templates and documentation

### Compliance Review

- **Per Pull Request**: Reviewers MUST verify all four core principles are satisfied
- **Monthly Audit**: Project maintainers audit recent merges for constitution compliance
- **Quarterly Review**: Full constitution review for relevance and effectiveness

### Enforcement

- Constitution violations discovered post-merge MUST be addressed within one sprint
- Repeated violations may result in reversion of changes and additional review requirements
- Intentional violations without explicit justification are grounds for rejection

### Guidance Integration

For runtime development guidance, developers should consult:
- This constitution for non-negotiable principles
- `.specify/templates/plan-template.md` for feature planning structure
- `.specify/templates/spec-template.md` for specification requirements
- `.specify/templates/tasks-template.md` for implementation task organization

**Version**: 1.0.0 | **Ratified**: 2025-11-12 | **Last Amended**: 2025-11-12
