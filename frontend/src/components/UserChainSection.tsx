import { useEffect, useState } from 'react';
import { ChainTopology } from './ChainTopology';
import { getNodes, updateUserChain } from '../services/api';
import type { Node } from '../types/node';
import type { User } from '../types/user';

export function UserChainSection({
  user,
  onUpdated,
}: {
  user: User;
  onUpdated: (u: User) => void;
}) {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [entryId, setEntryId] = useState(user.entry_node_id?.toString() ?? '');
  const [exitId, setExitId] = useState(user.exit_node_id?.toString() ?? '');
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
    setExitId(user.exit_node_id?.toString() ?? '');
  }, [user.entry_node_id, user.exit_node_id]);

  const entry = nodes.find((n) => n.id === Number(entryId)) ?? null;
  const exit = nodes.find((n) => n.id === Number(exitId)) ?? null;

  return (
    <section data-testid="user-chain-section" className="card-pad space-y-4">
      <div>
        <h2 className="text-lg font-medium">Multi-hop chain</h2>
        <p className="text-xs text-slate-500 mt-1">
          Assign entry/exit nodes for this user. Empty = auto by priority.
        </p>
      </div>
      <ChainTopology entry={entry} exit={exit} />
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
                </option>
              ))}
          </select>
        </div>
        <div>
          <label className="label">Exit node</label>
          <select
            data-testid="exit-node-select"
            value={exitId}
            onChange={(e) => setExitId(e.target.value)}
            className="input"
          >
            <option value="">Auto exit</option>
            {nodes
              .filter((n) => n.role === 'exit')
              .map((n) => (
                <option key={n.id} value={n.id}>
                  {n.name}
                </option>
              ))}
          </select>
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
              exit_node_id: exitId ? Number(exitId) : null,
            });
            onUpdated(updated);
            setMessage('Chain saved');
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
