import React, { useState } from 'react';
import { ChevronRight, ChevronDown, FolderOpen, User } from 'lucide-react';
import { useOrganization } from '../../hooks/useOrganization';
import { useAccountDetail } from '../../hooks/useAccountDetail';
import LoadingIndicator from '../../components/LoadingIndicator';
import AccountDetailPanel from './AccountDetailPanel';

function TreeNode({ node, depth = 0, selectedId, onSelectAccount }) {
  const [expanded, setExpanded] = useState(depth < 2);
  const isOu = node.type === 'ou';
  const isAccount = node.type === 'account';
  const hasChildren = isOu && node.children && node.children.length > 0;
  const isSelected = isAccount && selectedId === node.id;

  const handleClick = () => {
    if (isAccount) {
      onSelectAccount(node);
    } else if (hasChildren) {
      setExpanded(!expanded);
    }
  };

  return (
    <div>
      <button
        onClick={handleClick}
        className={`w-full flex items-center gap-1 px-2 py-1 text-sm text-left
          hover:bg-gray-100 dark:hover:bg-gray-700
          ${isSelected ? 'bg-blue-50 dark:bg-blue-900/30 font-medium' : ''}
          ${hasChildren || isAccount ? 'cursor-pointer' : 'cursor-default'}`}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
      >
        {hasChildren ? (
          expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />
        ) : (
          <span className="w-3.5" />
        )}
        {isOu ? (
          <FolderOpen size={14} className="text-yellow-600 dark:text-yellow-400 flex-shrink-0" />
        ) : (
          <User size={14} className="text-blue-600 dark:text-blue-400 flex-shrink-0" />
        )}
        <span className="truncate">{node.name}</span>
        <span className="text-xs text-gray-400 dark:text-gray-500 ml-1 flex-shrink-0">{node.id}</span>
        {isAccount && node.status && (
          <span className={`text-xs ml-1 px-1 rounded ${
            node.status === 'ACTIVE'
              ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
              : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
          }`}>
            {node.status}
          </span>
        )}
      </button>
      {expanded && hasChildren && node.children.map((child, i) => (
        <TreeNode key={child.id || i} node={child} depth={depth + 1} selectedId={selectedId} onSelectAccount={onSelectAccount} />
      ))}
    </div>
  );
}

export default function OrgTreeView() {
  const { tree, loading, error } = useOrganization();
  const [selectedAccount, setSelectedAccount] = useState(null);
  const { detail, loading: detailLoading, error: detailError } = useAccountDetail(selectedAccount?.id);

  if (loading) return <LoadingIndicator message="Loading organization tree..." />;
  if (error) return <p className="p-4 text-red-500">{error}</p>;
  if (!tree) return <p className="p-4 text-gray-500 dark:text-gray-400">No organization data available.</p>;

  return (
    <div className="flex h-full">
      <div className="flex-1 overflow-auto">
        <h2 className="px-4 py-2 text-sm font-semibold border-b border-gray-200 dark:border-gray-700">
          Organization Tree
        </h2>
        <TreeNode
          node={tree}
          selectedId={selectedAccount?.id}
          onSelectAccount={(node) => setSelectedAccount(selectedAccount?.id === node.id ? null : node)}
        />
      </div>
      {selectedAccount && (
        <AccountDetailPanel
          detail={detail}
          loading={detailLoading}
          error={detailError}
          onClose={() => setSelectedAccount(null)}
        />
      )}
    </div>
  );
}
