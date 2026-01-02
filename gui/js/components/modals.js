/**
 * Gas Town GUI - Modals Component
 *
 * Handles all modal dialogs in the application.
 */

import { api } from '../api.js';
import { showToast } from './toast.js';

// Modal registry
const modals = new Map();

// References
let overlay = null;

/**
 * Initialize modals system
 */
export function initModals() {
  overlay = document.getElementById('modal-overlay');

  // Register built-in modals
  registerModal('new-convoy', {
    element: document.getElementById('new-convoy-modal'),
    onOpen: initNewConvoyModal,
    onSubmit: handleNewConvoySubmit,
  });

  registerModal('sling', {
    element: document.getElementById('sling-modal'),
    onOpen: initSlingModal,
    onSubmit: handleSlingSubmit,
  });

  registerModal('mail-compose', {
    element: document.getElementById('mail-compose-modal'),
    onOpen: initMailComposeModal,
    onSubmit: handleMailComposeSubmit,
  });

  // Close on overlay click
  overlay?.addEventListener('click', (e) => {
    if (e.target === overlay) {
      closeAllModals();
    }
  });

  // Close buttons
  document.querySelectorAll('[data-modal-close]').forEach(btn => {
    btn.addEventListener('click', closeAllModals);
  });

  // Open buttons
  document.querySelectorAll('[data-modal-open]').forEach(btn => {
    btn.addEventListener('click', () => {
      const modalId = btn.dataset.modalOpen;
      openModal(modalId);
    });
  });

  // Form submissions
  document.querySelectorAll('.modal form').forEach(form => {
    form.addEventListener('submit', async (e) => {
      e.preventDefault();
      const modal = form.closest('.modal');
      if (!modal) return;

      const modalId = modal.id.replace('-modal', '');
      const config = modals.get(modalId);
      if (config?.onSubmit) {
        await config.onSubmit(form);
      }
    });
  });

  // Listen for custom modal events
  document.addEventListener('convoy:detail', (e) => {
    showConvoyDetailModal(e.detail.convoyId);
  });

  document.addEventListener('agent:detail', (e) => {
    showAgentDetailModal(e.detail.agentId);
  });

  document.addEventListener('agent:nudge', (e) => {
    showNudgeModal(e.detail.agentId);
  });

  document.addEventListener('mail:detail', (e) => {
    showMailDetailModal(e.detail.mailId, e.detail.mail);
  });
}

/**
 * Register a modal
 */
export function registerModal(id, config) {
  modals.set(id, config);
}

/**
 * Open a modal by ID
 */
export function openModal(modalId, data = {}) {
  const config = modals.get(modalId);
  if (!config?.element) {
    console.warn(`Modal not found: ${modalId}`);
    return;
  }

  // Hide all modals first
  document.querySelectorAll('.modal').forEach(m => m.classList.add('hidden'));

  // Show overlay and modal
  overlay?.classList.remove('hidden');
  config.element.classList.remove('hidden');

  // Call onOpen callback
  if (config.onOpen) {
    config.onOpen(config.element, data);
  }

  // Focus first input
  const firstInput = config.element.querySelector('input, textarea, select');
  if (firstInput) {
    setTimeout(() => firstInput.focus(), 100);
  }
}

/**
 * Close all modals
 */
export function closeAllModals() {
  overlay?.classList.add('hidden');
  document.querySelectorAll('.modal').forEach(m => m.classList.add('hidden'));

  // Reset forms
  document.querySelectorAll('.modal form').forEach(form => form.reset());
}

/**
 * Close specific modal
 */
export function closeModal(modalId) {
  const config = modals.get(modalId);
  if (config?.element) {
    config.element.classList.add('hidden');
  }

  // Check if any modal is still open
  const openModals = document.querySelectorAll('.modal:not(.hidden)');
  if (openModals.length === 0) {
    overlay?.classList.add('hidden');
  }
}

// === New Convoy Modal ===

function initNewConvoyModal(element, data) {
  // Clear any previous state
  const form = element.querySelector('form');
  if (form) form.reset();
}

