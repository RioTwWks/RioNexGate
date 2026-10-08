import type { SyncStatus } from '../types/device';

const STATUS_CONFIG: Record<
  SyncStatus,
  { label: string; className: string; dotClass: string }
> = {
  synced: {
    label: 'Synced',
    className: 'badge-ok',
    dotClass: 'bg-emerald-400',
  },
  stale: {
    label: 'Stale',
    className: 'badge-warn',
    dotClass: 'bg-amber-400',
  },
  never: {
    label: 'Never synced',
    className: 'badge-neutral',
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
      className={config.className}
      title={lastSeenAt ? `Last seen: ${new Date(lastSeenAt).toLocaleString()}` : undefined}
    >
      <span className={`w-1.5 h-1.5 rounded-full ${config.dotClass}`} />
      {config.label}
    </span>
  );
}
