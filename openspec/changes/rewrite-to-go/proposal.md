## Why

目前 `statusline.sh` 依賴 bash + `jq`，在 Windows 上無法原生執行，且 macOS 專屬的 `stat -f %m` 語法導致 Linux 快取功能失效。改寫為 Go 並編譯為各平台 binary，可讓任何平台的使用者無需安裝任何 runtime 即可使用。

## What Changes

- 以 Go 重寫 `statusline.sh` 的全部功能，輸出行為保持一致
- 採用 `cmd/statusline/` + `internal/` 的 Go 慣用專案結構
- 以 TDD 方式開發：每個 `internal/` 子包先寫測試，再實作
- 移除對 `jq` 的外部依賴（改用 Go 標準庫 `encoding/json`）
- 修正 `stat -f %m` 僅限 macOS 的 bug（改用 `os.Stat().ModTime()`）
- 以 GitHub Actions 自動編譯並發布四個平台 binary：
  - `statusline-darwin-amd64`
  - `statusline-darwin-arm64`
  - `statusline-linux-amd64`
  - `statusline-windows-amd64.exe`
- 更新 `install.sh` 支援自動偵測平台並下載對應 binary
- **舊版 bash 檔案移至 `reference/` 暫存，Go 版本確認後刪除**（不保留 legacy）
- 以 `ldflags -X main.version` 在 build time 注入版本號，release binary 自動帶上 git tag（本地開發預設為 `dev`）

## Capabilities

### New Capabilities

- `go-statusline`: Go 實作的 statusline 主程式，功能對等 bash 版本，支援全平台
- `cross-platform-release`: GitHub Actions workflow，自動交叉編譯並發布各平台 binary

### Modified Capabilities

- （無 spec 層級的需求變更，行為規格維持不變）

## Impact

- 新增 `cmd/statusline/main.go`、`internal/model/`、`internal/renderer/`、`internal/gitcache/`
- 新增 `go.mod`、`go.sum`
- 新增 `.github/workflows/release.yml`
- 修改 `install.sh` — 加入平台偵測與 binary 下載邏輯
- 修改 `README.md` / `README.zh-TW.md` — 更新安裝說明
- 移動舊版檔案：`statusline.sh`、`install.sh`（舊）、`examples/`、`docs/slides.js` → `reference/`（最終刪除）
- 移除 `jq` 作為必要依賴
