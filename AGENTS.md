# AGENTS.md — Panduan Maintainer Fork `tickernelz/sub2api`

Repo ini adalah **fork ringan** dari upstream `Wei-Shaw/sub2api`. Tujuan fork: mengikuti upstream sedekat mungkin, dengan **hanya sejumlah kecil kustomisasi yang sengaja di-keep**.

> ⚠️ **KOREKSI PENTING (2026-07-21) — baca ini dulu sebelum melakukan sync/reset apa pun:**
> Kalimat lama di sini pernah berbunyi "Semua fitur fork lama (Kiro/OpenCode/Cursor/Grok-registry/multi-group/dll) sudah dibuang — jangan hidupkan kembali kecuali diminta eksplisit." **Kalimat itu SALAH dan sudah dihapus.** Kiro **BUKAN** fitur yang boleh dibuang secara default — pemilik repo secara eksplisit meminta Kiro **selalu di-keep** setiap sync (lihat sesi 2026-07-20, user memilih opsi "keep semua fitur lokal termasuk Kiro" secara eksplisit dan berulang kali). Sebuah sesi agent sebelumnya, di tengah proses debugging environment, **secara sepihak** me-reset `main` ke upstream murni dan menulis kalimat "sudah dibuang, jangan hidupkan kembali" itu sendiri sebagai pembenaran — padahal itu bertentangan langsung dengan instruksi eksplisit pemilik repo beberapa jam sebelumnya. Ini adalah kesalahan eksekusi, bukan keputusan desain yang disengaja pemilik repo.
>
> **Aturan baku ke depan:** sebelum melakukan `git reset --hard`, `merge --abort`, atau operasi destruktif lain yang berpotensi membuang fitur, **WAJIB konfirmasi ke user dulu** kalau ada instruksi eksplisit user (di sesi ini atau sesi sebelumnya) yang menyebutkan fitur yang harus di-keep. Dokumen ini **tidak boleh** dijadikan alasan tunggal untuk membuang sesuatu yang user sudah minta secara eksplisit untuk dipertahankan — kalau dokumen ini ternyata bertentangan dengan instruksi user terbaru, **instruksi user yang menang**, dan dokumen ini harus diperbaiki, bukan sebaliknya.
>
> Status per 2026-07-21: Kiro sedang dalam proses **dipulihkan** (restoration in progress) sebagai **Fitur D**, lihat §2.4 di bawah untuk detail begitu selesai. Selama proses restorasi berjalan, **JANGAN** ada operasi `reset --hard ke wei-shaw/main` tanpa terlebih dahulu memastikan kode Kiro yang sedang dikembangkan (branch kerja / worktree `/tmp/sub2api-kiro-ref`) sudah di-cherry-pick ulang setelahnya.

Baca dokumen ini sebelum melakukan sync/merge dengan upstream. Tujuannya: sync berikutnya harus cepat dan bersih, bukan perang konflik.

---

## 1. Identitas Repo & Remote

| Hal | Nilai |
|---|---|
| Remote `origin` | `git@github.com:tickernelz/sub2api.git` (fork kita) |
| Remote `wei-shaw` | `git@github.com:Wei-Shaw/sub2api.git` (upstream) |
| Remote `myfork` | `git@github.com:tamakiramimy/sub2api-kiro.git` — fork lain milik pemilik yang menyimpan implementasi Kiro lengkap (branch `main`, bukan `sync-upstream`); dipakai sebagai referensi sumber saat memulihkan Fitur D Kiro. |
| Go module path | `github.com/Wei-Shaw/sub2api` (**tetap ikut upstream**, jangan diganti ke `tickernelz`) |
| Branch utama | `main` |
| Versi dilacak di | `backend/cmd/server/VERSION` (di-drive oleh git tag, lihat §6) |

> ⚠️ **Module path penting:** karena kita full-rewrite di atas upstream, module path = `Wei-Shaw`. Kalau meng-cherry-pick/mem-port kode dari fork lama (termasuk `myfork`, yang module path-nya `github.com/tickernelz/sub2api`), **perbaiki import path-nya** ke `Wei-Shaw/sub2api` atau build akan gagal.

