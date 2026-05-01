import { useState, useEffect, useCallback } from 'react';
import { apiGet } from '../api/client';

export function useCatalogServices() {
  const [services, setServices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchServices = useCallback(async () => {
    setLoading(true);
    const { data, error: err } = await apiGet('/api/catalog/services');
    if (err) {
      setError(err);
    } else {
      setServices(data);
      setError(null);
    }
    setLoading(false);
  }, []);

  useEffect(() => { fetchServices(); }, [fetchServices]);

  return { services, loading, error };
}
