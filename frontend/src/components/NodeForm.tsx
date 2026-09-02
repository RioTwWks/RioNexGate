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
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    try {
      await onSubmit({
        name, address, port, role, protocol, region: region || undefined, priority, active,
        credentials: stringifyCredentials({ uuid: uuid || undefined, public_key: publicKey || undefined, short_id: shortId || undefined }),
      });
    } catch { setError('Failed to save node'); } finally { setSaving(false); }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {error && <p className="text-red-400 text-sm">{error}</p>}
      <input value={name} onChange={(e) => setName(e.target.value)} required placeholder="Name" className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm" data-testid="node-name" />
      <select value={role} onChange={(e) => setRole(e.target.value as NodeRole)} className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm" data-testid="node-role">
        <option value="entry">Entry</option><option value="exit">Exit</option>
      </select>
      <input value={address} onChange={(e) => setAddress(e.target.value)} required placeholder="Address" className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm" data-testid="node-address" />
      <input type="number" value={port} onChange={(e) => setPort(Number(e.target.value))} required placeholder="Port" className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm" data-testid="node-port" />
      <input value={region} onChange={(e) => setRegion(e.target.value)} placeholder="Region" className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm" />
      <select value={protocol} onChange={(e) => setProtocol(e.target.value)} className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm">
        <option value="vless">VLESS</option><option value="vmess">VMess</option><option value="trojan">Trojan</option>
      </select>
      <input type="number" value={priority} onChange={(e) => setPriority(Number(e.target.value))} placeholder="Priority" className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm" />
      <label className="flex items-center gap-2"><input type="checkbox" checked={active} onChange={(e) => setActive(e.target.checked)} /><span className="text-sm">Active</span></label>
      <details><summary className="text-sm cursor-pointer">Credentials</summary>
        <input value={uuid} onChange={(e) => setUuid(e.target.value)} placeholder="UUID" className="w-full mt-2 px-2 py-1.5 rounded bg-slate-800 text-xs font-mono" />
        <input value={publicKey} onChange={(e) => setPublicKey(e.target.value)} placeholder="Public key" className="w-full mt-2 px-2 py-1.5 rounded bg-slate-800 text-xs font-mono" />
        <input value={shortId} onChange={(e) => setShortId(e.target.value)} placeholder="Short ID" className="w-full mt-2 px-2 py-1.5 rounded bg-slate-800 text-xs font-mono" />
      </details>
      <div className="flex justify-end gap-2">
        <button type="button" onClick={onCancel} className="px-4 py-2 text-sm text-slate-400">Cancel</button>
        <button type="submit" disabled={saving} className="px-4 py-2 rounded bg-sky-600 text-sm">{saving ? 'Saving…' : submitLabel}</button>
      </div>
    </form>
  );
}
