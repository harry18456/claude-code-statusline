# Statusline 下一步討論

> 起草：2026-06-15。定位：**討論起點草稿**，不是定案計畫。內容來自對「專案還能做什麼」的評估，待逐項討論後再決定要不要走 `/spectra:propose`。
> 前一版改善計畫（2026-06-13 全面 review 的逐項待辦）已全數處理完畢——CRITICAL / HIGH 全清、執行模式顯示 + flag 化 config 兩大功能 + cache 命中率皆完成，故清空重起。歷史記錄保留在 git log。

---

## 現況判斷

核心本職（把 Claude Code payload 渲染成兩行）已成熟紮實：~4142 行 Go、測試密、三層渲染（truecolor / ANSI / ASCII）、payload 容錯框架、跨平台 release workflow。**目前沒有進行中的 change。**

「差不多了」成立——以一個**聚焦的小工具**而言。剩下的不是「缺功能」，而是三種不同性質的事：

---

## A. 收尾與硬化（有限、值得做，因為每天在用）

- **L-2　NerdFont money glyph 遺失**（`internal/renderer/renderer.go`）：cost 段在 NerdFont 模式會雙空格且無圖示（`od -c` 確認 glyph 被吃掉，隔壁 time 段有 ``）。**肉眼可見的對齊 bug。** 修法是補 ``（nf-fa-money），卡在需要真 NerdFont 終端驗證 glyph 顯示，故未盲改。
- **測試缺口**（前一版 review 點名，需先核對是否仍存在）：
  - ANSI golden test——目前多用 `strings.Contains` 鬆散斷言，抓不到漏 reset 造成的顏色溢出；建議對 normal / danger 兩情境做完整 byte 比對。
  - `Render` 對 `nil *model.Payload` 的防禦 / contract test。
  - `main.go` / `debug-tee` 無測試（parse-error fallback、`isenv` 可抽成可測函式）。

> 註：前一版計畫中的 M-4（`filepath.Clean`）、M-7（時鐘回撥註解）依其「進度總表」已於 2026-06-14 處理，此處不再列為待辦；若要重新確認可查 git。

---

## B. 散布與採用（若目標是給別人用）

已有：跨平台 Go binary、release workflow、雙語 README。
缺的是「讓人願意裝」：

- README 放實際**截圖 / GIF**（現在是純文字描述，看不到成品長相）。
- 套件管理打包：scoop / winget（Windows）、Homebrew（macOS）。
- 一鍵安裝腳本。

---

## C. 新能力（可選，需過「值不值得佔兩行寬度」這關）

- **燃燒率指標**：cost/hour 或推估本輪成本，延續 cache% 建立的「成本感知」主軸，是最有延伸價值的一項。
- **不建議：config 檔**——才剛刻意從 env var 轉 flag，再加 config 檔等於把剛甩掉的複雜度迎回來。
- 其餘（可自訂段落順序、主題色等）容易讓工具變肥，門檻要拉高。

---

## 我的建議

別為了做而做。

- 日常依賴它　→　**A 價值最高**（修可見 bug + 補測試，保護天天用的東西）。
- 想推廣　→　做 **B**。
- 想加新功能　→　只挑 **C 的燃燒率**試水溫，其餘先放。

---

## 待討論

### D. 能力邊界查證（2026-06-15，9-agent workflow + 對抗驗證 + 本機 transcript 勘查）

針對「statusline 能否感知 ultracode/workflow、subagents、session 結束」三問的查證定論：

| 需求 | 結論 | 機制 / 限制 |
|------|------|-------------|
| 高 effort / thinking / fast | ✅ **已做到** | payload `effort.level`+`thinking.enabled`+`fast_mode`，line 1 已渲染 |
| ultracode / workflow 進行中 | ❌ 不可靠 | 官方把 ultracode 當行為非 effort level，payload 不外露；`/effort xhigh` 與 ultracode 的 xhigh 無法區分（啟發式 100% 誤判）；唯一痕跡 `workflow_keyword_request` 在主 transcript 內容、只證「曾觸發」 |
| active subagent 名稱 | ❌ payload 不給 | 提案欄位未合併；`agent.name` 只在 `--agent` session 有值 |
| 「N 個 subagent 在跑」 | ⚠ 可 hack、風險不可接受 | subagent 活動在 `<session>/subagents/` 子目錄、**不在** payload 的主 transcript；主 transcript 已 406KB、競態、事件 schema 官方未定義 |
| 知道 session 結束 | ✅ 但非 statusLine | statusLine 觸發時機不含結束；用 `SessionEnd` hook（只能清理/寫檔/通知）|
| session 結束時對使用者留訊息 | ✅ 用 `Stop` hook | Stop hook 能 block/注入（現有 `/goal` 即活例）；SessionEnd 已太晚 |

**共同洞察 → 候選新主軸**：三題都超出 stdin payload，但都能用同一招達成——
**Claude Code hook（SubagentStop / Stop / SessionEnd）寫狀態檔 → statusLine 讀檔渲染**。
這會把專案從「被動渲染 stdin」升級成「statusLine + 配套 hook 的狀態感知工具組」。工程量大（動 settings.json hook + 新 `cmd/`），待定要不要走。

**待確認項**：`SubagentStart` hook 是否存在（查證 agent 間說法不一）；statusLine binary 看到的環境變數集合是否完整等同 Bash subprocess（COLORTERM 可讀證明至少部分繼承，但未全面實測）。
