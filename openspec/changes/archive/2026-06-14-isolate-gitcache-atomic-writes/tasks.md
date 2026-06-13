## 1. Git branch with dirty-check caching 測試契約

- [x] 1.1 建立 `Git branch with dirty-check caching` 的 `DefaultCacheFile(dir string)` 目錄 hash 契約測試：同一個 dir 產生穩定 cache path、不同 dir 產生不同 cache path、cache path 位於 `os.TempDir()`，並驗證 `go test ./internal/gitcache -run TestDefaultCacheFileDerivesDirectoryHash` 通過。
- [x] 1.2 建立跨 repo 隔離測試：兩個 temp git repos 使用不同 branch，測試內以 `t.Setenv` 將 temp 根目錄導向 `t.TempDir()`，在 5 秒 TTL 內各自呼叫公開 `Get` 後再次讀取第一個 repo 時不得回傳第二個 repo 的 branch，並驗證 `go test ./internal/gitcache -run TestGetUsesDirectoryIsolatedCache` 通過。
- [x] 1.3 建立併發刷新測試：多個 goroutine 對同一 temp git repo 以 stale TTL 反覆呼叫公開 `Get`，測試內以 `t.Setenv` 將 temp 根目錄導向 `t.TempDir()` 並確保每次都觸發 refresh+write 競爭，每次成功讀取都回傳完整且正確 branch，並驗證 `go test -race ./internal/gitcache -run TestGetConcurrentRefreshesUseAtomicCache` 通過。

## 2. Gitcache API 與寫入語意

- [x] 2.1 實作 per-directory cache filename：`DefaultCacheFile(dir string)` 使用 `filepath.Clean`、可用時使用 `filepath.Abs`、以 `sha256` 前 16 bytes 的 lowercase hex 作為檔名 suffix，並驗證 `go test ./internal/gitcache -run TestDefaultCacheFileDerivesDirectoryHash` 通過。
- [x] 2.2 收斂公開 cache path 選擇到 `Get(dir string, maxAge time.Duration)`，並保留私有 `getCached(dir, cacheFile string, maxAge time.Duration)` 作為可測核心：公開 `Get` 一律使用 `DefaultCacheFile(dir)`，既有 `TestNonGitDir` / `TestGetFromRealRepo` 透過 `getCached` 將 cache 導向 `t.TempDir()`，並驗證 `go test ./internal/gitcache -run 'Test(GetUsesDirectoryIsolatedCache|NonGitDir|GetFromRealRepo)'` 通過。
- [x] 2.3 實作 atomic cache write：`writeCache` 使用 `os.CreateTemp(filepath.Dir(cacheFile), ...)` 在 final cache 同目錄建立 temp file，完整寫入並 close 後以 `os.Rename` publish，任何 write、close、rename 失敗都清除 temp file，並驗證 `go test -race ./internal/gitcache -run TestGetConcurrentRefreshesUseAtomicCache` 通過且 `rg -n "os.WriteFile" internal/gitcache/gitcache.go` 無輸出。
- [x] 2.4 更新 statusline 呼叫端：`cmd/statusline/main.go` 使用新 `gitcache.Get(p.Workspace.CurrentDir, 5*time.Second)` API 並保持 renderer input 不變，驗證 `go build ./cmd/statusline` 通過。

## 3. 範圍控管與完整驗證

- [x] 3.1 做 scope review：diff 只包含 `internal/gitcache/gitcache.go`、`internal/gitcache/gitcache_test.go`、`cmd/statusline/main.go`，且未修改 `git diff --quiet` exit-code 判斷、`ModTime` TTL 判斷、symlink hardening、debug-tee、renderer 或 payload parsing；以 `git diff -- internal/gitcache/gitcache.go internal/gitcache/gitcache_test.go cmd/statusline/main.go` 內容審查驗證。
- [x] 3.2 執行完整驗證：`gofmt -l .` 無輸出、`go build ./...` 通過、`go vet ./...` 通過、`staticcheck ./...` 通過、`go test -race ./...` 通過。
