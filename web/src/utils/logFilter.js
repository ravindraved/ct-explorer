/**
 * Log level filtering utilities for the OutputPanel.
 *
 * Extracted as pure functions so they can be tested independently
 * without rendering React components.
 */

const LEVEL_HIERARCHY = { error: 0, warning: 1, info: 2 };

/**
 * Return the effective log level for a log entry.
 * Request-type entries (log_type === "request") are treated as "info".
 */
export function getEffectiveLevel(entry) {
  if (entry.log_type === 'request') return 'info';
  return (entry.level || 'info').toLowerCase();
}

/**
 * Filter an array of log entries by the selected level threshold.
 * "all" returns everything; otherwise only entries at or above the
 * threshold severity are included (error=0, warning=1, info=2).
 */
export function filterLogsByLevel(logs, levelFilter) {
  if (levelFilter === 'all') return logs;
  const threshold = LEVEL_HIERARCHY[levelFilter] ?? 2;
  return logs.filter(entry => {
    const effective = getEffectiveLevel(entry);
    return (LEVEL_HIERARCHY[effective] ?? 2) <= threshold;
  });
}

/**
 * Filter log entries by log_type.
 * Returns only entries where log_type matches the given type.
 */
export function filterByLogType(logs, logType) {
  return logs.filter(entry => entry.log_type === logType);
}
