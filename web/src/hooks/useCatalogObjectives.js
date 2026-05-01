import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useCatalogObjectives(domainArn = null) {
  const [objectives, setObjectives] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchObjectives = useCallback(async () => {
    setLoading(true);
    const path = domainArn
      ? `/api/catalog/objectives?domain_arn=${encodeURIComponent(domainArn)}`
      : '/api/catalog/objectives';
    const { data, error: err } = await apiGet(path);
    if (err) {
      setError(err);
    } else {
      setObjectives(data);
      setError(null);
    }
    setLoading(false);
  }, [domainArn]);

  useEffect(() => { fetchObjectives(); }, [fetchObjectives]);

  return { objectives, loading, error };
}
