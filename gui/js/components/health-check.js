/**
 * Gas Town GUI - Health Check Component
 *
 * Displays system health diagnostics from gt doctor command.
 */

import { api } from '../api.js';
import { showToast } from './toast.js';

let container = null;
let refreshBtn = null;

/**
 * Initialize health check component
 */
export function initHealthCheck() {
  container = document.getElementById('health-check-container');
  refreshBtn = document.getElementById('health-refresh');

  if (refreshBtn) {
    refreshBtn.addEventListener('click', () => {
      loadHealthCheck();
    });
  }
}

/**
 * Load health check data
 */
export async function loadHealthCheck() {
  if (!container) return;

  // Show loading state
  container.innerHTML = `
    <div class="health-loading">
      <span class="loading-spinner"></span>
      <p>Running health diagnostics...</p>
    </div>
  `;

  if (refreshBtn) {
    refreshBtn.disabled = true;
    refreshBtn.innerHTML = '<span class="material-icons spinning">sync</span> Running...';
  }

  try {
    const result = await api.runDoctor();
    renderHealthResults(result);
  } catch (err) {
    showToast(`Failed to run health check: ${err.message}`, 'error');
    container.innerHTML = `
      <div class="health-error">
        <span class="material-icons">error</span>
        <h3>Health Check Failed</h3>
        <p>${escapeHtml(err.message)}</p>
        <button class="btn btn-secondary" onclick="document.getElementById('health-refresh').click()">
          <span class="material-icons">refresh</span>
          Retry
        </button>
      </div>
    `;
  } finally {
    if (refreshBtn) {
      refreshBtn.disabled = false;
      refreshBtn.innerHTML = '<span class="material-icons">refresh</span> Run Doctor';
    }
  }
}

/**
 * Render health check results
 */
function renderHealthResults(data) {
  if (!container) return;

  // Handle raw text response (if not JSON)
  if (data.raw) {
    container.innerHTML = `
      <div class="health-raw">
        <pre>${escapeHtml(data.raw)}</pre>
      </div>
    `;
    return;
  }

  // Parse structured results
  const checks = data.checks || data.results || [];
  const summary = data.summary || {};

  // Calculate overall status
  let passCount = 0;
  let warnCount = 0;
  let failCount = 0;

  checks.forEach(check => {
    const status = (check.status || check.result || '').toLowerCase();
    if (status === 'pass' || status === 'ok' || status === 'success') passCount++;
    else if (status === 'warn' || status === 'warning') warnCount++;
    else if (status === 'fail' || status === 'error') failCount++;
  });

  const overallStatus = failCount > 0 ? 'fail' : (warnCount > 0 ? 'warn' : 'pass');
  const statusLabels = {
    pass: 'All Systems Healthy',
    warn: 'Some Warnings',
    fail: 'Issues Detected'
  };
  const statusIcons = {
    pass: 'check_circle',
    warn: 'warning',
    fail: 'error'
  };

  container.innerHTML = `
    <div class="health-summary health-${overallStatus}">
      <div class="health-summary-icon">
        <span class="material-icons">${statusIcons[overallStatus]}</span>
      </div>
      <div class="health-summary-info">
        <h2>${statusLabels[overallStatus]}</h2>
        <div class="health-summary-stats">
          <span class="health-stat pass">
            <span class="material-icons">check_circle</span>
            ${passCount} Passed
          </span>
          <span class="health-stat warn">
            <span class="material-icons">warning</span>
            ${warnCount} Warnings
          </span>
          <span class="health-stat fail">
            <span class="material-icons">error</span>
            ${failCount} Failed
          </span>
        </div>
      </div>
      <div class="health-summary-time">
        Last checked: ${new Date().toLocaleTimeString()}
      </div>
    </div>

    <div class="health-checks">
      ${checks.length > 0 ? checks.map(check => renderCheckItem(check)).join('') : `
        <div class="health-empty">
          <span class="material-icons">info</span>
          <p>No diagnostic checks returned.</p>
        </div>
      `}
    </div>

    ${summary.recommendations ? `
      <div class="health-recommendations">
        <h3>
          <span class="material-icons">tips_and_updates</span>
          Recommendations
        </h3>
        <ul>
          ${summary.recommendations.map(r => `<li>${escapeHtml(r)}</li>`).join('')}
        </ul>
      </div>
    ` : ''}
  `;
}

/**
 * Render a single check item
 */
function renderCheckItem(check) {
  const status = (check.status || check.result || 'unknown').toLowerCase();
  const statusMap = {
    pass: { icon: 'check_circle', class: 'pass', label: 'Pass' },
    ok: { icon: 'check_circle', class: 'pass', label: 'OK' },
    success: { icon: 'check_circle', class: 'pass', label: 'Success' },
    warn: { icon: 'warning', class: 'warn', label: 'Warning' },
    warning: { icon: 'warning', class: 'warn', label: 'Warning' },
    fail: { icon: 'error', class: 'fail', label: 'Failed' },
    error: { icon: 'error', class: 'fail', label: 'Error' },
  };

  const statusInfo = statusMap[status] || { icon: 'help', class: 'unknown', label: status };

  return `
    <div class="health-check-item health-${statusInfo.class}">
      <div class="health-check-status">
        <span class="material-icons">${statusInfo.icon}</span>
      </div>
      <div class="health-check-info">
        <div class="health-check-name">${escapeHtml(check.name || check.check || 'Unknown Check')}</div>
        ${check.message || check.description ? `
          <div class="health-check-message">${escapeHtml(check.message || check.description)}</div>
        ` : ''}
        ${check.details ? `
          <div class="health-check-details">
            <pre>${escapeHtml(typeof check.details === 'string' ? check.details : JSON.stringify(check.details, null, 2))}</pre>
          </div>
        ` : ''}
      </div>
      <div class="health-check-label">${statusInfo.label}</div>
    </div>
  `;
}

/**
 * Escape HTML entities
 */
function escapeHtml(str) {
  if (!str) return '';
  const div = document.createElement('div');
  div.textContent = str;
  return div.innerHTML;
}
