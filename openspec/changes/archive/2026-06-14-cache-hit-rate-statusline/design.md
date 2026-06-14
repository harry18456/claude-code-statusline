## Context

目前 Go binary 透過 `model.ParsePayload` 解析 Claude Code statusLine JSON，再由 `renderer.Render` 輸出兩行 ANSI。`ContextWindow` 已有自訂 `UnmarshalJSON`，並使用 `tolerantInt64` 讓整數欄位遇到 decimal、scientific notation 或錯誤 scalar 時不讓整包 payload 失敗。

`context_window.current_usage` 目前完全未解析。實際 payload 會帶最近一次 API request 的 token breakdown，且在 session 開始前與 `/compact` 後會是 `null`。這不是 session 累計值，renderer 不得把它當長期統計。

## Goals / Non-Goals

**Goals:**

- 解析 `context_window.current_usage`，保留「有值」與「不可用」狀態。
- 用 input-side token breakdown 計算最近一次 request 的 prompt cache hit rate。
- 在資料不可用或分母為零時完全不顯示該段。
- 保持 payload parsing 的 field-level tolerance。
- 讓 cache hit rate 顯示設計可在 review 後拍板，再進入 apply。

**Non-Goals:**

- 不計算 session 累計 cache hit rate。
- 不納入 `output_tokens`。
- 不估算 dollar savings。
- 不改動 rate limit、progress bar、context label、line 2 的既有語意。
- 不在本 proposal 階段修改 Go 實作碼。

## Decisions

### Parse current_usage as nullable data

在 `ContextWindow` 下加入 `CurrentUsage *CurrentUsage` 或等價 presence 結構。`null` 與 absent 都代表 unavailable；object 代表 available。這能區分「無最近 request 用量」與「最近 request token 全為 0」。

`CurrentUsage` token 欄位使用既有 `tolerantInt64`：

- `InputTokens`
- `CacheCreationInputTokens`
- `CacheReadInputTokens`

可選擇保留 `OutputTokens` 作為解析欄位，但 cache hit rate 計算不得使用它。若為了最小 surface，也可不暴露 `OutputTokens`；parser 仍會忽略未知 JSON 欄位。

### Keep cache hit rate scoped to latest request

計算公式固定為：

```text
denominator = input_tokens + cache_creation_input_tokens + cache_read_input_tokens
numerator   = cache_read_input_tokens
hit_rate    = numerator / denominator
```

`current_usage` 只代表最近一次 API request，所以 renderer 顯示的是瞬時訊號。名稱、測試與文件要避免使用 `session`、`total`、`cumulative` 這類字眼描述此指標。

### Suppress unusable cache data

`formatCacheHit` 或等價 helper 接收 nullable current usage 與 render options。以下情境直接回空字串：

- `current_usage` 為 `nil` 或 unavailable。
- denominator 等於 `0`。
- token 欄位因 tolerant parsing 失敗而歸零後造成 denominator 等於 `0`。

renderer 不輸出佔位符，不顯示 `0%`，也不讓除以零發生。

### Keep renderer formatting isolated

新增 `formatCacheHit`，避免把計算、顏色與符號分散在 `Render` 主流程內。`Render` 只負責把回傳的非空 segment 插入 line 1。

測試層次：

- model tests：解析 object、`null`、absent、decimal/scientific token、wrong scalar token。
- renderer helper tests：公式、rounding、denominator zero、nil current usage、ASCII/default/NerdFont segment。
- render integration tests：line 1 在有資料時含 cache segment，在無資料時不含 cache segment。

### Gate final display through review

proposal 內提供兩個可拍板方案：

- Option A：接在 rate limits 後。
- Option B：靠近 cost。

兩者共用預設/Nerd Font `⚡<pct>%` 與 ASCII `cache:<pct>%`。顏色門檻建議為 `>=80%` gray、`50%..79%` yellow、`<50%` red；若產品負責人要降低視覺噪音，可改成只有 `<50%` red，其餘 gray。

apply 階段只採用 review 後的單一決策，不同時實作多種 layout。

## Implementation Contract

Observable behavior:

- 當最新 request 的 `current_usage` 可用且 denominator 大於 0，line 1 顯示一段 prompt cache hit rate。
- 當 `current_usage` 為 `null`、absent 或 denominator 為 0，line 1 不顯示 cache hit rate，其他段落照常輸出。
- ASCII mode 使用 `cache:<pct>%`，且該 segment 不含 ANSI color escape 與 Unicode glyph。
- Default/Nerd Font mode 使用 review 拍板的非 ASCII indicator，例如 `⚡99%`。

Data shape:

```json
"context_window": {
  "context_window_size": 200000,
  "used_percentage": 73,
  "current_usage": {
    "input_tokens": 1,
    "output_tokens": 374,
    "cache_creation_input_tokens": 1302,
    "cache_read_input_tokens": 144198
  }
}
```

Failure modes:

- Malformed top-level JSON remains a hard parse error and preserves the existing parse-error fallback.
- Field-level numeric conversion failures in current usage token fields are recoverable and zero only that field.
- `current_usage: null` is normal and silent.
- denominator zero is normal and silent.

Acceptance criteria:

- `internal/model/payload_test.go` covers current usage object, `null`, absent, tolerant numeric forms, and wrong scalar values.
- `internal/renderer/renderer_test.go` covers correct hit-rate calculation, omission for unavailable data, omission for denominator zero, ASCII output, and default/NerdFont output.
- Verification commands pass in apply phase: `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test -race ./...`, `gofmt -l .`.

Scope boundaries:

- In scope: payload model extension, renderer helper, line 1 segment insertion, focused tests.
- Out of scope: cost math, cumulative aggregation, persistent history, docs beyond any user-facing README note explicitly requested during apply.

## Risks / Trade-offs

- [Risk] Users may read the metric as session-wide efficiency. → Mitigation: name tests and comments around latest-request/current usage, and avoid cumulative wording.
- [Risk] Additional line 1 segment increases width. → Mitigation: choose one compact placement during review and suppress the segment when unavailable.
- [Risk] Low cache hit color could compete with rate-limit warnings. → Mitigation: use red only below the selected low threshold; high hit rates stay gray.
- [Risk] `current_usage` null after `/compact` may look like missing feature. → Mitigation: omit silently rather than showing a misleading `0%`.

## Open Questions

- Display placement: Option A after rate limits, or Option B near cost.
- Non-ASCII indicator: use `⚡` as proposed, or select a Nerd Font-specific cache glyph.
- Color threshold: use three bands (`>=80` gray, `50..79` yellow, `<50` red), or only red below `50%`.
