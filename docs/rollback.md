# OS Update Rollback

Vertex takes a ZFS snapshot of the root dataset before applying an OS update
(`rpool/ROOT@pre-update`). If an update breaks the system:

1. Reboot the machine.
2. In the bootloader menu, select the previous boot environment (the snapshot
   created before the update).
3. Boot into it; the system returns to the pre-update state.
4. From the recovered system, investigate and re-apply the update or roll
   forward as needed.

Channels: `stable` (default) and `testing`. Switch channels via the
`update_channel` setting.