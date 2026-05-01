import React, { useState, useMemo } from 'react';
import { Loader2, AlertCircle, ChevronRight, Filter, ArrowUp, ArrowDown } from 'lucide-react';

const BEHAVIOR_COLORS = {
  PREVENTIVE: { bg: 'bg-blue-500', label: 'Preventive' },
  DETECTIVE: { bg: 'bg-amber-500', label: 'Detective' },
  PROACTIVE: { bg: 'bg-purple-500', label: 'Proactive' },
};

const SOURCE_COLORS = {
  SH: 'bg-indigo-500',
  CONFIG: 'bg-teal-500',
  CT: 'bg-sky-500',
  'AWS-GR': 'bg-orange-500',
};

function getSource(aliases) {
  if (!aliases || aliases.length === 0) return { source: 'OTHER', code: '' };
  const alias = aliases[0];
  if (alias.startsWith('SH.')) return { source: 'SH', code: alias };
  if (alias.startsWith('CONFIG.')) return { source: 'CONFIG', code: alias };
  if (alias.startsWith('CT.')) return { source: 'CT', code: alias };
  if (alias.startsWith('AWS-GR_')) return { source: 'AWS-GR', code: alias };
  return { source: 'OTHER', code: alias };
}

function Breadcrumb({ items }) {
  if (!items || items.length === 0) return null;
  return (
    <div className="flex items-center gap-1 text-xs text-gray-400 dark:text-gray-500 mb-2">
      {items.map((item, i) => (
        <React.Fragment key={i}>
          {i > 0 && <ChevronRight size={10} />}
          <span>{item}</span>
        </React.Fragment>
      ))}
    </div>
  );
}

function CoverageSummary({ controlCount, enabledCount, distinctOuCount }) {
  return (
    <div className="rounded-lg border border-gray-200 dark:border-gray-700 p-3 mb-4">
      <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-2">
        Coverage Summary
      </h3>
      <p className="text-sm">
        <span className="font-medium">{enabledCount}</span> of{' '}
        <span className="font-medium">{controlCount}</span> controls enabled
        {distinctOuCount > 0 && (
          <> across <span className="font-medium">{distinctOuCount}</span> OUs</>
        )}
      </p>
    </div>
  );
}

function BehaviorBar({ breakdown }) {
  const total = Object.values(breakdown).reduce((s, v) => s + v, 0);
  if (total === 0) return null;

  return (
    <div className="mb-4">
      <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-2">
        Behavior Distribution
      </h3>
      <div className="flex h-4 rounded overflow-hidden">
        {Object.entries(BEHAVIOR_COLORS).map(([key, { bg }]) => {
          const count = breakdown[key] || 0;
          if (count === 0) return null;
          const pct = (count / total) * 100;
          return (
            <div key={key} className={`${bg} relative group`} style={{ width: `${pct}%` }}>
              <span className="absolute inset-0 flex items-center justify-center text-[9px] text-white font-medium opacity-0 group-hover:opacity-100">
                {count}
              </span>
            </div>
          );
        })}
      </div>
      <div className="flex gap-3 mt-1">
        {Object.entries(BEHAVIOR_COLORS).map(([key, { bg, label }]) => {
          const count = breakdown[key] || 0;
          if (count === 0) return null;
          return (
            <div key={key} className="flex items-center gap-1 text-[10px] text-gray-500 dark:text-gray-400">
              <span className={`w-2 h-2 rounded-full ${bg}`} />
              {label} ({count})
            </div>
          );
        })}
      </div>
    </div>
  );
}

const SORT_COLUMNS = [
  { key: 'source', label: 'Source', getValue: c => getSource(c.aliases).source },
  { key: 'code', label: 'Code', getValue: c => getSource(c.aliases).code },
  { key: 'name', label: 'Name', getValue: c => c.name || '' },
  { key: 'behavior', label: 'Behavior', getValue: c => c.behavior || '' },
  { key: 'severity', label: 'Severity', getValue: c => c.severity || '' },
  { key: 'enabled_ous', label: 'Enabled OUs', getValue: c => c.enabled_ous.length },
];

function SortHeader({ column, sortCol, sortDir, onSort }) {
  const active = sortCol === column.key;
  return (
    <th
      className="py-2 pr-2 font-medium cursor-pointer select-none hover:text-gray-700 dark:hover:text-gray-200"
      onClick={() => onSort(column.key)}
    >
      <span className="inline-flex items-center gap-0.5">
        {column.label}
        {active && (sortDir === 'asc'
          ? <ArrowUp size={10} />
          : <ArrowDown size={10} />
        )}
      </span>
    </th>
  );
}

