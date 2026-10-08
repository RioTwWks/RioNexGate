import { useCallback, useEffect, useMemo, useState } from 'react';
import { ChainTopology } from '../components/ChainTopology';
import { NodeForm } from '../components/NodeForm';
import { EmptyState, Modal, PageHeader, Spinner } from '../components/ui';
import { checkNodeHealth, createNode, deleteNode, getNodes, updateNode } from '../services/api';
import type { CreateNodeInput, Node, NodeHealthResult } from '../types/node';

export function Nodes() {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showAdd, setShowAdd] = useState(false);
  const [editNode, setEditNode] = useState<Node | null>(null);
  const [healthResults, setHealthResults] = useState<Record<number, NodeHealthResult>>({});
  const [checking, setChecking] = useState<number | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      setNodes(await getNodes());
    } catch {
      setError('Failed to load nodes');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const entryNodes = useMemo(() => nodes.filter((n) => n.role === 'entry'), [nodes]);
  const exitNodes = useMemo(() => nodes.filter((n) => n.role === 'exit'), [nodes]);
  const bestEntry = useMemo(
    () => [...entryNodes].filter((n) => n.active).sort((a, b) => a.priority - b.priority)[0],
    [entryNodes],
  );
  const bestExit = useMemo(
    () => [...exitNodes].filter((n) => n.active).sort((a, b) => a.priority - b.priority)[0],
    [exitNodes],
  );

  const handleHealthCheck = async (node: Node) => {
    setChecking(node.id);
    try {
      const result = await checkNodeHealth(node.id);
      setHealthResults((p) => ({ ...p, [node.id]: result }));
    } catch {
      setHealthResults((p) => ({
        ...p,
        [node.id]: { reachable: false, check_type: 'tcp', error: 'Health check failed' },
      }));
    } finally {
      setChecking(null);
    }
  };

  const renderTable = (title: string, items: Node[], empty: string) => (
    <section className="card-pad">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-lg font-medium">{title}</h2>
        <span className="badge-neutral">{items.length}</span>
      </div>
      {items.length === 0 ? (
        <EmptyState title={empty} />
      ) : (
        <div className="overflow-x-auto -mx-1">
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Endpoint</th>
                <th>Region</th>
                <th>Status</th>
                <th className="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {items.map((n) => (
                <tr key={n.id} data-testid={`node-row-${n.id}`}>
                  <td className="font-medium">{n.name}</td>
                  <td className="font-mono text-xs text-slate-300">
                    {n.address}:{n.port}
                  </td>
                  <td>{n.region || '—'}</td>
                  <td>
                    <button
                      data-testid={`toggle-active-${n.id}`}
                      onClick={async () => {
                        await updateNode(n.id, { active: !n.active });
                        await load();
                      }}
                      className={n.active ? 'badge-ok' : 'badge-neutral'}
                      type="button"
                    >
                      {n.active ? 'Active' : 'Inactive'}
                    </button>
                    {healthResults[n.id] && (
                      <div
                        data-testid={`health-result-${n.id}`}
                        className={`text-xs mt-1.5 ${
                          healthResults[n.id].reachable ? 'text-emerald-400' : 'text-red-400'
                        }`}
                      >
                        {healthResults[n.id].reachable
                          ? `TCP OK (${healthResults[n.id].latency_ms}ms)`
                          : healthResults[n.id].error}
                      </div>
                    )}
                  </td>
                  <td className="text-right whitespace-nowrap space-x-3">
                    <button
                      data-testid={`health-check-${n.id}`}
                      onClick={() => handleHealthCheck(n)}
                      disabled={checking === n.id}
                      className="link-action disabled:opacity-50"
                      type="button"
                    >
                      {checking === n.id ? 'Checking…' : 'Health check'}
                    </button>
                    <button type="button" onClick={() => setEditNode(n)} className="text-slate-300 hover:underline text-sm">
                      Edit
                    </button>
                    <button
                      type="button"
                      onClick={async () => {
                        if (confirm('Delete?')) {
                          await deleteNode(n.id);
                          await load();
                        }
                      }}
                      className="text-red-400 hover:underline text-sm"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );

  return (
    <div className="page">
      <PageHeader
        title="Multi-hop nodes"
        subtitle="Entry and exit relays for RU → EU chaining"
        actions={
          <button data-testid="add-node" type="button" onClick={() => setShowAdd(true)} className="btn-primary">
            Add node
          </button>
        }
      />

      {error && <p className="alert-error">{error}</p>}

      {loading ? (
        <Spinner />
      ) : (
        <>
          <section className="card-pad">
            <h2 className="text-lg font-medium mb-1">Topology</h2>
            <p className="text-xs text-slate-500 mb-2">Best active entry and exit by priority</p>
            <ChainTopology entry={bestEntry} exit={bestExit} />
          </section>
          {renderTable('Entry nodes', entryNodes, 'No entry nodes yet')}
          {renderTable('Exit nodes', exitNodes, 'No exit nodes yet')}
        </>
      )}

      {showAdd && (
        <Modal title="Add node" onClose={() => setShowAdd(false)} wide>
          <NodeForm
            submitLabel="Create"
            onCancel={() => setShowAdd(false)}
            onSubmit={async (d: CreateNodeInput) => {
              await createNode(d);
              setShowAdd(false);
              await load();
            }}
          />
        </Modal>
      )}

      {editNode && (
        <Modal title="Edit node" onClose={() => setEditNode(null)} wide>
          <NodeForm
            initial={editNode}
            submitLabel="Save"
            onCancel={() => setEditNode(null)}
            onSubmit={async (d: CreateNodeInput) => {
              await updateNode(editNode.id, d);
              setEditNode(null);
              await load();
            }}
          />
        </Modal>
      )}
    </div>
  );
}
