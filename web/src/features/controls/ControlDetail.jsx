import React from 'react';

function Field({ label, value }) {
  return (
    <div className="mb-3">
      <dt className="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide">{label}</dt>
      <dd className="text-sm mt-0.5 break-all">{value}</dd>
    </div>
  );
}

export default function ControlDetail({ control }) {
  return (
    <div className="p-4">
      <h3 className="text-sm font-semibold mb-4 border-b border-gray-200 dark:border-gray-700 pb-2">
        Control Detail
      </h3>
      <dl>
        <Field label="Name" value={control.name} />
        <Field label="ARN" value={control.arn} />
        <Field label="Control ID" value={control.control_id} />
        <Field label="Type" value={control.control_type} />
        <Field label="Enforcement" value={control.enforcement} />
        <Field label="Target" value={control.target_id} />
        <Field label="Description" value={control.description || '—'} />
      </dl>
    </div>
  );
}