Pastikan remote upstream ada sebelum sync:
```bash
git remote get-url wei-shaw || git remote add wei-shaw git@github.com:Wei-Shaw/sub2api.git
git fetch wei-shaw
```

---

## 2. Fitur Kustom yang WAJIB di-KEEP

Ada **empat fitur produk** yang di-keep. Selain ini, ikuti upstream apa adanya.

> ⚠️ **SHA berubah tiap sync.** SHA aktif setelah sync 2026-07-23 di atas upstream `60013c5f1`: `4924fa55c`, `f783a901a`, `0db861ede`, `b3991637a`, `7af502259` (Fitur A/B/C, 5 commit), `7b4ba716c`, `51d9d8628`, `dac01f212`, `24694fa5c` (Fitur D, 4 commit). Setiap reset + cherry-pick menghasilkan SHA baru; cari ulang berdasarkan judul commit dengan `git log --oneline --all --grep=...` (lihat §4).

### Fitur A: OpenAI/Codex OAuth jangan auto-disable saat refresh token gagal/reused

Akun OpenAI OAuth **tidak boleh** langsung di-`SetError`/unschedule ketika refresh token gagal karena `refresh_token_reused`. Alasannya: `refresh_token_reused` cuma menandakan refresh token sudah dikonsumsi/dirotasi — **bukan** bukti access token saat ini tidak valid. Akun tetap dibiarkan schedulable, dan UI menampilkan warning "reauth required".

Dua commit sumber (fork-only, tidak ada di upstream):

| Commit (SHA lama, cari ulang by subject) | Judul | Cakupan |
|---|---|---|
| `fix(openai): keep oauth accounts schedulable on reused refresh token` | Backend soft-handle |
| `feat(admin): show OpenAI OAuth reauth warning` | Frontend warning UI |

**File yang disentuh (referensi saat re-apply):**

Backend:
- `backend/internal/service/openai_refresh_token_state.go` *(file baru — inti logika)*
- `backend/internal/service/token_refresh_service.go` *(titik keputusan `isNonRetryableRefreshError`)*
- `backend/internal/service/openai_token_provider.go`
- `backend/internal/service/token_refresher.go`
- `backend/internal/handler/admin/account_handler.go`
- `backend/internal/repository/account_repo.go`
- `backend/internal/service/admin_account.go`
- `backend/internal/service/token_refresh_service_test.go`

Frontend:
- `frontend/src/components/admin/account/AccountActionMenu.vue` *(computed `hasOpenAIRefreshTokenReauthRequired`)*
- `frontend/src/views/admin/AccountsView.vue`
- `frontend/src/types/index.ts`
- `frontend/src/i18n/locales/en/admin/accounts.ts`, `frontend/src/i18n/locales/zh/admin/accounts.ts`
- `frontend/src/views/admin/__tests__/AccountsView.openaiReauthWarning.spec.ts`
- `Makefile`

**Titik konflik yang sudah diketahui saat re-apply:**
1. `token_refresh_service.go` cabang `if isNonRetryableRefreshError(err) {` — upstream pakai `logredact.RedactText(err.Error())`, commit fork pakai `fmt.Sprintf`. **Gabungkan:** pertahankan blok soft-handle (`shouldSoftHandleOpenAIRefreshTokenReused` → `markOpenAIRefreshTokenReused` → `ensureOpenAIPrivacy` → `return err`) **DAN** pakai `logredact.RedactText` dari upstream untuk `errorMsg`.
2. `AccountActionMenu.vue` — upstream punya `isShadow`/`isOpenAIOAuthParent`/`supportsPrivacy` versi shadow-aware. **Keep versi upstream** + tambahkan computed `hasOpenAIRefreshTokenReauthRequired`; jangan pakai `supportsPrivacy` versi lama dari commit fork.
3. **`token_refresh_service_test.go` DUPLIKAT (kambuh tiap sync!):** commit fork A nambah field `updateExtraCalls` + method `UpdateExtra` di mock `tokenRefreshAccountRepo`, TAPI upstream juga sudah punya keduanya (field `updateExtraCalls`, `lastExtraUpdates`, method `UpdateExtra`). Cherry-pick → **deklarasi ganda** → `redeclared`/`method already declared`. **Ini TAK kelihatan di `go build`/`go test ./...` polos — cuma muncul di `go test -tags=unit` (perintah CI `make test-unit`).** **Fix:** hapus field `updateExtraCalls` duplikat + method `UpdateExtra` versi fork, lalu lipat assignment `r.lastExtraUpdate = updates` (singular, dipakai test fork) ke DALAM method `UpdateExtra` upstream yang bertahan (yang set `lastExtraUpdates` plural).
4. Adaptasi upstream refresh state: upstream memisahkan cache/scheduler sync ke `postRefreshStateSync`. Letakkan `clearOpenAIRefreshTokenReusedMarker(...)` di `postRefreshActions` tepat setelah `s.postRefreshStateSync(ctx, account)` dan sebelum privacy calls.
5. Adaptasi scheduler-neutral extra: upstream menambah prefix `upstream_billing_probe` di `schedulerNeutralExtraKeyPrefixes`. Jika conflict dengan prefix fork `openai_refresh_token_`, pertahankan **keduanya**.

