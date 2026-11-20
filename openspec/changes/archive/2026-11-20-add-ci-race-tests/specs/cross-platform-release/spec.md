## MODIFIED Requirements

### Requirement: Triggered by git tag push

The release workflow SHALL be triggered automatically when a version tag is pushed. It SHALL run the test jobs first, and SHALL build and release only after every test job succeeds.

#### Scenario: Tag format

- **WHEN** a tag matching `v*` (e.g., `v1.0.0`) is pushed to the repository
- **THEN** the GitHub Actions workflow SHALL start the test jobs
- **THEN** the workflow SHALL build and release only after every test job succeeds
