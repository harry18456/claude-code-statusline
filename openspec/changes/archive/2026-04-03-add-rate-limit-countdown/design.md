## Context

Claude Code 的 statusLine JSON payload 包含 `rate_limits.five_hour.resets_at` 與 `rate_limits.seven_day.resets_at`（Unix timestamp），目前 `model.RateLimit` struct 只解析 `used_percentage`，倒數資訊被忽略。

現有 renderer 的 `formatRate()` 只輸出百分比；要加倒數，需要在 render 時計算 `resets_at - time.Now()`。

## Goals / Non-Goals

**Goals:**

- 解析並儲存 `resets_at` 欄位
- 只要 `resets_at` 存在，一律於百分比後附加倒數（常駐顯示）
- 支援 `(Xd Yh)` / `(Xh Ym)` / `(Ym)` / `(now)` 四段倒數格式
- 不改動 render 函式的外部 signature（`Render` 仍接收 `*model.Payload`）

**Non-Goals:**

- 不加 Line 3 或其他 verbose 模式
- 不顯示 cache / token 資訊

## Decisions

### ResetsAt 直接存入 RateLimit struct

`model.RateLimit` 新增 `ResetsAt int64`（0 表示缺席）。`payloadJSON` 的 `rateLimitRaw` 同步新增對應欄位。

**替代方案：** 單獨傳 `resets_at` map 進 renderer。拒絕——讓 struct 自帶完整資料更乾淨，呼叫端不需知道細節。

### 倒數計算在 renderer 內部執行

`formatRate()` 接收 `model.RateLimit`（含 `ResetsAt`），內部呼叫 `time.Now()` 計算差值。

**替代方案：** main.go 預先計算並傳入 `time.Duration`。拒絕——增加呼叫端複雜度，且測試時可透過 `ResetsAt=0` 跳過倒數。

### 倒數格式規則

| 剩餘時間 | 顯示 |
|---------|------|
| ≥ 24 小時 | `(Xd Yh)` |
| ≥ 60 分 | `(Xh Ym)` |
| 1–59 分 | `(Ym)` |
| ≤ 0 | `(now)` |
| ResetsAt == 0 | 不顯示倒數 |

只要 `resets_at` 不為零即顯示倒數，不限用量百分比（常駐顯示）。

## Risks / Trade-offs

- [Risk] 時鐘偏差：`time.Now()` 與 API server 的時間可能有小幅差距 → 倒數僅為估計值，可接受
- [Trade-off] 只在 ≥80% 顯示倒數：低用量時不顯示，但低用量時也不需要知道重置時間
