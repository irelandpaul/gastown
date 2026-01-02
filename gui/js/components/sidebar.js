/**
 * Gas Town GUI - Sidebar Component
 *
 * Renders the agent tree and quick status in the sidebar.
 */

// Agent status icons
const STATUS_ICONS = {
  idle: 'schedule',
  working: 'sync',
  waiting: 'hourglass_empty',
  error: 'error',
  complete: 'check_circle',
};

// Role colors (matches CSS variables)
const ROLE_CLASSES = {
  mayor: 'role-mayor',
  deacon: 'role-deacon',
  polecat: 'role-polecat',
  witness: 'role-witness',
  refinery: 'role-refinery',
};

/**
 * Render the sidebar with agent tree
 * @param {HTMLElement} container - The sidebar container
 * @param {Object} status - Current town status
 */
export function renderSidebar(container, status) {
  if (!container || !status) return;

  const agents = status.agents || [];
  const hook = status.hook;

  // Group agents by role
  const agentsByRole = groupByRole(agents);

  container.innerHTML = `
    <div class="sidebar-section">
      <h3 class="sidebar-title">
        <span class="material-icons">account_tree</span>
        Agents
      </h3>
      ${renderAgentTree(agentsByRole)}
    </div>

    ${hook ? renderHookSection(hook) : ''}

    <div class="sidebar-section">
      <h3 class="sidebar-title">
        <span class="material-icons">insights</span>
        Stats
      </h3>
      ${renderStats(status)}
    </div>
  `;
}

/**
 * Group agents by their role
 */
function groupByRole(agents) {
  const groups = {
    mayor: [],
    deacon: [],
    polecat: [],
    witness: [],
    refinery: [],
    other: [],
  };

  agents.forEach(agent => {
    const role = agent.role?.toLowerCase() || 'other';
    if (groups[role]) {
      groups[role].push(agent);
    } else {
      groups.other.push(agent);
    }
  });

  return groups;
}

/**
 * Render the hierarchical agent tree
 */
function renderAgentTree(agentsByRole) {
  const roles = ['mayor', 'deacon', 'polecat', 'witness', 'refinery', 'other'];

  let html = '<ul class="tree-view">';

  roles.forEach(role => {
    const agents = agentsByRole[role];
    if (!agents || agents.length === 0) return;

    html += `
      <li class="tree-node expandable expanded">
        <div class="tree-node-content">
          <span class="material-icons tree-icon">folder_open</span>
          <span class="tree-label ${ROLE_CLASSES[role] || ''}">${capitalize(role)}s</span>
          <span class="tree-badge">${agents.length}</span>
        </div>
        <ul class="tree-children">
          ${agents.map(agent => renderAgentNode(agent)).join('')}
        </ul>
      </li>
    `;
  });

  html += '</ul>';
  return html;
}

/**
 * Render a single agent node
 */
function renderAgentNode(agent) {
  const status = agent.status || 'idle';
  const icon = STATUS_ICONS[status] || 'help';
  const roleClass = ROLE_CLASSES[agent.role?.toLowerCase()] || '';

  return `
    <li class="tree-node">
      <div class="tree-node-content agent-node" data-agent-id="${agent.id || ''}">
        <span class="material-icons tree-icon status-${status}">${icon}</span>
        <span class="tree-label ${roleClass}">${agent.name || agent.id || 'Unknown'}</span>
        ${agent.current_task ? `<span class="tree-task">${truncate(agent.current_task, 20)}</span>` : ''}
      </div>
    </li>
  `;
}

/**
 * Render the hook section (currently hooked work)
 */
function renderHookSection(hook) {
  return `
    <div class="sidebar-section hook-section">
      <h3 class="sidebar-title">
        <span class="material-icons">anchor</span>
        Hook
      </h3>
      <div class="hook-card">
        <div class="hook-bead">${hook.bead_id || 'Unknown'}</div>
        <div class="hook-meta">
          <span class="hook-status status-${hook.status || 'idle'}">${hook.status || 'idle'}</span>
        </div>
        ${hook.title ? `<div class="hook-title">${truncate(hook.title, 40)}</div>` : ''}
      </div>
    </div>
  `;
}

/**
 * Render stats section
 */
function renderStats(status) {
  const stats = [
    { label: 'Convoys', value: status.convoy_count || 0, icon: 'local_shipping' },
    { label: 'Active', value: status.active_agents || 0, icon: 'person' },
    { label: 'Pending', value: status.pending_tasks || 0, icon: 'pending' },
  ];

  return `
    <div class="stats-grid">
      ${stats.map(stat => `
        <div class="stat-item">
          <span class="material-icons stat-icon">${stat.icon}</span>
          <div class="stat-content">
            <div class="stat-value">${stat.value}</div>
            <div class="stat-label">${stat.label}</div>
          </div>
        </div>
      `).join('')}
    </div>
  `;
}

// Utility functions
function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

function truncate(str, length) {
  if (!str) return '';
  return str.length > length ? str.slice(0, length) + '...' : str;
}

// Tree node toggle functionality
document.addEventListener('click', (e) => {
  const nodeContent = e.target.closest('.tree-node-content');
  if (!nodeContent) return;

  const node = nodeContent.closest('.tree-node.expandable');
  if (node) {
    node.classList.toggle('expanded');
    const icon = nodeContent.querySelector('.tree-icon');
    if (icon) {
      icon.textContent = node.classList.contains('expanded') ? 'folder_open' : 'folder';
    }
  }
});
