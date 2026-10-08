import { useEffect, useState } from 'react';
import api from '../services/api';
import { Modal } from './ui';

const PROTOCOLS = ['vless', 'vmess', 'trojan'] as const;
type Protocol = (typeof PROTOCOLS)[number];

interface Props {
  userId: number;
  onClose: () => void;
}

export function LinkModal({ userId, onClose }: Props) {
  const [protocol, setProtocol] = useState<Protocol>('vless');
  const [link, setLink] = useState('');
  const [qrUrl, setQrUrl] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    let objectUrl: string | null = null;
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setError('');
      setQrUrl(null);
      setLink('');
      setCopied(false);

      try {
        const linkRes = await api.get<{ link: string }>(`/users/${userId}/link`, {
          params: { proto: protocol },
        });
        if (cancelled) return;
        setLink(linkRes.data.link);

        const qrRes = await api.get(`/users/${userId}/qr`, {
          params: { proto: protocol },
          responseType: 'blob',
        });
        if (cancelled) return;
        objectUrl = URL.createObjectURL(qrRes.data);
        setQrUrl(objectUrl);
      } catch {
        if (!cancelled) setError('Failed to load link');
      } finally {
        if (!cancelled) setLoading(false);
      }
    };

    load();

    return () => {
      cancelled = true;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [userId, protocol]);

  const copyLink = async () => {
    if (!link) return;
    await navigator.clipboard.writeText(link);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Modal title="Connection link" onClose={onClose} wide>
      <div className="mb-4">
        <label className="label">Protocol</label>
        <select
          value={protocol}
          onChange={(e) => setProtocol(e.target.value as Protocol)}
          className="input"
        >
          {PROTOCOLS.map((p) => (
            <option key={p} value={p}>
              {p.toUpperCase()}
            </option>
          ))}
        </select>
      </div>

      {error ? (
        <p className="alert-error">{error}</p>
      ) : (
        <>
          <textarea readOnly value={link} className="input h-24 font-mono mb-3 resize-none" />
          <div className="flex gap-2 mb-4">
            <button type="button" onClick={copyLink} disabled={!link} className="btn-primary flex-1">
              {copied ? 'Copied!' : 'Copy link'}
            </button>
          </div>
          <div className="flex justify-center mb-4 min-h-[256px] items-center rounded-xl bg-slate-950/50 border border-surface-border">
            {loading ? (
              <p className="text-slate-500 text-sm">Loading...</p>
            ) : qrUrl ? (
              <img src={qrUrl} alt="QR code" className="bg-white p-3 rounded-lg" />
            ) : (
              <p className="text-slate-500 text-sm">Loading QR...</p>
            )}
          </div>
        </>
      )}

      <button type="button" onClick={onClose} className="btn-secondary w-full">
        Close
      </button>
    </Modal>
  );
}
