import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useScps() {
  const [scps, setScps] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchScps = useCallback(async () => {
    setLoading(true);
    const { data, error: err } = await apiGet('/api/scps');
    if (err) {
      setError(err);
    } else {
      setScps(data);
      setError(null);
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchScps(); }, [fetchScps]);

  return { scps, loading, error, refetch: fetchScps };
}
