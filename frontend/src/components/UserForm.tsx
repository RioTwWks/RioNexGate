import { useState } from 'react';

export interface UserFormData {
  email: string;
  traffic_gb: number;
  expire_days: number;
  active?: boolean;
}

interface Props {
  initial?: Partial<UserFormData>;
  onSubmit: (data: UserFormData) => Promise<void>;
  onCancel: () => void;
  submitLabel?: string;
  showExpireDays?: boolean;
}

export function UserForm({
  initial,
  onSubmit,
  onCancel,
  submitLabel = 'Save',
  showExpireDays = true,
}: Props) {
  const [email, setEmail] = useState(initial?.email ?? '');
  const [trafficGb, setTrafficGb] = useState(initial?.traffic_gb ?? 50);
  const [expireDays, setExpireDays] = useState(initial?.expire_days ?? 30);
  const [active, setActive] = useState(initial?.active ?? true);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await onSubmit({ email, traffic_gb: trafficGb, expire_days: expireDays, active });
    } catch (err: unknown) {
      const msg =
        err && typeof err === 'object' && 'response' in err
          ? String((err as { response?: { data?: { error?: string } } }).response?.data?.error || '')
          : '';
      setError(msg || (err instanceof Error ? err.message : 'Failed to save'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && <p className="alert-error">{error}</p>}
      <div>
        <label className="label">Email</label>
        <input
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="input"
        />
      </div>
      <div>
        <label className="label">Traffic limit (GB)</label>
        <input
          type="number"
          min={1}
          value={trafficGb}
          onChange={(e) => setTrafficGb(Number(e.target.value))}
          className="input"
        />
      </div>
      {showExpireDays && (
        <div>
          <label className="label">Expire in (days)</label>
          <input
            type="number"
            min={1}
            value={expireDays}
            onChange={(e) => setExpireDays(Number(e.target.value))}
            className="input"
          />
        </div>
      )}
      {initial?.email !== undefined && (
        <label className="flex items-center gap-2 text-sm text-slate-300">
          <input
            type="checkbox"
            checked={active}
            onChange={(e) => setActive(e.target.checked)}
            className="rounded border-surface-border"
          />
          Active
        </label>
      )}
      <div className="flex gap-2 justify-end pt-2">
        <button type="button" onClick={onCancel} className="btn-secondary">
          Cancel
        </button>
        <button type="submit" disabled={loading} className="btn-primary">
          {loading ? 'Saving...' : submitLabel}
        </button>
      </div>
    </form>
  );
}