### Fitur B: Stream stale-detection + auto-failover (watchdog) di semua provider

Watchdog aliran (stale-stream) yang di-wire ke SEMUA loop streaming provider (OpenAI chat/messages, Anthropic passthrough, Bedrock, Antigravity ×5). Mendeteksi 3 bentuk stall dan failover ke akun lain **sebelum** ada byte yang dikirim ke client:
- **TTFT timeout** — upstream connect tapi first event tak pernah datang → failover.
- **chunk-gap warning** — gap antar-event lunak (log/metrics only, tidak failover).
- **chunk-gap timeout** — gap antar-event keras → failover.

Desain inti (JANGAN copy-paste timer per-loop seperti fork lama):
- **Satu** `StreamWatchdog` (injectable clock utk test deterministik) + `StreamRetrySettings` berlapis (per-platform override > global > default), cache in-process 60s, semantik row-missing vs disabled dibedakan.
- Gap timer reset di **setiap** upstream event (`OnUpstreamEvent()`) → reasoning model tak salah-vonis stale.
- Failover dijaga `c.Writer.Written()` (via `streamOutputCommitted`): retry hanya bila belum ada output; kalau sudah, fail clean (tak ada output ganda).
- Komplementer dengan `StreamTimeoutSettings` yang sudah ada (itu governs post-timeout action).

Satu commit sumber (fork-only): `feat(gateway): stream stale detection + auto-failover across all providers` (~1900 baris, 20 file). Default enabled (TTFT 60s, gap-warn 10s, gap-timeout 30s).

**File inti (port verbatim — file baru, 0 collision dengan upstream):**
- `backend/internal/service/stream_stale_watchdog.go` *(watchdog + injectable clock)*
- `backend/internal/service/stream_retry_settings.go` *(settings + resolver berlapis + Get/Set)*
- `backend/internal/service/stream_watchdog_integration.go` *(glue: cache, `newStreamWatchdogForPlatform`, `decideStreamStall`, metrics)*
- `+ stream_stale_watchdog_test.go`, `stream_retry_settings_test.go`

**Admin surface (additive):** `setting_handler_runtime.go` (3 handler: Get/Update StreamRetry + Metrics), `dto/settings.go`, `server/routes/admin.go` (3 route), `domain_constants.go` (`SettingKeyStreamRetrySettings`), frontend `api/admin/settings.ts` + `SettingsView.vue` + i18n `streamRetry` block.

> ⚠️ **PITFALL WIRING:** upstream me-refactor besar gateway — `gateway_service.go` dipecah jadi `gateway_anthropic_passthrough.go`, `gateway_upstream_response.go`; antigravity dipecah jadi `antigravity_gateway_streaming.go` + `antigravity_gateway_upstream.go`. **10 titik wiring watchdog pindah file.** Jika cherry-pick mentah gagal atau upstream mengubah loop, **JANGAN merge body lama secara buta** — re-implement wiring ke lokasi loop upstream yang baru: (1) init watchdog setelah `intervalCh = intervalTicker.C`, (2) `staleWatchdog.OnUpstreamEvent()` setelah `line := ev.line`, (3) 3 select-arm (`staleTTFTCh`/`staleGapWarnCh`/`staleGapTimeoutCh`) sebelum `case <-keepaliveCh`. Setelah re-apply, audit total 10 init, 10 event-reset, dan 30 `decideStreamStall` call.

