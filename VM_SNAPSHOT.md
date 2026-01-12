# Windows VM Snapshot Documentation (sctest2)

This document describes the clean state of the Windows VM (sctest2) configured for MCP automation.

## VM Information
- **IP Address**: `192.168.122.90`
- **User**: `ScreenCoach`
- **Auth**: SSH with key authentication
- **Operating System**: Windows

## Setup Details
1. **Windows-MCP**: Installed at `C:\Windows-MCP-main`.
2. **Python Environment**: Managed by `uv`.
   - Python is installed.
   - Dependencies are synced using `uv sync`.
3. **MCP Server**:
   - Transport: `stdio`
   - Command: `uv run windows-mcp`
   - Tested successfully via SSH from the Linux host.

## Verification
The server can be verified by running the following command from the Linux host:
```bash
ssh ScreenCoach@192.168.122.90 "cd /d C:\Windows-MCP-main && uv run windows-mcp --help"
```

## Recommended Snapshot State
- No active applications running.
- `C:\Windows-MCP-main` is clean and `uv sync` has been run.
- SSH server is active and reachable.
