import React, { useState } from 'react';
import { useScps } from '../../hooks/useScps';
import { apiGet } from '../../api/client';
import LoadingIndicator from '../../components/LoadingIndicator';
import PolicyViewer from './PolicyViewer';

export default function ScpListView() {
  const { scps, loading, error } = useScps();
  const [selected, setSelected] = useState(null);
  const [detail, setDetail] = useState(null);
  const [detailLoading, setDetailLoading] = useState(false);

  const handleSelect = async (scp) => {
    setSelected(scp);
    setDetailLoading(true);
    const { data } = await apiGet(`/api/scps/${scp.id}`);
    setDetail(data);
    setDetailLoading(false);
  };

  if (loading) return <LoadingIndicator message="Loading SCPs..." />;
  if (error) return <p className="p-4 text-red-500">{error}</p>;

  return (
    <div className="flex h-full">
      {/* List */}
      <div className="w-1/2 overflow-auto border-r border-gray-200 dark:border-gray-700">
        <h2 className="px-4 py-2 text-sm font-semibold border-b border-gray-200 dark:border-gray-700">
          SCPs ({scps.length})
        </h2>
        {scps.length === 0 && (
          <p className="p-4 text-gray-500 dark:text-gray-400">No SCPs found.</p>
        )}
        {scps.map(s => (
          <button
            key={s.id}
            onClick={() => handleSelect(s)}
            className={`w-full text-left px-4 py-2 text-sm border-b border-gray-100 dark:border-gray-700
              hover:bg-gray-100 dark:hover:bg-gray-700
              ${selected?.id === s.id ? 'bg-blue-50 dark:bg-blue-900/30' : ''}`}
          >
            <div className="font-medium truncate">{s.name}</div>
            <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">{s.description}</div>
            <div className="text-xs text-gray-400 mt-0.5">{s.target_ids.length} targets</div>
          </button>
        ))}
      </div>
      {/* Detail / Policy Viewer */}
      <div className="w-1/2 overflow-auto">
        {detailLoading ? (
          <LoadingIndicator message="Loading policy..." />
        ) : detail ? (
          <PolicyViewer detail={detail} />
        ) : (
          <p className="p-4 text-gray-500 dark:text-gray-400">Select an SCP to view its policy.</p>
        )}
      </div>
    </div>
  );
}
