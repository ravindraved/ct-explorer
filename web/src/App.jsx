import React, { useState, useCallback } from 'react';
import { Loader2 } from 'lucide-react';
import { useTheme } from './hooks/useTheme';
import { useAuthStatus } from './hooks/useAuthStatus';
import { useAuth } from './hooks/useAuth';
import Sidebar from './components/Sidebar';
import OutputPanel from './components/OutputPanel';
import StatusBar from './components/StatusBar';
import OrgTreeView from './features/organization/OrgTreeView';
import ControlsListView from './features/controls/ControlsListView';
import ScpListView from './features/scps/ScpListView';
import SearchView from './features/search/SearchView';
import SettingsView from './features/settings/SettingsView';
import CatalogView from './features/catalog/CatalogView';
import OntologyView from './features/ontology/OntologyView';

export default function App() {
  useTheme();
  const { loading: authLoading, authenticated } = useAuth();
  const auth = useAuthStatus();
  const [activeView, setActiveView] = useState('organization');
  const [pendingNavigation, setPendingNavigation] = useState(null);

  const navigateTo = useCallback((view, payload) => {
    setPendingNavigation(payload);
    setActiveView(view);
  }, []);

  const clearPendingNavigation = useCallback(() => {
    setPendingNavigation(null);
  }, []);

  // Show loading while auth is initializing
  if (authLoading || !authenticated) {
    return (
      <div className="flex items-center justify-center h-screen bg-white dark:bg-gray-900 text-gray-400">
        <Loader2 size={24} className="animate-spin mr-2" />
        <span>Authenticating...</span>
      </div>
    );
  }

  if (!auth.backendReady && auth.loading) {
    return (
      <div className="h-screen flex items-center justify-center bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
        <div className="text-center space-y-4">
          <Loader2 className="w-10 h-10 animate-spin mx-auto text-blue-500" />
          <p className="text-lg font-medium">Initializing CT Explorer...</p>
          <p className="text-sm text-gray-500 dark:text-gray-400">Waiting for backend to start</p>
        </div>
      </div>
    );
  }

  const viewProps = {};
  if (activeView === 'ontology') {
    viewProps.pendingNavigation = pendingNavigation;
    viewProps.onNavigationHandled = clearPendingNavigation;
  }
  if (activeView === 'catalog' || activeView === 'search') {
    viewProps.navigateTo = navigateTo;
  }

  const VIEWS = {
    organization: OrgTreeView,
    controls: ControlsListView,
    catalog: CatalogView,
    ontology: OntologyView,
    scps: ScpListView,
    search: SearchView,
    settings: SettingsView,
  };

  const ActiveComponent = VIEWS[activeView] || OrgTreeView;

  return (
    <div className="h-screen flex flex-col bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100">
      <div className="flex flex-1 min-h-0">
        <Sidebar
          activeView={activeView}
          onNavigate={setActiveView}
        />
        <main className="flex-1 overflow-hidden">
          <ActiveComponent {...viewProps} />
        </main>
      </div>
      <OutputPanel />
      <StatusBar auth={auth} />
    </div>
  );
}
