## Context

此專案是 Claude Code `statusLine` hook 的 Go binary：stdin 讀 JSON payload，stdout 輸出兩行 ANSI statusline。Line 1 已承載 model、context bar、cost、cache hit rate、duration、rate limits；其中 cost 與 cache hit rate 都是花費因素的可視化。

Claude Code 2.1.177+ 新增 top-level execution-mode payload：

```json
{
  "effort": { "level": "max" },
  "thinking": { "enabled": true },
  "fast_mode": false
}
```

舊版 Claude Code 與不支援 effort 的模型會缺少這些欄位。現有 `context_window.current_usage` 已用 `*CurrentUsage` 表達 object availability，並用 tolerant field parsing 避免單一欄位異常拖垮整份 payload。本 change 沿用相同方向。

## Goals / Non-Goals

**Goals:**

- 解析 `effort.level`、`thinking.enabled`、`fast_mode`，並保留 absent/null 與 explicit false 的差異。
- 讓 older payload 或 unsupported model 的 absent fields 靜默省略顯示。
- 讓 malformed execution-mode fields 不造成 parse error fallback。
- 把 effort 作為主要 cost-risk signal 顯示在 Line 1。
- 在 apply 前先由產品 review 決定 final display design。

**Non-Goals:**

- 不新增 env var、CLI flag、hide list 或 config。
- 不修改 `cmd/statusline/main.go` 的 env var parsing。
- 不改 cache、cost、duration、rate-limit、context label、progress bar、Line 2。
- 不以 model name 推斷 effort。
- 不在 proposal review 前寫入 runtime implementation。

## Decisions

### Use presence-aware execution-mode fields

`effort` 與 `thinking` 是 nullable top-level objects，使用 pointer 欄位承接：

- `Payload.Effort *Effort`
- `Payload.Thinking *Thinking`

`effort = null`、`thinking = null`、欄位 absent 均解析為 unavailable。`Effort.Level` 保留原始 string；renderer 只接受 `low`、`medium`、`high`、`xhigh`、`max`，未知值不顯示。

`thinking.enabled` 與 `fast_mode` 需要辨識 explicit false，因此 boolean 使用 presence-aware wrapper，而非單純 bool：

```go
type OptionalBool struct {
    Value   bool
    Present bool
}
```

建議形狀：

```go
type Effort struct {
    Level string
}

type Thinking struct {
    Enabled OptionalBool
}

type Payload struct {
    Effort   *Effort
    Thinking *Thinking
    FastMode OptionalBool
}
```

`OptionalBool.UnmarshalJSON` 對 `true`/`false` 設定 `Present=true`；對 `null` 或 wrong scalar type 設定 `Present=false` 並回傳 nil error。這讓 `fast_mode:false` 能被辨識為 present false，也讓 malformed field 不破壞其他 payload。

### Keep effort as primary mode signal

`effort.level` 直接影響 thinking token 用量，是高訊號。default display 應以 effort 為主；`thinking.enabled` 與 `fast_mode` 是輔助訊號，預設是否顯示留給 review gate 決定。

Renderer validation rules：

- effort unavailable：不顯示 execution-mode segment。
- effort level empty：不顯示。
- effort level unknown：不顯示。
- effort known：依 final design 顯示。
- thinking/fast absent：不顯示 placeholder。
- thinking/fast malformed：不顯示 placeholder。

### Gate line 1 display design through review

Proposal 先保留兩個具體方案，不在 apply 前鎖死：

Option A：model-adjacent compact badge

```text
◆ Claude Opus 4.6 ⚙max │ ███████--- 73% 200k │ $0.85 ⚡99% │ 3m42s │ 5h:15% 7d:8%
<> Claude Opus 4.6 effort:max | #######--- 73% 200k | $0.85 cache:99% | 3m42s | 5h:15% 7d:8%
```

Option B：cost-adjacent combined mode segment

