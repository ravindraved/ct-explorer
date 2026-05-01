import React from 'react';

export default function StatusBar({ auth, refreshStatus }) {
  const connLabel = auth.authenticated
    ? `Connected — ${auth.accountId} (${auth.region})`
    : auth.error || 'Disconnected';

  const connColor = auth.authenticated
    ? 'text-green-400'
    : 'text-red-400';

  return (
    <div className="h-6 flex items-center justify-between px-3 text-xs bg-blue-600 dark:bg-blue-800 text-white select-none">
      <span className={connColor}>{connLabel}</span>
      {refreshStatus && refreshStatus !== 'idle' && (
        <span className="text-yellow-200">
          Refresh: {refreshStatus}
        </span>
      )}
    </div>
  );
}
