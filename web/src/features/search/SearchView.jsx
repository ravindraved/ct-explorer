import React from 'react';
import { Search } from 'lucide-react';
import { useSearchEngine } from '../../hooks/useSearchEngine';
import QueryPanel from './QueryPanel';
import ResultPanel from './ResultPanel';

export default function SearchView({ navigateTo }) {
  const engine = useSearchEngine();

  return (
    <div className="flex flex-col h-full bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex-shrink-0">
        <Search size={20} />
        <h1 className="text-lg font-semibold">Search</h1>
      </div>
      <div className="flex flex-1 min-h-0">
        <div className="w-80 flex-shrink-0 border-r border-gray-200 dark:border-gray-700 overflow-y-auto">
          <QueryPanel engine={engine} />
        </div>
        <div className="flex-1 min-w-0 overflow-y-auto">
          <ResultPanel engine={engine} navigateTo={navigateTo} />
        </div>
      </div>
    </div>
  );
}
