import { useState, useEffect } from 'react';
import { apiGet } from '../api/client';

export function useAccountDetail(accountId) {
  const [detail, setDetail] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    if (!accountId) {
      setDetail(null);
      return;
    }
    setLoading(true);
    apiGet(`/api/organization/account/${accountId}`)
      .then(({ data, error: err }) => {
        if (err) setError(err);
        else { setDetail(data); setError(null); }
      })
      .finally(() => setLoading(false));
  }, [accountId]);

  return { detail, loading, error };
}
