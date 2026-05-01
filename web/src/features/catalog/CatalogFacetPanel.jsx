import React, { useState, useMemo } from 'react';
import { ChevronDown, ChevronRight } from 'lucide-react';

const BEHAVIORS = ['PREVENTIVE', 'DETECTIVE', 'PROACTIVE'];
const SEVERITIES = ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW'];

function CheckboxGroup({ title, totalCount, items, selected, onToggle }) {
  const [open, setOpen] = useState(true);
  return (
    <div className="mb-3">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-1 w-full"
      >
        {open ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        {title}
        {totalCount != null && <span className="ml-auto font-normal normal-case tracking-normal text-gray-400">{totalCount}</span>}
      </button>
      {open && (
        <div className="space-y-1 pl-1">
          {items.map(({ value, count }) => (
            <label key={value} className="flex items-center gap-1.5 text-xs cursor-pointer">
              <input
                type="checkbox"
                checked={selected.includes(value)}
                onChange={() => onToggle(value)}
                className="rounded border-gray-300 dark:border-gray-600"
              />
              <span className="truncate">{value}</span>
              <span className="text-gray-400 ml-auto">{count}</span>
            </label>
          ))}
        </div>
      )}
    </div>
  );
}

export default function CatalogFacetPanel({
  services, controls,
  selectedBehaviors, onToggleBehavior,
  selectedSeverities, onToggleSeverity,
  selectedServices, onToggleService,
}) {
  // Compute counts from the full (unfiltered) controls list
  const behaviorCounts = useMemo(() => {
    const counts = {};
    BEHAVIORS.forEach(b => { counts[b] = 0; });
    controls.forEach(c => { counts[c.behavior] = (counts[c.behavior] || 0) + 1; });
    return counts;
  }, [controls]);

  const severityCounts = useMemo(() => {
    const counts = {};
    SEVERITIES.forEach(s => { counts[s] = 0; });
    controls.forEach(c => { counts[c.severity] = (counts[c.severity] || 0) + 1; });
    return counts;
  }, [controls]);

  // Sort services alphabetically, "Cross-Service" last
  const sortedServices = useMemo(() => {
    const sorted = [...services].sort((a, b) => {
      if (a.service === 'Cross-Service') return 1;
      if (b.service === 'Cross-Service') return -1;
      return a.service.localeCompare(b.service);
    });
    return sorted;
  }, [services]);

  return (
    <div className="w-52 flex-shrink-0 border-r border-gray-200 dark:border-gray-700 p-2 overflow-y-auto text-sm">
      <CheckboxGroup
        title="Behavior"
        items={BEHAVIORS.map(b => ({ value: b, count: behaviorCounts[b] || 0 }))}
        selected={selectedBehaviors}
        onToggle={onToggleBehavior}
      />
      <CheckboxGroup
        title="Severity"
        items={SEVERITIES.map(s => ({ value: s, count: severityCounts[s] || 0 }))}
        selected={selectedSeverities}
        onToggle={onToggleSeverity}
      />
      <CheckboxGroup
        title="Services"
        totalCount={sortedServices.length}
        items={sortedServices.map(s => ({ value: s.service, count: s.count }))}
        selected={selectedServices}
        onToggle={onToggleService}
      />
    </div>
  );
}
