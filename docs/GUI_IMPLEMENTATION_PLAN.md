# Gas Town GUI Implementation Plan

## Overview

Modern web-based GUI for Gas Town multi-agent orchestrator. Replaces/complements existing Bubbletea TUI with a rich browser experience.

## Technology Stack

### Frontend
- **Framework**: Vanilla JavaScript with Web Components (lightweight, no build step)
- **Styling**: CSS with custom properties for theming
- **Animations**: CSS transitions + GSAP for complex sequences
- **Icons**: Material Icons + custom SVG
- **Charts**: D3.js for progress/dependency visualization

### Backend Integration
- **CLI Bridge**: Node.js server executing `gt` commands via child_process
- **Real-time**: WebSocket for event streaming (`bd activity --follow`)
- **Data**: JSON output from all `gt` commands

### Testing
- **E2E**: Puppeteer for browser automation
- **Unit**: Vitest for JavaScript logic
- **Visual**: Percy for screenshot comparison

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   Browser UI                     │
│  ┌─────────┐ ┌─────────┐ ┌─────────────────┐   │
│  │ Sidebar │ │ Convoy  │ │ Activity Feed   │   │
│  │ (Tree)  │ │ Panel   │ │ (Real-time)     │   │
│  └─────────┘ └─────────┘ └─────────────────┘   │
└─────────────────────────────────────────────────┘
                        │
                        │ WebSocket / HTTP
                        ▼
┌─────────────────────────────────────────────────┐
│              Node.js Bridge Server               │
│  ┌───────────────────────────────────────────┐  │
│  │ CLI Executor (child_process.spawn)         │  │
│  │ - gt convoy list --json                    │  │
│  │ - gt status --json                         │  │
│  │ - bd activity --follow                     │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
                        │
                        │ Shell
                        ▼
┌─────────────────────────────────────────────────┐
│              Gas Town CLI (gt)                   │
│              Beads CLI (bd)                      │
└─────────────────────────────────────────────────┘
```

## File Structure

```
gui/
├── package.json
├── server.js                 # Node bridge server
├── index.html                # Main entry
├── css/
│   ├── reset.css
│   ├── variables.css         # CSS custom properties
│   ├── layout.css            # Grid/flexbox structure
│   ├── components.css        # Component styles
│   └── animations.css        # Keyframes & transitions
├── js/
│   ├── app.js                # Main application
│   ├── api.js                # Server communication
│   ├── state.js              # State management
│   ├── components/
│   │   ├── sidebar.js        # Agent tree
│   │   ├── convoy-list.js    # Convoy dashboard
│   │   ├── convoy-detail.js  # Convoy detail view
│   │   ├── agent-card.js     # Agent status card
│   │   ├── activity-feed.js  # Real-time events
│   │   ├── mail-inbox.js     # Mail interface
│   │   ├── sling-modal.js    # Work dispatch modal
│   │   └── status-bar.js     # Bottom status
│   └── utils/
│       ├── formatting.js     # Time, numbers
│       └── animations.js     # GSAP helpers
├── tests/
│   ├── e2e/
│   │   ├── convoy.test.js
│   │   ├── sling.test.js
│   │   └── mail.test.js
│   └── unit/
│       └── state.test.js
└── assets/
    └── icons/
```

## Implementation Phases

### Phase 1: Foundation (Current)
- [x] Repository setup
- [x] Codebase analysis
- [ ] Node bridge server
- [ ] Basic HTML shell
- [ ] CSS framework with animations
- [ ] WebSocket connection

### Phase 2: Core Dashboard
- [ ] Sidebar with agent tree
- [ ] Convoy list (main view)
- [ ] Status bar
- [ ] Real-time event stream

### Phase 3: Convoy Management
- [ ] Convoy detail view
- [ ] Issue tree with status
- [ ] Progress visualization
- [ ] Worker assignment panel

### Phase 4: Work Dispatch
- [ ] Sling modal
- [ ] Issue/formula search
- [ ] Target selection
- [ ] Confirmation & result

### Phase 5: Communication
- [ ] Mail inbox
- [ ] Compose modal
- [ ] Nudge interface
- [ ] Escalation form

### Phase 6: Polish
- [ ] Animations for state transitions
- [ ] Keyboard shortcuts
- [ ] Themes (dark/light)
- [ ] Performance optimization

### Phase 7: Testing
- [ ] Puppeteer E2E tests
- [ ] Unit tests
- [ ] Visual regression tests

## Key Components

### 1. Convoy Dashboard
The primary view showing all active convoys.

**Features:**
- Sortable/filterable table
- Progress bars with animation
- Expandable issue tree
- Quick actions (sling, close, add issues)
- Real-time status updates

**Data Source:** `gt convoy list --json`

### 2. Agent Tree (Sidebar)
Hierarchical view of all agents.

**Structure:**
```
Town
├── Infrastructure
│   ├── Mayor [running]
│   ├── Deacon [running]
├── Rigs
│   ├── greenplace
│   │   ├── Witness [running]
│   │   ├── Refinery [idle]
│   │   ├── Polecats
│   │   │   ├── Toast [working] → gp-123
│   │   │   └── Nux [done]
│   │   └── Crew
│   │       └── dave [idle]
```

**Data Source:** `gt status --json`

### 3. Activity Feed
Real-time event stream.

**Event Types:**
- ✓ Task complete (green)
- → In progress (yellow)
- ✗ Failed (red)
- 📧 Mail (blue)
- 🚨 Escalation (orange)

**Data Source:** WebSocket from `bd activity --follow`

### 4. Sling Modal
Work dispatch interface.

**Fields:**
- Issue/formula search (autocomplete)
- Target selector (rig/agent)
- Quality level dropdown
- Molecule template picker
- Args textarea

**Data Source:** `gt sling <issue> <target> --json`

## Animation Specifications

### State Transitions
| Element | Trigger | Animation | Duration |
|---------|---------|-----------|----------|
| Convoy row | Status change | Background pulse | 400ms |
| Progress bar | Value change | Width slide | 300ms |
| Agent status | Online/offline | Fade + scale | 250ms |
| Issue status | Done | Checkmark draw | 500ms |
| Modal | Open/close | Fade + slide up | 300ms |

### Real-time Updates
| Event | Animation | Notes |
|-------|-----------|-------|
| New event | Slide in from right | Stagger if multiple |
| Convoy update | Highlight row | Yellow flash 200ms |
| Agent spawn | Grow in | From 0 to 1 scale |
| Agent done | Shrink out | To 0 scale, then remove |

### CSS Keyframes
```css
@keyframes convoy-update {
  0% { background-color: var(--bg-default); }
  50% { background-color: var(--yellow-highlight); }
  100% { background-color: var(--bg-default); }
}