async function handleNewConvoySubmit(form) {
  const name = form.querySelector('[name="name"]')?.value;
  const issuesText = form.querySelector('[name="issues"]')?.value || '';
  const notify = form.querySelector('[name="notify"]')?.value || null;

  if (!name) {
    showToast('Please enter a convoy name', 'warning');
    return;
  }

  // Parse issues (comma or newline separated)
  const issues = issuesText
    .split(/[,\n]/)
    .map(s => s.trim())
    .filter(Boolean);

  try {
    const result = await api.createConvoy(name, issues, notify);
    showToast(`Convoy "${name}" created`, 'success');
    closeAllModals();

    // Dispatch event for refresh
    document.dispatchEvent(new CustomEvent('convoy:created', { detail: result }));
  } catch (err) {
    showToast(`Failed to create convoy: ${err.message}`, 'error');
  }
}

// === Sling Modal ===

function initSlingModal(element, data) {
  // Pre-fill if data provided
  if (data.bead) {
    const beadInput = element.querySelector('[name="bead"]');
    if (beadInput) beadInput.value = data.bead;
  }
  if (data.target) {
    const targetInput = element.querySelector('[name="target"]');
    if (targetInput) targetInput.value = data.target;
  }
}

async function handleSlingSubmit(form) {
  const bead = form.querySelector('[name="bead"]')?.value;
  const target = form.querySelector('[name="target"]')?.value;
  const molecule = form.querySelector('[name="molecule"]')?.value || undefined;
  const quality = form.querySelector('[name="quality"]')?.value || undefined;

  if (!bead || !target) {
    showToast('Please enter both bead and target', 'warning');
    return;
  }

  try {
    const result = await api.sling(bead, target, { molecule, quality });
    showToast(`Work slung: ${bead} → ${target}`, 'success');
    closeAllModals();

    // Dispatch event
    document.dispatchEvent(new CustomEvent('work:slung', { detail: result }));
  } catch (err) {
    showToast(`Failed to sling work: ${err.message}`, 'error');
  }
}

// === Mail Compose Modal ===

function initMailComposeModal(element, data) {
  // Pre-fill if replying
  if (data.replyTo) {
    const toInput = element.querySelector('[name="to"]');
    if (toInput) toInput.value = data.replyTo;
  }
  if (data.subject) {
    const subjectInput = element.querySelector('[name="subject"]');
    if (subjectInput) subjectInput.value = `Re: ${data.subject}`;
  }
}

async function handleMailComposeSubmit(form) {
  const to = form.querySelector('[name="to"]')?.value;
  const subject = form.querySelector('[name="subject"]')?.value;
  const message = form.querySelector('[name="message"]')?.value;
  const priority = form.querySelector('[name="priority"]')?.value || 'normal';

  if (!to || !subject || !message) {
    showToast('Please fill in all fields', 'warning');
    return;
  }

  try {
    await api.sendMail(to, subject, message, priority);
    showToast('Mail sent', 'success');
    closeAllModals();
  } catch (err) {
    showToast(`Failed to send mail: ${err.message}`, 'error');
  }
}

// === Dynamic Modals ===

async function showConvoyDetailModal(convoyId) {
  try {
    const convoy = await api.getConvoy(convoyId);
    const content = `
      <div class="modal-header">
        <h2>Convoy: ${escapeHtml(convoy.name || convoy.id)}</h2>
        <button class="btn btn-icon" data-modal-close>
          <span class="material-icons">close</span>
        </button>
      </div>
      <div class="modal-body">
        <div class="detail-grid">
          <div class="detail-item">
            <label>ID</label>
            <span>${convoyId}</span>
          </div>
          <div class="detail-item">
            <label>Status</label>
            <span class="status-badge status-${convoy.status || 'pending'}">${convoy.status || 'pending'}</span>
          </div>
          <div class="detail-item">
            <label>Created</label>
            <span>${new Date(convoy.created_at).toLocaleString()}</span>
          </div>
          ${convoy.issues?.length ? `
            <div class="detail-item full-width">
              <label>Issues</label>
              <ul class="issue-list">
                ${convoy.issues.map(i => `<li>${escapeHtml(typeof i === 'string' ? i : i.title)}</li>`).join('')}
              </ul>
            </div>
          ` : ''}
        </div>
      </div>
    `;
    showDynamicModal('convoy-detail', content);
  } catch (err) {
    showToast(`Failed to load convoy: ${err.message}`, 'error');
  }
}

