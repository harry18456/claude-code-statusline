## MODIFIED Requirements

### Requirement: Context window size label

The program SHALL display a context window size label based solely on `context_window_size`. When `context_window_size >= 1,000,000` and the payload indicates usage has crossed the 200k token pricing threshold, the `1M` label SHALL be colored red to warn of elevated per-token pricing. The model name SHALL NOT affect whether the label is displayed; the program MUST NOT suppress the label based on substrings of the model name.

#### Scenario: 1M context window within base pricing tier

- **WHEN** context_window_size >= 1,000,000 AND `exceeds_200k_tokens` is false
- **THEN** the program SHALL display ` 1M` in gray after the percentage

#### Scenario: 1M context window in premium pricing tier

- **WHEN** context_window_size >= 1,000,000 AND `exceeds_200k_tokens` is true
- **THEN** the program SHALL display ` 1M` in red after the percentage

#### Scenario: 200k context window

- **WHEN** context_window_size >= 200,000 AND context_window_size < 1,000,000
- **THEN** the program SHALL display ` 200k` in gray after the percentage

#### Scenario: Context window size below 200k

- **WHEN** context_window_size < 200,000 (including zero when the field is absent)
- **THEN** the program SHALL NOT display any context window size label

#### Scenario: Model name containing the substring "context" does not suppress the label

- **WHEN** context_window_size >= 1,000,000 AND the model `display_name` contains the substring "context" (case-insensitive), for example `Opus 4.7 (1M context)`
- **THEN** the program SHALL still display ` 1M` after the percentage, even though the label text duplicates information already present in the model name
