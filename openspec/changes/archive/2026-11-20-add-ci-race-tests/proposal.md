## Why

專案目前沒有任何在 push 或 pull request 時執行的測試。`release.yml` 只在推 `v*` tag 時 build，而且完全不跑測試，所以壞掉的程式碼也能被發版。

另外，`TestLookupSkipsSymlinks` 在本機 Windows 因為沒有建立 symlink 的權限而一直 SKIP，「略過 symlink」這條規則從來沒有被實際執行過。

## What Changes

- 新增 `.github/workflows/test.yml`：在 push 到 `main`、pull request，以及被其他 workflow 呼叫時，於 `ubuntu-latest` 與 `windows-latest` 執行 `go vet ./...` 與 `go test -race -count=1 -v ./...`。
- 兩個作業系統是獨立的 job（`fail-fast: false`），一邊失敗不會取消另一邊。token 權限只有 `contents: read`。
- `release.yml` 新增 `test` job 呼叫 `test.yml`，`build` 必須等 `test` 成功才執行。測試失敗就不 build，也不發 GitHub Release。
- `CLAUDE.md` 補上 CI 說明。

## Non-Goals

- 不加 staticcheck、gofmt 檢查、coverage 或 macOS runner。
- 不改 build matrix、binary 命名或 release 發佈方式。
- 不加 CI badge。
- 不在本機模擬 GitHub Actions（例如 `act`）；runner 上的行為以實際 push 後的 workflow run 為準。

## Capabilities

### New Capabilities

- `ci-tests`: push 與 pull request 時在 ubuntu 與 windows 執行 vet 與 race 測試，並作為 release build 的前置條件。

### Modified Capabilities

- `cross-platform-release`: 推 tag 觸發後先跑測試，全部通過才 build 與發佈。

## Impact

- Affected specs: `openspec/specs/ci-tests/spec.md`（新增）、`openspec/specs/cross-platform-release/spec.md`
- Affected code: `.github/workflows/test.yml`（新增）、`.github/workflows/release.yml`、`CLAUDE.md`
- 執行成本：每次 push `main` 多跑兩個 runner；發版要先等測試完成。
- 風險：Windows 的 `-race` 需要 cgo 與 gcc。`windows-latest` 映像內建 MinGW gcc，但要等第一次 workflow run 才能確認。
