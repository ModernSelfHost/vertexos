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
- [ ] Initialize React + TypeScript SPA under `web/` (Vite)
- [ ] Add ESLint + Prettier config for `web/`
- [ ] Add CI workflow: lint + test on PR
- [ ] Add CI workflow: build ISO (placeholder, wired in M3)
- [ ] Write `README.md` with dev setup

### vertexd skeleton
- [ ] `cmd/vertexd/main.go` entrypoint
- [ ] Config loading (file + env + flags)
- [ ] Structured logging via `slog`
- [ ] Graceful shutdown
- [ ] HTTP router with `/api/v1` prefix
- [ ] `GET /api/v1/health` endpoint
- [ ] Self-signed TLS cert generation on first run
- [ ] Serve static SPA bundle from daemon
- [ ] systemd unit file for `vertexd`
- [ ] Document root-running + scoped internal packages

### Auth & users (basic)
- [ ] SQLite schema: `users` (id, username, password_hash, role, created_at)
- [ ] Argon2id password hashing helper
- [ ] Bootstrap first admin account
- [ ] `POST /api/v1/auth/login`
- [ ] `POST /api/v1/auth/logout`
- [ ] Short-lived session token issuance
- [ ] Secure cookies + CSRF protection
- [ ] Auth middleware
- [ ] Role-based authorization middleware (admin/user/guest)
- [ ] Rate limiting on login
- [ ] Account lockout after N failures

### Job queue
- [ ] SQLite schema: `jobs` (id, type, status, progress, created_at, updated_at)
- [ ] Job manager: enqueue, run, cancel
- [ ] Status transitions (queued/running/succeeded/failed/cancelled)
- [ ] Progress reporting API
- [ ] Per-job log storage + streaming
- [ ] WebSocket event for job progress
- [ ] `GET /api/v1/jobs` (list/history)
- [ ] Safe retry on failure

### State store & reconciliation
- [ ] SQLite migrations framework
- [ ] Desired-state tables: `settings`, `shares` (placeholder), `users`
- [ ] Reconciler interface invoked on boot
- [ ] Detect direct edits to managed config files, flag them
- [ ] `GET /api/v1/system/export` (config export)
- [ ] `POST /api/v1/system/import` (config import)

### Wrappers
- [ ] Exec helper with timeout + structured output
- [ ] `zpool`/`zfs` JSON parser helpers
- [ ] `smartctl` parser helpers
- [ ] Network/systemd helpers (stubs)

### vertex-cli
- [ ] `cmd/vertexctl` entrypoint
- [ ] HTTP client to local API
- [ ] `auth login` command
- [ ] `health` command

### Discovery
- [ ] mDNS advertisement of `vertex.local`
- [ ] Print console URL on boot

## Milestone 1 — Storage

### Drive detection
- [ ] List block devices via `lsblk`
- [ ] Resolve stable IDs via `/dev/disk/by-id`
- [ ] Capture model/size/serial
- [ ] SMART data via `smartctl`
- [ ] Drive health classification
- [ ] `GET /api/v1/drives`
- [ ] WebSocket drive status updates

### Pool management (advanced)
- [ ] Parse `zpool status`
- [ ] `GET /api/v1/pools`
- [ ] `GET /api/v1/pools/{name}`
- [ ] Create pool (manual vdev spec) endpoint → job
- [ ] Add vdev endpoint → job
- [ ] Trigger scrub endpoint → job
- [ ] Scrub schedule CRUD
- [ ] Destroy pool endpoint (typed confirm) → job
- [ ] Pool property get/set

### Simple-mode layout proposal
- [ ] Layout algorithm: 1→single (warn), 2→mirror, 3+→RAIDZ1, 5+→offer RAIDZ2
- [ ] Usable-space + fault-tolerance calculation
- [ ] Plain-language explanation strings
- [ ] `POST /api/v1/pools/propose`
- [ ] Create pool from proposal endpoint

### SMART monitoring
- [ ] Periodic SMART poll
- [ ] Alert generation on SMART failure
- [ ] Alert storage + `GET /api/v1/alerts`
- [ ] Alert dismissal

### Dashboard storage view
- [ ] Storage usage aggregation (pool capacity)
- [ ] `GET /api/v1/dashboard/storage`
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
