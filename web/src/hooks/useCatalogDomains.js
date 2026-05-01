import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useCatalogDomains() {
  const [domains, setDomains] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchDomains = useCallback(async () => {
    setLoading(true);
    const { data, error: err } = await apiGet('/api/catalog/domains');
    if (err) {
      setError(err);
    } else {
      setDomains(data);
      setError(null);
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchDomains(); }, [fetchDomains]);

  return { domains, loading, error };
}
