/**
 * Gas Town GUI - Convoy List Component
 *
 * Renders the list of convoys with status, progress, and actions.
 */

// Status icons for convoys
const STATUS_ICONS = {
  pending: 'hourglass_empty',
  running: 'sync',
  complete: 'check_circle',
  failed: 'error',
  cancelled: 'cancel',
};

// Priority colors
const PRIORITY_CLASSES = {
  high: 'priority-high',
  normal: 'priority-normal',
  low: 'priority-low',
};

/**
 * Render the convoy list
 * @param {HTMLElement} container - The list container
 * @param {Array} convoys - Array of convoy objects
 */
export function renderConvoyList(container, convoys) {
  if (!container) return;

  if (!convoys || convoys.length === 0) {
    container.innerHTML = `
      <div class="empty-state">
        <span class="material-icons empty-icon">local_shipping</span>
        <h3>No Convoys</h3>
        <p>Create a new convoy to start organizing work</p>
        <button class="btn btn-primary" id="empty-new-convoy">
          <span class="material-icons">add</span>
          New Convoy
        </button>
      </div>
    `;

    // Add event listener for empty state button
    const btn = container.querySelector('#empty-new-convoy');
    if (btn) {
      btn.addEventListener('click', () => {
        document.getElementById('new-convoy-btn')?.click();
      });
    }
    return;
  }

  container.innerHTML = convoys.map((convoy, index) => renderConvoyCard(convoy, index)).join('');

  // Add event listeners
  container.querySelectorAll('.convoy-card').forEach(card => {
    card.addEventListener('click', (e) => {
      if (!e.target.closest('button')) {
        const convoyId = card.dataset.convoyId;
        showConvoyDetail(convoyId);
      }
    });
  });
}

/**
 * Render a single convoy card
 */
function renderConvoyCard(convoy, index) {
  const status = convoy.status || 'pending';
  const statusIcon = STATUS_ICONS[status] || 'help';
  const priorityClass = PRIORITY_CLASSES[convoy.priority] || '';
  const progress = calculateProgress(convoy);

  return `
    <div class="convoy-card animate-spawn stagger-${Math.min(index, 6)}"
         data-convoy-id="${convoy.id}">
      <div class="convoy-header">
        <div class="convoy-status status-${status}">
          <span class="material-icons">${statusIcon}</span>
        </div>
        <div class="convoy-info">
          <h3 class="convoy-name">${escapeHtml(convoy.name || convoy.id)}</h3>
          <div class="convoy-meta">
            <span class="convoy-id">#${convoy.id?.slice(0, 8) || 'unknown'}</span>
            ${convoy.priority ? `<span class="convoy-priority ${priorityClass}">${convoy.priority}</span>` : ''}
          </div>
        </div>
        <div class="convoy-actions">
          <button class="btn btn-icon" title="View Details" data-action="view">
            <span class="material-icons">visibility</span>
          </button>
          <button class="btn btn-icon" title="More Actions" data-action="menu">
            <span class="material-icons">more_vert</span>
          </button>
        </div>
      </div>

      ${convoy.issues?.length ? renderIssueList(convoy.issues) : ''}

      <div class="convoy-progress">
        <div class="progress-bar">
          <div class="progress-fill" style="width: ${progress}%"></div>
        </div>
        <span class="progress-text">${progress}%</span>
      </div>

      <div class="convoy-footer">
        <div class="convoy-stats">
          ${renderConvoyStats(convoy)}
        </div>
        <div class="convoy-time">
          ${formatTime(convoy.created_at || convoy.timestamp)}
        </div>
      </div>
    </div>
  `;
}

/**
 * Render issue list within a convoy
 */
function renderIssueList(issues) {
  const maxVisible = 3;
  const visible = issues.slice(0, maxVisible);
  const remaining = issues.length - maxVisible;

  return `
    <div class="convoy-issues">
      ${visible.map(issue => `
        <div class="issue-chip" title="${escapeHtml(issue.title || issue)}">
          <span class="material-icons">assignment</span>
          ${escapeHtml(truncate(issue.title || issue, 25))}
        </div>
      `).join('')}
      ${remaining > 0 ? `
        <div class="issue-chip more">+${remaining} more</div>
      ` : ''}
    </div>
  `;
}

/**
 * Render convoy statistics
 */
function renderConvoyStats(convoy) {
  const stats = [];

  if (convoy.agent_count !== undefined) {
    stats.push(`<span title="Agents"><span class="material-icons">person</span>${convoy.agent_count}</span>`);
  }
  if (convoy.task_count !== undefined) {
    stats.push(`<span title="Tasks"><span class="material-icons">task</span>${convoy.task_count}</span>`);
  }
  if (convoy.bead_count !== undefined) {
    stats.push(`<span title="Beads"><span class="material-icons">bubble_chart</span>${convoy.bead_count}</span>`);
  }

  return stats.join('');
}

/**
 * Calculate progress percentage
 */
function calculateProgress(convoy) {
  if (convoy.progress !== undefined) {
    return Math.round(convoy.progress * 100);
  }
  if (convoy.completed && convoy.total) {
    return Math.round((convoy.completed / convoy.total) * 100);
  }
  if (convoy.status === 'complete') return 100;
  if (convoy.status === 'pending') return 0;
  return 50; // Default for running
}

/**
 * Show convoy detail modal
 */
function showConvoyDetail(convoyId) {
  const event = new CustomEvent('convoy:detail', { detail: { convoyId } });
  document.dispatchEvent(event);
}

// Utility functions
function escapeHtml(str) {
  if (!str) return '';
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}

function truncate(str, length) {
  if (!str) return '';
  return str.length > length ? str.slice(0, length) + '...' : str;
}

function formatTime(timestamp) {
  if (!timestamp) return '';
  const date = new Date(timestamp);
  const now = new Date();
  const diff = now - date;

  // Less than 1 minute
  if (diff < 60000) return 'Just now';
  // Less than 1 hour
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
  // Less than 24 hours
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;
  // Otherwise show date
  return date.toLocaleDateString();
}
