import React, { useState, useEffect } from 'react';
import { RefreshCw, Loader2, Linkedin, ExternalLink, Settings } from 'lucide-react';
import { useTheme } from '../../hooks/useTheme';
import { useAuthStatus } from '../../hooks/useAuthStatus';
import { useRefreshStatus } from '../../hooks/useRefreshStatus';
import { apiGet, apiPost } from '../../api/client';

function Card({ title, icon: Icon, children, className = '' }) {
  return (
    <div className={`rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-4 ${className}`}>
      <div className="flex items-center gap-2 mb-3 pb-2 border-b border-gray-100 dark:border-gray-700">
        {Icon && <Icon size={15} className="text-gray-400 dark:text-gray-500" />}
        <h3 className="text-sm font-semibold">{title}</h3>
      </div>
      {children}
    </div>
  );
}

function ReadOnlyField({ label, value }) {
  return (
    <div className="flex items-center justify-between py-1 text-sm">
      <span className="text-gray-500 dark:text-gray-400">{label}</span>
      <span className="text-right">{value || '—'}</span>
    </div>
  );
}

function formatTimestamp(iso) {
  if (!iso) return '—';
  try { return new Date(iso).toLocaleString(); } catch { return iso; }
}

function RefreshButton({ label, onClick, busy, disabled }) {
  return (
    <button
      onClick={onClick}
      disabled={busy || disabled}
      className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded border
        border-gray-300 dark:border-gray-600
        hover:bg-gray-100 dark:hover:bg-gray-700
        disabled:opacity-50 disabled:cursor-not-allowed"
    >
      {busy ? <Loader2 size={14} className="animate-spin" /> : <RefreshCw size={14} />}
      {label}
    </button>
  );
}

/* ── Data Refresh Card ── */
function DataRefreshCard() {
  const {
    orgStatus, orgLastRefreshed, orgError,
    catalogStatus, catalogLastRefreshed, catalogError,
    refreshOrg, refreshCatalog, refreshAll,
  } = useRefreshStatus();

  const anyBusy = orgStatus === 'in_progress' || catalogStatus === 'in_progress';

  return (
    <Card title="Data Refresh" icon={RefreshCw}>
      <div className="flex flex-wrap gap-2 mb-3">
        <RefreshButton label="Org Data" onClick={refreshOrg} busy={orgStatus === 'in_progress'} disabled={anyBusy && orgStatus !== 'in_progress'} />
        <RefreshButton label="Catalog" onClick={refreshCatalog} busy={catalogStatus === 'in_progress'} disabled={anyBusy && catalogStatus !== 'in_progress'} />
        <RefreshButton label="Refresh All" onClick={refreshAll} busy={anyBusy} disabled={anyBusy} />
      </div>
      <div className="text-xs space-y-1.5 text-gray-500 dark:text-gray-400">
        <div className="flex justify-between">
          <span>Org Data</span>
          <span>{orgStatus === 'in_progress' ? 'Refreshing…' : formatTimestamp(orgLastRefreshed)}</span>
        </div>
        <div className="flex justify-between">
          <span>Catalog</span>
          <span>{catalogStatus === 'in_progress' ? 'Refreshing…' : formatTimestamp(catalogLastRefreshed)}</span>
        </div>
      </div>
      {orgError && <p className="text-xs text-red-500 mt-2">Org: {orgError}</p>}
      {catalogError && <p className="text-xs text-red-500 mt-1">Catalog: {catalogError}</p>}
    </Card>
  );
}

