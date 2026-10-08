import { useEffect, useState } from 'react';
import { PageHeader } from '../components/ui';
import api, { rotateApiKey, setApiKey } from '../services/api';

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

  const [newKeyPreview, setNewKeyPreview] = useState('');
  const [rotating, setRotating] = useState(false);

  const rotateKey = async () => {
    if (
      !confirm(
        'Generate and apply a new API key? The current key stops working immediately. Copy the new key after rotation.',
      )
    ) {
      return;
    }
    setRotating(true);
    setMessage('');
    setError('');
    setNewKeyPreview('');
    try {
      const key = await rotateApiKey();
      setApiKey(key);
      setNewKeyPreview(key);
      setMessage('API key rotated. Stored in this browser; copy it for other clients.');
    } catch {
      setError('Failed to rotate API key');
    } finally {
      setRotating(false);
    }
  };

  return (
    <div className="page max-w-3xl">
      <PageHeader
        title="Settings"
        subtitle="Active proxy core, API key, and configuration reload"
      />

      {message && <p className="alert-ok">{message}</p>}
      {error && <p className="alert-error">{error}</p>}

      <section className="card-pad space-y-4">
        <div>
          <p className="text-sm font-medium text-slate-200 mb-1">Panel API key</p>
          <p className="text-xs text-slate-500 mb-3">
            Rotates <code className="text-xs">server.api_key</code> in config.yaml and applies it live.
          </p>
          <button
            type="button"
            disabled={loading || rotating}
            onClick={rotateKey}
            className="btn-secondary"
            data-testid="rotate-api-key"
          >
            {rotating ? 'Rotating…' : 'Rotate API key'}
          </button>
          {newKeyPreview && (
            <div className="mt-3 space-y-2" data-testid="new-api-key">
              <input readOnly value={newKeyPreview} className="input font-mono text-xs" />
              <button
                type="button"
                className="btn-primary"
                onClick={async () => {
                  await navigator.clipboard.writeText(newKeyPreview);
                  setMessage('New API key copied');
                }}
              >
                Copy new key
              </button>
            </div>
          )}
        </div>
      </section>

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