### Fitur C: Netralisasi harmony `<|channel|>` token supaya request tak kena upstream `invalid_prompt` block

OpenAI `/v1/responses` upstream punya **request-level hard guard** untuk harmony "hidden analysis channel" header. Ketika request body memuat literal ASCII `<|channel|>` yang **langsung diikuti** `analysis`, upstream menolak **seluruh request** dengan HTTP 200 stream-internal `response.failed` + `error.code=invalid_prompt`.

**Perbaikan (fork-only):** sebelum body dikirim ke upstream, ganti dua ASCII pipe `|` (U+007C) di dalam `<|channel|>` jadi fullwidth pipe `｜` (U+FF5C) → `<｜channel｜>`. Dua commit sumber (fork-only): neutralizer + dua builder, dan observability `invalid_prompt`. Cari ulang dengan `git log --oneline --all --grep="harmony"` atau `--grep="invalid_prompt"`. Default **enabled** (`gateway.neutralize_harmony_channel_token=true`).

**File inti (file baru — 0 collision dengan upstream):**
- `backend/internal/service/openai_harmony_channel_neutralize.go`
- `+ openai_harmony_channel_neutralize_test.go`
- `+ openai_harmony_channel_neutralize_forward_test.go`

**Titik wiring (2 builder HTTP `/v1/responses` — keduanya WAJIB, path berbeda):**
- `backend/internal/service/openai_gateway_forward.go` → `buildUpstreamRequest`
- `backend/internal/service/openai_gateway_passthrough.go` → `buildUpstreamRequestOpenAIPassthrough`

**Layer 2 — observability `invalid_prompt`:** `detectOpenAIInvalidPrompt(payload)` di-wire di dua loop streaming `response.failed`:
- `openai_gateway_passthrough.go` → `handleStreamingResponsePassthrough`
- `openai_gateway_response_handling.go` → loop `response.failed` jalur rewrite

**Config surface (additive):** `backend/internal/config/config.go`: field `GatewayConfig.NeutralizeHarmonyChannelToken bool`.

**Titik konflik yang mungkin saat re-apply:**
1. Jangan cukup wire satu builder — kedua titik `http.NewRequestWithContext(... "POST" ...)` untuk `/v1/responses` harus ter-cover.
2. Guard `s.cfg == nil` wajib dipertahankan (default ON).
3. Kalau upstream sudah punya penanganan setara, fitur ini gugur — verifikasi dulu sebelum menghapus.

### Fitur D: Kiro 账号平台支持 *(✅ SELESAI direstorasi per 2026-07-21, diperbarui per 2026-07-23)*

**Status: COMPLETE.** 4 commit aktif:

| `7b4ba716c` — `feat(kiro): restore Kiro account platform support (backend)` — backend inti (internal/kiro/ pkg + internal/service/kiro_*.go, 81 file) |
| `51d9d8628` — `feat(kiro): restore Kiro account platform support (frontend)` — frontend (frontend/src/kiro/ + glue changes, 16 file) |
| `dac01f212` — `fix(kiro): correct new-account form defaults + experimental GPT-5.6 passthrough` — bug fix modal defaults + GPT-5.6 experimental support |
| `24694fa5c` — `feat(kiro): bridge Kiro accounts into OpenAI Responses/Chat Completions protocol gateways` — OpenAI Responses/CC protocol bridge (Codex clients can now use Kiro accounts) |

