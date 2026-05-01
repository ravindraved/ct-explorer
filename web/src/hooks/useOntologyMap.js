import { useState, useEffect } from 'react';
import { apiGet } from '../api/client';

export function useOntologyMap() {
  const [ontologyMap, setOntologyMap] = useState({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiGet('/api/ontology/control-map')
      .then(({ data, error }) => {
        if (!error && data) setOntologyMap(data);
      })
      .finally(() => setLoading(false));
  }, []);

  return { ontologyMap, loading };
}
