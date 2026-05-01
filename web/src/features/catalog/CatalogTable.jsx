import { useState, useMemo, useCallback } from 'react';
import { ArrowUp, ArrowDown, ArrowUpDown } from 'lucide-react';

const BEHAVIOR_COLORS = {
  PREVENTIVE: 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
  DETECTIVE: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  PROACTIVE: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300',
};

const SEVERITY_COLORS = {
  CRITICAL: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  HIGH: 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
  MEDIUM: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  LOW: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
};

function Badge({ text, colorMap }) {
  return (
    <span className={`text-xs px-1.5 py-0.5 rounded whitespace-nowrap ${colorMap[text] || ''}`}>
      {text}
    </span>
  );
}

function ServicesCell({ services }) {
  if (!services || services.length === 0) return <span className="text-gray-400">—</span>;
  const shown = services.slice(0, 2);
  const extra = services.length - 2;
  return (
    <span className="text-xs">
      {shown.join(', ')}
      {extra > 0 && <span className="text-gray-400"> +{extra} more</span>}
    </span>
  );
}

function EnabledIndicator({ enabledMap, arn }) {
  const ous = enabledMap?.[arn];
  if (!ous || ous.length === 0) return null;
  return (
    <span className="text-xs px-1.5 py-0.5 rounded bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300">
      {ous.length} OU{ous.length > 1 ? 's' : ''}
    </span>
  );
}

function OntologyRefs({ refs }) {
  if (!refs || refs.length === 0) return <span className="text-gray-400">—</span>;
  return (
    <div className="flex flex-wrap gap-1">
      {refs.map(r => (
        <span
          key={r.number}
          title={r.name}
          className="text-[10px] px-1.5 py-0.5 rounded bg-violet-100 text-violet-700 dark:bg-violet-900 dark:text-violet-300 whitespace-nowrap cursor-default"
        >
          {r.number}
        </span>
      ))}
    </div>
  );
}

const DEFAULT_COL_WIDTHS = { name: 280, behavior: 112, severity: 96, services: 130, impl: 112, enabled: 80, ontology: 120 };

function ResizeHandle({ colKey, colWidths, setColWidths, minWidth = 50 }) {
  const onMouseDown = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    const startX = e.clientX;
    const startW = colWidths[colKey];
    const onMove = (ev) => {
      setColWidths(prev => ({ ...prev, [colKey]: Math.max(minWidth, startW + ev.clientX - startX) }));
    };
    const onUp = () => {
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  }, [colKey, colWidths, setColWidths, minWidth]);

  return (
    <div
      onMouseDown={onMouseDown}
      className="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-blue-400 active:bg-blue-500"
    />
  );
}

export default function CatalogTable({ controls, totalCount, selected, onSelect, enabledMap, ontologyMap }) {
  const [sortKey, setSortKey] = useState(null);
  const [sortDir, setSortDir] = useState('asc');
  const [colWidths, setColWidths] = useState(DEFAULT_COL_WIDTHS);

  const toggleSort = (key) => {
    if (sortKey === key) {
      setSortDir(prev => prev === 'asc' ? 'desc' : 'asc');
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const sorted = useMemo(() => {
    if (!sortKey) return controls;
    return [...controls].sort((a, b) => {
      let av, bv;
      if (sortKey === 'enabled') {
        av = enabledMap?.[a.arn]?.length || 0;
        bv = enabledMap?.[b.arn]?.length || 0;
      } else {
        av = (a[sortKey] || '').toString().toLowerCase();
        bv = (b[sortKey] || '').toString().toLowerCase();
      }
      if (av < bv) return sortDir === 'asc' ? -1 : 1;
      if (av > bv) return sortDir === 'asc' ? 1 : -1;
      return 0;
    });
  }, [controls, sortKey, sortDir, enabledMap]);

  const SortIcon = ({ col }) => {
    if (sortKey !== col) return <ArrowUpDown size={12} className="text-gray-400" />;
    return sortDir === 'asc' ? <ArrowUp size={12} /> : <ArrowDown size={12} />;
  };

  const thBase = "px-4 py-2 font-medium cursor-pointer select-none hover:bg-gray-200 dark:hover:bg-gray-700 relative";
  const rh = (key, min) => <ResizeHandle colKey={key} colWidths={colWidths} setColWidths={setColWidths} minWidth={min} />;

  return (
    <div className="flex-1 flex flex-col min-h-0 overflow-auto">
      <div className="px-4 py-1.5 text-xs text-gray-500 dark:text-gray-400 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex-shrink-0">
        {controls.length} of {totalCount} controls
      </div>

      <table className="w-full text-sm table-fixed">
        <thead className="sticky top-0 bg-gray-100 dark:bg-gray-800 text-left">
          <tr>
            <th className={thBase} style={{ width: colWidths.name }} onClick={() => toggleSort('name')}>
              <span className="flex items-center gap-1">Name <SortIcon col="name" /></span>
              {rh('name', 100)}
            </th>
            <th className={thBase} style={{ width: colWidths.behavior }} onClick={() => toggleSort('behavior')}>
              <span className="flex items-center gap-1">Behavior <SortIcon col="behavior" /></span>
              {rh('behavior', 70)}
            </th>
            <th className={thBase} style={{ width: colWidths.severity }} onClick={() => toggleSort('severity')}>
              <span className="flex items-center gap-1">Severity <SortIcon col="severity" /></span>
              {rh('severity', 70)}
            </th>
            <th className={`${thBase} cursor-default`} style={{ width: colWidths.services }}>
              <span className="flex items-center gap-1">Services</span>
              {rh('services', 80)}
            </th>
            <th className={thBase} style={{ width: colWidths.impl }} onClick={() => toggleSort('implementation_type')}>
              <span className="flex items-center gap-1">Impl Type <SortIcon col="implementation_type" /></span>
              {rh('impl', 60)}
            </th>
            <th className={thBase} style={{ width: colWidths.enabled }} onClick={() => toggleSort('enabled')}>
              <span className="flex items-center gap-1">Enabled <SortIcon col="enabled" /></span>
              {rh('enabled', 50)}
            </th>
            <th className={`${thBase} cursor-default`} style={{ width: colWidths.ontology }}>
              <span className="flex items-center gap-1">Ontology</span>
              {rh('ontology', 60)}
            </th>
          </tr>
        </thead>
        <tbody>
          {sorted.length === 0 && (
            <tr>
              <td colSpan={7} className="px-4 py-6 text-center text-gray-500 dark:text-gray-400">
                No controls match the current filters.
              </td>
            </tr>
          )}
          {sorted.map(c => (
            <tr
              key={c.arn}
              onClick={() => onSelect(selected?.arn === c.arn ? null : c)}
              className={`cursor-pointer border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50
                ${selected?.arn === c.arn ? 'bg-blue-50 dark:bg-blue-900/30' : ''}`}
            >
              <td className="px-4 py-2 break-words">{c.name}</td>
              <td className="px-4 py-2 break-words"><Badge text={c.behavior} colorMap={BEHAVIOR_COLORS} /></td>
              <td className="px-4 py-2 break-words"><Badge text={c.severity} colorMap={SEVERITY_COLORS} /></td>
              <td className="px-4 py-2 break-words"><ServicesCell services={c.services} /></td>
              <td className="px-4 py-2 text-xs text-gray-500 dark:text-gray-400 break-words">{c.implementation_type || '—'}</td>
              <td className="px-4 py-2 break-words"><EnabledIndicator enabledMap={enabledMap} arn={c.arn} /></td>
              <td className="px-4 py-2 break-words"><OntologyRefs refs={ontologyMap?.[c.arn]} /></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
