import React, { useState, useMemo } from 'react';
import { Library, Search, X, ChevronDown } from 'lucide-react';
import { useCatalogControls } from '../../hooks/useCatalogControls';
import { useEnabledMap } from '../../hooks/useEnabledMap';
import { useCatalogServices } from '../../hooks/useCatalogServices';
import { useOntologyMap } from '../../hooks/useOntologyMap';
import LoadingIndicator from '../../components/LoadingIndicator';
import CatalogFacetPanel from './CatalogFacetPanel';
import CatalogTable from './CatalogTable';
import CatalogDetailPanel from './CatalogDetailPanel';

function toggleArrayItem(arr, item) {
  return arr.includes(item) ? arr.filter(x => x !== item) : [...arr, item];
}

export default function CatalogView({ navigateTo }) {
  const { controls, loading, error, refetch } = useCatalogControls();
  const { enabledMap } = useEnabledMap();
  const { services } = useCatalogServices();
  const { ontologyMap } = useOntologyMap();

  const [selected, setSelected] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [selectedBehaviors, setSelectedBehaviors] = useState([]);
  const [selectedSeverities, setSelectedSeverities] = useState([]);
  const [selectedServices, setSelectedServices] = useState([]);
  const [behaviorFilter, setBehaviorFilter] = useState('ALL');
  const [severityFilter, setSeverityFilter] = useState('ALL');
  const [implTypeFilter, setImplTypeFilter] = useState('ALL');
  const [serviceFilter, setServiceFilter] = useState('ALL');
  const [enabledFilter, setEnabledFilter] = useState('ALL');

  // Derive unique values for quick-filter dropdowns
  const implTypes = useMemo(() => [...new Set(controls.map(c => c.implementation_type).filter(Boolean))].sort(), [controls]);
  const serviceNames = useMemo(() => services.map(s => s.service).sort((a, b) => {
    if (a === 'Cross-Service') return 1;
    if (b === 'Cross-Service') return -1;
    return a.localeCompare(b);
  }), [services]);

  // Apply all filters
  const filtered = useMemo(() => {
    return controls.filter(c => {
      // Facet panel checkbox filters
      if (selectedBehaviors.length > 0 && !selectedBehaviors.includes(c.behavior)) return false;
      if (selectedSeverities.length > 0 && !selectedSeverities.includes(c.severity)) return false;
      if (selectedServices.length > 0) {
        const controlServices = c.services || [];
        if (!selectedServices.some(s => controlServices.includes(s))) return false;
      }
      // Quick-filter dropdowns
      if (behaviorFilter !== 'ALL' && c.behavior !== behaviorFilter) return false;
      if (severityFilter !== 'ALL' && c.severity !== severityFilter) return false;
      if (implTypeFilter !== 'ALL' && c.implementation_type !== implTypeFilter) return false;
      if (serviceFilter !== 'ALL') {
        const controlServices = c.services || [];
        if (!controlServices.includes(serviceFilter)) return false;
      }
      // Text search
      if (searchText) {
        const q = searchText.toLowerCase();
        if (!c.name.toLowerCase().includes(q)
          && !c.description.toLowerCase().includes(q)
          && !(c.aliases || []).some(a => a.toLowerCase().includes(q))) return false;
      }
      // Enabled filter
      if (enabledFilter !== 'ALL') {
        const isEnabled = !!(enabledMap[c.arn] && enabledMap[c.arn].length > 0);
        if (enabledFilter === 'ENABLED' && !isEnabled) return false;
        if (enabledFilter === 'NOT_ENABLED' && isEnabled) return false;
      }
      return true;
    });
  }, [controls, selectedBehaviors, selectedSeverities, selectedServices, behaviorFilter, severityFilter, implTypeFilter, serviceFilter, enabledFilter, enabledMap, searchText]);

  if (loading) return <LoadingIndicator message="Loading catalog data..." />;
  if (error) return <p className="p-4 text-red-500">{error}</p>;

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
        <div className="flex items-center gap-2">
          <Library size={20} />
          <h1 className="text-lg font-semibold">Catalog</h1>
        </div>
      </div>

      {/* Filter bar */}
      <div className="flex items-center gap-3 px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex-shrink-0">
        <div className="relative flex-1 max-w-xs">
          <Search size={14} className="absolute left-2 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search controls..."
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

        <div className="relative">
          <select value={behaviorFilter} onChange={e => setBehaviorFilter(e.target.value)}
            className="appearance-none pl-2 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500">
            <option value="ALL">All Behaviors</option>
            <option value="PREVENTIVE">Preventive</option>
            <option value="DETECTIVE">Detective</option>
            <option value="PROACTIVE">Proactive</option>
          </select>
          <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
        </div>

        <div className="relative">
          <select value={severityFilter} onChange={e => setSeverityFilter(e.target.value)}
            className="appearance-none pl-2 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500">
            <option value="ALL">All Severities</option>
            <option value="CRITICAL">Critical</option>
            <option value="HIGH">High</option>
            <option value="MEDIUM">Medium</option>
            <option value="LOW">Low</option>
          </select>
          <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
        </div>

        <div className="relative">
          <select value={implTypeFilter} onChange={e => setImplTypeFilter(e.target.value)}
            className="appearance-none pl-2 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500">
            <option value="ALL">All Impl Types</option>
            {implTypes.map(t => <option key={t} value={t}>{t}</option>)}
          </select>
          <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
        </div>

        <div className="relative">
          <select value={serviceFilter} onChange={e => setServiceFilter(e.target.value)}
            className="appearance-none pl-2 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500">
            <option value="ALL">All Services</option>
            {serviceNames.map(s => <option key={s} value={s}>{s}</option>)}
          </select>
          <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
        </div>

        <div className="relative">
          <select value={enabledFilter} onChange={e => setEnabledFilter(e.target.value)}
            className="appearance-none pl-2 pr-7 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500">
            <option value="ALL">All Enabled</option>
            <option value="ENABLED">Enabled</option>
            <option value="NOT_ENABLED">Not Enabled</option>
          </select>
          <ChevronDown size={14} className="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
        </div>
      </div>

      {/* Main content: facet + table + detail */}
      <div className="flex flex-1 min-h-0">
        <CatalogFacetPanel
          services={services}
          controls={controls}
          selectedBehaviors={selectedBehaviors}
          onToggleBehavior={v => setSelectedBehaviors(prev => toggleArrayItem(prev, v))}
          selectedSeverities={selectedSeverities}
          onToggleSeverity={v => setSelectedSeverities(prev => toggleArrayItem(prev, v))}
          selectedServices={selectedServices}
          onToggleService={v => setSelectedServices(prev => toggleArrayItem(prev, v))}
        />

        <CatalogTable
          controls={filtered}
          totalCount={controls.length}
          selected={selected}
          onSelect={setSelected}
          enabledMap={enabledMap}
          ontologyMap={ontologyMap}
        />

        {selected && (
          <CatalogDetailPanel
            control={selected}
            enabledMap={enabledMap}
            ontologyMap={ontologyMap}
            navigateTo={navigateTo}
            onClose={() => setSelected(null)}
          />
        )}
      </div>
    </div>
  );
}
