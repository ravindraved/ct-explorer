import React from 'react';
import { Loader2, Search, Shield, Network, GitBranch, Zap } from 'lucide-react';
import FindResults from './FindResults';
import CoverageResults from './CoverageResults';
import PathResults from './PathResults';

export default function ResultPanel({ engine, navigateTo }) {
  const { mode, results, loading, error } = engine;

  if (loading) {
    return (
      <div className="flex items-center justify-center gap-2 p-8 text-gray-500 dark:text-gray-400">
        <Loader2 size={20} className="animate-spin" />
        <span className="text-sm">Searching...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-4 m-4 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 rounded border border-red-200 dark:border-red-800">
        {error}
      </div>
    );
  }

  if (!results) {
    return (
      <div className="flex flex-col items-center justify-center h-full text-gray-400 dark:text-gray-500 p-8 space-y-6">
        <Search size={48} strokeWidth={1} />
        <div className="text-center space-y-4 max-w-md">
          <p className="text-lg font-medium text-gray-600 dark:text-gray-300">Search your organization</p>
          <div className="space-y-3 text-sm text-left">
            <div className="flex items-start gap-2">
              <Search size={16} className="mt-0.5 flex-shrink-0" />
              <span><span className="font-medium text-gray-600 dark:text-gray-300">Find</span> — search across all entity types by name, ID, ARN, or description</span>
            </div>
            <div className="flex items-start gap-2">
              <Shield size={16} className="mt-0.5 flex-shrink-0" />
              <span><span className="font-medium text-gray-600 dark:text-gray-300">Coverage</span> — analyze security posture and gaps for an OU, account, or ontology node</span>
            </div>
            <div className="flex items-start gap-2">
              <GitBranch size={16} className="mt-0.5 flex-shrink-0" />
              <span><span className="font-medium text-gray-600 dark:text-gray-300">Path</span> — trace the OU inheritance chain for an account</span>
            </div>
            <div className="flex items-start gap-2">
              <Zap size={16} className="mt-0.5 flex-shrink-0" />
              <span><span className="font-medium text-gray-600 dark:text-gray-300">Quick Queries</span> — one-click preset compliance checks</span>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Coverage mode
  if (mode === 'coverage' && results.mode === 'coverage') {
    return <CoverageResults results={results} navigateTo={navigateTo} />;
  }

  // Path mode
  if (mode === 'path' && results.mode === 'path') {
    return <PathResults results={results} navigateTo={navigateTo} />;
  }

  // Find mode and quick_query mode both use FindResults
  return <FindResults results={results} engine={engine} navigateTo={navigateTo} />;
}