/* ── AWS Configuration Card ── */
function AwsConfigCard() {
  const auth = useAuthStatus();
  const [mode, setMode] = useState(auth.authMode || 'instance_metadata');
  const [imds, setImds] = useState(null);
  const [form, setForm] = useState({ accessKeyId: '', secretAccessKey: '', sessionToken: '', profileName: '', region: '' });
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState(null);

  useEffect(() => { if (auth.authMode) setMode(auth.authMode); }, [auth.authMode]);

  useEffect(() => {
    if (mode === 'instance_metadata') {
      apiGet('/api/auth/metadata').then(({ data }) => { if (data) setImds(data); });
    }
  }, [mode]);

  const handleConfigure = async () => {
    setSaving(true);
    setMessage(null);
    const body = { auth_mode: mode };
    if (mode === 'manual') {
      if (form.accessKeyId) body.access_key_id = form.accessKeyId;
      if (form.secretAccessKey) body.secret_access_key = form.secretAccessKey;
      if (form.sessionToken) body.session_token = form.sessionToken;
      if (form.profileName) body.profile_name = form.profileName;
      if (form.region) body.region = form.region;
    }
    const { data, error } = await apiPost('/api/auth/configure', body);
    setSaving(false);
    setForm(f => ({ ...f, accessKeyId: '', secretAccessKey: '', sessionToken: '' }));
    if (error) { setMessage({ type: 'error', text: error }); return; }
    if (data?.success) {
      setMessage({ type: 'success', text: `Connected — ${data.account_id} (${data.region})` });
      auth.refetch();
    } else {
      setMessage({ type: 'error', text: data?.error || 'Configuration failed' });
    }
  };

  const inputClass = "w-full px-2 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500";

  return (
    <Card title="AWS Configuration" icon={Settings} className="col-span-1 lg:col-span-2">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Left column: mode toggle + credentials form */}
        <div>
          <div className="flex gap-2 mb-3">
            {['instance_metadata', 'manual'].map(m => (
              <button key={m} onClick={() => setMode(m)}
                className={`px-3 py-1.5 text-sm rounded border ${mode === m
                  ? 'bg-blue-600 text-white border-blue-600'
                  : 'border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700'}`}>
                {m === 'instance_metadata' ? 'Instance Metadata' : 'Manual'}
              </button>
            ))}
          </div>

          {mode === 'instance_metadata' && imds && (
            <div className="space-y-0">
              <ReadOnlyField label="Available" value={imds.available ? 'Yes' : 'No'} />
              {imds.available && (<>
                <ReadOnlyField label="Instance ID" value={imds.instance_id} />
                <ReadOnlyField label="Instance Type" value={imds.instance_type} />
                <ReadOnlyField label="AZ" value={imds.availability_zone} />
                <ReadOnlyField label="Region" value={imds.region} />
                <ReadOnlyField label="IAM Role" value={imds.iam_role} />
                <ReadOnlyField label="Account ID" value={imds.account_id} />
              </>)}
            </div>
          )}

          {mode === 'manual' && (
            <div className="space-y-2">
              <input type="text" placeholder="Access Key ID" value={form.accessKeyId}
                onChange={e => setForm(f => ({ ...f, accessKeyId: e.target.value }))} className={inputClass} />
              <input type="password" placeholder="Secret Access Key" value={form.secretAccessKey}
                onChange={e => setForm(f => ({ ...f, secretAccessKey: e.target.value }))} className={inputClass} />
              <input type="password" placeholder="Session Token (optional)" value={form.sessionToken}
                onChange={e => setForm(f => ({ ...f, sessionToken: e.target.value }))} className={inputClass} />
              <div className="text-xs text-gray-400 dark:text-gray-500 text-center">— or —</div>
              <input type="text" placeholder="AWS Profile Name" value={form.profileName}
                onChange={e => setForm(f => ({ ...f, profileName: e.target.value }))} className={inputClass} />
              <input type="text" placeholder="Region (e.g. us-east-1)" value={form.region}
                onChange={e => setForm(f => ({ ...f, region: e.target.value }))} className={inputClass} />
            </div>
          )}

          <div className="flex gap-2 mt-3">
            <button onClick={handleConfigure} disabled={saving}
              className="px-3 py-1.5 text-sm rounded bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50">
              {saving ? 'Saving…' : 'Save & Connect'}
            </button>
            <button onClick={handleConfigure} disabled={saving}
              className="px-3 py-1.5 text-sm rounded border border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50">
              Test Connection
            </button>
          </div>

          {message && (
            <p className={`text-xs mt-2 ${message.type === 'error' ? 'text-red-500' : 'text-green-600 dark:text-green-400'}`}>
              {message.text}
            </p>
          )}
        </div>

        {/* Right column: connection status */}
        <div className="rounded-lg bg-gray-50 dark:bg-gray-900 p-3">
          <h4 className="text-xs font-semibold mb-2 text-gray-500 dark:text-gray-400 uppercase tracking-wide">Connection Status</h4>
          <div className="space-y-0">
            <ReadOnlyField label="Status" value={auth.authenticated ? '● Connected' : '○ Disconnected'} />
            <ReadOnlyField label="Account ID" value={auth.accountId} />
            <ReadOnlyField label="Region" value={auth.region} />
            <ReadOnlyField label="Auth Mode" value={auth.authMode} />
          </div>
        </div>
      </div>
    </Card>
  );
}

/* ── About Card ── */
function AboutCard() {
  return (
    <Card title="About" icon={Settings}>
      <div className="text-sm space-y-1">
        <p>Ravindra Ved</p>
        <p className="text-gray-500 dark:text-gray-400">Sr. Security SA @ AWS</p>
        <a
          href="https://www.linkedin.com/in/ravindraved"
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 text-blue-600 dark:text-blue-400 hover:underline"
        >
          <Linkedin size={14} />
          @ravindraved
          <ExternalLink size={12} />
        </a>
      </div>
    </Card>
  );
}

/* ── Main SettingsView ── */
export default function SettingsView() {
  const { theme, setTheme } = useTheme();

  return (
    <div className="p-4 overflow-auto h-full bg-gray-50 dark:bg-gray-900">
      <h2 className="text-lg font-semibold mb-4">Settings</h2>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {/* Row 1: Appearance + Data Refresh side by side */}
        <Card title="Appearance" icon={Settings}>
          <div className="flex gap-2">
            {['light', 'dark', 'system'].map(opt => (
              <button
                key={opt}
                onClick={() => setTheme(opt)}
                className={`px-3 py-1.5 text-sm rounded border
                  ${theme === opt
                    ? 'bg-blue-600 text-white border-blue-600'
                    : 'border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700'}`}
              >
                {opt.charAt(0).toUpperCase() + opt.slice(1)}
              </button>
            ))}
          </div>
        </Card>

        <DataRefreshCard />

        {/* Row 2: AWS Config spans full width */}
        <AwsConfigCard />

        {/* Row 3: About */}
        <AboutCard />
      </div>
    </div>
  );
}
