import { useCallback, useEffect, useState } from 'react';
import {
  getStealthSettings,
  testDestAvailability,
  updateStealthSettings,
} from '../services/api';
import { StealthWarnings } from '../components/StealthWarnings';
import type { DestTestResult, StealthSettings } from '../types/stealth';
import { FINGERPRINT_OPTIONS } from '../types/stealth';

const DEFAULT_SETTINGS: StealthSettings = {
  presets: {
    xhttp_reality: { enabled: true, port: 443 },
    vision_reality: { enabled: true, port: 8443 },
    tls: { enabled: false, port: 2053 },
    amneziawg: { enabled: false, port: 51820 },
  },
  reality: {
    dest: '',
    server_names: [],
    fingerprint: 'firefox',
    short_ids: [],
  },
  fragmentation: {
    enabled: false,
    strategy: 'serverhello',
    length: '50-100',
    delay: '10-20',
    max_split: '2-4',
  },
};

const PRESET_LABELS: Record<keyof StealthSettings['presets'], string> = {
  xhttp_reality: 'VLESS + Reality + XHTTP (primary)',
  vision_reality: 'VLESS + Reality + Vision (fallback)',
  tls: 'VLESS + TLS (mobile)',
  amneziawg: 'AmneziaWG (UDP reserve)',
};

