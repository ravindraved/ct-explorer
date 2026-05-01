import React from 'react';
import { Check, X, Shield, Building2, FileKey } from 'lucide-react';

function SummaryCard({ label, value, icon: Icon, color }) {
  return (
    <div className={`flex items-center gap-2 px-3 py-2 rounded border ${color}`}>
      {Icon && <Icon size={16} />}
      <div>
        <div className="text-lg font-semibold">{value}</div>
        <div className="text-xs">{label}</div>
      </div>
    </div>
  );
}

function getNavigationTarget(item) {
  switch (item.entity_type) {
    case 'control': return { view: 'controls', payload: { arn: item.arn } };
    case 'scp': return { view: 'scps', payload: { id: item.arn } };
    case 'catalog_control': return { view: 'catalog', payload: { arn: item.arn } };
    default: return null;
  }
}

export default function CoverageResults({ results, navigateTo }) {
  const { target_name, target_type, total_controls, enabled_count, gap_count, scp_count, account_count, items } = results;

  return (
    <div className="p-4 space-y-4">
      <div>
        <h2 className="text-sm font-semibold text-gray-600 dark:text-gray-300 uppercase tracking-wide">
          Coverage for {target_name}
        </h2>
        <p className="text-xs text-gray-400 dark:text-gray-500 mt-0.5">Type: {target_type}</p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-2 lg:grid-cols-5 gap-2">
        <SummaryCard
          label="Total Controls"
          value={total_controls}
          icon={Shield}
          color="border-gray-200 dark:border-gray-700 text-gray-700 dark:text-gray-300"
        />
        <SummaryCard
          label="Enabled"
          value={enabled_count}
          icon={Check}
          color="border-green-200 dark:border-green-800 text-green-700 dark:text-green-300"
        />
        <SummaryCard
          label="Gaps"
          value={gap_count}
          icon={X}
          color="border-red-200 dark:border-red-800 text-red-700 dark:text-red-300"
        />
        <SummaryCard
          label="SCPs"
          value={scp_count}
          icon={FileKey}
          color="border-orange-200 dark:border-orange-800 text-orange-700 dark:text-orange-300"
        />
        <SummaryCard
          label="Accounts"
          value={account_count}
          icon={Building2}
          color="border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300"
        />
      </div>

      {/* Items list */}
      <div className="border border-gray-200 dark:border-gray-700 rounded overflow-hidden">
        {items && items.length > 0 ? (
          items.map((item, i) => {
            const target = getNavigationTarget(item);
            return (
              <button
                key={item.arn + '-' + i}
                onClick={() => target && navigateTo?.(target.view, target.payload)}
                className="w-full text-left flex items-center gap-2 px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors border-b border-gray-100 dark:border-gray-800 last:border-b-0"
              >
                {item.is_covered ? (
                  <Check size={16} className="text-green-500 flex-shrink-0" />
                ) : (
                  <X size={16} className="text-red-500 flex-shrink-0" />
                )}
                <div className="min-w-0 flex-1">
                  <div className="text-sm font-medium truncate">{item.name}</div>
                  <div className="text-xs text-gray-500 dark:text-gray-400 truncate">{item.arn}</div>
                </div>
                <span className={`text-xs px-1.5 py-0.5 rounded font-medium flex-shrink-0 ${
                  item.is_covered
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
                    : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
                }`}>
                  {item.is_covered ? 'Covered' : 'Gap'}
                </span>
              </button>
            );
          })
        ) : (
          <div className="p-4 text-sm text-gray-400 dark:text-gray-500 text-center">
            No coverage items
          </div>
        )}
      </div>
    </div>
  );
}
