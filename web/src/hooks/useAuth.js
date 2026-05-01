import { useState, useEffect, useRef } from 'react';

/**
 * useAuth — Cognito authorization code + PKCE authentication hook.
 *
 * On mount:
 * 1. Fetches /api/auth/config to check if auth is enabled
 * 2. If disabled → returns { authenticated: true } immediately (no login)
 * 3. If enabled → checks for authorization code in URL query params
 * 4. If code present → exchanges it for tokens via Cognito /oauth2/token
 * 5. If no code → checks sessionStorage for existing token
 * 6. If no valid token → generates PKCE challenge and redirects to Cognito
 */
export function useAuth() {
  const [state, setState] = useState({ loading: true, authenticated: false, token: null, user: null });
  const initRef = useRef(false);

  useEffect(() => {
    if (initRef.current) return;
    initRef.current = true;

    async function init() {
      // 1. Check for authorization code in URL (Cognito callback)
      const urlParams = new URLSearchParams(window.location.search);
      const code = urlParams.get('code');

      if (code) {
        const config = await fetchAuthConfig();
        if (!config) {
          retry(init);
          return;
        }

        const verifier = sessionStorage.getItem('ct_pkce_verifier');
        if (verifier) {
          const tokens = await exchangeCode(config, code, verifier);
          sessionStorage.removeItem('ct_pkce_verifier');
          window.history.replaceState(null, '', window.location.pathname);

          if (tokens && tokens.id_token) {
            const user = parseToken(tokens.id_token);
            if (user && user.exp > Date.now() / 1000) {
              sessionStorage.setItem('ct_auth_token', tokens.id_token);
              setState({ loading: false, authenticated: true, token: tokens.id_token, user });
              return;
            }
          }
        }
        // Code exchange failed — clear URL and fall through to re-auth
        window.history.replaceState(null, '', window.location.pathname);
      }

      // 2. Check sessionStorage for existing token
      const stored = sessionStorage.getItem('ct_auth_token');
      if (stored) {
        const user = parseToken(stored);
        if (user && user.exp > Date.now() / 1000) {
          setState({ loading: false, authenticated: true, token: stored, user });
          return;
        }
        sessionStorage.removeItem('ct_auth_token');
      }

      // 3. Fetch auth config
      const config = await fetchAuthConfig();
      if (!config) {
        retry(init);
        return;
      }

      // 4. Auth disabled — skip login
      if (!config.enabled) {
        setState({ loading: false, authenticated: true, token: null, user: null });
        return;
      }

      // 5. No valid token — generate PKCE challenge and redirect to Cognito
      const { verifier, challenge } = await generatePKCE();
      sessionStorage.setItem('ct_pkce_verifier', verifier);

      const redirectUri = window.location.origin + '/';
      const loginUrl = config.cognito_domain + '/oauth2/authorize?' +
        'response_type=code&' +
        'client_id=' + config.client_id + '&' +
        'redirect_uri=' + encodeURIComponent(redirectUri) + '&' +
        'scope=openid+email+profile&' +
        'code_challenge_method=S256&' +
        'code_challenge=' + challenge;
      window.location.href = loginUrl;
    }

    init();
  }, []);

  const logout = () => {
    sessionStorage.removeItem('ct_auth_token');
    sessionStorage.removeItem('ct_pkce_verifier');
    fetch('/api/auth/config')
      .then(r => r.json())
      .then(config => {
        if (config.enabled) {
          const logoutUrl = config.cognito_domain + '/logout?' +
            'client_id=' + config.client_id + '&' +
            'logout_uri=' + encodeURIComponent(window.location.origin + '/');
          window.location.href = logoutUrl;
        } else {
          window.location.reload();
        }
      });
  };

  return { ...state, logout };
}

/**
 * getAuthToken returns the current JWT token from sessionStorage.
 */
export function getAuthToken() {
  return sessionStorage.getItem('ct_auth_token');
}

// ── Internal helpers ──

function retry(fn) {
  setTimeout(() => { fn(); }, 3000);
}

async function fetchAuthConfig() {
  try {
    const resp = await fetch('/api/auth/config');
    if (!resp.ok) return null;
    return await resp.json();
  } catch {
    return null;
  }
}

/**
 * Exchange authorization code for tokens via Cognito /oauth2/token endpoint.
 */
async function exchangeCode(config, code, verifier) {
  const tokenUrl = config.cognito_domain + '/oauth2/token';
  const redirectUri = window.location.origin + '/';

  const body = new URLSearchParams({
    grant_type: 'authorization_code',
    client_id: config.client_id,
    code: code,
    redirect_uri: redirectUri,
    code_verifier: verifier,
  });

  try {
    const resp = await fetch(tokenUrl, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: body.toString(),
    });
    if (!resp.ok) return null;
    return await resp.json();
  } catch {
    return null;
  }
}

/**
 * Generate PKCE code verifier and challenge (S256).
 * Uses Web Crypto API — available in all modern browsers.
 */
async function generatePKCE() {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  const verifier = base64UrlEncode(array);

  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  const challenge = base64UrlEncode(new Uint8Array(digest));

  return { verifier, challenge };
}

function base64UrlEncode(bytes) {
  let str = '';
  for (const b of bytes) {
    str += String.fromCharCode(b);
  }
  return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function parseToken(token) {
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));
    return { email: payload.email, sub: payload.sub, exp: payload.exp };
  } catch {
    return null;
  }
}
