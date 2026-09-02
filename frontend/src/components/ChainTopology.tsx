import type { Node } from '../types/node';
function NodeBox({ node, fallback }: { node?: Node | null; fallback: string }) {
  if (!node) return <div className="px-4 py-3 rounded-lg border border-dashed border-slate-600 text-slate-500 text-sm min-w-[140px] text-center">{fallback}</div>;
  return <div className={`px-4 py-3 rounded-lg border min-w-[140px] text-center ${node.active ? 'border-sky-600 bg-sky-950/40' : 'border-slate-600 opacity-60'}`} data-testid={`topology-node-${node.role}`}>
    <div className="text-xs uppercase text-slate-400">{node.role}</div><div className="font-medium text-sm">{node.name}</div>
    <div className="text-xs text-slate-500">{node.region || '—'} · {node.address}:{node.port}</div></div>;
}
export function ChainTopology({ entry, exit }: { entry?: Node | null; exit?: Node | null }) {
  return <div className="flex items-center justify-center gap-2 flex-wrap py-4" data-testid="chain-topology">
    <div className="px-3 py-2 rounded bg-slate-800 text-sm">Client</div><span>→</span>
    <NodeBox node={entry} fallback="Entry (auto)" /><span>→</span>
    <NodeBox node={exit} fallback="Exit (auto)" /><span>→</span>
    <div className="px-3 py-2 rounded bg-slate-800 text-sm">Internet</div></div>;
}
