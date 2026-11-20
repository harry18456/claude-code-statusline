## 1. CI workflow

- [x] [P] 1.1 新增 `.github/workflows/test.yml` 實作 `Run vet and race tests in CI`：觸發條件為 push `main`、`pull_request`、`workflow_call`；matrix 為 `ubuntu-latest` 與 `windows-latest`，`fail-fast: false`；`permissions` 只有 `contents: read`；步驟為 checkout、setup-go（`go-version-file: go.mod`）、`go vet ./...`、`go test -race -count=1 -v ./...`；驗證：人工審查 YAML，並在本機執行相同的兩個指令全數通過。
- [x] 1.2 修改 `.github/workflows/release.yml` 實作 `Gate release builds on tests` 與 `Triggered by git tag push`：新增 `test` job，以 `uses: ./.github/workflows/test.yml` 呼叫；`build` 加上 `needs: test`；`release` 維持 `needs: build`；驗證：人工審查 job 相依鏈為 test → build → release。

## 2. 文件

- [x] [P] 2.1 更新 `CLAUDE.md`：說明 test workflow 的觸發條件、平台與指令，以及 release 必須先通過測試才 build；驗證 `grep -n "test.yml" CLAUDE.md` 找得到新段落。

## 3. 驗收

- [x] 3.1 乾淨視角 subagent 審查 `test.yml` 與 `release.yml`；驗證：審查報告沒有未處理的 CRITICAL 或 HIGH。
- [x] 3.2 替代驗證：本機無法執行 GitHub Actions，archive 前以本機 `go vet ./...` 與 `go test -race -count=1 ./...` 通過作為最接近的驗證；runner 上的行為（Windows gcc、symlink 測試是否實際執行）於 push 後以 `gh run watch` 與 workflow log 確認並回報。
- [x] 3.3 審查修正：Windows job 先印出 `gcc --version` 與 `go env CGO_ENABLED CC`；setup-go 設 `cache: false`（repo 沒有 `go.sum`）；job 加 `timeout-minutes: 15`；checkout 設 `persist-credentials: false`；`release.yml` 加 top-level `permissions: contents: read`（`release` job 的 `contents: write` 仍會覆寫）；驗證：人工審查兩個 workflow 檔。
