## 1. Review Gate

- [x] 1.1 完成 `Gate line 1 display design through review`：已拍板 Option A，execution-mode segment 緊貼 model 名後；default/Nerd Font format 為 `⚙<level>`，`thinking.enabled=true` 後綴 `T`，`fast_mode=true` 後綴 `F`，兩者都 on 為 `TF`，off/absent 不顯示；ASCII fallback 為 `effort:<level>`，thinking/fast on 時追加 `think`/`fast`；effort color scale 為 low gray、medium cyan、high yellow、xhigh red、max bold red，T/F 後綴 gray；驗證方式為本次 review 決策已寫入 tasks 且 apply 階段依此實作。

## 2. Model Parsing

- [x] [P] 2.1 為 `Parse execution mode fields` 加上 model tests：涵蓋 all present、absent、null、malformed execution-mode fields、known effort levels、explicit `fast_mode:false`；驗證 `go test ./internal/model -run TestParsePayload_ExecutionMode` 已先失敗於實作前且通過於實作後。
- [x] 2.2 實作 `Use presence-aware execution-mode fields`：parsed payload 能辨識 `effort.level`、`thinking.enabled`、`fast_mode` availability，並保留 explicit false 與 absent 的差異；驗證 `go test ./internal/model -run TestParsePayload_ExecutionMode` 通過。

## 3. Renderer Behavior

- [x] [P] 3.1 為 `Execution mode display` 與 `Keep effort as primary mode signal` 加上 renderer tests：涵蓋 effort 各等級、unknown effort omission、absent omission、thinking true/false、fast true/false、ASCII/default/NerdFont output；驗證 `go test ./internal/renderer -run TestFormatExecutionMode` 已先失敗於實作前且通過於實作後。
- [x] 3.2 實作 `Keep renderer formatting isolated`：新增 `formatEffort`、`formatExecutionMode` 或等價 helper，讓 approved display format、effort validation、color scale、ASCII no ANSI/no Unicode contract 集中且可單測；驗證 `go test ./internal/renderer -run TestFormatExecutionMode` 通過。
- [x] 3.3 實作 `Render two-line ANSI output` insertion：Line 1 在 approved position 顯示 execution-mode segment，absent 或 invalid data 不顯示，既有 cost/cache/duration/rate-limit 順序與 Line 2 維持原狀；驗證 `go test ./internal/renderer -run TestRenderExecutionMode` 通過。

## 4. Scope Preservation and Verification

- [x] 4.1 落實 `Preserve existing config surface`：不修改 `cmd/statusline/main.go` env var parsing，不新增 `CLAUDE_STATUSLINE_*`、CLI flag、hide/config 機制；驗證 `git diff -- cmd/statusline/main.go` 無輸出且 renderer tests 覆蓋現有 options 行為。
- [x] 4.2 完成 final verification：真實 binary 與測試矩陣全綠；驗證 `go build ./...`、`go vet ./...`、`staticcheck ./...`、`go test -race ./...`、`gofmt -l .` 均通過且 `gofmt -l .` 無輸出。
