import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useControls(ouId = null) {
  const [controls, setControls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchControls = useCallback(async () => {
    setLoading(true);
    const path = ouId ? `/api/controls?ou_id=${ouId}` : '/api/controls';
    const { data, error: err } = await apiGet(path);
    if (err) {
      setError(err);
    } else {
      setControls(data);
      setError(null);
    }
    setLoading(false);
  }, [ouId]);

  useEffect(() => { fetchControls(); }, [fetchControls]);

  return { controls, loading, error, refetch: fetchControls };
}
