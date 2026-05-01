import React from 'react';
import { X, Loader2, AlertCircle } from 'lucide-react';

function Section({ title, children }) {
  return (
    <div className="mb-3">
      <p className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-1">{title}</p>
      {children}
    </div>
  );
}

function InfoRow({ label, value }) {
  return (
    <div className="flex text-xs py-0.5">
      <span className="w-24 flex-shrink-0 text-gray-500 dark:text-gray-400">{label}</span>
      <span className="text-gray-700 dark:text-gray-300 break-all">{value || '—'}</span>
    </div>
  );
}

function Badge({ text, className }) {
  return <span className={`text-xs px-1.5 py-0.5 rounded ${className}`}>{text}</span>;
}

export default function AccountDetailPanel({ detail, loading, error, onClose }) {
  return (
    <div className="w-2/5 flex-shrink-0 border-l border-gray-200 dark:border-gray-700 flex flex-col h-full overflow-y-auto">
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex-shrink-0">
        <span className="text-sm font-medium truncate">Account Detail</span>
        <button onClick={onClose} className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
          <X size={14} />
        </button>
      </div>

      {loading && (
        <div className="flex items-center justify-center h-32 text-gray-400">
          <Loader2 size={20} className="animate-spin" />
        </div>
      )}

      {error && (
        <div className="flex items-center gap-2 p-4 text-red-500 text-sm">
          <AlertCircle size={16} />
          <span>{error}</span>
        </div>
      )}

      {detail && !loading && (
        <div className="p-4 space-y-1 text-sm">
          <h2 className="font-semibold text-base mb-1">{detail.name}</h2>
          <Badge
            text={detail.status}
            className={detail.status === 'ACTIVE'
              ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
              : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'}
          />

          <Section title="Account Info">
            <InfoRow label="Account ID" value={detail.id} />
            <InfoRow label="Email" value={detail.email} />
            <InfoRow label="ARN" value={detail.arn} />
            <InfoRow label="Parent OU" value={`${detail.ou_name} (${detail.ou_id})`} />
          </Section>

          <Section title={`Inherited Controls (${detail.controls?.length || 0})`}>
            {detail.controls?.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="w-full text-xs">
                  <thead>
                    <tr className="border-b border-gray-200 dark:border-gray-700 text-left text-gray-500 dark:text-gray-400">
                      <th className="py-1.5 pr-2 font-medium">Name</th>
                      <th className="py-1.5 pr-2 font-medium">Type</th>
                      <th className="py-1.5 font-medium">Enforcement</th>
                    </tr>
                  </thead>
                  <tbody>
                    {detail.controls.map(c => (
                      <tr key={c.arn} className="border-b border-gray-100 dark:border-gray-800">
                        <td className="py-1 pr-2 max-w-xs truncate">{c.name}</td>
                        <td className="py-1 pr-2">
                          <Badge text={c.control_type} className="bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300" />
                        </td>
                        <td className="py-1">
                          <Badge text={c.enforcement} className={c.enforcement === 'ENABLED'
                            ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                            : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'} />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <p className="text-xs text-gray-400">No controls inherited from parent OU</p>
            )}
          </Section>

          <Section title={`SCPs (${detail.scps?.length || 0})`}>
            {detail.scps?.length > 0 ? (
              <div className="text-xs space-y-1">
                {detail.scps.map(s => (
                  <div key={s.id} className="p-2 rounded border border-gray-200 dark:border-gray-700">
                    <p className="font-medium">{s.name}</p>
                    <p className="text-gray-500 dark:text-gray-400 text-[10px]">{s.id}</p>
                    {s.description && <p className="text-gray-400 mt-0.5">{s.description}</p>}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-gray-400">No SCPs targeting this account</p>
            )}
          </Section>
        </div>
      )}
    </div>
  );
}