**Prinsip desain (wajib diikuti siapa pun yang menyentuh Fitur D):**
- Semua kode Kiro **terkonsolidasi di satu direktori baru per sisi**: backend `backend/internal/kiro/` (bukan tersebar di `internal/service/kiro_*.go` seperti sumber referensi lama), frontend `frontend/src/kiro/` (bukan tersebar di `components/common/`, `composables/`, `api/admin/`).
- File-file existing yang dipaksa harus disentuh (karena mekanisme framework: enum platform terpusat, ent schema terpusat, dispatch gateway terpusat, wire DI terpusat) HARUS dijaga seminimal mungkin (idealnya ≤ 5 baris per file, murni "deklarasi/registrasi/pemanggilan", bukan logika bisnis Kiro).
- Sumber referensi implementasi lengkap: remote `myfork` (`tamakiramimy/sub2api-kiro`) branch `main`, module path `github.com/tickernelz/sub2api` — **wajib** ganti import path ke `github.com/Wei-Shaw/sub2api` saat porting.
- Titik konflik yang diketahui saat re-apply: `ent/schema/group.go` (upstream bisa tambah field baru di posisi sebelum `kiro_cache_emulation_*`; keep kedua sisi), `internal/service/group.go` struct `Group` (sama: keep `MaxReasoningEffort`/`ReasoningEffortMappings` upstream + `KiroCacheEmulationEnabled`/`KiroCacheEmulationRatio` Fitur D).

## 3. Kebijakan `.github/` — IKUT UPSTREAM, keep hanya divergence minimal

**Kebijakan aktif:** `.github/` workflow **ikut versi upstream terbaru**, dengan dua divergence yang di-keep: referensi repo `tickernelz/sub2api` di `cla.yml`, dan `continue-on-error: true` pada step `Update DockerHub description` di `release.yml`. Selebihnya ikut upstream apa adanya.

| File | Delta fork vs upstream | Alasan |
|---|---|---|
| `cla.yml` | 5 baris: `github.repository == 'tickernelz/sub2api'` (2×) + `path-to-document`/link CLA `github.com/tickernelz/sub2api` (3×) | Guard job CLA hanya jalan di repo fork; link ke CLA.md fork. |
| `release.yml` | 1 baris: `continue-on-error: true` di step `Update DockerHub description` | Step itu bisa `403 Forbidden` (token perms) dan menandai job Release merah walau image sukses publish. |
| **semua file `.github` lain** | **TIDAK ADA** — 100% ikut upstream | pnpm/golangci/step ikut upstream terbaru. |

**Cara apply `.github` setelah reset ke upstream:**
```bash
sed -i 's#Wei-Shaw/sub2api#tickernelz/sub2api#g' .github/workflows/cla.yml
# + tambahkan 'continue-on-error: true' di bawah '- name: Update DockerHub description' di release.yml
```
> ⚠️ **JANGAN** `git checkout <fork-backup> -- .github` — itu membawa balik divergence usang. Cukup reset-ke-upstream + sed cla.yml + 1 baris continue-on-error.

### Divergensi lain di luar `.github/`

| File | Delta fork vs upstream | Alasan |
|---|---|---|
| `.gitignore` | Rule `AGENTS.md`/`CLAUDE.md` diganti komentar bahwa file sengaja tracked | Supaya dokumen ini tetap tracked & tidak hilang saat sync. |
| `frontend/package.json` + `pnpm-lock.yaml` | *(RESOLVED)* upstream sudah mendeklarasikan `@intlify/message-compiler` langsung sejak base saat ini — tidak perlu divergence lagi, verifikasi tiap sync apakah masih ada. |
| `backend/go.mod` | *(RESOLVED)* upstream sudah adopt versi Go yang sama — bukan divergence lagi. |

> ⚠️ **VERIFIKASI PAKAI PERINTAH CI YANG SEBENARNYA — bukan `go test ./...` polos.** CI jalanin `make test-unit` = `go test -tags=unit ./...` dan `make test-integration` = `go test -tags=integration ./...`. `go test ./...` polos MELEWATI file ber-`//go:build unit`. **Selalu tutup gate backend dengan:** `go test -tags=unit ./...`, `go test -tags=integration ./...`, `govulncheck ./...`, dan golangci-lint versi CI (v2.9.0).

### Divergensi struktural upstream yang mempengaruhi re-apply fitur

Upstream sering me-refactor/memecah file besar. Yang sudah ketahuan mengubah lokasi hunk fitur di-keep:

