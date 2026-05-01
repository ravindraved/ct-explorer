import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useEnabledMap() {
  const [enabledMap, setEnabledMap] = useState({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchEnabledMap = useCallback(async () => {
    setLoading(true);
    const { data, error: err } = await apiGet('/api/catalog/enabled-map');
    if (err) {
      setError(err);
    } else {
      setEnabledMap(data);
      setError(null);
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchEnabledMap(); }, [fetchEnabledMap]);

  return { enabledMap, loading, error };
}
