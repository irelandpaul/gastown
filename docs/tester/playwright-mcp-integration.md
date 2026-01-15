# Playwright MCP Integration Guide

This document describes how to set up and use the Playwright MCP server for AI User Testing in Gas Town.

## Overview

The Playwright MCP (Model Context Protocol) server provides browser automation capabilities to Claude agents. When testing applications, the tester agent uses Playwright MCP to:

- Navigate to URLs
- Click buttons and links
- Fill form fields
- Take screenshots
- Wait for page elements
- Record video and traces

## Prerequisites

1. **Node.js** (v18 or later)
2. **npm** (comes with Node.js)
3. **Claude Code** with MCP support

## Installation

### Option 1: Global Installation (Recommended for CLI)

```bash
# Install the official Microsoft Playwright MCP server
npm install -g @playwright/mcp

# Or use npx (no installation needed)
npx @playwright/mcp@latest --version
```

### Option 2: Project-Local Installation

```bash
# In your project directory
npm install @playwright/mcp

# Also install Playwright browsers
npx playwright install
```

## Configuration

### For Claude Code (CLI)

Add the Playwright MCP server to your Claude Code configuration:

```bash
# Add globally
claude mcp add playwright npx '@playwright/mcp@latest'

# Or add for a specific project
cd /path/to/project
claude mcp add --scope project playwright npx '@playwright/mcp@latest'
```

### Manual Configuration

Create or edit `~/.claude.json` (global) or `.mcp.json` (project):

```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp@latest"]
    }
  }
}
```

### With Recording Enabled

For video and trace recording, configure output directories:

```json
{
  "mcpServers": {
    "playwright": {
      "command": "npx",
      "args": ["@playwright/mcp@latest"],
      "env": {
        "PLAYWRIGHT_VIDEO_DIR": "./test-results/video",
        "PLAYWRIGHT_TRACES_DIR": "./test-results/trace",
        "PLAYWRIGHT_HEADLESS": "true"
      }
    }
  }
}
```

## Available Tools

Once configured, the Playwright MCP provides these tools to Claude:

### Navigation

| Tool | Description |
|------|-------------|
| `browser_navigate` | Navigate to a URL |
| `browser_navigate_back` | Go back in history |
| `browser_navigate_forward` | Go forward in history |
| `browser_reload` | Reload the current page |

### Actions

| Tool | Description |
|------|-------------|
| `browser_click` | Click an element |
| `browser_type` | Type text character by character |
| `browser_fill` | Fill a form field |
| `browser_press_key` | Press a keyboard key |
| `browser_select_option` | Select from a dropdown |
| `browser_check` | Check a checkbox |
| `browser_uncheck` | Uncheck a checkbox |
| `browser_hover` | Hover over an element |

### Observation

| Tool | Description |
|------|-------------|
| `browser_screenshot` | Capture the current page |
| `browser_get_text` | Get text from an element |
| `browser_get_html` | Get HTML content |
| `browser_get_url` | Get the current URL |
| `browser_get_title` | Get the page title |

### Waiting

| Tool | Description |
|------|-------------|
| `browser_wait_for_selector` | Wait for an element to appear |
| `browser_wait_for_navigation` | Wait for navigation to complete |
| `browser_wait_for_timeout` | Wait for a duration |
| `browser_wait_for_load_state` | Wait for network idle / DOM ready |

## Usage in Tester Agent

The tester agent uses Playwright MCP automatically when running scenarios. Here's how it works:

1. **Agent Spawn**: When `gt tester run` executes, it:
   - Creates an MCP configuration for the test run
   - Sets up recording directories
   - Spawns a Claude agent with the tester CLAUDE.md

2. **Browser Control**: The agent uses natural language to describe actions:
   ```
   I'll navigate to the signup page.
   [Uses browser_navigate to go to /signup]

   Now I'll click the "Sign Up" button.
   [Uses browser_click with text "Sign Up"]
   ```

3. **Screenshots**: When the agent notices confusion:
   ```
   I can't find the submit button... taking a screenshot.
   [Uses browser_screenshot]

   [OBSERVATION]
   Type: confusion
   Severity: P2
   Confidence: high
   Description: Submit button not visible on form
   Screenshot: confusion-submit-hidden.png
   [/OBSERVATION]
   ```

4. **Recording**: Video and traces are recorded automatically throughout the session.

## Headless vs Headed Mode

By default, tests run in headless mode (no visible browser window). To debug:

```bash
# Run with visible browser
gt tester run scenario.yaml --headed

# Or set environment variable
PLAYWRIGHT_HEADLESS=false gt tester run scenario.yaml
```

## Troubleshooting

### Browser not starting

```bash
# Ensure Playwright browsers are installed
npx playwright install chromium

# Verify MCP server works
npx @playwright/mcp@latest --version
```

### MCP server not found

```bash
# Check Claude Code MCP configuration
claude mcp list

# Re-add if missing
claude mcp add playwright npx '@playwright/mcp@latest'
```

### Video recording not working

1. Ensure the output directory exists and is writable
2. Check `PLAYWRIGHT_VIDEO_DIR` environment variable is set
3. Verify disk space is available

### Timeout errors

Increase the default timeout in your scenario:

```yaml
timeout: 900  # 15 minutes (in seconds)

wait_strategies:
  network_idle: true
  min_load_time: 2000  # Wait 2 seconds after navigation
```

## Best Practices

1. **Use text-based selectors**: Click "Sign Up" rather than CSS selectors
2. **Wait for network idle**: Pages often make API calls after appearing ready
3. **Take screenshots liberally**: They help debug issues later
4. **Keep scenarios focused**: One goal per scenario
5. **Use appropriate timeouts**: Some flows take longer than others

## Related Documentation

- [AI User Testing Proposal](../proposals/ai-user-testing-proposal.spec.md)
- [Tester CLAUDE.md Spec](../specs/tester-claude-md.spec.md)
- [Tester Commands Spec](../specs/tester-commands.spec.md)
- [Scenario Format Spec](../specs/tester-scenario-format.spec.md)

## References

- [Microsoft Playwright MCP](https://github.com/microsoft/playwright-mcp)
- [Playwright Documentation](https://playwright.dev/)
- [Model Context Protocol](https://modelcontextprotocol.io/)
