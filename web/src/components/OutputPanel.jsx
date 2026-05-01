import React, { useRef, useEffect, useState, useMemo, useCallback } from 'react';
import { ChevronDown, Copy, Check } from 'lucide-react';
import { useWebSocketLogs } from '../hooks/useWebSocketLogs';
import { formatLog, getLogColor, formatApiCall, getApiCallColor } from '../utils/formatLog';
import { filterLogsByLevel, filterByLogType } from '../utils/logFilter';

function CopyButton({ text }) {
  const [copied, setCopied] = useState(false);

  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    });
  }, [text]);

  return (
    <button
      onClick={handleCopy}
      title="Copy CLI command"
      className="ml-2 p-0.5 rounded opacity-0 group-hover:opacity-100 transition-opacity text-gray-400 hover:text-gray-200"
    >
      {copied ? <Check size={12} className="text-green-400" /> : <Copy size={12} />}
    </button>
  );
}

export default function OutputPanel() {
  const { logs, connected } = useWebSocketLogs();
  const outputRef = useRef(null);
  const apiRef = useRef(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const [levelFilter, setLevelFilter] = useState('all');
  const [activeTab, setActiveTab] = useState('output');

  const filteredLogs = useMemo(() => filterLogsByLevel(logs, levelFilter), [logs, levelFilter]);
  const apiCallLogs = useMemo(() => filterByLogType(logs, 'api_call'), [logs]);

  const activeRef = activeTab === 'output' ? outputRef : apiRef;
  const activeEntries = activeTab === 'output' ? filteredLogs : apiCallLogs;

  useEffect(() => {
    if (autoScroll && activeRef.current) {
      activeRef.current.scrollTop = activeRef.current.scrollHeight;
    }
  }, [activeEntries, autoScroll, activeRef]);

  const handleScroll = () => {
    const el = activeRef.current;
    if (!el) return;
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 30;
    setAutoScroll(atBottom);
  };

  const tabClass = (tab) =>
    `px-3 py-1 text-xs font-medium cursor-pointer border-b-2 ${
      activeTab === tab
        ? 'border-blue-500 text-blue-600 dark:text-blue-400'
        : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
    }`;

  return (
    <div className="h-48 border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 flex flex-col">
      <div className="flex items-center justify-between px-3 border-b border-gray-200 dark:border-gray-700 text-xs">
        <div className="flex items-center gap-0">
          <button className={tabClass('output')} onClick={() => setActiveTab('output')}>
            Output
          </button>
          <button className={tabClass('api_calls')} onClick={() => setActiveTab('api_calls')}>
            API Calls
            {apiCallLogs.length > 0 && (
              <span className="ml-1.5 px-1.5 py-0 rounded-full bg-cyan-100 dark:bg-cyan-900 text-cyan-700 dark:text-cyan-300 text-[10px]">
                {apiCallLogs.length}
              </span>
            )}
          </button>
        </div>
        <div className="flex items-center gap-3">
          {activeTab === 'output' && (
            <div className="relative">
              <select
                value={levelFilter}
                onChange={e => setLevelFilter(e.target.value)}
                className="appearance-none pl-2 pr-6 py-0.5 text-xs rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 focus:outline-none focus:ring-1 focus:ring-blue-500"
              >
                <option value="all">All Levels</option>
                <option value="error">Error</option>
                <option value="warning">Warning</option>
                <option value="info">Info</option>
              </select>
              <ChevronDown size={12} className="absolute right-1 top-1/2 -translate-y-1/2 text-gray-400 pointer-events-none" />
            </div>
          )}
          <span className={connected ? 'text-green-600 dark:text-green-400' : 'text-red-500 dark:text-red-400'}>
            {connected ? '● Connected' : '○ Disconnected'}
          </span>
        </div>
      </div>

      {/* Output tab */}
      <div
        ref={outputRef}
        onScroll={handleScroll}
        className={`flex-1 overflow-y-auto px-3 py-1 font-mono text-xs leading-5 text-gray-700 dark:text-gray-300 ${activeTab !== 'output' ? 'hidden' : ''}`}
      >
        {filteredLogs.length === 0 && (
          <p className="text-gray-400 dark:text-gray-500 italic">
            {logs.length === 0 ? 'Waiting for log entries...' : 'No entries match the selected filter.'}
          </p>
        )}
        {filteredLogs.map((entry, i) => (
          <div key={i} className={`whitespace-pre-wrap ${getLogColor(entry)}`}>{formatLog(entry)}</div>
        ))}
      </div>

      {/* API Calls tab */}
      <div
        ref={apiRef}
        onScroll={handleScroll}
        className={`flex-1 overflow-y-auto px-3 py-1 font-mono text-xs leading-5 text-gray-700 dark:text-gray-300 ${activeTab !== 'api_calls' ? 'hidden' : ''}`}
      >
        {apiCallLogs.length === 0 && (
          <p className="text-gray-400 dark:text-gray-500 italic">No AWS API calls recorded yet.</p>
        )}
        {apiCallLogs.map((entry, i) => (
          <div key={i} className={`group flex items-start whitespace-pre-wrap ${getApiCallColor(entry)}`}>
            <span className="flex-1">{formatApiCall(entry)}</span>
            {entry.cli_command && <CopyButton text={entry.cli_command} />}
          </div>
        ))}
      </div>
    </div>
  );
}
