import { useEffect, useState } from 'react';
import api from '../services/api';

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

  useEffect(() => {
    let objectUrl: string | null = null;
    let cancelled = false;

    const load = async () => {
      setLoading(true);
      setError('');
      setQrUrl(null);
      setLink('');

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

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={onClose}>
      <div
        className="bg-slate-900 border border-slate-700 rounded-lg p-6 max-w-lg w-full mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">Connection link</h2>
        <div className="mb-4">
          <label className="block text-sm text-slate-400 mb-1">Protocol</label>
          <select
            value={protocol}
            onChange={(e) => setProtocol(e.target.value as Protocol)}
            className="w-full px-3 py-2 rounded bg-slate-800 border border-slate-700"
          >
            {PROTOCOLS.map((p) => (
              <option key={p} value={p}>
                {p.toUpperCase()}
              </option>
            ))}
          </select>
        </div>
        {error ? (
          <p className="text-red-400">{error}</p>
        ) : (
          <>
            <textarea
              readOnly
              value={link}
              className="w-full h-24 px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm font-mono mb-4"
            />
            <div className="flex justify-center mb-4 min-h-[256px] items-center">
              {loading ? (
                <p className="text-slate-500">Loading...</p>
              ) : qrUrl ? (
                <img src={qrUrl} alt="QR code" className="bg-white p-2 rounded" />
              ) : (
                <p className="text-slate-500">Loading QR...</p>
              )}
            </div>
          </>
        )}
        <button onClick={onClose} className="w-full py-2 rounded bg-slate-700 hover:bg-slate-600">
          Close
        </button>
      </div>
    </div>
  );
}
