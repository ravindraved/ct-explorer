/**
 * Format a structured log entry for display in the OutputPanel.
 *
 * Request logs (log_type === "request"):
 *   [timestamp] METHOD /path → STATUS (Xms)
 *
 * Event logs (all others):
 *   [timestamp] LEVEL   component — event key=value key=value
 */
export function formatLog(entry) {
  const ts = entry.timestamp || '';
  const level = (entry.level || 'info').toUpperCase();

  if (entry.log_type === 'request') {
    const method = entry.method || '?';
    const endpoint = entry.endpoint || '?';
    const status = entry.status_code || '?';
    const duration = entry.duration_ms != null ? `${entry.duration_ms}ms` : '?';
    return `[${ts}] ${method} ${endpoint} → ${status} (${duration})`;
  }

  // Event log: [timestamp] LEVEL component — event key=value key=value
  const component = entry.component || '-';
  const event = entry.event || '';
  const skipKeys = new Set([
    'timestamp', 'level', 'component', 'event', 'log_type',
  ]);
  const extras = Object.entries(entry)
    .filter(([k]) => !skipKeys.has(k))
    .map(([k, v]) => `${k}=${v}`)
    .join(' ');
  const suffix = extras ? ` ${extras}` : '';
  return `[${ts}] ${level.padEnd(7)} ${component} — ${event}${suffix}`;
}

/**
 * Return a Tailwind CSS color class based on log level.
 */
export function getLogColor(entry) {
  const level = (entry.level || 'info').toLowerCase();
  if (level === 'error') return 'text-red-500 dark:text-red-400';
  if (level === 'warning') return 'text-yellow-600 dark:text-yellow-400';
  return 'text-green-700 dark:text-green-400';
}

/**
 * Format an api_call log entry for the API Calls tab.
 * Shows: [Q-001] aws organizations list-roots --region ap-south-1
 * On error: [Q-002] ERROR aws sts get-caller-identity → AccessDenied
 */
export function formatApiCall(entry) {
  const qid = entry.query_id || '?';
  const cli = entry.cli_command || 'unknown';
  const status = (entry.status || 'ok').toLowerCase();
  if (status === 'error') {
    const err = entry.error || 'unknown error';
    return `[${qid}] ERROR  ${cli}\n         → ${err}`;
  }
  return `[${qid}] ${cli}`;
}

/**
 * Return a Tailwind CSS color class for an API call entry.
 */
export function getApiCallColor(entry) {
  const status = (entry.status || 'ok').toLowerCase();
  const event = (entry.event || '').toLowerCase();
  if (status === 'error') return 'text-red-500 dark:text-red-400';
  if (event === 'throttle_retry') return 'text-yellow-600 dark:text-yellow-400';
  return 'text-cyan-600 dark:text-cyan-400';
}
