# Running vertexd as root

`vertexd` runs as root because it must manage ZFS, Samba, users and systemd.
To limit blast radius:

- All system-touching code lives under `internal/storage`, `internal/shares`,
  `internal/users` and `internal/system`. HTTP/auth/job/state code never shells
  out directly; it goes through `internal/wrappers`.
- Inputs from the API are validated before reaching wrappers.
- The daemon binds only to the configured listen address and serves TLS.
- Destructive operations require typed confirmation and become jobs.

Future work (M4): drop capabilities, run helpers via polkit, seccomp profile.
