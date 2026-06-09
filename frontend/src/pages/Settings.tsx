import { useEffect, useState } from 'react';
import api from '../services/api';

export function Settings() {
  const [coreType, setCoreType] = useState('xray');
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');

  useEffect(() => {
    api.get<{ type: string }>('/core/type').then((res) => setCoreType(res.data.type));
  }, []);

  const switchCore = async (type: string) => {
    setLoading(true);
    setMessage('');
    try {
      await api.put('/core/type', { type });
      setCoreType(type);
      setMessage(`Switched to ${type}`);
    } catch {
      setMessage('Failed to switch core');
    } finally {
      setLoading(false);
    }
  };

  const reload = async () => {
    setLoading(true);
    setMessage('');
    try {
      await api.post('/core/reload');
      setMessage('Core reloaded');
    } catch {
      setMessage('Reload failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Settings</h1>
      {message && <p className="mb-4 text-sky-400">{message}</p>}
      <div className="bg-slate-900 border border-slate-800 rounded-lg p-6 max-w-md space-y-4">
        <div>
          <p className="text-slate-400 text-sm mb-2">Active core</p>
          <div className="flex gap-2">
            <button
              disabled={loading}
              onClick={() => switchCore('xray')}
              className={`px-4 py-2 rounded ${
                coreType === 'xray' ? 'bg-sky-600' : 'bg-slate-700 hover:bg-slate-600'
              }`}
            >
              Xray
            </button>
            <button
              disabled={loading}
              onClick={() => switchCore('sing-box')}
              className={`px-4 py-2 rounded ${
                coreType === 'sing-box' ? 'bg-sky-600' : 'bg-slate-700 hover:bg-slate-600'
              }`}
            >
              sing-box
            </button>
          </div>
        </div>
        <button
          disabled={loading}
          onClick={reload}
          className="px-4 py-2 rounded bg-slate-700 hover:bg-slate-600 disabled:opacity-50"
        >
          Reload core config
        </button>
      </div>
    </div>
  );
}