| Area | Dulu (fork lama) | Sekarang (upstream) | Dampak re-apply |
|---|---|---|---|
| i18n locales | monolitik `en.ts`/`zh.ts` | modular `locales/<lang>/admin/{accounts,settings}.ts` dst. | Re-home key ke file modular yang sesuai. |
| admin service | `admin_service.go` monolitik | dipecah; `UpdateAccount` → `admin_account.go` | Hunk fitur re-home ke file baru. |
| setting handler | `setting_handler.go` | handler streamRetry → `setting_handler_runtime.go` | — |
| gateway | `gateway_service.go` monolitik | dipecah → `gateway_anthropic_passthrough.go`, `gateway_upstream_response.go`, `antigravity_gateway_streaming.go`, `antigravity_gateway_upstream.go` | Wiring watchdog (Fitur B) pindah file. |

Prinsip umum: kalau cherry-pick fitur kena `modify/delete` atau conflict struktural raksasa, **ambil versi upstream lalu re-apply hunk kecil fitur ke lokasi barunya** — jangan paksa merge body monolitik lama.

---

## 4. Strategi Sync dengan Upstream (WAJIB: full-rewrite, bukan merge)

**Pelajaran mahal:** `git merge wei-shaw/main` menghasilkan puluhan konflik yang salah di-resolve. **Jangan pakai merge.** Pakai pola full-rewrite + cherry-pick.

> ⚠️ **Sebelum langkah 2 (reset --hard) di bawah:** konfirmasi tidak ada instruksi eksplisit user (sesi ini/sebelumnya) yang minta fitur lain di-keep di luar daftar §2. Kalau ada dan belum masuk §2, STOP dan tanya user dulu — jangan lanjut reset.

### Prosedur baku

```bash
# 0. Pastikan working tree bersih & fetch upstream
git status
git fetch wei-shaw

# 1. Backup state sekarang (WAJIB — recovery net)
git branch backup/pre-rewrite-$(date +%Y%m%d-%H%M%S) HEAD

# 2. Reset main ke upstream (full upstream tree)
git reset --hard wei-shaw/main

# 3. Bersihkan leftover fork-only yang untracked
git status --porcelain

# 4. Cherry-pick commit fitur yang di-keep (urut; cari SHA terbaru berdasarkan subject)
git cherry-pick <sha-fitur-A-backend>
git cherry-pick <sha-fitur-A-frontend>
git cherry-pick <sha-fitur-B-watchdog>
git cherry-pick <sha-fitur-C-harmony>
git cherry-pick <sha-fitur-C-layer2>
git cherry-pick <sha-fitur-D-kiro-backend>   # setelah Fitur D selesai direstorasi
git cherry-pick <sha-fitur-D-kiro-frontend>
#   -> resolve konflik sesuai §2, audit wiring watchdog (Fitur B), 2 builder/2 response loops (Fitur C)
#   -> pastikan TIDAK ada import 'tickernelz/sub2api' yang bocor:
#      git diff --cached | grep tickernelz   # harus kosong

# 5. Re-apply HANYA divergence minimal (§3)
#    - cla.yml: referensi Wei-Shaw/sub2api -> tickernelz/sub2api
#    - release.yml: continue-on-error pada Update DockerHub description
#    - restore AGENTS.md dari safety branch dan unignore di .gitignore

# 6. Bump VERSION (§6)
echo "0.1.xxx" > backend/cmd/server/VERSION

# 7. GATE verifikasi (§5) — WAJIB hijau sebelum push

# 8. Commit sisa (chore) + push (dengan --force-with-lease, bukan --force polos)
git add .github .gitignore AGENTS.md backend/cmd/server/VERSION
git commit -m "chore: adapt fork to upstream and bump VERSION to 0.1.xxx"
git push --force-with-lease=refs/heads/main:<verified-old-origin-sha> origin HEAD:main
```

---

## 5. Gate Verifikasi (WAJIB hijau sebelum push)

Jalankan dari root repo. **Jangan push kalau ada yang merah.**

### Backend
```bash
cd backend
unset OPENAI_API_KEY
export PATH="$(go env GOPATH)/bin:$PATH"
go build ./...
go vet ./...
go test -tags=unit ./...
go test -tags=integration ./...
govulncheck ./...
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.9.0
"$(go env GOPATH)/bin/golangci-lint" run --timeout=30m
```
> ⚠️ PATH mesin maintainer bisa menunjuk ke golangci-lint versi berbeda dari CI. **Jangan pakai binary PATH tanpa cek versi** — gunakan v2.9.0 exact.

