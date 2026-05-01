import { useState, useRef, useCallback } from 'react';
import { apiPost, apiGet } from '../api/client';

export function useCatalogRefresh() {
  const [status, setStatus] = useState('idle');
  const [error, setError] = useState(null);
  const pollingRef = useRef(null);

  const stopPolling = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current);
      pollingRef.current = null;
    }
  }, []);

  const pollStatus = useCallback(() => {
    pollingRef.current = setInterval(async () => {
      const { data } = await apiGet('/api/catalog/refresh/status');
      if (!data) return;
      setStatus(data.status);
      setError(data.error);
      if (data.status === 'completed' || data.status === 'failed' || data.status === 'idle') {
        stopPolling();
      }
    }, 1500);
  }, [stopPolling]);

  const refreshCatalog = useCallback(async () => {
    const { error: err } = await apiPost('/api/catalog/refresh');
    if (err) {
      setError(err);
      return;
    }
    setStatus('in_progress');
    setError(null);
    pollStatus();
  }, [pollStatus]);

  return { refreshCatalog, status, error };
}
