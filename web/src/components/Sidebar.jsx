import React from 'react';
import {
  Sun, Moon, Network, ShieldCheck, Library, GitBranch, FileKey, Search,
  Settings,
} from 'lucide-react';
import { useTheme } from '../hooks/useTheme';

const NAV_ITEMS = [
  { id: 'organization', label: 'Organization', icon: Network },
  { id: 'controls', label: 'Controls', icon: ShieldCheck },
  { id: 'catalog', label: 'Catalog', icon: Library },
  { id: 'ontology', label: 'Ontology', icon: GitBranch },
  { id: 'scps', label: 'SCPs', icon: FileKey },
  { id: 'search', label: 'Search', icon: Search },
];

export default function Sidebar({ activeView, onNavigate }) {
  const { theme, setTheme } = useTheme();

  const toggleTheme = () => {
    const next = theme === 'dark' ? 'light' : theme === 'light' ? 'dark' : 'light';
    setTheme(next);
  };

  return (
    <aside className="w-56 flex-shrink-0 border-r border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex flex-col h-full">
      {/* Header: project name + theme toggle */}
      <div className="flex items-center justify-between px-3 py-3 border-b border-gray-200 dark:border-gray-700">
        <span className="text-sm font-semibold truncate">CT Explorer</span>
        <button
          onClick={toggleTheme}
          className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600"
          aria-label="Toggle theme"
        >
          {theme === 'dark' ? <Sun size={16} /> : <Moon size={16} />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 py-2" role="navigation" aria-label="Main navigation">
        {NAV_ITEMS.map(({ id, label, icon: Icon }) => (
          <button
            key={id}
            onClick={() => onNavigate(id)}
            className={`w-full flex items-center gap-2 px-3 py-2 text-sm text-left
              hover:bg-gray-200 dark:hover:bg-gray-700
              ${activeView === id ? 'bg-gray-200 dark:bg-gray-700 font-medium' : ''}`}
          >
            <Icon size={16} />
            {label}
          </button>
        ))}
      </nav>

      {/* Bottom section: settings */}
      <div className="border-t border-gray-200 dark:border-gray-700 py-2">
        <button
          onClick={() => onNavigate('settings')}
          className={`w-full flex items-center gap-2 px-3 py-2 text-sm
            hover:bg-gray-200 dark:hover:bg-gray-700
            ${activeView === 'settings' ? 'bg-gray-200 dark:bg-gray-700 font-medium' : ''}`}
        >
          <Settings size={16} />
          Settings
        </button>
      </div>
    </aside>
  );
}
