import React, { useState } from 'react';
import { Search, Filter, Zap, ChevronDown } from 'lucide-react';

const ENTITY_TYPES = [
  { value: 'ou', label: 'OU' },
  { value: 'account', label: 'Account' },
  { value: 'control', label: 'Control' },
  { value: 'scp', label: 'SCP' },
  { value: 'catalog_control', label: 'Catalog Control' },
  { value: 'catalog_domain', label: 'Catalog Domain' },
  { value: 'catalog_objective', label: 'Catalog Objective' },
  { value: 'catalog_common_control', label: 'Common Control' },
];

const BEHAVIORS = ['PREVENTIVE', 'DETECTIVE', 'PROACTIVE'];
const SEVERITIES = ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'];

const COVERAGE_TARGET_TYPES = [
  { value: 'ou', label: 'OU' },
  { value: 'account', label: 'Account' },
  { value: 'domain', label: 'Domain' },
  { value: 'objective', label: 'Objective' },
  { value: 'common_control', label: 'Common Control' },
];

const QUICK_QUERIES = [
  { preset: 'unenabled_critical_controls', label: 'Unenabled Critical Controls' },
  { preset: 'ous_without_detective_controls', label: 'OUs Without Detective Controls' },
  { preset: 'accounts_without_scps', label: 'Accounts Without SCPs' },
  { preset: 'uncovered_ontology_objectives', label: 'Uncovered Ontology Objectives' },
];

const MODES = [
  { value: 'find', label: 'Find' },
  { value: 'coverage', label: 'Coverage' },
  { value: 'path', label: 'Path' },
];

function toggleItem(arr, item) {
  return arr.includes(item) ? arr.filter(x => x !== item) : [...arr, item];
}

export default function QueryPanel({ engine }) {
  const { mode, setMode, find, coverage, path, quickQuery } = engine;

  // Find mode state
  const [searchText, setSearchText] = useState('');
  const [entityTypes, setEntityTypes] = useState([]);
  const [behaviors, setBehaviors] = useState([]);
  const [severities, setSeverities] = useState([]);

  // Coverage mode state
  const [coverageTargetType, setCoverageTargetType] = useState('ou');
  const [coverageTargetId, setCoverageTargetId] = useState('');

  // Path mode state
  const [pathAccountId, setPathAccountId] = useState('');

  const handleFind = () => {
    find(searchText, { entity_types: entityTypes, behaviors, severities }, 1);
  };

  const handleCoverage = () => {
    if (coverageTargetId.trim()) {
      coverage(coverageTargetId.trim(), coverageTargetType);
    }
  };

  const handlePath = () => {
    if (pathAccountId.trim()) {
      path(pathAccountId.trim());
    }
  };

  const handleModeChange = (newMode) => {
    setMode(newMode);
  };

  return (
    <div className="flex flex-col h-full">
      {/* Mode tabs */}
      <div className="flex border-b border-gray-200 dark:border-gray-700">
        {MODES.map(m => (
          <button
            key={m.value}
            onClick={() => handleModeChange(m.value)}
            className={`flex-1 px-3 py-2 text-sm font-medium transition-colors
              ${mode === m.value
                ? 'text-blue-600 dark:text-blue-400 border-b-2 border-blue-600 dark:border-blue-400'
                : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              }`}
          >
            {m.label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-3 space-y-4">
        {/* Find mode */}
        {mode === 'find' && (
          <>
            <div className="relative">
              <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                type="text"
                placeholder="Search all entities..."
                value={searchText}
                onChange={e => setSearchText(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleFind()}
                className="w-full pl-8 pr-3 py-2 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>

            <div>
              <div className="flex items-center gap-1 mb-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                <Filter size={12} />
                Entity Types
              </div>
              <div className="space-y-1">
                {ENTITY_TYPES.map(et => (
                  <label key={et.value} className="flex items-center gap-2 text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 px-1 py-0.5 rounded">
                    <input
                      type="checkbox"
                      checked={entityTypes.includes(et.value)}
                      onChange={() => setEntityTypes(prev => toggleItem(prev, et.value))}
                      className="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                    />
                    {et.label}
                  </label>
                ))}
              </div>
            </div>

            <div>
              <div className="flex items-center gap-1 mb-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                <Filter size={12} />
                Behavior
              </div>
              <div className="space-y-1">
                {BEHAVIORS.map(b => (
                  <label key={b} className="flex items-center gap-2 text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 px-1 py-0.5 rounded">
                    <input
                      type="checkbox"
                      checked={behaviors.includes(b)}
                      onChange={() => setBehaviors(prev => toggleItem(prev, b))}
                      className="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                    />
                    {b.charAt(0) + b.slice(1).toLowerCase()}
                  </label>
                ))}
              </div>
            </div>

            <div>
              <div className="flex items-center gap-1 mb-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                <Filter size={12} />
                Severity
              </div>
              <div className="space-y-1">
                {SEVERITIES.map(s => (
                  <label key={s} className="flex items-center gap-2 text-sm cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800 px-1 py-0.5 rounded">
                    <input
                      type="checkbox"
                      checked={severities.includes(s)}
                      onChange={() => setSeverities(prev => toggleItem(prev, s))}
                      className="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                    />
                    {s.charAt(0) + s.slice(1).toLowerCase()}
                  </label>
                ))}
              </div>
            </div>

            <button
              onClick={handleFind}
              className="w-full py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors"
            >
              Search
            </button>
          </>
        )}

        {/* Coverage mode */}
        {mode === 'coverage' && (
          <>
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                Target Type
              </label>
              <div className="relative">
                <select
                  value={coverageTargetType}
                  onChange={e => setCoverageTargetType(e.target.value)}
                  className="w-full appearance-none pl-3 pr-8 py-2 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
                >
                  {COVERAGE_TARGET_TYPES.map(t => (
                    <option key={t.value} value={t.value}>{t.label}</option>
                  ))}
                </select>
                <ChevronDown size={14} className="absolute right-2.5 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                ID or Name
              </label>
              <input
                type="text"
                placeholder="Enter ID or name..."
                value={coverageTargetId}
                onChange={e => setCoverageTargetId(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleCoverage()}
                className="w-full px-3 py-2 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <button
              onClick={handleCoverage}
              className="w-full py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors"
            >
              Analyze Coverage
            </button>
          </>
        )}

        {/* Path mode */}
        {mode === 'path' && (
          <>
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">
                Account ID
              </label>
              <input
                type="text"
                placeholder="Enter account ID..."
                value={pathAccountId}
                onChange={e => setPathAccountId(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handlePath()}
                className="w-full px-3 py-2 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <button
              onClick={handlePath}
              className="w-full py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded transition-colors"
            >
              Trace Path
            </button>
          </>
        )}
      </div>

      {/* Quick Queries — always visible */}
      <div className="border-t border-gray-200 dark:border-gray-700 p-3">
        <div className="flex items-center gap-1 mb-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wide">
          <Zap size={12} />
          Quick Queries
        </div>
        <div className="space-y-1.5">
          {QUICK_QUERIES.map(qq => (
            <button
              key={qq.preset}
              onClick={() => quickQuery(qq.preset, 1)}
              className="w-full text-left px-3 py-1.5 text-sm rounded border border-gray-200 dark:border-gray-600 hover:bg-blue-50 dark:hover:bg-gray-800 hover:border-blue-300 dark:hover:border-blue-500 transition-colors"
            >
              {qq.label}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
