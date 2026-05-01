import React, { useState, useEffect } from 'react';
import { ChevronRight, ChevronDown, Loader2, AlertCircle } from 'lucide-react';

const COVERAGE_DOT = {
  red: 'bg-red-500',
  yellow: 'bg-yellow-500',
  green: 'bg-green-500',
};

function BehaviorLabel({ breakdown }) {
  const p = breakdown.PREVENTIVE || 0;
  const d = breakdown.DETECTIVE || 0;
  const r = breakdown.PROACTIVE || 0;
  if (p + d + r === 0) return null;
  return (
    <span className="text-[10px] text-gray-400 dark:text-gray-500 whitespace-nowrap">
      {p}P / {d}D / {r}R
    </span>
  );
}

function EnabledFraction({ enabled, total }) {
  return (
    <span className="text-[10px] text-gray-400 dark:text-gray-500 whitespace-nowrap">
      {enabled}/{total}
    </span>
  );
}

const DEPTH_BORDER = [
  'border-l-2 border-indigo-400 dark:border-indigo-500',   // depth 0 — Domain
  'border-l-2 border-teal-400 dark:border-teal-500',       // depth 1 — Objective
  'border-l-2 border-amber-400 dark:border-amber-500',     // depth 2 — Common Control
];

function TreeNode({ node, depth, selectedArn, onSelect, expanded, onToggle }) {
  const hasChildren = node.children && node.children.length > 0;
  const isExpanded = expanded[node.arn];
  const isSelected = selectedArn === node.arn;
  const paddingLeft = depth * 16 + 4;
  const borderClass = DEPTH_BORDER[depth] || '';

  return (
    <>
      <div
        className={`flex items-center gap-1 py-1 pr-2 cursor-pointer text-xs hover:bg-gray-100 dark:hover:bg-gray-700 ${borderClass} ${
          isSelected ? 'bg-blue-100 dark:bg-blue-900/40 font-medium' : ''
        }`}
        style={{ paddingLeft }}
        onClick={() => {
          if (hasChildren) onToggle(node.arn);
          onSelect(node);
        }}
      >
        {/* Expand/collapse chevron */}
        {hasChildren ? (
          <span className="flex-shrink-0 w-3">
            {isExpanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
          </span>
        ) : (
          <span className="flex-shrink-0 w-3" />
        )}

        {/* Coverage dot */}
        <span className={`flex-shrink-0 w-2 h-2 rounded-full ${COVERAGE_DOT[node.coverage_status] || COVERAGE_DOT.red}`} />

        {/* Name */}
        <span className="truncate flex-1 min-w-0">
          <span className="text-gray-400 dark:text-gray-500 mr-1">{node.number}</span>
          {node.name}
        </span>

        {/* Behavior breakdown */}
        <BehaviorLabel breakdown={node.behavior_breakdown} />

        {/* Enabled fraction */}
        <EnabledFraction enabled={node.enabled_count} total={node.total_count} />
      </div>

      {/* Children */}
      {isExpanded && hasChildren && node.children.map(child => (
        <TreeNode
          key={child.arn}
          node={child}
          depth={depth + 1}
          selectedArn={selectedArn}
          onSelect={onSelect}
          expanded={expanded}
          onToggle={onToggle}
        />
      ))}
    </>
  );
}

export default function OntologyTreePanel({ tree, loading, error, selectedArn, onSelectNode, expandArns, onExpandHandled }) {
  const [expanded, setExpanded] = useState({});

  useEffect(() => {
    if (expandArns && expandArns.length > 0) {
      setExpanded(prev => {
        const next = { ...prev };
        for (const arn of expandArns) {
          next[arn] = true;
        }
        return next;
      });
      onExpandHandled?.();
    }
  }, [expandArns, onExpandHandled]);

  const handleToggle = (arn) => {
    setExpanded(prev => ({ ...prev, [arn]: !prev[arn] }));
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-gray-400">
        <Loader2 size={20} className="animate-spin" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center gap-2 p-3 text-red-500 text-sm">
        <AlertCircle size={16} />
        <span>{error}</span>
      </div>
    );
  }

  if (!tree || tree.length === 0) {
    return (
      <div className="p-3 text-sm text-gray-400">
        No ontology data available. Try refreshing the catalog.
      </div>
    );
  }

  return (
    <div className="overflow-y-auto h-full">
      {/* Hierarchy legend */}
      <div className="flex items-center gap-1.5 px-3 py-1.5 border-b border-gray-200 dark:border-gray-700 text-[10px] text-gray-400 dark:text-gray-500">
        <span className="w-2 h-2 rounded-sm bg-indigo-400 dark:bg-indigo-500 flex-shrink-0" />
        <span>Domain</span>
        <span className="mx-0.5">→</span>
        <span className="w-2 h-2 rounded-sm bg-teal-400 dark:bg-teal-500 flex-shrink-0" />
        <span>Objective</span>
        <span className="mx-0.5">→</span>
        <span className="w-2 h-2 rounded-sm bg-amber-400 dark:bg-amber-500 flex-shrink-0" />
        <span>Common Control</span>
      </div>
      <div className="py-1">
      {tree.map(domain => (
        <TreeNode
          key={domain.arn}
          node={domain}
          depth={0}
          selectedArn={selectedArn}
          onSelect={onSelectNode}
          expanded={expanded}
          onToggle={handleToggle}
        />
      ))}
      </div>
    </div>
  );
}
