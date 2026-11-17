# Vertex OS — Todo

Status legend: `[ ]` todo · `[~]` in progress · `[x]` done · `[!]` blocked.
Granularity: task-level. Each checkbox is a single concrete deliverable.
Module path: `github.com/ModernSelfHost/vertexos`.

## Milestone 0 — Foundations

### Repo & tooling
- [x] Initialize git repository in project root
- [x] Add `.gitignore` (Go, Node, OS, secrets)
- [x] Create top-level layout: `cmd/`, `internal/`, `web/`, `installer/`, `docs/`, `scripts/`
- [x] Initialize Go module (`go.mod`) with module `github.com/ModernSelfHost/vertexos`
- [x] Pin Go toolchain version in `go.mod`
- [x] Add `Makefile` with `build`, `test`, `lint`, `run` targets
- [x] Add `golangci-lint` config
- [x] Initialize React + TypeScript SPA under `web/` (Vite)
- [x] Add ESLint + Prettier config for `web/`
- [x] Add CI workflow: lint + test on PR
- [x] Add CI workflow: build ISO (placeholder, wired in M3)
- [x] Write `README.md` with dev setup

### vertexd skeleton
- [x] `cmd/vertexd/main.go` entrypoint
- [x] Config loading (file + env + flags)
- [x] Structured logging via `slog`
- [x] Graceful shutdown
- [x] HTTP router with `/api/v1` prefix
- [x] `GET /api/v1/health` endpoint
- [x] Self-signed TLS cert generation on first run
- [x] Serve static SPA bundle from daemon
- [x] systemd unit file for `vertexd`
- [x] Document root-running + scoped internal packages

### Auth & users (basic)
- [x] SQLite schema: `users` (id, username, password_hash, role, created_at)
- [x] Argon2id password hashing helper
- [x] Bootstrap first admin account
- [x] `POST /api/v1/auth/login`
- [x] `POST /api/v1/auth/logout`
- [x] Short-lived session token issuance
- [x] Secure cookies + CSRF protection
- [x] Auth middleware
- [x] Role-based authorization middleware (admin/user/guest)
- [x] Rate limiting on login
- [x] Account lockout after N failures

### Job queue
- [x] SQLite schema: `jobs` (id, type, status, progress, created_at, updated_at)
- [x] Job manager: enqueue, run, cancel
- [x] Status transitions (queued/running/succeeded/failed/cancelled)
- [x] Progress reporting API
- [x] Per-job log storage + streaming
- [x] WebSocket event for job progress
- [x] `GET /api/v1/jobs` (list/history)
- [x] Safe retry on failure

### State store & reconciliation
- [x] SQLite migrations framework
- [x] Desired-state tables: `settings`, `shares` (placeholder), `users`
- [x] Reconciler interface invoked on boot
- [x] Detect direct edits to managed config files, flag them
- [x] `GET /api/v1/system/export` (config export)
- [x] `POST /api/v1/system/import` (config import)

### Wrappers
- [x] Exec helper with timeout + structured output
- [x] `zpool`/`zfs` JSON parser helpers
- [x] `smartctl` parser helpers
- [x] Network/systemd helpers (stubs)

### vertex-cli
- [x] `cmd/vertexctl` entrypoint
- [x] HTTP client to local API
- [x] `auth login` command
- [x] `health` command

### Discovery
- [x] mDNS advertisement of `vertex.local`
- [x] Print console URL on boot

## Milestone 1 — Storage

### Drive detection
- [x] List block devices via `lsblk`
- [x] Resolve stable IDs via `/dev/disk/by-id`
- [x] Capture model/size/serial
- [x] SMART data via `smartctl`
- [x] Drive health classification
- [x] `GET /api/v1/drives`
- [x] WebSocket drive status updates

