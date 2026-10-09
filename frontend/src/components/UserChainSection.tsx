import { useEffect, useState } from 'react';
import { ChainTopology } from './ChainTopology';
import { getNodes, updateUserChain } from '../services/api';
import type { Node } from '../types/node';
import type { User } from '../types/user';

function initialExitIds(user: User): number[] {
  if (user.exit_node_ids && user.exit_node_ids.length > 0) {
    return [...user.exit_node_ids];
  }
  if (user.exit_node_id) {
    return [user.exit_node_id];
  }
  return [];
}

export function UserChainSection({
  user,
  onUpdated,
}: {
  user: User;
  onUpdated: (u: User) => void;
}) {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [entryId, setEntryId] = useState(user.entry_node_id?.toString() ?? '');
  const [exitIds, setExitIds] = useState<number[]>(() => initialExitIds(user));
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    getNodes()
      .then(setNodes)
      .catch(() => setNodes([]));
  }, []);

  useEffect(() => {
    setEntryId(user.entry_node_id?.toString() ?? '');
    setExitIds(initialExitIds(user));
  }, [user.entry_node_id, user.exit_node_id, user.exit_node_ids]);

  const entry = nodes.find((n) => n.id === Number(entryId)) ?? null;
  const exitNodes = nodes.filter((n) => n.role === 'exit');
  const selectedExits = exitIds
    .map((id) => exitNodes.find((n) => n.id === id))
    .filter((n): n is Node => Boolean(n));

  function toggleExit(id: number) {
    setExitIds((prev) => {
      if (prev.includes(id)) {
        return prev.filter((x) => x !== id);
      }
      return [...prev, id];
    });
  }

  function moveExit(id: number, dir: -1 | 1) {
    setExitIds((prev) => {
      const i = prev.indexOf(id);
      if (i < 0) return prev;
      const j = i + dir;
      if (j < 0 || j >= prev.length) return prev;
      const next = [...prev];
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });
  }

  return (
    <section data-testid="user-chain-section" className="card-pad space-y-4">
      <div>
        <h2 className="text-lg font-medium">Multi-hop chain</h2>
        <p className="text-xs text-slate-500 mt-1">
          Entry + one or more exit countries. Multiple exits appear as selectable servers in one
          subscription.
        </p>
      </div>
      <ChainTopology entry={entry} exits={selectedExits} />
      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label className="label">Entry node</label>
          <select
            data-testid="entry-node-select"
            value={entryId}
            onChange={(e) => setEntryId(e.target.value)}
            className="input"
          >
            <option value="">Auto entry</option>
            {nodes
              .filter((n) => n.role === 'entry')
              .map((n) => (
                <option key={n.id} value={n.id}>
                  {n.name}
                  {n.region ? ` (${n.region})` : ''}
                </option>
              ))}
          </select>
        </div>
        <div>
          <label className="label">Exit countries</label>
          <div
            data-testid="exit-node-multiselect"
            className="rounded-lg border border-surface-border bg-slate-950/40 divide-y divide-surface-border max-h-56 overflow-y-auto"
          >
            {exitNodes.length === 0 && (
              <p className="px-3 py-2 text-xs text-slate-500">No exit nodes</p>
            )}
            {exitNodes.map((n) => {
              const checked = exitIds.includes(n.id);
              const order = checked ? exitIds.indexOf(n.id) + 1 : null;
              return (
                <label
                  key={n.id}
                  className="flex items-center gap-2 px-3 py-2 text-sm cursor-pointer hover:bg-slate-900/60"
                >
                  <input
                    type="checkbox"
                    data-testid={`exit-node-${n.id}`}
                    checked={checked}
                    onChange={() => toggleExit(n.id)}
                    className="rounded border-slate-600"
                  />
                  <span className="flex-1 min-w-0">
                    <span className="font-medium">{n.region || n.name}</span>
                    <span className="text-slate-500 text-xs ml-1">
                      {n.name}
                      {!n.active ? ' · inactive' : ''}
                    </span>
                  </span>
                  {order != null && (
                    <span className="flex items-center gap-1 shrink-0">
                      <span className="text-[10px] uppercase tracking-wide text-slate-500">
                        #{order}
                      </span>
                      <button
                        type="button"
                        className="text-xs text-slate-400 hover:text-sky-400 px-1"
                        disabled={order === 1}
                        onClick={(e) => {
                          e.preventDefault();
                          moveExit(n.id, -1);
                        }}
                        aria-label="Move up"
                      >
                        ↑
                      </button>
                      <button
                        type="button"
                        className="text-xs text-slate-400 hover:text-sky-400 px-1"
                        disabled={order === exitIds.length}
                        onClick={(e) => {
                          e.preventDefault();
                          moveExit(n.id, 1);
                        }}
                        aria-label="Move down"
                      >
                        ↓
                      </button>
                    </span>
                  )}
                </label>
              );
            })}
          </div>
          <p className="text-[11px] text-slate-500 mt-1">
            First selected exit is primary (keeps user UUID). Empty = auto by priority.
          </p>
        </div>
      </div>
      {message && <p className="alert-ok">{message}</p>}
      {error && <p className="alert-error">{error}</p>}
      <button
        data-testid="save-chain"
        type="button"
        className="btn-primary"
        disabled={saving}
        onClick={async () => {
          setSaving(true);
          setMessage('');
          setError('');
          try {
            const updated = await updateUserChain(user.id, {
              entry_node_id: entryId ? Number(entryId) : null,
              exit_node_ids: exitIds,
            });
            onUpdated(updated);
            setMessage(
              exitIds.length > 1
                ? `Chain saved · ${exitIds.length} exits in subscription`
                : 'Chain saved',
            );
          } catch {
            setError('Failed to save chain');
          } finally {
            setSaving(false);
          }
        }}
      >
        {saving ? 'Saving…' : 'Save chain'}
      </button>
    </section>
  );
}
