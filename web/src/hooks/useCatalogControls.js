import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useCatalogControls() {
  const [controls, setControls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchControls = useCallback(async () => {
    setLoading(true);
    const { data, error: err } = await apiGet('/api/catalog/controls');
    if (err) {
      setError(err);
    } else {
      setControls(data);
      setError(null);
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchControls(); }, [fetchControls]);

  return { controls, loading, error, refetch: fetchControls };
}
