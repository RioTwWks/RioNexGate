import { useState } from 'react';
import type { CreateNodeInput, Node, NodeRole } from '../types/node';
import { parseCredentials, stringifyCredentials } from '../types/node';

interface Props {
  initial?: Node;
  submitLabel: string;
  onCancel: () => void;
  onSubmit: (data: CreateNodeInput) => Promise<void>;
}

export function NodeForm({ initial, submitLabel, onCancel, onSubmit }: Props) {
  const creds = initial ? parseCredentials(initial.credentials) : {};
  const [name, setName] = useState(initial?.name ?? '');
  const [address, setAddress] = useState(initial?.address ?? '');
  const [port, setPort] = useState(initial?.port ?? 443);
  const [role, setRole] = useState<NodeRole>(initial?.role ?? 'entry');
  const [protocol, setProtocol] = useState(initial?.protocol ?? 'vless');
  const [region, setRegion] = useState(initial?.region ?? '');
  const [priority, setPriority] = useState(initial?.priority ?? 100);
  const [active, setActive] = useState(initial?.active ?? true);
  const [uuid, setUuid] = useState(creds.uuid ?? '');
  const [publicKey, setPublicKey] = useState(creds.public_key ?? '');
  const [shortId, setShortId] = useState(creds.short_id ?? '');
  const [sni, setSni] = useState(creds.sni ?? '');
  const [flow, setFlow] = useState(creds.flow ?? 'xtls-rprx-vision');
  const [security, setSecurity] = useState(creds.security ?? 'reality');
  const [network, setNetwork] = useState(creds.network ?? 'tcp');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      const baseCreds = initial ? parseCredentials(initial.credentials) : {};
      const credentialFields: Record<string, unknown> = { ...baseCreds };
      if (role === 'exit') {
        Object.assign(credentialFields, {
          uuid: uuid || undefined,
          public_key: publicKey || undefined,
          short_id: shortId || undefined,
          sni: sni || undefined,
          flow: network === 'tcp' ? flow || undefined : undefined,
          security: security || undefined,
          network: network || undefined,
          fingerprint: credentialFields.fingerprint ?? 'firefox',
        });
      } else {
        Object.assign(credentialFields, {
          uuid: uuid || undefined,
          public_key: publicKey || undefined,
          short_id: shortId || undefined,
        });
      }
      await onSubmit({
        name,
        address,
        port,
        role,
        protocol,
        region: region || undefined,
        priority,
        active,
        credentials: stringifyCredentials(credentialFields),
      });
    } catch {
      setError('Failed to save node');
    } finally {
      setSaving(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && <p className="alert-error">{error}</p>}
      <div>
        <label className="label">Name</label>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
          className="input"
          data-testid="node-name"
        />
      </div>
      <div>
        <label className="label">Role</label>
        <select
          value={role}
          onChange={(e) => setRole(e.target.value as NodeRole)}
          className="input"
          data-testid="node-role"
        >
          <option value="entry">Entry</option>
          <option value="exit">Exit</option>
        </select>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label className="label">Address</label>
          <input
            value={address}
            onChange={(e) => setAddress(e.target.value)}
            required
            className="input"
            data-testid="node-address"
          />
        </div>
        <div>
          <label className="label">Port</label>
          <input
            type="number"
            value={port}
            onChange={(e) => setPort(Number(e.target.value))}
            required
            className="input"
            data-testid="node-port"
          />
        </div>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        <div>
          <label className="label">Region</label>
          <input
            value={region}
            onChange={(e) => setRegion(e.target.value)}
            placeholder="RU / NL / …"
            className="input"
          />
        </div>
        <div>
          <label className="label">Priority</label>
          <input
            type="number"
            value={priority}
            onChange={(e) => setPriority(Number(e.target.value))}
            className="input"
          />
        </div>
      </div>
      <div>
        <label className="label">Protocol</label>
        <select
          value={protocol}
          onChange={(e) => setProtocol(e.target.value)}
          className="input"
        >
          <option value="vless">VLESS</option>
          <option value="vmess">VMess</option>
          <option value="trojan">Trojan</option>
        </select>
      </div>
      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          checked={active}
          onChange={(e) => setActive(e.target.checked)}
          className="rounded border-surface-border"
        />
        Active
      </label>
      <details open={role === 'exit'} className="rounded-xl border border-surface-border p-3">
        <summary className="text-sm cursor-pointer text-slate-300">Credentials</summary>
        <div className="mt-3 space-y-2">
          <input
            value={uuid}
            onChange={(e) => setUuid(e.target.value)}
            placeholder="Relay UUID (EU inbound user)"
            className="input font-mono text-xs"
          />
          <input
            value={publicKey}
            onChange={(e) => setPublicKey(e.target.value)}
            placeholder="EU Reality public key (pbk)"
            className="input font-mono text-xs"
          />
          <input
            value={shortId}
            onChange={(e) => setShortId(e.target.value)}
            placeholder="EU Reality short ID"
            className="input font-mono text-xs"
          />
          {role === 'exit' && (
            <>
              <input
                value={sni}
                onChange={(e) => setSni(e.target.value)}
                placeholder="SNI (EU serverNames, e.g. www.cloudflare.com)"
                className="input font-mono text-xs"
                data-testid="node-sni"
              />
              <select
                value={security}
                onChange={(e) => setSecurity(e.target.value)}
                className="input text-xs"
              >
                <option value="reality">reality</option>
                <option value="tls">tls</option>
                <option value="none">none</option>
              </select>
              <select
                value={network}
                onChange={(e) => setNetwork(e.target.value)}
                className="input text-xs"
              >
                <option value="tcp">tcp (Vision)</option>
                <option value="xhttp">xhttp</option>
              </select>
              {network === 'tcp' && (
                <input
                  value={flow}
                  onChange={(e) => setFlow(e.target.value)}
                  placeholder="flow (xtls-rprx-vision)"
                  className="input font-mono text-xs"
                />
              )}
              <p className="text-xs text-slate-400">
                SNI must match EU inbound Reality serverNames — not rio2skadi.pro.
              </p>
            </>
          )}
        </div>
      </details>
      <div className="flex justify-end gap-2 pt-1">
        <button type="button" onClick={onCancel} className="btn-secondary">
          Cancel
        </button>
        <button type="submit" disabled={saving} className="btn-primary">
          {saving ? 'Saving…' : submitLabel}
        </button>
      </div>
    </form>
  );
}
