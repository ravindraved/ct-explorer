import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useCatalogCommonControls(objectiveArn = null) {
  const [commonControls, setCommonControls] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchCommonControls = useCallback(async () => {
    setLoading(true);
    const path = objectiveArn
      ? `/api/catalog/common-controls?objective_arn=${encodeURIComponent(objectiveArn)}`
      : '/api/catalog/common-controls';
    const { data, error: err } = await apiGet(path);
    if (err) {
      setError(err);
    } else {
      setCommonControls(data);
      setError(null);
    }
    setLoading(false);
  }, [objectiveArn]);

  useEffect(() => { fetchCommonControls(); }, [fetchCommonControls]);

  return { commonControls, loading, error };
}
