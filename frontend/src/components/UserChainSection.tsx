import { useEffect, useState } from 'react';
import { ChainTopology } from './ChainTopology';
import { getNodes, updateUserChain } from '../services/api';
import type { Node } from '../types/node';
import type { User } from '../types/user';
export function UserChainSection({ user, onUpdated }: { user: User; onUpdated: (u: User) => void }) {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [entryId, setEntryId] = useState(user.entry_node_id?.toString() ?? '');
  const [exitId, setExitId] = useState(user.exit_node_id?.toString() ?? '');
  useEffect(() => { getNodes().then(setNodes).catch(() => setNodes([])); }, []);
  const entry = nodes.find(n => n.id === Number(entryId)) ?? null;
  const exit = nodes.find(n => n.id === Number(exitId)) ?? null;
  return <section data-testid="user-chain-section" className="bg-slate-900 border border-slate-800 rounded-lg p-5">
    <h2 className="text-lg font-medium">Multi-hop chain</h2>
    <ChainTopology entry={entry} exit={exit} />
    <select data-testid="entry-node-select" value={entryId} onChange={e=>setEntryId(e.target.value)} className="mt-4 w-full px-3 py-2 rounded bg-slate-800">
      <option value="">Auto entry</option>{nodes.filter(n=>n.role==='entry').map(n=><option key={n.id} value={n.id}>{n.name}</option>)}</select>
    <select data-testid="exit-node-select" value={exitId} onChange={e=>setExitId(e.target.value)} className="mt-2 w-full px-3 py-2 rounded bg-slate-800">
      <option value="">Auto exit</option>{nodes.filter(n=>n.role==='exit').map(n=><option key={n.id} value={n.id}>{n.name}</option>)}</select>
    <button data-testid="save-chain" className="mt-4 px-4 py-2 rounded bg-sky-600" onClick={async()=>{onUpdated(await updateUserChain(user.id,{entry_node_id:entryId?Number(entryId):null,exit_node_id:exitId?Number(exitId):null}));}}>Save chain</button>
  </section>;
}
