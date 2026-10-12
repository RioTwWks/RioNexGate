import type { Node } from '../types/node';

function NodeBox({
  node,
  fallback,
  compact,
}: {
  node?: Node | null;
  fallback: string;
  compact?: boolean;
}) {
  const box = compact
    ? 'px-2.5 py-1.5 rounded-lg border text-xs min-w-[96px] text-center'
    : 'px-4 py-3 rounded-xl border text-sm min-w-[140px] text-center';
  if (!node) {
    return (
      <div className={`${box} border-dashed border-slate-600 text-slate-500 bg-slate-950/30`}>
        {fallback}
      </div>
    );
  }
  return (
    <div
      className={`${box} ${
        node.active
          ? 'border-sky-500/40 bg-sky-950/40 shadow-glow'
          : 'border-slate-600 opacity-60 bg-slate-950/30'
      }`}
      data-testid={`topology-node-${node.role}`}
    >
      <div className="text-[10px] uppercase tracking-wider text-slate-400">{node.role}</div>
      <div className={`font-medium mt-0.5 ${compact ? 'text-xs' : 'text-sm'}`}>
        {node.region || node.name}
      </div>
      {!compact && (
        <div className="text-xs text-slate-500 mt-1">
          {node.name} · {node.address}:{node.port}
        </div>
      )}
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

export function ChainTopology({
  entry,
  exit,
  exits,
  compact,
}: {
  entry?: Node | null;
  /** @deprecated use exits */
  exit?: Node | null;
  exits?: Node[] | null;
  compact?: boolean;
}) {
  const list = exits && exits.length > 0 ? exits : exit ? [exit] : [];

  return (
    <div
      className={`flex items-center justify-center gap-2 flex-wrap ${compact ? 'py-2' : 'py-4'}`}
      data-testid="chain-topology"
    >
      <div
        className={`rounded-lg bg-slate-800/80 border border-surface-border ${
          compact ? 'px-2 py-1 text-xs' : 'px-3 py-2 text-sm'
        }`}
      >
        Client
      </div>
      <Arrow />
      <NodeBox node={entry} fallback="Entry (auto)" compact={compact} />
      <Arrow />
      {list.length === 0 ? (
        <NodeBox fallback="Exit (auto)" compact={compact} />
      ) : list.length === 1 ? (
        <NodeBox node={list[0]} fallback="Exit (auto)" compact={compact} />
      ) : (
        <div
          className={`flex flex-wrap items-center justify-center gap-2 ${compact ? 'max-w-xs' : 'max-w-md'}`}
          data-testid="topology-exits"
        >
          {list.map((n) => (
            <NodeBox key={n.id} node={n} fallback="Exit" compact={compact} />
          ))}
        </div>
      )}
      <Arrow />
      <div
        className={`rounded-lg bg-slate-800/80 border border-surface-border ${
          compact ? 'px-2 py-1 text-xs' : 'px-3 py-2 text-sm'
        }`}
      >
        Internet
      </div>
    </div>
  );
}
