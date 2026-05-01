import React from 'react';
import { Network, Shield, FileKey } from 'lucide-react';

function getNavigationTarget(item) {
  switch (item.entity_type) {
    case 'control': return { view: 'controls', payload: { arn: item.arn } };
    case 'scp': return { view: 'scps', payload: { id: item.id } };
    default: return null;
  }
}

function NodeItem({ item, navigateTo }) {
  const target = getNavigationTarget(item);
  return (
    <button
      onClick={() => target && navigateTo?.(target.view, target.payload)}
      className="w-full text-left flex items-center gap-2 px-2 py-1 text-xs hover:bg-gray-50 dark:hover:bg-gray-800 rounded transition-colors"
    >
      {item.entity_type === 'control' ? (
        <Shield size={12} className="text-green-500 flex-shrink-0" />
      ) : (
        <FileKey size={12} className="text-orange-500 flex-shrink-0" />
      )}
      <span className="truncate">{item.name}</span>
    </button>
  );
}

function PathNode({ node, isLast, navigateTo }) {
  return (
    <div className="relative pl-6">
      {/* Vertical connector line */}
      {!isLast && (
        <div className="absolute left-2.5 top-8 bottom-0 w-px bg-gray-300 dark:bg-gray-600" />
      )}
      {/* Node dot */}
      <div className="absolute left-1 top-3 w-3 h-3 rounded-full border-2 border-blue-500 bg-white dark:bg-gray-900" />

      <div className="pb-4">
        <div className="flex items-center gap-2 mb-1">
          <Network size={14} className="text-blue-500" />
          <span className="text-sm font-medium">{node.ou_name}</span>
          <span className="text-xs text-gray-400 dark:text-gray-500">{node.ou_id}</span>
        </div>

        {node.controls.length > 0 && (
          <div className="ml-4 mb-1">
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-0.5">
              Controls ({node.controls.length})
            </div>
            {node.controls.map((c, i) => (
              <NodeItem key={c.id + '-' + i} item={c} navigateTo={navigateTo} />
            ))}
          </div>
        )}

        {node.scps.length > 0 && (
          <div className="ml-4">
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-0.5">
              SCPs ({node.scps.length})
            </div>
            {node.scps.map((s, i) => (
              <NodeItem key={s.id + '-' + i} item={s} navigateTo={navigateTo} />
            ))}
          </div>
        )}

        {node.controls.length === 0 && node.scps.length === 0 && (
          <div className="ml-4 text-xs text-gray-400 dark:text-gray-500 italic">
            No controls or SCPs at this level
          </div>
        )}
      </div>
    </div>
  );
}

export default function PathResults({ results, navigateTo }) {
  const { account_name, account_id, chain, total_controls, total_scps } = results;

  return (
    <div className="p-4 space-y-4">
      <div>
        <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wide">
          Inheritance Path
        </h2>
        <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">
          Account: {account_name} ({account_id})
        </p>
      </div>

      <div className="border border-gray-200 dark:border-gray-700 rounded p-3">
        {chain && chain.length > 0 ? (
          chain.map((node, i) => (
            <PathNode
              key={node.ou_id}
              node={node}
              isLast={i === chain.length - 1}
              navigateTo={navigateTo}
            />
          ))
        ) : (
          <div className="text-sm text-gray-400 dark:text-gray-500 text-center py-4">
            No OU chain found
          </div>
        )}
      </div>

      {/* Cumulative totals */}
      <div className="flex gap-4 text-sm">
        <div className="flex items-center gap-1.5 text-gray-600 dark:text-gray-300">
          <Shield size={14} className="text-green-500" />
          <span>Total Controls: <span className="font-semibold">{total_controls}</span></span>
        </div>
        <div className="flex items-center gap-1.5 text-gray-600 dark:text-gray-300">
          <FileKey size={14} className="text-orange-500" />
          <span>Total SCPs: <span className="font-semibold">{total_scps}</span></span>
        </div>
      </div>
    </div>
  );
}