async function showAgentDetailModal(agentId) {
  // For now show a simple modal - can be expanded later
  const content = `
    <div class="modal-header">
      <h2>Agent Details</h2>
      <button class="btn btn-icon" data-modal-close>
        <span class="material-icons">close</span>
      </button>
    </div>
    <div class="modal-body">
      <p>Agent ID: <code>${agentId}</code></p>
      <p>Detailed agent view coming soon...</p>
    </div>
  `;
  showDynamicModal('agent-detail', content);
}

function showNudgeModal(agentId) {
  const content = `
    <div class="modal-header">
      <h2>Nudge Agent</h2>
      <button class="btn btn-icon" data-modal-close>
        <span class="material-icons">close</span>
      </button>
    </div>
    <div class="modal-body">
      <form id="nudge-form">
        <input type="hidden" name="agent_id" value="${agentId}">
        <div class="form-group">
          <label for="nudge-message">Message</label>
          <textarea id="nudge-message" name="message" rows="3" placeholder="Enter a message to send to the agent..."></textarea>
        </div>
        <div class="form-actions">
          <button type="button" class="btn btn-secondary" data-modal-close>Cancel</button>
          <button type="submit" class="btn btn-primary">
            <span class="material-icons">send</span>
            Send Nudge
          </button>
        </div>
      </form>
    </div>
  `;

  const modal = showDynamicModal('nudge', content);

  // Handle form submission
  const form = modal.querySelector('#nudge-form');
  form?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const message = form.querySelector('[name="message"]')?.value;

    try {
      await api.nudge(agentId, message);
      showToast('Nudge sent', 'success');
      closeAllModals();
    } catch (err) {
      showToast(`Failed to nudge agent: ${err.message}`, 'error');
    }
  });
}

function showMailDetailModal(mailId, mail) {
  const content = `
    <div class="modal-header">
      <h2>${escapeHtml(mail.subject || '(No Subject)')}</h2>
      <button class="btn btn-icon" data-modal-close>
        <span class="material-icons">close</span>
      </button>
    </div>
    <div class="modal-body">
      <div class="mail-detail-meta">
        <div><strong>From:</strong> ${escapeHtml(mail.from || 'System')}</div>
        <div><strong>Date:</strong> ${new Date(mail.timestamp).toLocaleString()}</div>
        ${mail.priority && mail.priority !== 'normal' ? `<div><strong>Priority:</strong> ${mail.priority}</div>` : ''}
      </div>
      <div class="mail-detail-body">
        ${escapeHtml(mail.message || mail.body || '(No content)')}
      </div>
    </div>
    <div class="modal-footer">
      <button class="btn btn-secondary" onclick="document.dispatchEvent(new CustomEvent('mail:reply', { detail: { mail: ${JSON.stringify(mail)} } }))">
        <span class="material-icons">reply</span>
        Reply
      </button>
    </div>
  `;
  showDynamicModal('mail-detail', content);
}

/**
 * Show a dynamic modal with custom content
 */
function showDynamicModal(id, content) {
  // Remove existing dynamic modal if present
  const existing = document.getElementById(`${id}-modal`);
  if (existing) existing.remove();

  // Create new modal
  const modal = document.createElement('div');
  modal.id = `${id}-modal`;
  modal.className = 'modal';
  modal.innerHTML = content;

  // Add to document
  document.body.appendChild(modal);

  // Register and show
  registerModal(id, { element: modal });
  overlay?.classList.remove('hidden');
  modal.classList.remove('hidden');

  // Wire up close buttons
  modal.querySelectorAll('[data-modal-close]').forEach(btn => {
    btn.addEventListener('click', closeAllModals);
  });

  return modal;
}

// Utility
function escapeHtml(str) {
  if (!str) return '';
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
