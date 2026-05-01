import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useNodeControls(arn) {
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const fetchDetail = useCallback(async () => {
    if (!arn) {
      setDetail(null);
      return;
    }
    setLoading(true);
    const { data, error: err } = await apiGet(`/api/ontology/node/${arn}/controls`);
    if (err) {
      setError(err);
    } else {
      setDetail(data);
      setError(null);
    }
    setLoading(false);
  }, [arn]);

  useEffect(() => { fetchDetail(); }, [fetchDetail]);

  return { detail, loading, error };
}
