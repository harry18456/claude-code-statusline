## Context

此 binary 是 Claude Code `statusLine` hook：Claude Code 透過 stdin 傳 JSON payload，程式輸出兩行 ANSI statusline。現有 config 在 `cmd/statusline/main.go` 由專案自訂環境變數轉成 `renderer.Options`：

- `CLAUDE_STATUSLINE_ASCII=1`
- `CLAUDE_STATUSLINE_NERDFONT=1`
- `CLAUDE_STATUSLINE_POWERLINE=1`
- `COLORTERM=truecolor|24bit`

前三者是本專案 config；`COLORTERM` 是終端標準能力訊號。Claude Code 的 statusLine 本來就在 `settings.json` 以 command 字串設定，因此專案 config 放在 command flags 中比放在環境變數中更可見。

Renderer 目前靠資料 zero-value 自動省略部分段落；使用者無法主動隱藏有資料但暫時不需要的段落。

Claude Code 官方 statusLine 文件說明：statusLine 顯示 script 印到 stdout 的內容；troubleshooting 明確指出 script non-zero exit 或無輸出會讓 status line blank，stderr 需透過 `claude --debug` 才能查。這使 CLI config error 不能用 non-zero exit 或 stdout 空白處理，否則一個 typo 會讓狀態列直接消失且一般使用者看不到原因。參考：https://code.claude.com/docs/en/statusline

## Goals / Non-Goals

**Goals:**

- 用 command flags 取代專案自訂 env var：`--ascii`、`--nerdfont`、`--powerline`。
- 直接廢棄 `CLAUDE_STATUSLINE_ASCII`、`CLAUDE_STATUSLINE_NERDFONT`、`CLAUDE_STATUSLINE_POWERLINE`。
- 保留 `COLORTERM` true color 偵測。
- 新增 `--hide <keys>`，讓使用者主動隱藏任意 statusline 段落。
- CLI config 問題採容錯渲染：stdout 仍輸出有效 statusline，程式 exit 0，stderr 只作 debug warning。
- 讓 `--version` 透過同一套 flag parser 運作，並在讀 stdin 前返回版本。
- 更新英文與繁中 README，明確標示 breaking upgrade path。

**Non-Goals:**

- 不改各段落的顯示內容、格式、顏色、計算方式。
- 不改 model parsing、payload model、gitcache、rate-limit countdown 與 pace 演算法。
- 不新增 config file、YAML、JSON config 或新的 env var。
- 不保留舊 `CLAUDE_STATUSLINE_*` env var 相容層。
- 不新增段落排序、自訂顏色、自訂文字格式或顯示模板。

## Decisions

### Use the Go standard flag package

採用 `flag.NewFlagSet` 搭配 `flag.ContinueOnError`，不使用 package-level `flag.CommandLine`。解析會在讀 stdin 前執行，產出一個小型 config 結構，再轉成 `renderer.Options`。

理由：

- `flag` 是標準函式庫，無新增 dependency。
- bool flags 與 string/list flag 的語義清楚，支援 `--hide effort,duration` 與 `--hide=effort,duration`。
- `ContinueOnError` 讓 main 可以捕捉錯誤並轉成容錯渲染，而不是讓標準 `flag` 自行 `os.Exit`。
- `--version` 可以是同一個 `FlagSet` 裡的 bool flag，避免手動 `os.Args[1] == "--version"` 和新 flags 分岔。

替代方案：手動 parse `os.Args`。不採用，因為 unknown flag、missing value、bool value、`--flag=value`、usage 錯誤都要自行維護，且很容易讓 `--version` 形成第二套語義。

### Tolerant CLI config never blanks the statusline

CLI config 問題一律容錯渲染。程式可以把警告寫到 stderr 供 `claude --debug` 使用，但 stdout 必須輸出有效 statusline，exit code 必須是 0。

理由：

- Claude Code statusLine 只顯示 stdout；script non-zero 或無輸出會讓 status line blank。
- statusLine stderr 不會在一般 UI 中顯示，依賴 stderr 命名 invalid key 無法幫使用者定位 typo。
- 本專案核心原則是 statusline 不因局部問題消失。CLI config typo 屬於局部問題，不能比 payload parse error 更致命。

