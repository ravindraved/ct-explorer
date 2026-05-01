import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useOrganization() {
  const [tree, setTree] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchTree = useCallback(async () => {
    setLoading(true);
    const { data, error: err } = await apiGet('/api/organization/tree');
    if (err) {
      setError(err);
    } else {
      setTree(data.tree);
      setError(null);
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchTree(); }, [fetchTree]);

  return { tree, loading, error, refetch: fetchTree };
}
