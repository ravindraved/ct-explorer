import { useState, useEffect, useRef, useCallback } from 'react';
import { apiGet, apiPost } from '../api/client';

const POLL_INTERVAL = 2000;

export function useRefreshStatus() {
  const [orgStatus, setOrgStatus] = useState('idle');
  const [orgLastRefreshed, setOrgLastRefreshed] = useState(null);
  const [orgError, setOrgError] = useState(null);

  const [catalogStatus, setCatalogStatus] = useState('idle');
  const [catalogLastRefreshed, setCatalogLastRefreshed] = useState(null);
  const [catalogError, setCatalogError] = useState(null);

  const pollingRef = useRef(null);

  const fetchStatuses = useCallback(async () => {
    const [org, cat] = await Promise.all([
      apiGet('/api/refresh/status'),
      apiGet('/api/catalog/refresh/status'),
    ]);
    if (org.data) {
      setOrgStatus(org.data.status);
      setOrgLastRefreshed(org.data.last_refreshed_at);
      setOrgError(org.data.error);
    }
    if (cat.data) {
      setCatalogStatus(cat.data.status);
      setCatalogLastRefreshed(cat.data.last_refreshed_at);
      setCatalogError(cat.data.error);
    }
    return { org: org.data, cat: cat.data };
  }, []);

  const startPolling = useCallback(() => {
    if (pollingRef.current) return;
    pollingRef.current = setInterval(async () => {
      const { org, cat } = await fetchStatuses();
      const orgDone = !org || org.status !== 'in_progress';
      const catDone = !cat || cat.status !== 'in_progress';
      if (orgDone && catDone) {
        clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    }, POLL_INTERVAL);
  }, [fetchStatuses]);

  const refreshOrg = useCallback(async () => {
    const { error } = await apiPost('/api/refresh');
    if (error) { setOrgError(error); return; }
    setOrgStatus('in_progress');
    setOrgError(null);
    startPolling();
  }, [startPolling]);

  const refreshCatalog = useCallback(async () => {
    const { error } = await apiPost('/api/catalog/refresh');
    if (error) { setCatalogError(error); return; }
    setCatalogStatus('in_progress');
    setCatalogError(null);
    startPolling();
  }, [startPolling]);

  const refreshAll = useCallback(async () => {
    const [orgRes, catRes] = await Promise.all([
      apiPost('/api/refresh'),
      apiPost('/api/catalog/refresh'),
    ]);
    if (orgRes.error) setOrgError(orgRes.error);
    else { setOrgStatus('in_progress'); setOrgError(null); }
    if (catRes.error) setCatalogError(catRes.error);
    else { setCatalogStatus('in_progress'); setCatalogError(null); }
    startPolling();
  }, [startPolling]);

  // Fetch initial statuses on mount
  useEffect(() => {
    fetchStatuses();
    return () => { if (pollingRef.current) clearInterval(pollingRef.current); };
  }, [fetchStatuses]);

  return {
    orgStatus, orgLastRefreshed, orgError,
    catalogStatus, catalogLastRefreshed, catalogError,
    refreshOrg, refreshCatalog, refreshAll,
  };
}