容錯規則：

- Unknown `--hide` key：忽略該 key；同一 `--hide` 中已知 key 照常生效；stderr warning 命名 invalid key。
- `--ascii` 與 `--nerdfont` 或 `--powerline` 衝突：ASCII 優先，因為純 ASCII 是最安全 fallback；stderr warning 說明衝突與採用結果。
- Unknown flag、`--hide` 缺值、非空 positional arg：視為 flag 語法錯誤；main 捕捉 parse error 後仍讀 stdin 並渲染。實作應盡量保留已成功解析的 flags；若 `flag` 套件無法可靠提供部分解析結果，退回 default renderer options 並輸出 stderr warning。
- `--version` 不是 config error；仍印版本、exit 0、不讀 stdin、不渲染。

### Preserve COLORTERM as system capability detection

`COLORTERM=truecolor|24bit` 繼續設定 `renderer.Options.TrueColor`。它不屬於本專案 config，不列入 breaking removal。

### Preserve current powerline implication

`--nerdfont` 會啟用 Nerd Font 模式，且維持既有語義：Powerline follows NerdFont。`--powerline` 可在不啟用 Nerd Font glyph set 的情況下只切換 separator。

`--ascii` 是純 ASCII 渲染模式。若同時給 `--ascii` 與 `--nerdfont` 或 `--powerline`，採 ASCII 優先並警告；不讓狀態列消失。

### Define canonical hide section keys

`--hide` 接受逗號分隔的 canonical lowercase keys。解析時 trim 空白、忽略空 token、去重；任何非空未知 key 都被忽略並寫 stderr warning。多次提供 `--hide` 時合併 key 集合。

Canonical keys：

| Key | Hidden section |
| --- | --- |
| `model` | Line 1 brand/model segment, including brand symbol and `model.display_name` |
| `effort` | Line 1 execution-mode segment |
| `bar` | Line 1 context progress bar, percentage, and high-usage warning symbol |
| `size` | Line 1 context size label, such as `200k` or `1M` |
| `cost` | Line 1 total cost display |
| `cache` | Line 1 prompt cache hit-rate display |
| `duration` | Line 1 duration display |
| `rate` | Line 1 five-hour and seven-day rate-limit display, including countdown and pace indicator |
| `branch` | Line 2 branch display and dirty marker |
| `lines` | Line 2 `+added/-removed` display |
| `dir` | Line 2 directory display |
| `agent` | Line 2 agent or worktree indicator |

No aliases are defined. Examples such as `rate_limits`, `directory`, or `worktree` are ignored with warnings rather than treated as accepted aliases.

### Apply hide at segment assembly boundaries

Hide logic belongs at renderer segment assembly boundaries, not inside low-level formatters. Existing helpers such as cache hit, execution mode, duration, rate limit, directory display, and bar rendering keep their current behavior; `Render` only decides whether a finished segment participates in the output.

Segment assembly rules:

- Default hidden set is empty, so all currently available segments render as before.
- Existing zero-value/unavailable suppression still applies before or alongside hide checks.
- Hidden or empty segments do not emit separators.
- If every segment on a line is hidden or unavailable, that line is an empty string; the program still emits the two-line protocol.
- Cost and cache remain adjacent when both are visible. If `cost` is hidden but `cache` is visible and available, cache renders as a standalone segment in the cost/cache position. If `cache` is hidden but `cost` is visible, cost renders exactly as the current cost-only segment.
- Hiding `bar` also hides the percentage and high-context warning symbol; `size` is separate.
- Hiding `agent` suppresses both subagent and worktree indicator because the current output has one agent/worktree segment.

### Represent hidden sections as renderer options

`renderer.Options` gains a hidden-section collection. Apply can use a typed key set such as `map[SectionKey]bool` with constants, or an immutable `map[string]struct{}` built by main and read by renderer. The durable contract is that renderer receives a canonical hidden set and never parses raw `--hide` strings itself.

## Implementation Contract

Observable CLI behavior:

