# ◆ claude-code-statusline

為 [Claude Code](https://docs.anthropic.com/en/docs/claude-code) 打造的即時狀態列。每次回應後顯示模型、上下文用量、花費、時間、Git 分支與速率限制。

[English](README.md) | [繁體中文](README.zh-TW.md)

---

## 畫面說明

```
◆ Sonnet 4.6 │ ████████░░ 78% │ $1.23 │ 14m32s │ 5h:42% 7d:15%
⎇ main* │ +84/-12 │ my-project │ ⚙ code-reviewer
```

### 第一行

| 區段 | 範例 | 說明 |
|------|------|------|
| `◆` | `◆` | Anthropic 品牌菱形（紫色）。ASCII 模式顯示 `<>` |
| 模型 | `Sonnet 4.6` | 目前使用的 Claude 模型名稱 |
| 進度條 | `████████░░` | 10 格上下文視窗使用量 |
| 百分比 | `78%` | 上下文使用率。綠色 < 70%，黃色 70–89%，紅色 ≥ 90% |
| ⚠ 警告 | `⚠` | 上下文 ≥ 90% 時才出現 |
| 視窗大小 | `200k` / `1M` | 僅在模型名稱未包含此資訊時顯示 |
| 花費 | `$1.23` | 本次 session 累積 token 費用（估算值）。黃色 > $0，紅色 ≥ $10，$0.00 時顯示為灰色 |
| 時間 | `14m32s` | Session 總時長。不足 1 秒時隱藏 |
| 速率限制 | `5h:42%` `7d:15%` | 5 小時與 7 天配額使用率（僅 Claude Pro/Max）。≥ 80% 時轉紅 |

### 第二行

| 區段 | 範例 | 說明 |
|------|------|------|
| 分支 | `⎇ main*` | 目前 Git 分支。`*` 表示有未提交的變更 |
| 行數 | `+84/-12` | 本次 session 中 Claude 新增／刪除的行數 |
| 目錄 | `my-project` | 目前工作目錄名稱 |
| 指示器 | `⚙ code-reviewer` | 執行中的子 agent 名稱，或 worktree 時顯示 `⚙ worktree:name`。Worktree 優先於 agent |

數值為零的區段會完全隱藏（`+0/-0`、`0m0s`、未提供的速率限制）。

---

## 安裝

### 第一步 — 下載執行檔

前往 [Releases](https://github.com/harry18456/claude-code-statusline/releases/latest) 下載對應平台的檔案：

| 平台 | 檔案名稱 |
|------|---------|
| macOS（Apple Silicon） | `statusline-darwin-arm64` |
| macOS（Intel） | `statusline-darwin-amd64` |
| Linux（x86_64） | `statusline-linux-amd64` |
| Windows（x86_64） | `statusline-windows-amd64.exe` |

### 第二步 — 放置執行檔

**macOS / Linux**

```bash
# 將檔名替換為你下載的版本
mv statusline-darwin-arm64 ~/.claude/statusline
chmod +x ~/.claude/statusline
```

**Windows**（PowerShell）

```powershell
Move-Item statusline-windows-amd64.exe "$env:USERPROFILE\.claude\statusline.exe"
```

### 第三步 — 設定 Claude Code

編輯 `~/.claude/settings.json`（若不存在請自行建立）。

**macOS / Linux**

```json
{
  "statusLine": {
    "type": "command",
    "command": "/Users/你的使用者名稱/.claude/statusline"
  }
}
```

**Windows**

```json
{
  "statusLine": {
    "type": "command",
    "command": "C:/Users/你的使用者名稱/.claude/statusline.exe"
  }
}
```

將 `你的使用者名稱` 替換為實際的使用者名稱。Windows 上也請使用正斜線 `/`。

若 `settings.json` 已有內容，在既有物件內新增 `"statusLine"` 即可：

```json
{
  "someOtherSetting": true,
  "statusLine": {
    "type": "command",
    "command": "/Users/你的使用者名稱/.claude/statusline"
  }
}
```

### 第四步 — 驗證

重新啟動 Claude Code。第一次回應後，狀態列應出現在終端機底部。

---

## 環境變數

在 shell 設定檔（`~/.zshrc`、`~/.bashrc` 等）或 Claude Code 的 `env` 設定中加入。

| 變數 | 預設值 | 效果 |
|------|--------|------|
| `CLAUDE_STATUSLINE_ASCII` | `0` | 設為 `1` 啟用純 ASCII 輸出（`#---`）。適用於 Unicode 不可用的環境 |
| `CLAUDE_STATUSLINE_NERDFONT` | `0` | 設為 `1` 啟用 Nerd Font 圖示（, 󰔟, ）。需要終端機已安裝 [Nerd Font](https://www.nerdfonts.com/) |
| `CLAUDE_STATUSLINE_POWERLINE` | 跟隨 `NERDFONT` | 設為 `1` 使用 Powerline 箭頭分隔符（``）取代 `│`。`NERDFONT=1` 時自動啟用 |
| `COLORTERM` | 系統設定 | 設為 `truecolor` 或 `24bit` 啟用 RGB 漸層進度條。大多數現代終端機會自動設定 |

### 渲染層級

執行檔根據環境自動選擇最佳渲染方式：

| 層級 | 條件 | 進度條樣式 |
|------|------|-----------|
| True color | `COLORTERM=truecolor` 或 `24bit` | 每格獨立 RGB 漸層，綠 → 黃 → 紅 |
| ANSI | 預設 | 依整體百分比顯示單一顏色 |
| ASCII | `CLAUDE_STATUSLINE_ASCII=1` | `#` 已填，`-` 未填 |

### 範例：Nerd Font + true color

```bash
# 加入 ~/.zshrc 或 ~/.bashrc
export CLAUDE_STATUSLINE_NERDFONT=1
export COLORTERM=truecolor
```

---

## 從原始碼編譯

需要 [Go](https://go.dev/) 1.21 以上版本。

```bash
git clone https://github.com/harry18456/claude-code-statusline.git
cd claude-code-statusline
go build -o ~/.claude/statusline ./cmd/statusline/
chmod +x ~/.claude/statusline
```

---

## 關於花費顯示

`$cost` 是根據 Claude API token 費率計算的本次 session **估算值**。若你使用 Claude Pro 或 Max 訂閱方案，實際上不是按 token 計費，此數字僅供參考，不會反映在帳單頁面。

---

## 授權

[MIT](LICENSE)
