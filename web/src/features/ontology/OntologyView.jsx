import React, { useState, useEffect, useCallback } from 'react';
import { GitBranch } from 'lucide-react';
import { usePostureTree } from '../../hooks/usePostureTree';
import { useNodeControls } from '../../hooks/useNodeControls';
import OntologyTreePanel from './OntologyTreePanel';
import OntologyDetailPanel from './OntologyDetailPanel';

function findAncestorArns(tree, targetArn) {
  // Walk the tree to find the path of ARNs leading to targetArn
  for (const domain of tree) {
    if (domain.arn === targetArn) return [domain.arn];
    for (const obj of domain.children || []) {
      if (obj.arn === targetArn) return [domain.arn, obj.arn];
      for (const cc of obj.children || []) {
        if (cc.arn === targetArn) return [domain.arn, obj.arn, cc.arn];
      }
    }
  }
  return [];
}

export default function OntologyView({ pendingNavigation, onNavigationHandled }) {
  const { tree, loading, error } = usePostureTree();
  const [selectedNode, setSelectedNode] = useState(null);
  const [expandArns, setExpandArns] = useState(null);
  const { detail, loading: detailLoading, error: detailError } = useNodeControls(selectedNode?.arn);

  useEffect(() => {
    if (pendingNavigation?.arn && tree && tree.length > 0) {
      const targetArn = pendingNavigation.arn;
      const path = findAncestorArns(tree, targetArn);
      if (path.length > 0) {
        // Expand all ancestors
        setExpandArns(path);
        // Find the target node in the tree
        for (const domain of tree) {
          if (domain.arn === targetArn) { setSelectedNode(domain); break; }
          for (const obj of domain.children || []) {
            if (obj.arn === targetArn) { setSelectedNode(obj); break; }
            for (const cc of obj.children || []) {
              if (cc.arn === targetArn) { setSelectedNode(cc); break; }
            }
          }
        }
      }
      onNavigationHandled?.();
    }
  }, [pendingNavigation, tree, onNavigationHandled]);

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
        <GitBranch size={20} />
        <h1 className="text-lg font-semibold">Ontology Posture</h1>
      </div>
      <div className="flex flex-1 min-h-0">
        <div className="w-80 flex-shrink-0 border-r border-gray-200 dark:border-gray-700">
          <OntologyTreePanel
            tree={tree}
            loading={loading}
            error={error}
            selectedArn={selectedNode?.arn}
            onSelectNode={setSelectedNode}
            expandArns={expandArns}
            onExpandHandled={() => setExpandArns(null)}
          />
        </div>
        <div className="flex-1 min-w-0">
          <OntologyDetailPanel
            detail={detail}
            loading={detailLoading}
            error={detailError}
          />
        </div>
      </div>
    </div>
  );
}