export function Stealth() {
  const [settings, setSettings] = useState<StealthSettings>(DEFAULT_SETTINGS);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [testResult, setTestResult] = useState<DestTestResult | null>(null);
  const [testing, setTesting] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const data = await getStealthSettings();
      setSettings(data);
    } catch {
      setError('Failed to load stealth settings. Using defaults until backend is ready.');
      setSettings(DEFAULT_SETTINGS);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const save = async () => {
    setSaving(true);
    setMessage('');
    setError('');
    try {
      const updated = await updateStealthSettings(settings);
      setSettings(updated);
      setMessage('Settings saved. Core reload may be required.');
    } catch {
      setError('Failed to save settings. Backend endpoint may not be ready yet.');
    } finally {
      setSaving(false);
    }
  };

  const testDest = async () => {
    if (!settings.reality.dest) return;
    setTesting(true);
    setTestResult(null);
    try {
      const result = await testDestAvailability(settings.reality.dest);
      setTestResult(result);
    } catch {
      setTestResult({ reachable: false, error: 'Test request failed. Backend endpoint may not be ready yet.' });
    } finally {
      setTesting(false);
    }
  };

  const updatePreset = (
    key: keyof StealthSettings['presets'],
    field: 'enabled' | 'port',
    value: boolean | number,
  ) => {
    setSettings((prev) => ({
      ...prev,
      presets: {
        ...prev.presets,
        [key]: { ...prev.presets[key], [field]: value },
      },
    }));
  };

  const updateReality = (field: keyof StealthSettings['reality'], value: string | string[]) => {
    setSettings((prev) => ({
      ...prev,
      reality: { ...prev.reality, [field]: value },
    }));
  };

  if (loading) {
    return <p className="text-slate-400">Loading...</p>;
  }

  return (
    <div className="space-y-6 max-w-3xl">
      <h1 className="text-2xl font-semibold">Transports / Stealth</h1>
      <p className="text-slate-400 text-sm">
        Configure anti-DPI transport presets, Reality masking, and fingerprint defaults.
      </p>

      {message && <p className="text-emerald-400">{message}</p>}
      {error && <p className="text-amber-400">{error}</p>}

      <StealthWarnings settings={settings} />

      {/* Transport presets */}
      <section className="bg-slate-900 border border-slate-800 rounded-lg p-5 space-y-4">
        <h2 className="text-lg font-medium">Transport presets</h2>
        {(Object.keys(settings.presets) as Array<keyof StealthSettings['presets']>).map((key) => (
          <div key={key} className="flex items-center gap-4 border-b border-slate-800 pb-3 last:border-0 last:pb-0">
            <label className="flex items-center gap-2 flex-1">
              <input
                type="checkbox"
                checked={settings.presets[key].enabled}
                onChange={(e) => updatePreset(key, 'enabled', e.target.checked)}
                className="rounded"
              />
              <span className="text-sm">{PRESET_LABELS[key]}</span>
            </label>
            <div className="flex items-center gap-2">
              <span className="text-xs text-slate-500">Port</span>
              <input
                type="number"
                value={settings.presets[key].port}
                onChange={(e) => updatePreset(key, 'port', Number(e.target.value))}
                disabled={!settings.presets[key].enabled}
                className="w-20 px-2 py-1 rounded bg-slate-800 border border-slate-700 text-sm disabled:opacity-50"
              />
            </div>
          </div>
        ))}
      </section>

      {/* Reality settings */}
      <section className="bg-slate-900 border border-slate-800 rounded-lg p-5 space-y-4">
        <h2 className="text-lg font-medium">Reality masking</h2>

        <div>
          <label className="block text-sm text-slate-400 mb-1">Dest (host:port)</label>
          <div className="flex gap-2">
            <input
              value={settings.reality.dest}
              onChange={(e) => updateReality('dest', e.target.value)}
              placeholder="cdn.example.com:443"
              className="flex-1 px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm"
              data-testid="reality-dest"
            />
            <button
              onClick={testDest}
              disabled={testing || !settings.reality.dest}
              className="px-4 py-2 rounded bg-slate-700 hover:bg-slate-600 text-sm disabled:opacity-50 whitespace-nowrap"
              data-testid="test-dest"
            >
              {testing ? 'Testing…' : 'Test availability'}
            </button>
          </div>
          {testResult && (
            <div
              className={`mt-2 text-sm p-2 rounded ${
                testResult.reachable
                  ? 'bg-emerald-900/30 text-emerald-300'
                  : 'bg-red-900/30 text-red-300'
              }`}
              data-testid="dest-test-result"
            >
              {testResult.reachable
                ? `Reachable${testResult.status_code ? ` (HTTP ${testResult.status_code})` : ''}${testResult.latency_ms ? ` — ${testResult.latency_ms}ms` : ''}`
                : testResult.error || 'Destination unreachable'}
            </div>
          )}
        </div>

        <div>
          <label className="block text-sm text-slate-400 mb-1">Server names (comma-separated)</label>
          <input
            value={settings.reality.server_names.join(', ')}
            onChange={(e) =>
              updateReality(
                'server_names',
                e.target.value.split(',').map((s) => s.trim()).filter(Boolean),
              )
            }
            placeholder="cdn.example.com"
            className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm"
          />
        </div>

        <div>
          <label className="block text-sm text-slate-400 mb-1">Default fingerprint</label>
          <select
            value={settings.reality.fingerprint}
            onChange={(e) => updateReality('fingerprint', e.target.value)}
            className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm"
          >
            {FINGERPRINT_OPTIONS.map((fp) => (
              <option key={fp} value={fp}>
                {fp}
              </option>
            ))}
          </select>
        </div>
      </section>

      {/* Fragmentation */}
      <section className="bg-slate-900 border border-slate-800 rounded-lg p-5 space-y-4">
        <h2 className="text-lg font-medium">ServerHello fragmentation</h2>
        <label className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={settings.fragmentation.enabled}
            onChange={(e) =>
              setSettings((prev) => ({
                ...prev,
                fragmentation: { ...prev.fragmentation, enabled: e.target.checked },
              }))
            }
            className="rounded"
          />
          <span className="text-sm">Enable ServerHello fragmentation</span>
        </label>
        {settings.fragmentation.enabled && (
          <div className="space-y-3">
            <div>
              <label className="block text-sm text-slate-400 mb-1">Strategy</label>
              <select
                value={settings.fragmentation.strategy || 'serverhello'}
                onChange={(e) =>
                  setSettings((prev) => ({
                    ...prev,
                    fragmentation: { ...prev.fragmentation, strategy: e.target.value },
                  }))
                }
                className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm"
              >
                <option value="serverhello">ServerHello only (recommended)</option>
                <option value="all">All packets (aggressive)</option>
              </select>
            </div>
            <div className="grid gap-3 sm:grid-cols-3">
              <div>
                <label className="block text-sm text-slate-400 mb-1">Length (bytes)</label>
                <input
                  type="text"
                  value={settings.fragmentation.length || '50-100'}
                  onChange={(e) =>
                    setSettings((prev) => ({
                      ...prev,
                      fragmentation: { ...prev.fragmentation, length: e.target.value },
                    }))
                  }
                  placeholder="50-100"
                  className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Delay (ms)</label>
                <input
                  type="text"
                  value={settings.fragmentation.delay || '10-20'}
                  onChange={(e) =>
                    setSettings((prev) => ({
                      ...prev,
                      fragmentation: { ...prev.fragmentation, delay: e.target.value },
                    }))
                  }
                  placeholder="10-20"
                  className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm"
                />
              </div>
              <div>
                <label className="block text-sm text-slate-400 mb-1">Max split</label>
                <input
                  type="text"
                  value={settings.fragmentation.max_split || '2-4'}
                  onChange={(e) =>
                    setSettings((prev) => ({
                      ...prev,
                      fragmentation: { ...prev.fragmentation, max_split: e.target.value },
                    }))
                  }
                  placeholder="2-4"
                  className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm"
                />
              </div>
            </div>
            <p className="text-xs text-slate-500">
              Applied only on VLESS+TLS inbound. REALITY presets ignore fragmentation until upstream fix.
            </p>
          </div>
        )}
      </section>

      <button
        onClick={save}
        disabled={saving}
        className="px-6 py-2 rounded bg-sky-600 hover:bg-sky-500 disabled:opacity-50"
      >
        {saving ? 'Saving…' : 'Save settings'}
      </button>
    </div>
  );
}
