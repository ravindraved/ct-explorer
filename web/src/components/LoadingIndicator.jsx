import React from 'react';
import { Loader2 } from 'lucide-react';

export default function LoadingIndicator({ message = 'Loading...' }) {
  return (
    <div className="flex items-center justify-center gap-2 p-8 text-gray-500 dark:text-gray-400">
      <Loader2 size={20} className="animate-spin" />
      <span className="text-sm">{message}</span>
    </div>
  );
}
