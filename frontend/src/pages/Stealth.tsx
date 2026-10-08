import { useCallback, useEffect, useState } from 'react';
import {
  getStealthSettings,
  testDestAvailability,
  updateStealthSettings,
} from '../services/api';
import { StealthWarnings } from '../components/StealthWarnings';
import { PageHeader, Spinner } from '../components/ui';
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
      setTestResult({
        reachable: false,
        error: 'Test request failed. Backend endpoint may not be ready yet.',
      });
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
    return (
      <div className="page">
        <PageHeader title="Transports / Stealth" />
        <Spinner />
      </div>
    );
  }

  return (
    <div className="page max-w-3xl">
      <PageHeader
        title="Transports / Stealth"
        subtitle="Configure anti-DPI transport presets, Reality masking, and fingerprint defaults."
        actions={
          <button type="button" onClick={save} disabled={saving} className="btn-primary">
            {saving ? 'Saving…' : 'Save settings'}
          </button>
        }
      />

      {message && <p className="alert-ok">{message}</p>}
      {error && <p className="alert-warn">{error}</p>}

      <StealthWarnings settings={settings} />

      <section className="card-pad space-y-4">
        <h2 className="text-lg font-medium">Transport presets</h2>
        {(Object.keys(settings.presets) as Array<keyof StealthSettings['presets']>).map((key) => (
          <div
            key={key}
            className="flex flex-col sm:flex-row sm:items-center gap-3 border-b border-surface-border pb-3 last:border-0 last:pb-0"
          >
            <label className="flex items-center gap-2 flex-1">
              <input
                type="checkbox"
                checked={settings.presets[key].enabled}
                onChange={(e) => updatePreset(key, 'enabled', e.target.checked)}
                className="rounded border-surface-border"
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
                className="input w-24 disabled:opacity-50"
              />
            </div>
          </div>
        ))}
      </section>

      <section className="card-pad space-y-4">
        <h2 className="text-lg font-medium">Reality masking</h2>

        <div>
          <label className="label">Dest (host:port)</label>
          <div className="flex flex-col sm:flex-row gap-2">
            <input
              value={settings.reality.dest}
              onChange={(e) => updateReality('dest', e.target.value)}
              placeholder="cdn.example.com:443"
              className="input flex-1"
              data-testid="reality-dest"
            />
            <button
              type="button"
              onClick={testDest}
              disabled={testing || !settings.reality.dest}
              className="btn-secondary whitespace-nowrap"
              data-testid="test-dest"
            >
              {testing ? 'Testing…' : 'Test availability'}
            </button>
          </div>
          {testResult && (
            <div
              className={`mt-2 text-sm p-3 rounded-lg ${
                testResult.reachable ? 'alert-ok' : 'alert-error'
              }`}
              data-testid="dest-test-result"
            >
              {testResult.reachable
                ? `Reachable${testResult.status_code ? ` (HTTP ${testResult.status_code})` : ''}${
                    testResult.latency_ms ? ` — ${testResult.latency_ms}ms` : ''
                  }`
                : testResult.error || 'Destination unreachable'}
            </div>
          )}
        </div>

        <div>
          <label className="label">Server names (comma-separated)</label>
          <input
            value={settings.reality.server_names.join(', ')}
            onChange={(e) =>
              updateReality(
                'server_names',
                e.target.value
                  .split(',')
                  .map((s) => s.trim())
                  .filter(Boolean),
              )
            }
            placeholder="cdn.example.com"
            className="input"
          />
        </div>

        <div>
          <label className="label">Default fingerprint</label>
          <select
            value={settings.reality.fingerprint}
            onChange={(e) => updateReality('fingerprint', e.target.value)}
            className="input"
          >
            {FINGERPRINT_OPTIONS.map((fp) => (
              <option key={fp} value={fp}>
                {fp}
              </option>
            ))}
          </select>
        </div>
      </section>

      <section className="card-pad space-y-4">
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
            className="rounded border-surface-border"
          />
          <span className="text-sm">Enable ServerHello fragmentation</span>
        </label>
        {settings.fragmentation.enabled && (
          <div className="space-y-3">
            <div>
              <label className="label">Strategy</label>
              <select
                value={settings.fragmentation.strategy || 'serverhello'}
                onChange={(e) =>
                  setSettings((prev) => ({
                    ...prev,
                    fragmentation: { ...prev.fragmentation, strategy: e.target.value },
                  }))
                }
                className="input"
              >
                <option value="serverhello">ServerHello only (recommended)</option>
                <option value="all">All packets (aggressive)</option>
              </select>
            </div>
            <div className="grid gap-3 sm:grid-cols-3">
              <div>
                <label className="label">Length (bytes)</label>
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
                  className="input"
                />
              </div>
              <div>
                <label className="label">Delay (ms)</label>
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
                  className="input"
                />
              </div>
              <div>
                <label className="label">Max split</label>
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
                  className="input"
                />
              </div>
            </div>
            <p className="text-xs text-slate-500">
              Applied only on VLESS+TLS inbound. REALITY presets ignore fragmentation until upstream
              fix.
            </p>
          </div>
        )}
      </section>

      <button type="button" onClick={save} disabled={saving} className="btn-primary">
        {saving ? 'Saving…' : 'Save settings'}
      </button>
    </div>
  );
}
