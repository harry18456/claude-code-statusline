## MODIFIED Requirements

### Requirement: Context window size label

The program SHALL display context window size only when not already implied by the model name. When `context_window_size >= 1,000,000` and the payload indicates usage has crossed the 200k token pricing threshold, the `1M` label SHALL be colored red to warn of elevated per-token pricing.

#### Scenario: 1M context window within base pricing tier

- **WHEN** context_window_size >= 1,000,000 AND `exceeds_200k_tokens` is false AND model name does not contain "context" or "Context"
- **THEN** the program SHALL display ` 1M` in gray after the percentage

#### Scenario: 1M context window in premium pricing tier

- **WHEN** context_window_size >= 1,000,000 AND `exceeds_200k_tokens` is true AND model name does not contain "context" or "Context"
- **THEN** the program SHALL display ` 1M` in red after the percentage

#### Scenario: 200k context window

- **WHEN** context_window_size >= 200,000 AND context_window_size < 1,000,000 AND model name does not contain "context" or "Context"
- **THEN** the program SHALL display ` 200k` in gray after the percentage

## ADDED Requirements

### Requirement: Directory path display on line 2

The program SHALL display the working directory on line 2 using a project root resolved by a three-step fallback: (1) use `workspace.project_dir` when it is a strict ancestor of `workspace.current_dir`; (2) otherwise walk upward from `workspace.current_dir` looking for a `.git` entry (file or directory) and use the first matching directory as root; (3) otherwise fall back to `filepath.Base(workspace.current_dir)`. The `.git` lookup SHALL use filesystem stat only and SHALL NOT invoke the `git` CLI.

When a project root is resolved, the program SHALL display `filepath.Base(root)` in blue when `workspace.current_dir` equals the root, and `filepath.Base(root) + "/" + relative_path` (forward slashes) when `workspace.current_dir` is a strict descendant of the root.

#### Scenario: Payload project_dir is a strict ancestor of current directory

- **WHEN** `workspace.project_dir` is non-empty AND `workspace.current_dir` is a strict descendant of `workspace.project_dir`
- **THEN** line 2 SHALL display `<project_base>/<relative_path>` in blue, where `project_base = filepath.Base(workspace.project_dir)` and `relative_path` uses forward slashes

#### Scenario: Payload project_dir equals current directory but inside a git repository

- **WHEN** `workspace.project_dir` equals `workspace.current_dir` (or is empty) AND an ancestor directory of `workspace.current_dir` contains a `.git` entry (file or directory)
- **THEN** line 2 SHALL display `<git_root_base>` in blue when `workspace.current_dir` equals the resolved git root, or `<git_root_base>/<relative_path>` in blue when it is a strict descendant, where `git_root_base = filepath.Base(<resolved git root>)` and `relative_path` uses forward slashes

#### Scenario: Payload project_dir unusable and no git repository detected

- **WHEN** `workspace.project_dir` cannot be used as an ancestor (empty, equals current, or not an ancestor) AND no `.git` entry is found walking upward from `workspace.current_dir` to the filesystem root
- **THEN** line 2 SHALL fall back to `filepath.Base(workspace.current_dir)` in blue

#### Scenario: Current directory is empty or "."

- **WHEN** `workspace.current_dir` is an empty string or `.`
- **THEN** line 2 SHALL display `.` in blue

#### Scenario: Git worktree with `.git` as a file

- **WHEN** an ancestor of `workspace.current_dir` contains `.git` as a regular file (git worktree layout) rather than a directory AND no usable `workspace.project_dir` is available
- **THEN** the program SHALL treat that ancestor as the project root for display purposes

#### Scenario: Git submodule treated as its own root

- **WHEN** `workspace.current_dir` is located inside a git submodule (a `.git` entry — typically a file with `gitdir:` metadata — exists at the submodule directory AND an outer `.git` directory also exists further up in the parent repository) AND no usable `workspace.project_dir` is available
- **THEN** the program SHALL resolve the submodule directory as the project root (first `.git` wins) and SHALL NOT continue walking upward to the parent repository

### Requirement: Parse exceeds_200k_tokens field

The program SHALL parse the top-level `exceeds_200k_tokens` boolean field from the JSON payload and store it on the parsed `Payload` struct for use by the renderer.

#### Scenario: Field present and true

- **WHEN** the JSON payload contains `"exceeds_200k_tokens": true`
- **THEN** the parsed payload SHALL expose the value `true` for this field

#### Scenario: Field present and false

- **WHEN** the JSON payload contains `"exceeds_200k_tokens": false`
- **THEN** the parsed payload SHALL expose the value `false` for this field

#### Scenario: Field absent

- **WHEN** the JSON payload omits the `exceeds_200k_tokens` field
- **THEN** the parsed payload SHALL expose the value `false` for this field (zero-value default)