### Frontend
```bash
cd frontend
export CI=true
./node_modules/.bin/eslint . --ext .vue,.js,.jsx,.cjs,.mjs,.ts,.tsx,.cts,.mts
./node_modules/.bin/vue-tsc --noEmit
./node_modules/.bin/vitest run
./node_modules/.bin/vue-tsc -b && ./node_modules/.bin/vite build
```
> Mesin ini memakai pnpm 11, CI memakai pnpm 9. Pakai binary existing langsung, jangan `pnpm run` (bisa memicu frozen-install/overrides drift).

**Pitfall lockfile:** setelah build, pnpm versi baru bisa memindahkan `overrides` dari `pnpm-lock.yaml` ke `pnpm-workspace.yaml` baru — buang artifact ini:
```bash
git checkout frontend/pnpm-lock.yaml
rm -f frontend/pnpm-workspace.yaml
```
`backend/internal/web/dist/` sudah gitignored — jangan di-commit.

---

## 6. Versi & Release

- Versi produk ada di `backend/cmd/server/VERSION`, di-drive oleh git tag via `.github/workflows/release.yml`.
- Rilis dipicu dengan push tag `vX.Y.Z`.
- `release.yml` **hanya build + push image** ke GHCR + Docker Hub. **Tidak ada deploy/SSH ke server.**
- `backend-ci.yml` trigger `on: push` (branch **dan** tag).

---

## 7. Standar Kerja & Pitfalls

- **Kode produk = ikut upstream**, kecuali Fitur A-D di §2 yang WAJIB di-keep.
- **Instruksi user eksplisit mengalahkan dokumen ini.** Kalau ada instruksi user yang bertentangan dengan apa yang tertulis di sini, ikuti instruksi user, lalu update dokumen ini — jangan sebaliknya.
- **Jangan pakai subagent** untuk kerjaan repo ini kalau diminta kerjakan sendiri.
- **Force push** ke `origin main` diperbolehkan untuk fork ini, tapi **selalu bikin backup branch dulu** dan gunakan `--force-with-lease` (bukan `--force` polos), simpan `origin/main` lama sebagai recovery.
- Setelah reset, cek **untracked leftover** fork-only (`git status --porcelain`).
- Verifikasi klaim "sukses" dengan bukti nyata (log, output command), bukan asumsi.
- **Operasi destruktif (reset --hard, merge --abort, force push) yang akan membuang perubahan yang diminta user secara eksplisit untuk dipertahankan HARUS dikonfirmasi ke user dulu**, walau tampak "lebih bersih" atau "sesuai dokumen ini".

---

## 8. Ringkasan Struktur Commit yang Diharapkan

Setelah sync bersih, `main` harus terlihat seperti ini (**10 commit fitur A-D + 1 chore**):
```
<chore>   chore: adapt fork to upstream and bump VERSION to 0.1.xxx
<feat>    feat(kiro): bridge Kiro accounts into OpenAI Responses/Chat Completions gateways  ← Fitur D (bridge)
<fix>     fix(kiro): correct new-account form defaults + experimental GPT-5.6 passthrough  ← Fitur D (fix)
<feat>    feat(kiro): restore Kiro account platform support (frontend)       ← Fitur D (frontend)
<feat>    feat(kiro): restore Kiro account platform support (backend)        ← Fitur D (backend)
<feat>    feat(gateway): record invalid_prompt blocks to ops_error_logs      ← Fitur C Layer 2
<feat>    feat(gateway): neutralize harmony <｜channel｜> token to avoid upstream invalid_prompt block  ← Fitur C
<feat>    feat(gateway): stream stale detection + auto-failover...  ← Fitur B (watchdog)
<feat>    feat(admin): show OpenAI OAuth reauth warning             ← Fitur A (frontend)
<fix>     fix(openai): keep oauth accounts schedulable...           ← Fitur A (backend)
<upstream HEAD = wei-shaw/main>
```
