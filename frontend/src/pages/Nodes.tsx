import { useCallback, useEffect, useMemo, useState } from 'react';
import { ChainTopology } from '../components/ChainTopology';
import { NodeForm } from '../components/NodeForm';
import { checkNodeHealth, createNode, deleteNode, getNodes, updateNode } from '../services/api';
import type { CreateNodeInput, Node, NodeHealthResult } from '../types/node';

export function Nodes() {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [editNode, setEditNode] = useState<Node | null>(null);
  const [healthResults, setHealthResults] = useState<Record<number, NodeHealthResult>>({});
  const [checking, setChecking] = useState<number | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try { setNodes(await getNodes()); } finally { setLoading(false); }
  }, []);

  useEffect(() => { load(); }, [load]);

  const entryNodes = useMemo(() => nodes.filter((n) => n.role === 'entry'), [nodes]);
  const exitNodes = useMemo(() => nodes.filter((n) => n.role === 'exit'), [nodes]);
  const bestEntry = useMemo(() => [...entryNodes].filter((n) => n.active).sort((a, b) => a.priority - b.priority)[0], [entryNodes]);
  const bestExit = useMemo(() => [...exitNodes].filter((n) => n.active).sort((a, b) => a.priority - b.priority)[0], [exitNodes]);

  const handleHealthCheck = async (node: Node) => {
    setChecking(node.id);
    try {
      const result = await checkNodeHealth(node.id);
      setHealthResults((p) => ({ ...p, [node.id]: result }));
    } finally { setChecking(null); }
  };

  const renderTable = (title: string, items: Node[]) => (
    <section className="bg-slate-900 border border-slate-800 rounded-lg p-5">
      <h2 className="text-lg font-medium mb-3">{title}</h2>
      <table className="w-full text-sm">
        <tbody>
          {items.map((n) => (
            <tr key={n.id} className="border-t border-slate-800" data-testid={`node-row-${n.id}`}>
              <td className="p-2">{n.name}</td>
              <td className="p-2 font-mono text-xs">{n.address}:{n.port}</td>
              <td className="p-2">{n.region || '—'}</td>
              <td className="p-2">
                <button data-testid={`toggle-active-${n.id}`} onClick={async () => { await updateNode(n.id, { active: !n.active }); await load(); }} className="text-xs px-2 py-0.5 rounded bg-slate-800">
                  {n.active ? 'Active' : 'Inactive'}
                </button>
                {healthResults[n.id] && (
                  <div data-testid={`health-result-${n.id}`} className={`text-xs mt-1 ${healthResults[n.id].reachable ? 'text-emerald-400' : 'text-red-400'}`}>
                    {healthResults[n.id].reachable ? `TCP OK (${healthResults[n.id].latency_ms}ms)` : healthResults[n.id].error}
                  </div>
                )}
              </td>
              <td className="p-2 text-right space-x-2">
                <button data-testid={`health-check-${n.id}`} onClick={() => handleHealthCheck(n)} disabled={checking === n.id} className="text-sky-400">Health check</button>
                <button onClick={() => setEditNode(n)} className="text-slate-300">Edit</button>
                <button onClick={async () => { if (confirm('Delete?')) { await deleteNode(n.id); await load(); } }} className="text-red-400">Delete</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-semibold">Multi-hop nodes</h1>
        <button data-testid="add-node" onClick={() => setShowAdd(true)} className="px-4 py-2 rounded bg-sky-600">Add node</button>
      </div>
      {loading ? <p>Loading...</p> : (
        <>
          <section className="bg-slate-900 border border-slate-800 rounded-lg p-5">
            <h2 className="text-lg font-medium mb-2">Topology</h2>
            <ChainTopology entry={bestEntry} exit={bestExit} />
          </section>
          {renderTable('Entry nodes', entryNodes)}
          {renderTable('Exit nodes', exitNodes)}
        </>
      )}
      {showAdd && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-slate-900 border border-slate-700 rounded-lg p-6 w-full max-w-lg mx-4">
            <NodeForm submitLabel="Create" onCancel={() => setShowAdd(false)} onSubmit={async (d: CreateNodeInput) => { await createNode(d); setShowAdd(false); await load(); }} />
          </div>
        </div>
      )}
      {editNode && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-slate-900 border border-slate-700 rounded-lg p-6 w-full max-w-lg mx-4">
            <NodeForm initial={editNode} submitLabel="Save" onCancel={() => setEditNode(null)} onSubmit={async (d: CreateNodeInput) => { await updateNode(editNode.id, d); setEditNode(null); await load(); }} />
          </div>
        </div>
      )}
    </div>
  );
}
