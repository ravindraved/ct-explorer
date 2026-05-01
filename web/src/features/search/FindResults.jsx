import React, { useState } from 'react';
import { ChevronDown, ChevronRight, ArrowLeft, ArrowRight } from 'lucide-react';

const TYPE_COLORS = {
  ou: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
  account: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
  control: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300',
  scp: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-300',
  catalog_control: 'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-300',
  catalog_domain: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300',
  catalog_objective: 'bg-cyan-100 text-cyan-700 dark:bg-cyan-900/30 dark:text-cyan-300',
  catalog_common_control: 'bg-pink-100 text-pink-700 dark:bg-pink-900/30 dark:text-pink-300',
};

const TYPE_LABELS = {
  ou: 'OU',
  account: 'Account',
  control: 'Control',
  scp: 'SCP',
  catalog_control: 'Catalog Control',
  catalog_domain: 'Domain',
  catalog_objective: 'Objective',
  catalog_common_control: 'Common Control',
};

function getNavigationTarget(item) {
  switch (item.entity_type) {
    case 'ou': return { view: 'organization', payload: { type: 'ou', id: item.id } };
    case 'account': return { view: 'organization', payload: { type: 'account', id: item.id } };
    case 'control': return { view: 'controls', payload: { arn: item.arn } };
    case 'scp': return { view: 'scps', payload: { id: item.id } };
    case 'catalog_control': return { view: 'catalog', payload: { arn: item.arn } };
    case 'catalog_domain':
    case 'catalog_objective':
    case 'catalog_common_control':
      return { view: 'ontology', payload: { arn: item.arn } };
    default: return null;
  }
}

function MetadataChips({ metadata }) {
  if (!metadata || Object.keys(metadata).length === 0) return null;
  const entries = Object.entries(metadata).filter(([, v]) => v != null && v !== '');
  if (entries.length === 0) return null;

  return (
    <div className="flex flex-wrap gap-1 mt-1">
      {entries.slice(0, 3).map(([key, value]) => (
        <span key={key} className="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-400">
          {key}: {typeof value === 'object' ? JSON.stringify(value) : String(value)}
        </span>
      ))}
    </div>
  );
}

function ResultItem({ item, navigateTo }) {
  const target = getNavigationTarget(item);
  const colorClass = TYPE_COLORS[item.entity_type] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300';

  return (
    <button
      onClick={() => target && navigateTo?.(target.view, target.payload)}
      className="w-full text-left px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors border-b border-gray-100 dark:border-gray-800 last:border-b-0"
    >
      <div className="flex items-center gap-2">
        <span className={`text-xs px-1.5 py-0.5 rounded font-medium ${colorClass}`}>
          {TYPE_LABELS[item.entity_type] || item.entity_type}
        </span>
        <span className="text-sm font-medium truncate">{item.name}</span>
      </div>
      {item.description && (
        <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5 truncate">{item.description}</p>
      )}
      <MetadataChips metadata={item.metadata} />
    </button>
  );
}

function GroupSection({ group, navigateTo }) {
  const [expanded, setExpanded] = useState(true);

  return (
    <div className="border-b border-gray-200 dark:border-gray-700 last:border-b-0">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-2 px-3 py-2 text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors"
      >
        {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
        <span className={`text-xs px-1.5 py-0.5 rounded font-medium ${TYPE_COLORS[group.entity_type] || ''}`}>
          {TYPE_LABELS[group.entity_type] || group.entity_type}
        </span>
        <span className="text-gray-500 dark:text-gray-400">({group.count})</span>
      </button>
      {expanded && (
        <div>
          {group.items.map((item, i) => (
            <ResultItem key={item.id + '-' + i} item={item} navigateTo={navigateTo} />
          ))}
        </div>
      )}
    </div>
  );
}

export default function FindResults({ results, engine, navigateTo }) {
  const { page, pageCount, setPage } = engine;
  const groups = results?.groups || [];
  const items = results?.items || [];
  const total = results?.total || 0;

  if (total === 0) {
    return (
      <div className="flex items-center justify-center p-8 text-gray-400 dark:text-gray-500 text-sm">
        No results found
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <div className="px-3 py-2 text-xs text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700">
        {total} result{total !== 1 ? 's' : ''} found
      </div>

      <div className="flex-1 overflow-y-auto">
        {groups.length > 0 ? (
          groups.map(group => (
            <GroupSection key={group.entity_type} group={group} navigateTo={navigateTo} />
          ))
        ) : (
          items.map((item, i) => (
            <ResultItem key={item.id + '-' + i} item={item} navigateTo={navigateTo} />
          ))
        )}
      </div>

      {pageCount > 1 && (
        <div className="flex items-center justify-center gap-3 px-3 py-2 border-t border-gray-200 dark:border-gray-700 flex-shrink-0">
          <button
            onClick={() => setPage(page - 1)}
            disabled={page <= 1}
            className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            <ArrowLeft size={16} />
          </button>
          <span className="text-sm text-gray-600 dark:text-gray-400">
            Page {page} of {pageCount}
          </span>
          <button
            onClick={() => setPage(page + 1)}
            disabled={page >= pageCount}
            className="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors"
          >
            <ArrowRight size={16} />
          </button>
        </div>
      )}
    </div>
  );
}
