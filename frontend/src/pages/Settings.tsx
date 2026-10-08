import { useEffect, useState } from 'react';
import { PageHeader } from '../components/ui';
import api from '../services/api';

const CORES = [
  {
    id: 'xray',
    label: 'Xray',
    description: 'Full VLESS Reality + Vision / XHTTP. Default for multihop.',
  },
  {
    id: 'sing-box',
    label: 'sing-box',
    description: 'Alternative core with JSON config generation.',
  },
  {
    id: 'skadi',
    label: 'SkadiCore',
    description: 'Single listen (Reality ± XHTTP). Vision flow not supported.',
  },
] as const;

export function Settings() {
  const [coreType, setCoreType] = useState('xray');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  useEffect(() => {
    api
      .get<{ type: string }>('/core/type')
      .then((res) => setCoreType(res.data.type))
      .catch(() => setError('Failed to load core type'));
  }, []);

  const switchCore = async (type: string) => {
    setLoading(true);
    setMessage('');
    setError('');
    try {
      await api.put('/core/type', { type });
      setCoreType(type);
      setMessage(`Switched to ${type}`);
    } catch {
      setError('Failed to switch core');
    } finally {
      setLoading(false);
    }
  };

  const reload = async () => {
    setLoading(true);
    setMessage('');
    setError('');
    try {
      await api.post('/core/reload');
      setMessage('Core reloaded');
    } catch {
      setError('Reload failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page max-w-3xl">
      <PageHeader
        title="Settings"
        subtitle="Active proxy core and configuration reload"
      />

      {message && <p className="alert-ok">{message}</p>}
      {error && <p className="alert-error">{error}</p>}

      <section className="card-pad space-y-5">
        <div>
          <p className="text-sm font-medium text-slate-200 mb-1">Active core</p>
          <p className="text-xs text-slate-500 mb-4">
            Switching regenerates configs for the selected engine. Reload after stealth or node changes.
          </p>
          <div className="grid gap-3">
            {CORES.map((core) => {
              const selected = coreType === core.id;
              return (
                <button
                  key={core.id}
                  type="button"
                  disabled={loading}
                  onClick={() => switchCore(core.id)}
                  className={`text-left rounded-xl border px-4 py-3 transition-colors ${
                    selected
                      ? 'border-sky-500/40 bg-sky-600/15 shadow-glow'
                      : 'border-surface-border bg-slate-950/40 hover:bg-slate-800/50'
                  }`}
                >
                  <div className="flex items-center justify-between gap-3">
                    <span className="font-medium">{core.label}</span>
                    {selected && <span className="badge-info">Selected</span>}
                  </div>
                  <p className="text-xs text-slate-500 mt-1">{core.description}</p>
                </button>
              );
            })}
          </div>
          {coreType === 'skadi' && (
            <p className="text-slate-500 text-xs mt-3">
              SkadiCore: single listen port (Reality ± XHTTP). Vision flow is not supported;
              client links use TCP+Reality without xtls-rprx-vision.
            </p>
          )}
        </div>

        <div className="pt-2 border-t border-surface-border">
          <button type="button" disabled={loading} onClick={reload} className="btn-secondary">
            Reload core config
          </button>
        </div>
      </section>
    </div>
  );
}