### Pool management (advanced)
- [x] Parse `zpool status`
- [x] `GET /api/v1/pools`
- [x] `GET /api/v1/pools/{name}`
- [x] Create pool (manual vdev spec) endpoint → job
- [x] Add vdev endpoint → job
- [x] Trigger scrub endpoint → job
- [x] Scrub schedule CRUD
- [x] Destroy pool endpoint (typed confirm) → job
- [x] Pool property get/set

### Simple-mode layout proposal
- [x] Layout algorithm: 1→single (warn), 2→mirror, 3+→RAIDZ1, 5+→offer RAIDZ2
- [x] Usable-space + fault-tolerance calculation
- [x] Plain-language explanation strings
- [x] `POST /api/v1/pools/propose`
- [x] Create pool from proposal endpoint

### SMART monitoring
- [x] Periodic SMART poll
- [x] Alert generation on SMART failure
- [x] Alert storage + `GET /api/v1/alerts`
- [x] Alert dismissal

### Dashboard storage view
- [x] Storage usage aggregation (pool capacity)
- [x] `GET /api/v1/dashboard/storage`
- [ ] Frontend storage card (usage bar, health)

## Milestone 2 — Shares and users

### Users
- [ ] User CRUD endpoints
- [ ] Role assignment
- [ ] Password change (self + admin reset)
- [ ] Map web accounts to OS system accounts
- [ ] SMB password sync

### Shares
- [ ] Create dataset per share
- [ ] Share CRUD endpoints (name, dataset, quota, access list)
- [ ] Quota set via ZFS
- [ ] Access list → Samba config regeneration
- [ ] Managed Samba config file
- [ ] NFS export management (advanced)
- [ ] Guest-accessible flag
- [ ] Time Machine preset (later)

### Permissions
- [ ] Permission matrix (user/role → read/write)
- [ ] Apply permissions to dataset ACLs
- [ ] `GET /api/v1/shares/{id}/permissions`

### Frontend
- [ ] Users screen
- [ ] Shares list + create wizard
- [ ] Permissions editor

## Milestone 3 — Installer and first-run

### Installer
- [ ] live-build/mkosi pipeline
- [ ] Debian base package set
- [ ] Include OpenZFS, Samba, NFS, smartmontools, nftables
- [ ] Boot disk selection
- [ ] First admin creation in installer
- [ ] Hostname setting
- [ ] Reboot + console URL

### First-run wizard (web)
- [ ] Welcome screen
- [ ] Name server
- [ ] Create admin (if not done in installer)
- [ ] Drive detection
- [ ] Pool proposal + creation
- [ ] First share creation
- [ ] Completion screen

### Config export/import
- [ ] Export (zip with SQLite + configs)
- [ ] Import with validation

## Milestone 4 — Hardening and v1

### Advanced storage mode
- [ ] UI toggle + persisted preference
- [ ] Pools/vdevs/datasets views
- [ ] Properties editor (compression, recordsize, quotas)
- [ ] Snapshots list/create/destroy
- [ ] Scrub schedule editor

### OS updates
- [ ] Update detection via apt
- [ ] Approval flow endpoint
- [ ] ZFS boot-environment snapshot before update
- [ ] Apply update job with status
- [ ] Rollback documentation (boot menu)
- [ ] Channels: stable + testing

### Logging & alerts
- [ ] Central log collection
- [ ] Logs view endpoint + UI
- [ ] Alert rules (SMART, pool degraded, capacity)
- [ ] In-app alert notifications

### Security hardening
- [ ] Input validation review
- [ ] Least-privilege review for daemon
- [ ] nftables default rules
- [ ] Security headers

### Docs
- [ ] User guide
- [ ] Tested-hardware list
- [ ] Name/trademark check (open question 5)
- [ ] License selection (open question 6)

## Open questions (design doc §12)
- [ ] App packaging decision (deferred; keep container-runtime seam)
- [ ] ZFS on root vs simpler boot filesystem
- [ ] Mixed-size drive support stance
- [ ] Hardware compatibility publishing
- [ ] Name/trademark check for "Vertex"
- [ ] Open-source license choice
