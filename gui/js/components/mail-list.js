/**
 * Gas Town GUI - Mail List Component
 *
 * Renders the mail inbox with messages from the Gastown system.
 */

// Priority icons and colors
const PRIORITY_CONFIG = {
  high: { icon: 'priority_high', class: 'priority-high' },
  normal: { icon: 'mail', class: 'priority-normal' },
  low: { icon: 'mail_outline', class: 'priority-low' },
};

/**
 * Render the mail list
 * @param {HTMLElement} container - The mail list container
 * @param {Array} mail - Array of mail objects
 */
export function renderMailList(container, mail) {
  if (!container) return;

  if (!mail || mail.length === 0) {
    container.innerHTML = `
      <div class="empty-state">
        <span class="material-icons empty-icon">mail</span>
        <h3>No Mail</h3>
        <p>Your inbox is empty</p>
      </div>
    `;
    return;
  }

  // Sort by date (newest first), then by read status
  const sorted = [...mail].sort((a, b) => {
    // Unread first
    if (a.read !== b.read) return a.read ? 1 : -1;
    // Then by date
    return new Date(b.timestamp || 0) - new Date(a.timestamp || 0);
  });

  container.innerHTML = sorted.map((item, index) => renderMailItem(item, index)).join('');

  // Add click handlers
  container.querySelectorAll('.mail-item').forEach(item => {
    item.addEventListener('click', () => {
      const mailId = item.dataset.mailId;
      showMailDetail(mailId, mail.find(m => m.id === mailId));
    });
  });
}

/**
 * Render a single mail item
 */
function renderMailItem(mail, index) {
  const priority = mail.priority || 'normal';
  const priorityConfig = PRIORITY_CONFIG[priority] || PRIORITY_CONFIG.normal;
  const isUnread = !mail.read;

  return `
    <div class="mail-item ${isUnread ? 'unread' : ''} animate-spawn stagger-${Math.min(index, 6)}"
         data-mail-id="${mail.id}">
      <div class="mail-status">
        <span class="material-icons ${priorityConfig.class}">${priorityConfig.icon}</span>
      </div>

      <div class="mail-content">
        <div class="mail-header">
          <span class="mail-from">${escapeHtml(mail.from || 'System')}</span>
          <span class="mail-time">${formatTime(mail.timestamp)}</span>
        </div>
        <div class="mail-subject ${isUnread ? 'unread' : ''}">${escapeHtml(mail.subject || '(No Subject)')}</div>
        <div class="mail-preview">${escapeHtml(truncate(mail.message || mail.body || '', 80))}</div>

        ${mail.tags?.length ? `
          <div class="mail-tags">
            ${mail.tags.map(tag => `
              <span class="mail-tag">${escapeHtml(tag)}</span>
            `).join('')}
          </div>
        ` : ''}
      </div>

      <div class="mail-actions">
        <button class="btn btn-icon btn-sm" title="Archive" data-action="archive">
          <span class="material-icons">archive</span>
        </button>
        <button class="btn btn-icon btn-sm" title="Delete" data-action="delete">
          <span class="material-icons">delete</span>
        </button>
      </div>
    </div>
  `;
}

/**
 * Show mail detail modal
 */
function showMailDetail(mailId, mail) {
  if (!mail) return;

  // Mark as read
  const event = new CustomEvent('mail:read', { detail: { mailId } });
  document.dispatchEvent(event);

  // Show modal
  const modalEvent = new CustomEvent('mail:detail', {
    detail: { mailId, mail }
  });
  document.dispatchEvent(modalEvent);
}

/**
 * Format timestamp for display
 */
function formatTime(timestamp) {
  if (!timestamp) return '';

  const date = new Date(timestamp);
  const now = new Date();
  const diff = now - date;

  // Today - show time
  if (date.toDateString() === now.toDateString()) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  }

  // This week - show day
  if (diff < 7 * 86400000) {
    return date.toLocaleDateString([], { weekday: 'short' });
  }

  // Older - show date
  return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
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
