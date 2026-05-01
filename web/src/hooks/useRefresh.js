import { useState, useRef, useCallback } from 'react';
import { apiPost, apiGet } from '../api/client';

export function useRefresh() {
  const [status, setStatus] = useState('idle');
  const [phase, setPhase] = useState(null);
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
      const { data } = await apiGet('/api/refresh/status');
      if (!data) return;
      setStatus(data.status);
      setPhase(data.phase);
      setError(data.error);
      if (data.status === 'completed' || data.status === 'failed' || data.status === 'idle') {
        stopPolling();
      }
    }, 1500);
  }, [stopPolling]);

  const refresh = useCallback(async () => {
    const { data, error: err } = await apiPost('/api/refresh');
    if (err) {
      setError(err);
      return;
    }
    setStatus('in_progress');
    setPhase('starting');
    setError(null);
    pollStatus();
  }, [pollStatus]);

  return { refresh, status, phase, error };
}
