# Vertex OS — Design Doc

*Draft v0.1 · November 2026 · Status: design*

## 1. Overview

Vertex OS is a simplified home NAS operating system. It installs from its own ISO onto a Debian base with OpenZFS, and is managed through a local web dashboard. A first-time user should go from blank machine to working network shares in minutes without learning ZFS. A power user should be able to open an advanced mode and see everything.

It is inspired by HexOS: a friendly layer over proven storage technology. Vertex differs in being fully local-first, open source and account-free.

## 2. Goals and non-goals

**Goals**

- Guided install and first-run wizard that ends with a healthy pool and a first share.
- Simple mode that hides ZFS terminology; advanced mode that exposes pools, vdevs and datasets.
- Multi-user with roles (admin, user, guest).
- A dashboard that answers three questions at a glance: is the system healthy, is storage filling up, is anything misbehaving.
- Fully functional with no internet connection and no cloud account.

**Non-goals for v1**

- App catalog and container management.
- Remote access / VPN.
- Backups, replication, virtual machines.
- ARM64 support.
- A custom filesystem. Storage is OpenZFS.

## 3. Decisions so far

| Area | Decision |
| --- | --- |
| Base | Debian + OpenZFS, shipped as its own distro/ISO |
| Backend | Go daemon |
| Web UI | SPA in React and TypeScript |
| Target users | Beginners and enthusiasts, with an advanced mode |
| Storage UX | Simple mode hides ZFS; advanced mode exposes it |
| Users | Multi-user with roles: admin, user, guest |
| Updates | Apps auto-update; OS updates only on approval |
| Remote access | Built-in WireGuard/Tailscale-style VPN (post-v1) |
| v1 scope | Core only: install, storage, shares, users, dashboard |
| Hardware | x86\_64 only |
| App packaging | **Open question** (see section 12) |

## 4. Architecture

```
+-------------------------------+
|  Browser: React/TS SPA        |
+---------------+---------------+
                | HTTPS (REST + WebSocket)
+---------------v---------------+
|  vertexd (Go)                 |
|  API · auth · job queue       |
|  storage · shares · users     |
|  health · updates             |
+---+---------+---------+-------+
    |         |         |
  ZFS CLI   Samba/NFS  systemd, apt, smartctl, networkd
```

The design principle is that `vertexd` is the single source of truth. The UI never touches the system directly, and the OS is configured only through the daemon. This keeps the system predictable and makes an advanced mode possible without the two modes fighting each other.

### Components

- **vertexd:** the long-running Go service, run under systemd as root with a tightly scoped internal structure.
- **vertex-ui:** a static React/TypeScript bundle served by `vertexd` itself, so there is no separate web server to manage.
- **vertex-cli:** a small CLI that talks to the same local API. Useful for debugging, scripting and recovery.
- **vertex-installer:** the installer image flow, covered in section 5.

### Daemon design

- **Job queue:** every state-changing action (create pool, add share, apply update) becomes a job with an ID, status, progress and logs. The UI shows jobs live. This also gives auditability and safe retry.
- **Reconciliation:** the daemon stores desired state (shares, users, settings) in a local database and reconciles the system to it on boot and on change. Direct edits to managed config files are detected and flagged.
- **Wrapping tools, not reimplementing them:** use `zpool`, `zfs`, `smartctl`, Samba and systemd rather than rewriting them. Parse structured output (`-j` / `-p` / `-H` flags) wherever the tools provide it.
- **State store:** SQLite for desired state, job history and audit log, kept on the boot drive and included in config export.

## 5. Base OS and installer

- **Base:** current Debian stable, with a minimal package set, OpenZFS from the distro packages, Samba, NFS server, smartmontools and nftables.
- **Image build:** a reproducible build pipeline (for example live-build or mkosi) that produces the installer ISO in CI.
- **Boot drive:** the installer targets a dedicated boot device and keeps it separate from the data pool. Consider ZFS on root so that boot environments can be used for update rollback.
- **Installer flow:** boot ISO, choose the boot disk, set hostname and the first admin account, install, reboot. After reboot the machine prints its URL (for example `http://vertex.local`) on the console.
- **Discovery:** advertise through mDNS so `vertex.local` works on most home networks.

## 6. Storage model

OpenZFS is the only supported storage backend in v1.

### Simple mode

The user sees drives, capacity and health, not vdevs. When drives are added, Vertex proposes a layout automatically:

- 1 drive: single, with a clear warning that there is no redundancy.
- 2 drives: mirror.
- 3 or more drives of similar size: RAIDZ1 by default, with RAIDZ2 offered when there are 5 or more drives.

The wizard explains the trade-off in plain language: how much usable space, and how many drive failures the pool can survive.

### Advanced mode

Exposes pools, vdevs, datasets, properties (compression, recordsize, quotas), snapshots and scrub schedules. Actions still go through `vertexd`, so simple mode stays coherent with whatever the advanced user does.

