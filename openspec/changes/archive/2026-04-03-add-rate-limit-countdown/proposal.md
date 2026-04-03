## Why

Rate limit 顯示目前只有使用百分比（如 `5h:79%`），無法判斷「還要等多久才能重置」。JSON payload 已提供 `resets_at` Unix timestamp，加上倒數計時可讓使用者立即知道是否要等待或繼續使用。

## What Changes

- Rate limit 顯示從 `5h:79%` 改為 `5h:79% (1h23m)`，附上距離重置的剩餘時間
- 同時套用於 `five_hour` 與 `seven_day` 兩個 rate limit 欄位
- **常駐顯示**：只要 `resets_at` 存在即顯示倒數，不限用量百分比
- 倒數格式：≥24h 顯示 `(Xd Yh)`；≥60min 顯示 `(Xh Ym)`；<60min 顯示 `(Ym)`；已過期顯示 `(now)`
- `model.RateLimit` struct 新增 `ResetsAt int64` 欄位
- `formatRate` renderer function 接收 `ResetsAt` 並計算剩餘時間

## Non-Goals

- 不加 Line 3（verbose token/cache 資訊）
- 不顯示 cache hit rate
- 不顯示 token 計數（`total_input_tokens` / `total_output_tokens`）

## Capabilities

### New Capabilities

- `rate-limit-countdown`: 解析 `resets_at` timestamp，計算剩餘時間並附加於 rate limit 顯示

### Modified Capabilities

- `go-statusline`: `RateLimit` struct 及 renderer 的 rate limit 顯示邏輯有行為變更

## Impact

- Affected specs: `rate-limit-countdown`（新增）、`go-statusline`（修改）
- Affected code: `internal/model/payload.go`、`internal/renderer/renderer.go`、`internal/model/payload_test.go`、`internal/renderer/renderer_test.go`
