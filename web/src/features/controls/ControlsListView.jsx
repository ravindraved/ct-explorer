import { useState, useMemo, useCallback } from 'react';
import { Search, X, ChevronDown } from 'lucide-react';
import { useControls } from '../../hooks/useControls';
import LoadingIndicator from '../../components/LoadingIndicator';
import ControlDetail from './ControlDetail';

function TypeBadge({ type }) {
  const colors = {
    PREVENTIVE: 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
    DETECTIVE: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
    PROACTIVE: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300',
  };
  return <span className={`text-xs px-1.5 py-0.5 rounded ${colors[type] || ''}`}>{type}</span>;
}

function EnforcementBadge({ status }) {
  const isEnabled = status === 'ENABLED';
  return (
    <span className={`text-xs px-1.5 py-0.5 rounded ${
      isEnabled
        ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
        : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
    }`}>
      {status}
    </span>
  );
}

const DEFAULT_COL_WIDTHS = { name: 300, type: 112, enforcement: 112, ou: 160 };

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

export default function ControlsListView() {
  const { controls, loading, error } = useControls();
  const [selected, setSelected] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [typeFilter, setTypeFilter] = useState('ALL');
  const [ouFilter, setOuFilter] = useState('ALL');
  const [colWidths, setColWidths] = useState(DEFAULT_COL_WIDTHS);

  const rh = (key, min) => <ResizeHandle colKey={key} colWidths={colWidths} setColWidths={setColWidths} minWidth={min} />;

  // Derive unique types and OUs for filter dropdowns
  const controlTypes = useMemo(() => [...new Set(controls.map(c => c.control_type))].sort(), [controls]);
  const targetOUs = useMemo(() => [...new Set(controls.map(c => c.target_id))].sort(), [controls]);

  // Apply filters
  const filtered = useMemo(() => {
    return controls.filter(c => {
      if (typeFilter !== 'ALL' && c.control_type !== typeFilter) return false;
      if (ouFilter !== 'ALL' && c.target_id !== ouFilter) return false;
      if (searchText) {
        const q = searchText.toLowerCase();
        return c.name.toLowerCase().includes(q) || c.control_id.toLowerCase().includes(q);
      }
      return true;
    });
  }, [controls, typeFilter, ouFilter, searchText]);

  if (loading) return <LoadingIndicator message="Loading controls..." />;
  if (error) return <p className="p-4 text-red-500">{error}</p>;

  return (
    <div className="flex flex-col h-full">
      {/* Filter bar */}
      <div className="flex items-center gap-3 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex-shrink-0">
        {/* Search */}
        <div className="relative flex-1 max-w-xs">
          <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Filter controls..."
            value={searchText}
            onChange={e => setSearchText(e.target.value)}
            className="w-full pl-7 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
          {searchText && (
            <button onClick={() => setSearchText('')} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600">
              <X size={14} />
            </button>
          )}
        </div>

        {/* Type dropdown */}
        <div className="relative">
          <select
            value={typeFilter}
            onChange={e => setTypeFilter(e.target.value)}
            className="appearance-none pl-2 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            <option value="ALL">All Types</option>
            {controlTypes.map(t => <option key={t} value={t}>{t}</option>)}
          </select>
          <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
        </div>

        {/* OU dropdown */}
        <div className="relative">
          <select
            value={ouFilter}
            onChange={e => setOuFilter(e.target.value)}
            className="appearance-none pl-2 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
          >
            <option value="ALL">All Targets</option>
            {targetOUs.map(ou => <option key={ou} value={ou}>{ou}</option>)}
          </select>
          <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
        </div>

        {/* Count */}
        <span className="text-xs text-gray-500 dark:text-gray-400 whitespace-nowrap">
          {filtered.length} of {controls.length}
        </span>
      </div>

      {/* Table + Detail split */}
      <div className="flex flex-1 min-h-0">
        {/* Table */}
        <div className={`overflow-auto ${selected ? 'w-3/5 border-r border-gray-200 dark:border-gray-700' : 'w-full'}`}>
          <table className="w-full text-sm table-fixed">
            <thead className="sticky top-0 bg-gray-100 dark:bg-gray-800 text-left">
              <tr>
                <th className="px-4 py-2 font-medium relative" style={{ width: colWidths.name }}>Name{rh('name', 100)}</th>
                <th className="px-4 py-2 font-medium relative" style={{ width: colWidths.type }}>Type{rh('type', 70)}</th>
                <th className="px-4 py-2 font-medium relative" style={{ width: colWidths.enforcement }}>Enforcement{rh('enforcement', 70)}</th>
                <th className="px-4 py-2 font-medium relative" style={{ width: colWidths.ou }}>Target{rh('ou', 80)}</th>
              </tr>
            </thead>
            <tbody>
              {filtered.length === 0 && (
                <tr><td colSpan={4} className="px-4 py-6 text-center text-gray-500 dark:text-gray-400">No controls match the current filters.</td></tr>
              )}
              {filtered.map(c => (
                <tr
                  key={c.arn}
                  onClick={() => setSelected(selected?.arn === c.arn ? null : c)}
                  className={`cursor-pointer border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50
                    ${selected?.arn === c.arn ? 'bg-blue-50 dark:bg-blue-900/30' : ''}`}
                >
                  <td className="px-4 py-2 break-words">{c.name}</td>
                  <td className="px-4 py-2 break-words"><TypeBadge type={c.control_type} /></td>
                  <td className="px-4 py-2 break-words"><EnforcementBadge status={c.enforcement} /></td>
                  <td className="px-4 py-2 text-xs text-gray-500 dark:text-gray-400 break-words">{c.target_id}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Detail panel (slides in from right) */}
        {selected && (
          <div className="w-2/5 overflow-auto">
            <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
              <span className="text-sm font-medium">Detail</span>
              <button onClick={() => setSelected(null)} className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
                <X size={14} />
              </button>
            </div>
            <ControlDetail control={selected} />
          </div>
        )}
      </div>
    </div>
  );
}