### Defaults

- Compression on (lz4 or zstd).
- Scheduled scrubs.
- SMART monitoring with alerts.
- Each share is its own dataset, so quotas and snapshots work per share.

### Safety rails

- Destructive actions require typed confirmation.
- Drive identification is by stable ID (`/dev/disk/by-id`), never by `/dev/sdX`.
- Pool operations that cannot be undone show exactly what will happen first.

## 7. Shares and file services

- **SMB** is the primary protocol, for Windows, macOS and Linux clients.
- **NFS** is available as an advanced option.
- Each share has a name, a dataset, a quota and an access list based on users and roles.
- Time Machine support can be added later as a share preset.
- Guest-accessible shares are explicit and clearly labeled in the UI.

## 8. Users, roles and authentication

| Role | Capabilities |
| --- | --- |
| Admin | Full system control: storage, users, updates, settings |
| User | Access to shares they are granted; manage own password |
| Guest | Access only to shares explicitly marked as guest-accessible |

- Web accounts map to system accounts so that SMB permissions are enforced by the OS.
- Passwords are hashed with a modern algorithm (argon2id).
- Sessions use short-lived tokens, with secure cookies and CSRF protection.
- Rate limiting and lockout apply to login attempts.
- Two-factor authentication and SSO are good follow-ups, but are not v1.

## 9. Web UI

A React and TypeScript single-page app.

**Primary screens**

1. **Setup wizard:** welcome, name the server, create admin, detect drives, propose a pool layout, create the first share.
2. **Dashboard:** system health, storage usage, CPU/RAM/network, drive status, recent jobs and alerts.
3. **Storage:** drives and pools; simple view first, with a toggle for advanced.
4. **Shares:** list, create, permissions.
5. **Users:** accounts and roles.
6. **System:** updates, network, time, logs, config export.

**Principles**

- Mobile-responsive from day one.
- Plain language first; technical names appear in advanced mode or in tooltips.
- Live updates over WebSocket for jobs and health.
- The UI works entirely on the LAN without touching the internet.
- Types for the API are generated from a single schema (OpenAPI), so the Go server and the TypeScript client cannot drift apart.

## 10. API

- REST over HTTPS for resources (`/api/v1/pools`, `/shares`, `/users`, `/jobs`, `/system`).
- WebSocket for live events (job progress, health changes).
- OpenAPI schema as the contract, used to generate the TypeScript client and server stubs.
- The same API serves the UI, the CLI and, later, any third-party integration.
- Versioned from the start.

## 11. Updates

- **Apps:** auto-update (relevant once apps exist after v1).
- **OS:** updates are detected and surfaced, but only applied after an admin approves them.
- **Mechanism:** apt-based updates from a Vertex-managed repository, with a ZFS boot environment snapshot taken before each OS update so a failed update can be rolled back from the boot menu.
- **Update visibility:** while an update runs, the UI shows clear status and explains which parts of the system are temporarily unavailable.
- **Channels:** stable and testing.

## 12. Open questions

1. **App packaging and installation (deliberately deferred).** Options include Docker Compose with a curated catalog, a Vertex-specific app manifest layered on Compose, and rootless Podman with Quadlets. This decision is not needed for v1 but will shape the post-v1 architecture, so the daemon should keep a clean seam where a container runtime can be added later.
2. ZFS on root versus a simpler boot filesystem with ZFS for data only.
3. Mixed-size drive support: ZFS handles it poorly, so decide whether to document the limitation or add an alternative later.
4. Hardware compatibility: where to publish a tested-hardware list.
5. Product name check: confirm there is no naming or trademark conflict for "Vertex".
6. Licensing: pick an open-source license for the daemon, UI and image build.

## 13. Roadmap

**Milestone 0, Foundations:** repo layout, image build pipeline producing a bootable Debian + ZFS ISO, `vertexd` skeleton with auth and the job queue.

**Milestone 1, Storage:** drive detection, pool creation in simple mode, SMART health, scrubs, dashboard storage view.

**Milestone 2, Shares and users:** SMB shares as datasets, users and roles, permissions, share UI.

**Milestone 3, Installer and first-run:** polished installer, setup wizard, mDNS discovery, config export and import.

**Milestone 4, Hardening and v1:** advanced storage mode, OS update flow with boot-environment rollback, logging and alerts, documentation, tested-hardware list.

**After v1:** app system (resolve open question 1), built-in VPN remote access, backups and replication, ARM64.

## 14. Risks

- **Data loss from storage mistakes:** mitigated with confirmations, stable drive IDs and conservative defaults.
- **Scope creep:** v1 stays strictly at core features.
- **Update breakage:** mitigated with boot environment snapshots and approval-gated OS updates.
- **Security surface:** a root daemon exposed on a LAN needs careful authentication, input validation and minimal privileges from the start.
- **Maintenance burden of a distro:** keeping the package set small and tracking Debian stable limits this.
