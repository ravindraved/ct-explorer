import { useState, useEffect, useCallback, useRef } from 'react';
import { apiGet } from '../api/client';

export function useAuthStatus() {
  const [state, setState] = useState({
    authenticated: false,
    accountId: null,
    region: null,
    authMode: null,
    loading: true,
    error: null,
    backendReady: false,
  });
  const retryRef = useRef(null);

  const fetchStatus = useCallback(async () => {
    const { data, error } = await apiGet('/api/auth/status');
    if (error) {
      // Backend not reachable yet — retry in 2s
      const isNetworkError = error.includes('Network error') || error.includes('unreachable');
      if (isNetworkError) {
        setState(s => ({ ...s, loading: true, backendReady: false, error: null }));
        retryRef.current = setTimeout(fetchStatus, 2000);
        return;
      }
      setState(s => ({ ...s, loading: false, backendReady: true, error }));
    } else {
      setState({
        authenticated: data.authenticated,
        accountId: data.account_id,
        region: data.region,
        authMode: data.auth_mode || null,
        loading: false,
        error: data.error,
        backendReady: true,
      });
    }
  }, []);

  useEffect(() => {
    fetchStatus();
    return () => clearTimeout(retryRef.current);
  }, [fetchStatus]);

  return { ...state, refetch: fetchStatus };
}
