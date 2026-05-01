import React from 'react';
import { X } from 'lucide-react';

function Section({ title, children }) {
  return (
    <div className="mb-3">
      <p className="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400 mb-1">{title}</p>
      {children}
    </div>
  );
}

function Badge({ text, className }) {
  return <span className={`text-xs px-1.5 py-0.5 rounded ${className}`}>{text}</span>;
}

export default function CatalogDetailPanel({ control, enabledMap, ontologyMap, navigateTo, onClose }) {
  if (!control) return null;

  const ous = enabledMap?.[control.arn] || [];
  const ontologyRefs = ontologyMap?.[control.arn] || [];

  return (
    <div className="w-2/5 flex-shrink-0 border-l border-gray-200 dark:border-gray-700 flex flex-col h-full overflow-y-auto">
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex-shrink-0">
        <span className="text-sm font-medium truncate">Detail</span>
        <button onClick={onClose} className="p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-600">
          <X size={14} />
        </button>
      </div>

      <div className="p-4 space-y-1 text-sm">
        <h2 className="font-semibold text-base mb-2">{control.name}</h2>
        <p className="text-gray-600 dark:text-gray-300 text-xs mb-3">{control.description}</p>

        <Section title="Properties">
          <div className="flex gap-2 flex-wrap">
            <Badge text={control.behavior} className="bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300" />
            <Badge text={control.severity} className="bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300" />
            {control.implementation_type && (
              <Badge text={control.implementation_type} className="bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300" />
            )}
          </div>
        </Section>

        {ontologyRefs.length > 0 && (
          <Section title="Ontology Mapping">
            <div className="text-xs space-y-1">
              {ontologyRefs.map(r => (
                <div key={r.number} className="flex items-center gap-2">
                  <button
                    onClick={() => navigateTo?.('ontology', { arn: r.arn })}
                    className="px-1.5 py-0.5 rounded bg-violet-100 text-violet-700 dark:bg-violet-900 dark:text-violet-300 font-medium whitespace-nowrap hover:bg-violet-200 dark:hover:bg-violet-800 cursor-pointer"
                    title="Go to Ontology view"
                  >
                    {r.number}
                  </button>
                  <span className="text-gray-600 dark:text-gray-300">{r.name}</span>
                </div>
              ))}
            </div>
          </Section>
        )}

        {control.services?.length > 0 && (
          <Section title="Services">
            <div className="flex gap-1 flex-wrap">
              {control.services.map(s => (
                <span key={s} className="text-xs px-1.5 py-0.5 rounded bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300">{s}</span>
              ))}
            </div>
          </Section>
        )}

        <Section title="Enabled OUs">
          {ous.length > 0 ? (
            <div className="text-xs space-y-0.5">
              {ous.map(ou => (
                <p key={ou} className="text-gray-600 dark:text-gray-300">{ou}</p>
              ))}
            </div>
          ) : (
            <p className="text-xs text-gray-400">Not enabled on any OU</p>
          )}
        </Section>

        {control.aliases?.length > 0 && (
          <Section title="Aliases">
            <div className="text-xs space-y-0.5">
              {control.aliases.map((a, i) => (
                <p key={i} className="text-gray-600 dark:text-gray-300 break-all">{a}</p>
              ))}
            </div>
          </Section>
        )}

        {control.governed_resources?.length > 0 && (
          <Section title="Governed Resources">
            <div className="text-xs space-y-0.5">
              {control.governed_resources.map((r, i) => (
                <p key={i} className="text-gray-600 dark:text-gray-300">{r}</p>
              ))}
            </div>
          </Section>
        )}

        {control.implementation_identifier && (
          <Section title="Implementation">
            <p className="text-xs text-gray-600 dark:text-gray-300 break-all">{control.implementation_identifier}</p>
          </Section>
        )}

        {control.create_time && (
          <Section title="Created">
            <p className="text-xs text-gray-600 dark:text-gray-300">{control.create_time}</p>
          </Section>
        )}
      </div>
    </div>
  );
}
