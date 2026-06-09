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
    setApiKey(key);
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
    <div className="min-h-screen flex items-center justify-center p-4">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm bg-slate-900 border border-slate-800 rounded-lg p-6 space-y-4"
      >
        <h1 className="text-xl font-semibold text-center text-sky-400">proxy-mgr</h1>
        <p className="text-sm text-slate-400 text-center">Enter your API key to continue</p>
        {error && <p className="text-red-400 text-sm text-center">{error}</p>}
        <input
          type="password"
          required
          placeholder="API key"
          value={key}
          onChange={(e) => setKey(e.target.value)}
          className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700"
        />
        <button
          type="submit"
          disabled={loading}
          className="w-full py-2 rounded bg-sky-600 hover:bg-sky-500 disabled:opacity-50"
        >
          {loading ? 'Checking...' : 'Login'}
        </button>
      </form>
    </div>
  );
}