function ControlsTable({ controls }) {
  const [sourceFilter, setSourceFilter] = useState('ALL');
  const [sortCol, setSortCol] = useState(null);
  const [sortDir, setSortDir] = useState('asc');

  const handleSort = (key) => {
    if (sortCol === key) {
      setSortDir(d => d === 'asc' ? 'desc' : 'asc');
    } else {
      setSortCol(key);
      setSortDir('asc');
    }
  };

  const sources = useMemo(() => {
    const s = new Set();
    (controls || []).forEach(c => s.add(getSource(c.aliases).source));
    return ['ALL', ...Array.from(s).sort()];
  }, [controls]);

  const filtered = useMemo(() => {
    if (!controls) return [];
    let result = sourceFilter === 'ALL' ? [...controls] : controls.filter(c => getSource(c.aliases).source === sourceFilter);
    if (sortCol) {
      const col = SORT_COLUMNS.find(c => c.key === sortCol);
      if (col) {
        result = [...result].sort((a, b) => {
          const va = col.getValue(a);
          const vb = col.getValue(b);
          if (typeof va === 'number') return sortDir === 'asc' ? va - vb : vb - va;
          return sortDir === 'asc' ? String(va).localeCompare(String(vb)) : String(vb).localeCompare(String(va));
        });
      }
    }
    return result;
  }, [controls, sourceFilter, sortCol, sortDir]);

  if (!controls || controls.length === 0) {
    return (
      <p className="text-sm text-gray-400 italic">
        No controls implement this ontology node.
      </p>
    );
  }

  return (
    <div>
      {/* Source filter */}
      {sources.length > 2 && (
        <div className="flex items-center gap-2 mb-2">
          <Filter size={12} className="text-gray-400" />
          <select
            value={sourceFilter}
            onChange={e => setSourceFilter(e.target.value)}
            className="text-xs bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded px-2 py-1 text-gray-700 dark:text-gray-300"
          >
            {sources.map(s => (
              <option key={s} value={s}>
                {s === 'ALL' ? `All Sources (${controls.length})` : `${s} (${controls.filter(c => getSource(c.aliases).source === s).length})`}
              </option>
            ))}
          </select>
        </div>
      )}

      <div className="overflow-x-auto">
        <table className="w-full text-xs">
          <thead>
            <tr className="border-b border-gray-200 dark:border-gray-700 text-left text-gray-500 dark:text-gray-400">
              {SORT_COLUMNS.map(col => (
                <SortHeader key={col.key} column={col} sortCol={sortCol} sortDir={sortDir} onSort={handleSort} />
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.map(c => {
              const { source, code } = getSource(c.aliases);
              return (
                <tr key={c.arn} className="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                  <td className="py-1.5 pr-2">
                    <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] font-medium text-white ${SOURCE_COLORS[source] || 'bg-gray-400'}`}>
                      {source}
                    </span>
                  </td>
                  <td className="py-1.5 pr-2 font-mono text-[10px] text-gray-500 dark:text-gray-400 whitespace-nowrap">{code}</td>
                  <td className="py-1.5 pr-3 max-w-xs truncate">{c.name}</td>
                  <td className="py-1.5 pr-2">
                    <span className={`inline-block px-1.5 py-0.5 rounded text-[10px] font-medium text-white ${BEHAVIOR_COLORS[c.behavior]?.bg || 'bg-gray-400'}`}>
                      {c.behavior}
                    </span>
                  </td>
                  <td className="py-1.5 pr-2">{c.severity}</td>
                  <td className="py-1.5 text-gray-400">
                    {c.enabled_ous.length > 0 ? c.enabled_ous.join(', ') : '—'}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default function OntologyDetailPanel({ detail, loading, error }) {
  if (!detail && !loading && !error) {
    return (
      <div className="flex items-center justify-center h-full text-gray-400 text-sm">
        Select a node to view details
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-gray-400">
        <Loader2 size={20} className="animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center gap-2 p-4 text-red-500 text-sm">
        <AlertCircle size={16} />
        <span>{error}</span>
      </div>
    );
  }

  return (
    <div className="p-4 overflow-y-auto h-full">
      <Breadcrumb items={detail.breadcrumb} />
      <h2 className="text-base font-semibold mb-1">
        {detail.number && <span className="text-gray-400 dark:text-gray-500 mr-1.5">{detail.number}</span>}
        {detail.name}
      </h2>
      {detail.description && (
        <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">{detail.description}</p>
      )}

      <CoverageSummary
        controlCount={detail.control_count}
        enabledCount={detail.enabled_count}
        distinctOuCount={detail.distinct_ou_count}
      />

      <BehaviorBar breakdown={detail.behavior_breakdown} />

      <h3 className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-2">
        Implementing Controls ({detail.control_count})
      </h3>
      <ControlsTable controls={detail.controls} />
    </div>
  );
}
