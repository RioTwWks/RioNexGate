import { FormEvent, useState } from 'react';
import api, { setApiKey } from '../services/api';

export function Login() {
  const [key, setKey] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setApiKey(key.trim());
    try {
      await api.get('/users');
      window.location.href = '/';
    } catch {
      setApiKey('');
      setError('Invalid API key');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4 relative overflow-hidden">
      <div
        className="pointer-events-none absolute inset-0 opacity-70"
        style={{
          background:
            'radial-gradient(circle at 20% 20%, rgba(56,189,248,0.18), transparent 40%), radial-gradient(circle at 80% 70%, rgba(52,211,153,0.1), transparent 35%)',
        }}
      />
      <form
        onSubmit={handleSubmit}
        className="relative w-full max-w-md card p-8 space-y-5 shadow-glow animate-fade-up"
      >
        <div className="text-center space-y-2">
          <p className="text-xs uppercase tracking-[0.25em] text-sky-400/80">Proxy control plane</p>
          <h1 className="text-3xl font-semibold text-sky-300">RioNexGate</h1>
          <p className="text-sm text-surface-muted">Enter your API key to continue</p>
        </div>

        {error && <p className="alert-error text-center">{error}</p>}

        <div>
          <label className="label" htmlFor="api-key">
            API key
          </label>
          <input
            id="api-key"
            type="password"
            required
            autoFocus
            placeholder="API key"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            className="input font-mono"
          />
        </div>

        <button type="submit" disabled={loading} className="btn-primary w-full py-2.5">
          {loading ? 'Checking...' : 'Login'}
        </button>

        <p className="text-xs text-center text-slate-500">
          Auth header: <span className="font-mono text-slate-400">X-API-Key</span>
        </p>
      </form>
    </div>
  );
}