@keyframes agent-spawn {
  from { transform: scale(0); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

@keyframes event-slide-in {
  from { transform: translateX(100%); opacity: 0; }
  to { transform: translateX(0); opacity: 1; }
}
```

## API Endpoints (Node Bridge)

### REST
```
GET  /api/status           # Town overview
GET  /api/convoys          # List convoys
GET  /api/convoy/:id       # Convoy detail
POST /api/convoy           # Create convoy
POST /api/sling            # Dispatch work
GET  /api/mail             # Inbox
POST /api/mail             # Send message
GET  /api/agents           # Agent list
POST /api/nudge            # Nudge agent
```

### WebSocket
```
Connection: ws://localhost:3000/ws

Events:
- activity: { type, actor, target, message, timestamp }
- convoy_update: { id, status, progress }
- agent_status: { name, status, hook }
```

## Testing Strategy

### Puppeteer E2E Tests

```javascript
// tests/e2e/convoy.test.js
describe('Convoy Dashboard', () => {
  it('displays active convoys', async () => {
    await page.goto('http://localhost:3000');
    await page.waitForSelector('.convoy-list');
    const rows = await page.$$('.convoy-row');
    expect(rows.length).toBeGreaterThan(0);
  });

  it('expands convoy to show issues', async () => {
    await page.click('.convoy-row:first-child .expand-btn');
    await page.waitForSelector('.issue-tree');
    const issues = await page.$$('.issue-item');
    expect(issues.length).toBeGreaterThan(0);
  });

  it('creates new convoy via modal', async () => {
    await page.click('[data-action="new-convoy"]');
    await page.waitForSelector('.modal-convoy');
    await page.type('#convoy-name', 'Test Convoy');
    await page.type('#convoy-issues', 'gt-123');
    await page.click('.btn-create');
    await page.waitForSelector('.toast-success');
  });
});
```

### Mock Server for Testing

```javascript
// tests/mocks/server.js
const mockConvoys = [
  { id: 'hq-cv-test', name: 'Test Convoy', status: 'active', progress: { done: 2, total: 5 } }
];

app.get('/api/convoys', (req, res) => {
  res.json(mockConvoys);
});
```

## Theming

### CSS Variables

```css
:root {
  /* Dark theme (default) */
  --bg-primary: #1a1a2e;
  --bg-secondary: #16213e;
  --bg-tertiary: #0f3460;
  --text-primary: #eaeaea;
  --text-secondary: #b8b8b8;

  /* Status colors */
  --status-running: #10b981;
  --status-working: #f59e0b;
  --status-done: #6366f1;
  --status-stuck: #ef4444;
  --status-idle: #6b7280;

  /* Accent */
  --accent-primary: #6366f1;
  --accent-secondary: #8b5cf6;

  /* Spacing */
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;
}

[data-theme="light"] {
  --bg-primary: #f8fafc;
  --bg-secondary: #e2e8f0;
  --text-primary: #1e293b;
  /* ... */
}
```

## Commit Cadence

- Commit after each component completion
- Push at end of each phase
- Create feature branch for major changes
- PR for merge to work1

## Success Metrics

1. All CLI commands executable from GUI
2. Real-time updates < 500ms latency
3. Animations smooth at 60fps
4. Puppeteer tests pass on all core flows
5. Works on Chrome, Firefox, Safari
6. Responsive down to 1024px width

## Notes

- Steve Yegge's style: Make it fun, make it work, make it fast
- Gas Town metaphor: Industrial, gritty, efficient
- Animations should feel mechanical/precise (not bouncy)
- Error states should be clear and actionable
