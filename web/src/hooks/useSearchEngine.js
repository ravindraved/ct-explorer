import { useState, useCallback, useRef } from 'react';
import { apiGet } from '../api/client';

const PAGE_SIZE = 50;

export function useSearchEngine() {
  const [mode, setModeState] = useState('find');
  const [results, setResults] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [page, setPageState] = useState(1);
  const [pageCount, setPageCount] = useState(0);
  const [total, setTotal] = useState(0);

  // Track last query so setPage can re-fetch
  const lastQuery = useRef(null);

  const find = useCallback(async (q, filters = {}, pg = 1) => {
    setLoading(true);
    setError(null);
    lastQuery.current = { type: 'find', q, filters, page: pg };

    const params = new URLSearchParams();
    if (q) params.set('q', q);
    if (filters.entity_types?.length) params.set('entity_types', filters.entity_types.join(','));
    if (filters.behaviors?.length) params.set('behaviors', filters.behaviors.join(','));
    if (filters.severities?.length) params.set('severities', filters.severities.join(','));
    params.set('page', String(pg));
    params.set('page_size', String(PAGE_SIZE));

    const { data, error: err } = await apiGet(`/api/search/find?${params}`);
    if (err) {
      setError(err);
      setResults(null);
    } else {
      setResults(data);
      setPageState(data.page);
      setPageCount(data.page_count);
      setTotal(data.total);
    }
    setLoading(false);
  }, []);

  const coverage = useCallback(async (targetId, targetType) => {
    setLoading(true);
    setError(null);
    lastQuery.current = { type: 'coverage', targetId, targetType };

    const params = new URLSearchParams({ target_id: targetId, target_type: targetType });
    const { data, error: err } = await apiGet(`/api/search/coverage?${params}`);
    if (err) {
      setError(err);
      setResults(null);
    } else {
      setResults(data);
    }
    setLoading(false);
  }, []);

  const path = useCallback(async (accountId) => {
    setLoading(true);
    setError(null);
    lastQuery.current = { type: 'path', accountId };

    const params = new URLSearchParams({ account_id: accountId });
    const { data, error: err } = await apiGet(`/api/search/path?${params}`);
    if (err) {
      setError(err);
      setResults(null);
    } else {
      setResults(data);
    }
    setLoading(false);
  }, []);

  const quickQuery = useCallback(async (preset, pg = 1) => {
    setLoading(true);
    setError(null);
    lastQuery.current = { type: 'quick_query', preset, page: pg };

    const params = new URLSearchParams({
      preset,
      page: String(pg),
      page_size: String(PAGE_SIZE),
    });
    const { data, error: err } = await apiGet(`/api/search/quick-query?${params}`);
    if (err) {
      setError(err);
      setResults(null);
    } else {
      setResults(data);
      setPageState(data.page);
      setPageCount(data.page_count);
      setTotal(data.total);
    }
    setLoading(false);
  }, []);

  const setMode = useCallback((newMode) => {
    setModeState(newMode);
    setResults(null);
    setError(null);
    setPageState(1);
    setPageCount(0);
    setTotal(0);
    lastQuery.current = null;
  }, []);

  const setPage = useCallback((pg) => {
    if (!lastQuery.current) return;
    const lq = lastQuery.current;
    if (lq.type === 'find') {
      find(lq.q, lq.filters, pg);
    } else if (lq.type === 'quick_query') {
      quickQuery(lq.preset, pg);
    }
  }, [find, quickQuery]);

  return {
    mode, results, loading, error, page, pageCount, total,
    find, coverage, path, quickQuery, setMode, setPage,
  };
}
