import type { ReactNode } from 'react';

export function PageHeader({
  title,
  subtitle,
  actions,
}: {
  title: string;
  subtitle?: string;
  actions?: ReactNode;
}) {
  return (
    <div className="page-header">
      <div>
        <h1 className="page-title">{title}</h1>
        {subtitle && <p className="page-subtitle">{subtitle}</p>}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </div>
  );
}

export function StatCard({
  label,
  value,
  hint,
  accent = 'sky',
}: {
  label: string;
  value: ReactNode;
  hint?: string;
  accent?: 'sky' | 'emerald' | 'amber' | 'rose';
}) {
  const accentClass = {
    sky: 'text-sky-400',
    emerald: 'text-emerald-400',
    amber: 'text-amber-400',
    rose: 'text-rose-400',
  }[accent];

  return (
    <div className="stat-card">
      <p className="text-xs uppercase tracking-wider text-surface-muted">{label}</p>
      <p className={`mt-2 text-2xl sm:text-3xl font-semibold tabular-nums ${accentClass}`}>{value}</p>
      {hint && <p className="mt-1 text-xs text-slate-500">{hint}</p>}
    </div>
  );
}

export function EmptyState({ title, description }: { title: string; description?: string }) {
  return (
    <div className="px-6 py-12 text-center">
      <p className="text-slate-300 font-medium">{title}</p>
      {description && <p className="mt-1 text-sm text-slate-500">{description}</p>}
    </div>
  );
}

export function Modal({
  title,
  children,
  onClose,
  wide,
}: {
  title: string;
  children: ReactNode;
  onClose: () => void;
  wide?: boolean;
}) {
  return (
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div
        className={`modal-panel ${wide ? 'max-w-lg' : ''}`}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label={title}
      >
        <div className="flex items-start justify-between gap-3 mb-4">
          <h2 className="text-lg font-semibold">{title}</h2>
          <button type="button" onClick={onClose} className="btn-ghost px-2 py-1 text-slate-400" aria-label="Close">
            ✕
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

export function TrafficBar({ used, limit }: { used: number; limit: number }) {
  const pct = limit > 0 ? Math.min(100, (used / limit) * 100) : 0;
  const tone = pct >= 90 ? 'bg-rose-500' : pct >= 70 ? 'bg-amber-400' : 'bg-sky-500';
  return (
    <div className="space-y-1.5 min-w-[140px]">
      <div className="flex justify-between text-xs text-slate-400 tabular-nums">
        <span>
          {used.toFixed(2)} / {limit} GB
        </span>
        <span>{pct.toFixed(0)}%</span>
      </div>
      <div className="progress">
        <span className={tone} style={{ width: `${pct}%` }} />
      </div>
    </div>
  );
}

export function StatusBadge({ active, expired }: { active: boolean; expired?: boolean }) {
  if (expired) return <span className="badge-danger">Expired</span>;
  if (active) return <span className="badge-ok">Active</span>;
  return <span className="badge-neutral">Inactive</span>;
}

export function Spinner({ label = 'Loading…' }: { label?: string }) {
  return (
    <div className="flex items-center gap-3 text-surface-muted py-8">
      <span className="w-4 h-4 rounded-full border-2 border-sky-500/30 border-t-sky-400 animate-spin" />
      <span className="text-sm">{label}</span>
    </div>
  );
}
