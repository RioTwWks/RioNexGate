import type { Node } from '../types/node';

function NodeBox({ node, fallback }: { node?: Node | null; fallback: string }) {
  if (!node) {
    return (
      <div className="px-4 py-3 rounded-xl border border-dashed border-slate-600 text-slate-500 text-sm min-w-[140px] text-center bg-slate-950/30">
        {fallback}
      </div>
    );
  }
  return (
    <div
      className={`px-4 py-3 rounded-xl border min-w-[140px] text-center ${
        node.active
          ? 'border-sky-500/40 bg-sky-950/40 shadow-glow'
          : 'border-slate-600 opacity-60 bg-slate-950/30'
      }`}
      data-testid={`topology-node-${node.role}`}
    >
      <div className="text-[10px] uppercase tracking-wider text-slate-400">{node.role}</div>
      <div className="font-medium text-sm mt-0.5">{node.name}</div>
      <div className="text-xs text-slate-500 mt-1">
        {node.region || '—'} · {node.address}:{node.port}
      </div>
    </div>
  );
}

function Arrow() {
  return (
    <span className="text-sky-500/70 text-lg font-light px-1" aria-hidden>
      →
    </span>
  );
}

export function ChainTopology({ entry, exit }: { entry?: Node | null; exit?: Node | null }) {
  return (
    <div
      className="flex items-center justify-center gap-2 flex-wrap py-4"
      data-testid="chain-topology"
    >
      <div className="px-3 py-2 rounded-lg bg-slate-800/80 border border-surface-border text-sm">
        Client
      </div>
      <Arrow />
      <NodeBox node={entry} fallback="Entry (auto)" />
      <Arrow />
      <NodeBox node={exit} fallback="Exit (auto)" />
      <Arrow />
      <div className="px-3 py-2 rounded-lg bg-slate-800/80 border border-surface-border text-sm">
        Internet
      </div>
    </div>
  );
}