```text
◆ Claude Opus 4.6 │ ███████--- 73% 200k │ $0.85 ⚡99% │ mode:max think fast:off │ 3m42s │ 5h:15% 7d:8%
<> Claude Opus 4.6 | #######--- 73% 200k | $0.85 cache:99% | mode:max think fast:off | 3m42s | 5h:15% 7d:8%
```

Review 必須決定：

- insertion point：model adjacent 或 cost/cache adjacent。
- default/nerdfont text：icon badge 或 explicit text。
- ASCII fallback text。
- effort color scale。
- `thinking.enabled` 與 `fast_mode` 的 default visibility。

Apply 階段第一個 task 記錄此決策，再進入測試與 implementation。

### Keep renderer formatting isolated

新增 `formatExecutionMode` 或等價 helper，集中處理：

- effort validation。
- final text/glyph format。
- low-to-max color mapping。
- ASCII no ANSI/no Unicode contract。
- Nerd Font/default output。
- thinking/fast visibility rule。
- absent/null/invalid suppression。

`Render` 只負責把 helper 回傳的 non-empty segment 插入 Line 1 的 approved position。這避免 parsing、formatting、placement、color policy 混在主 render flow。

### Preserve existing config surface

本 change 不改 `Options` 的來源，不新增 env var，不新增 CLI flag。Renderer 只沿用現有 `ASCIIMode`、`NerdFont`、`Powerline`、`TrueColor`。

## Implementation Contract

Observable behavior:

- Payload 有 known `effort.level` 時，Line 1 顯示一段 execution-mode segment。
- Payload 缺少 execution-mode fields、fields 為 null、或 effort level 不可用時，Line 1 不新增該段，其餘段落維持原狀。
- `thinking.enabled` 與 `fast_mode` 的顯示由 review-approved design 決定；無論 final design 是否顯示，parser 都必須保留 present false 與 absent 的差異。
- ASCII mode 的 execution-mode segment 不含 ANSI escape 與 Unicode glyph。
- Default/Nerd Font mode 的 effort level 依 approved low-to-max intensity scale 上色。

Data shape:

```json
{
  "effort": { "level": "max" },
  "thinking": { "enabled": true },
  "fast_mode": false
}
```

Failure modes:

- Top-level malformed JSON 仍維持既有 parse-error fallback。
- `effort`, `thinking`, `fast_mode` absent：正常且靜默。
- `effort`, `thinking`, `fast_mode` null：正常且靜默。
- execution-mode field wrong type：不顯示該 mode data，不影響其他 payload fields。
- unknown effort level：不顯示 execution-mode segment。

Acceptance criteria:

- Model tests 覆蓋 all present、absent、null、malformed、known effort levels、explicit `fast_mode:false`。
- Renderer helper tests 覆蓋 all effort levels、unknown effort、absent effort、thinking true/false、fast true/false、ASCII/default/NerdFont output。
- Render integration tests 覆蓋 approved Line 1 insertion point、absent fields 不顯示、既有 cost/cache/rate-limit 順序不被破壞。
- Verification commands pass：`go build ./...`、`go vet ./...`、`staticcheck ./...`、`go test -race ./...`、`gofmt -l .`。

Scope boundaries:

- In scope：payload model extension、tolerant parsing、renderer helper、Line 1 optional display、focused tests。
- Out of scope：main env parsing、new config flags、README rewrite、cache/cost/rate-limit changes、gitcache changes。

## Risks / Trade-offs

- [Risk] Line 1 width increases. Mitigation: gate final display through review and prefer compact effort-only default unless product selects explicit mode text.
- [Risk] `thinking.enabled` and `fast_mode` create visual noise. Mitigation: parse them now, but make their default display a review decision.
- [Risk] Unknown future effort levels could mislead users if rendered blindly. Mitigation: whitelist known levels and suppress unknown values.
- [Risk] `fast_mode:false` can be lost if represented as plain bool. Mitigation: use presence-aware boolean wrapper.

## Open Questions

- Review decision: Option A or Option B.
- Review decision: effort color scale exact ANSI mapping.
- Review decision: thinking/fast default visibility.