- `statusline --version` prints the injected version to stdout, exits 0, and does not read stdin.
- `statusline --ascii` enables ASCII mode.
- `statusline --nerdfont` enables Nerd Font mode and Powerline separators.
- `statusline --powerline` enables Powerline separators without changing the symbol set to Nerd Font.
- `statusline --hide effort,duration,rate` suppresses those segments when rendering.
- `statusline --hide effort,bogus,rate` suppresses `effort` and `rate`, ignores `bogus`, writes a warning to stderr, renders stdout normally, and exits 0.
- `statusline --ascii --nerdfont` renders in ASCII mode, writes a warning to stderr, and exits 0.
- `statusline --unknown-flag` or `statusline --hide` still reads stdin, renders with default or successfully parsed safe options, writes a warning to stderr, and exits 0.
- `statusline` with no config flags renders all available segments by default.
- `CLAUDE_STATUSLINE_ASCII=1`, `CLAUDE_STATUSLINE_NERDFONT=1`, and `CLAUDE_STATUSLINE_POWERLINE=1` have no effect.
- `COLORTERM=truecolor` or `COLORTERM=24bit` still enables true color progress bar rendering.

Failure modes:

- Unknown hide key: ignored; known hide keys still apply; stderr warning names the invalid key; stdout renders; exit 0.
- Display mode conflict: ASCII wins; stderr warning names the conflict; stdout renders; exit 0.
- Unknown flag, missing `--hide` value, or positional arg: parser warning to stderr; stdout renders with safe options; exit 0.
- Invalid JSON on stdin remains the existing parse-error fallback and exit 0.

Acceptance criteria:

- Main-level parsing tests cover all flags, `--version`, default options, `COLORTERM`, removed env vars, invalid flags, missing `--hide`, unknown hide key, positional args, and ASCII/NerdFont-Powerline conflicts.
- Main-level execution tests cover malformed config still producing stdout statusline and exit 0.
- Renderer tests cover each hide key against a payload that would otherwise render the segment.
- Renderer tests cover no orphan separators when adjacent segments are hidden.
- Renderer tests cover `cost`/`cache` combinations: both visible, cost hidden only, cache hidden only, both hidden.
- README.md and README.zh-TW.md include breaking migration from env vars to command flags and document tolerant config warnings.
- Real binary checks cover `--nerdfont`, `--hide effort`, unknown hide key tolerance, display-mode conflict tolerance, flag syntax tolerance, and no-flag default rendering.
- Verification passes: `go build ./...`, `go vet ./...`, `staticcheck ./...`, `go test -race ./...`, `gofmt -l .`.

Scope boundaries:

- In scope: CLI parser, `renderer.Options`, segment hide gating, focused tests, README updates.
- Out of scope: payload shape, formatter algorithms, rate-limit math, cache hit-rate math, execution-mode display format, gitcache, release workflow, install path.

## Risks / Trade-offs

- [Risk] Breaking env var removal can surprise existing users. Mitigation: README migration section and explicit examples mapping each env var to a flag.
- [Risk] Ignoring unknown hide keys can hide typos from normal UI. Mitigation: stderr warning is available in `claude --debug`, and preserving a visible statusline is higher priority than strict rejection.
- [Risk] Flag syntax errors can make later flags unavailable. Mitigation: keep successfully parsed values when reliable; otherwise fall back to default Options and render.
- [Risk] ASCII conflict precedence discards requested Nerd Font or Powerline visuals. Mitigation: ASCII is the safest fallback and the warning records the conflict.

## Migration Plan

- Replace README env var configuration with command examples:
  - `statusline.exe --ascii`
  - `statusline.exe --nerdfont`
  - `statusline.exe --powerline`
  - `statusline.exe --nerdfont --hide effort,duration,rate`
- Mark `CLAUDE_STATUSLINE_ASCII`, `CLAUDE_STATUSLINE_NERDFONT`, and `CLAUDE_STATUSLINE_POWERLINE` as removed in the breaking change note.
- Document that invalid config is tolerated: output remains visible, warnings go to stderr/debug logs.
- Leave `COLORTERM` documented only as terminal capability detection.

Rollback path during apply: revert the change before release if tests or real binary checks fail; no data migration exists.

## Open Questions

None. This revision fixes the parser choice, hide key set, tolerant config behavior, and `--version` coexistence contract for review.
