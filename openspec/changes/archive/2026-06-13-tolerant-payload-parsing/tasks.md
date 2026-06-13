## 1. 解析契約測試

- [x] 1.1 新增 table-driven tests 驗證 `Read JSON from stdin` 對 `context_window.context_window_size`, `cost.total_duration_ms`, and rate-limit `resets_at` 的 integer-like decimal JSON numbers 會成功 parse 並保留其他欄位；以 `go test ./internal/model -run TestParsePayload_TolerantIntegerNumbers` 驗證。
- [x] 1.2 新增 table-driven tests 驗證 `Read JSON from stdin` 對 `1e6` and `1.7e9` 形式的 scientific notation JSON numbers 會成功 parse；以 `go test ./internal/model -run TestParsePayload_TolerantIntegerNumbers` 驗證。
- [x] 1.3 新增 table-driven tests 驗證 `Read JSON from stdin` 在缺欄位時維持既有 zero-value 行為，在 syntactically malformed JSON 時仍回傳 error；以 `go test ./internal/model -run 'TestParsePayload_(TolerantIntegerNumbers|MissingFields|InvalidJSON|EmptyInput)'` 驗證。
- [x] 1.4 新增或保留 normal payload regression test，驗證既有 model, context, cost, workspace, worktree, agent, and rate limit fields 不回歸；以 `go test ./internal/model -run TestParsePayload_Normal` 驗證。
- [x] 1.5 新增 table-driven tests 驗證 `Read JSON from stdin` 的 wrong-type / unconvertible 欄位只歸零該欄位且不讓整包 parse fail：`context_window.context_window_size` 或 `cost.total_duration_ms` 為 string/bool 時該欄位為 0 且 model/cost 等其他欄位保留；`rate_limits.*.resets_at` 為 string 時 `ResetsAt` 為 0、entry 仍 `Present`、`used_percentage` 照常 parse；以 `go test ./internal/model -run TestParsePayload_UnconvertibleIntegerFields` 驗證。

## 2. 容錯解析實作

- [x] 2.1 實作 `Read JSON from stdin` 的 field-level numeric tolerance：`context_window.context_window_size` and `cost.total_duration_ms` 遇到 integer-like decimal/scientific JSON numbers 時成功轉成 `int64`，遇到 unconvertible value 時只讓該欄位歸零且不讓整包 parse fail；以 `go test ./internal/model -run TestParsePayload_TolerantIntegerNumbers` 驗證。
- [x] 2.2 實作 `Parse resets_at timestamp` 的 field-level numeric tolerance：`rate_limits.five_hour.resets_at` and `rate_limits.seven_day.resets_at` 遇到 integer-like decimal/scientific JSON numbers 時成功轉成 `int64`，遇到 unconvertible value 時 `ResetsAt` 為 0 且 `Present` 保持正確；以 `go test ./internal/model -run 'TestParsePayload_(TolerantIntegerNumbers|RateLimitAbsent|ResetsAtPresent|ResetsAtAbsent)'` 驗證。
- [x] 2.3 將 rate-limit presence 偵測移入 `RateLimit` / `RateLimits` unmarshalling，使新增 top-level payload 欄位只需更新 `Payload` 一處，不再維護 `payloadJSON` 鏡像 struct；以 `rg -n "type payloadJSON|payloadJSON" internal/model/payload.go` 無結果或僅剩非鏡像 adapter 註解驗證。
- [x] 2.4 保留 malformed or empty JSON 的 hard parse error contract：`ParsePayload` 對 malformed JSON and empty stdin 回傳 error，`cmd/statusline/main.go` 維持灰色 `─ │ parse error` fallback and exit code 0；以現有 `TestParsePayload_InvalidJSON` and `TestParsePayload_EmptyInput` 驗證，必要時補 main-level test 或手動命令驗證。

## 3. 全域驗證

- [x] 3.1 執行 `gofmt -l .`，驗證沒有未格式化 Go 檔案。
- [x] 3.2 執行 `go build ./...`，驗證所有 packages 可編譯。
- [x] 3.3 執行 `go vet ./...`，驗證標準靜態檢查乾淨。
- [x] 3.4 執行 `staticcheck ./...`，驗證 staticcheck 乾淨。
- [x] 3.5 執行 `go test -race ./...`，驗證既有與新增測試在 race detector 下全通過。
