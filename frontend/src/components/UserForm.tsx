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
}

export function UserForm({ initial, onSubmit, onCancel, submitLabel = 'Save' }: Props) {
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
      setError(err instanceof Error ? err.message : 'Failed to save');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && <p className="text-red-400 text-sm">{error}</p>}
      <div>
        <label className="block text-sm text-slate-400 mb-1">Email</label>
        <input
          type="email"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700"
        />
      </div>
      <div>
        <label className="block text-sm text-slate-400 mb-1">Traffic limit (GB)</label>
        <input
          type="number"
          min={1}
          value={trafficGb}
          onChange={(e) => setTrafficGb(Number(e.target.value))}
          className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700"
        />
      </div>
      <div>
        <label className="block text-sm text-slate-400 mb-1">Expire in (days)</label>
        <input
          type="number"
          min={1}
          value={expireDays}
          onChange={(e) => setExpireDays(Number(e.target.value))}
          className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700"
        />
      </div>
      {initial?.email !== undefined && (
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} />
          Active
        </label>
      )}
      <div className="flex gap-2 justify-end">
        <button type="button" onClick={onCancel} className="px-4 py-2 rounded bg-slate-700 hover:bg-slate-600">
          Cancel
        </button>
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 rounded bg-sky-600 hover:bg-sky-500 disabled:opacity-50"
        >
          {loading ? 'Saving...' : submitLabel}
        </button>
      </div>
    </form>
  );
}
