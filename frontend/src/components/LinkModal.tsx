import { useEffect, useState } from 'react';
import api from '../services/api';

interface Props {
  userId: number;
  onClose: () => void;
}

export function LinkModal({ userId, onClose }: Props) {
  const [link, setLink] = useState('');
  const [qrUrl, setQrUrl] = useState<string | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    let objectUrl: string | null = null;

    api
      .get<{ link: string }>(`/users/${userId}/link`, { params: { proto: 'vless' } })
      .then((res) => setLink(res.data.link))
      .catch(() => setError('Failed to load link'));

    api
      .get(`/users/${userId}/qr`, { params: { proto: 'vless' }, responseType: 'blob' })
      .then((res) => {
        objectUrl = URL.createObjectURL(res.data);
        setQrUrl(objectUrl);
      })
      .catch(() => setError('Failed to load QR code'));

    return () => {
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [userId]);

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50" onClick={onClose}>
      <div
        className="bg-slate-900 border border-slate-700 rounded-lg p-6 max-w-lg w-full mx-4"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">Connection link</h2>
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
              {qrUrl ? (
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
