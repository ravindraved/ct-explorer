import React from 'react';

function prettyJson(raw) {
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

export default function PolicyViewer({ detail }) {
  return (
    <div className="p-4">
      <h3 className="text-sm font-semibold mb-2">{detail.name}</h3>
      <p className="text-xs text-gray-500 dark:text-gray-400 mb-1">{detail.description}</p>
      <p className="text-xs text-gray-400 mb-3">Targets: {detail.target_ids.join(', ') || '—'}</p>
      <div className="border border-gray-200 dark:border-gray-700 rounded">
        <div className="px-3 py-1 text-xs font-medium bg-gray-100 dark:bg-gray-700 border-b border-gray-200 dark:border-gray-600">
          Policy Document
        </div>
        <pre className="p-3 text-xs overflow-auto bg-gray-50 dark:bg-gray-900 text-gray-800 dark:text-gray-200 leading-5">
          {prettyJson(detail.policy_document)}
        </pre>
      </div>
    </div>
  );
}
