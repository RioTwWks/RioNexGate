import type { SyncStatus } from '../types/device';

const STATUS_CONFIG: Record<
  SyncStatus,
  { label: string; className: string; dotClass: string }
> = {
  synced: {
    label: 'Synced',
    className: 'text-emerald-400 bg-emerald-400/10',
    dotClass: 'bg-emerald-400',
  },
  stale: {
    label: 'Stale',
    className: 'text-amber-400 bg-amber-400/10',
    dotClass: 'bg-amber-400',
  },
  never: {
    label: 'Never synced',
    className: 'text-slate-400 bg-slate-400/10',
    dotClass: 'bg-slate-500',
  },
};

interface Props {
  status: SyncStatus;
  lastSeenAt?: string | null;
}

export function SyncStatusBadge({ status, lastSeenAt }: Props) {
  const config = STATUS_CONFIG[status];

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-xs font-medium ${config.className}`}
      title={lastSeenAt ? `Last seen: ${new Date(lastSeenAt).toLocaleString()}` : undefined}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${config.dotClass}`} />
      {config.label}
    </span>
  );
}
