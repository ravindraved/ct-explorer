/**
 * Thin fetch wrapper for the Go backend.
 * Returns { data, error } tuples — no exceptions thrown.
 * Attaches Cognito JWT token from sessionStorage when available.
 * Base URL comes from VITE_API_URL env var (defaults to '' for Vite proxy).
 */

import { getAuthToken } from '../hooks/useAuth';

const BASE_URL = import.meta.env.VITE_API_URL || '';

function authHeaders() {
  const token = getAuthToken();
  if (token) {
    return { 'Authorization': `Bearer ${token}` };
  }
  return {};
}

export async function apiGet(path) {
  try {
    const response = await fetch(`${BASE_URL}${path}`, {
      headers: { ...authHeaders() },
    });
    if (response.status === 401) {
      // Token expired or invalid — clear and reload to trigger login
      sessionStorage.removeItem('ct_auth_token');
      window.location.reload();
      return { data: null, error: 'Session expired' };
    }
    if (!response.ok) {
      const body = await response.json().catch(() => ({}));
      return { data: null, error: body.detail || `Error ${response.status}` };
    }
    const data = await response.json();
    return { data, error: null };
  } catch (err) {
    return { data: null, error: 'Network error — backend unreachable' };
  }
}

export async function apiPost(path, body = null) {
  try {
    const options = {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders() },
    };
    if (body) options.body = JSON.stringify(body);
    const response = await fetch(`${BASE_URL}${path}`, options);
    if (response.status === 401) {
      sessionStorage.removeItem('ct_auth_token');
      window.location.reload();
      return { data: null, error: 'Session expired' };
    }
    if (!response.ok) {
      const data = await response.json().catch(() => ({}));
      return { data: null, error: data.detail || `Error ${response.status}` };
    }
    const data = await response.json();
    return { data, error: null };
  } catch (err) {
    return { data: null, error: 'Network error — backend unreachable' };
  }
}
