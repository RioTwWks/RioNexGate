import { useCallback, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  getUser,
  getUserDevices,
  getUserProfiles,
  revokeDevice,
} from '../services/api';
import { SyncStatusBadge } from '../components/SyncStatusBadge';
import { UserChainSection } from '../components/UserChainSection';
import type { Device } from '../types/device';
import { getSyncStatus } from '../types/device';
import type { ProfileLink } from '../types/stealth';
import type { User } from '../types/user';

interface Props {
  userId: number;
}

function maskToken(token: string): string {
  if (token.length <= 8) return token;
  return `${token.slice(0, 4)}…${token.slice(-4)}`;
}

export function UserDetail({ userId }: Props) {
  const [user, setUser] = useState<User | null>(null);
  const [devices, setDevices] = useState<Device[]>([]);
  const [profiles, setProfiles] = useState<ProfileLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);
  const [revoking, setRevoking] = useState<number | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const [userData, deviceData, profileData] = await Promise.all([
        getUser(userId),
        getUserDevices(userId).catch(() => [] as Device[]),
        getUserProfiles(userId).catch(() => [] as ProfileLink[]),
      ]);
      setUser(userData);
      setDevices(deviceData);
      setProfiles(profileData);
    } catch {
      setError('Failed to load user details');
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    load();
  }, [load]);

  const copySubscription = async () => {
    if (!user?.subscription_url) return;
    await navigator.clipboard.writeText(user.subscription_url);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleRevoke = async (deviceId: number) => {
    if (!confirm('Revoke this device token? The client will need to re-register.')) return;
    setRevoking(deviceId);
    try {
      await revokeDevice(userId, deviceId);
      await load();
    } catch {
      setError('Failed to revoke device');
    } finally {
      setRevoking(null);
    }
  };

  const copyProfileLink = async (link: string) => {
    await navigator.clipboard.writeText(link);
  };

  if (loading) {
    return <p className="text-slate-400">Loading...</p>;
  }

  if (error && !user) {
    return <p className="text-red-400">{error}</p>;
  }

  if (!user) return null;

  const overallSync = devices.length === 0
    ? 'never' as const
    : devices.some((d) => getSyncStatus(d.last_seen_at) === 'synced')
      ? 'synced' as const
      : 'stale' as const;

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link to="/users" className="text-slate-400 hover:text-white text-sm">
          ← Users
        </Link>
      </div>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{user.email}</h1>
          <p className="text-slate-400 text-sm mt-1">
            {user.used_gb.toFixed(2)} / {user.traffic_gb} GB · Expires{' '}
            {new Date(user.expires_at).toLocaleDateString()}
          </p>
        </div>
        <SyncStatusBadge status={overallSync} />
      </div>

      {error && <p className="text-red-400">{error}</p>}

      <UserChainSection user={user} onUpdated={setUser} />

      {/* Subscription URL */}
      <section className="bg-slate-900 border border-slate-800 rounded-lg p-5">
        <h2 className="text-lg font-medium mb-3">Subscription</h2>
        {user.subscription_url ? (
          <div className="space-y-3">
            <div className="flex gap-2">
              <input
                readOnly
                value={user.subscription_url}
                className="flex-1 px-3 py-2 rounded bg-slate-800 border border-slate-700 text-sm font-mono"
                data-testid="subscription-url"
              />
              <button
                onClick={copySubscription}
                className="px-4 py-2 rounded bg-sky-600 hover:bg-sky-500 text-sm whitespace-nowrap"
                data-testid="copy-subscription"
              >
                {copied ? 'Copied!' : 'Copy subscription'}
              </button>
            </div>
            {user.subscription_token && (
              <p className="text-xs text-slate-500">
                Token: <span className="font-mono">{maskToken(user.subscription_token)}</span>
              </p>
            )}
          </div>
        ) : (
          <p className="text-slate-500 text-sm">
            Subscription URL not available yet. Backend may still be provisioning the token.
          </p>
        )}
      </section>

      {/* Devices */}
      <section className="bg-slate-900 border border-slate-800 rounded-lg p-5">
        <h2 className="text-lg font-medium mb-3">Registered devices</h2>
        {devices.length === 0 ? (
          <p className="text-slate-500 text-sm">No devices registered via RioNexTunnel yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="text-slate-400">
                <tr>
                  <th className="text-left p-2">Label</th>
                  <th className="text-left p-2">Token</th>
                  <th className="text-left p-2">Last seen</th>
                  <th className="text-left p-2">Status</th>
                  <th className="text-right p-2">Actions</th>
                </tr>
              </thead>
              <tbody>
                {devices.map((d) => (
                  <tr key={d.id} className="border-t border-slate-800">
                    <td className="p-2">{d.label || '—'}</td>
                    <td className="p-2 font-mono text-xs">{maskToken(d.token)}</td>
                    <td className="p-2">
                      {d.last_seen_at
                        ? new Date(d.last_seen_at).toLocaleString()
                        : 'Never'}
                    </td>
                    <td className="p-2">
                      <SyncStatusBadge status={getSyncStatus(d.last_seen_at)} lastSeenAt={d.last_seen_at} />
                    </td>
                    <td className="p-2 text-right">
                      <button
                        onClick={() => handleRevoke(d.id)}
                        disabled={revoking === d.id}
                        className="text-red-400 hover:underline disabled:opacity-50"
                      >
                        {revoking === d.id ? 'Revoking…' : 'Revoke'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      {/* Profile links */}
      {profiles.length > 0 && (
        <section className="bg-slate-900 border border-slate-800 rounded-lg p-5">
          <h2 className="text-lg font-medium mb-3">Transport profiles</h2>
          <div className="space-y-3">
            {profiles
              .sort((a, b) => a.priority - b.priority)
              .map((p) => (
                <div key={p.id} className="border border-slate-800 rounded p-3">
                  <div className="flex items-center justify-between mb-2">
                    <div className="flex items-center gap-2">
                      <span className="font-medium">{p.name}</span>
                      <span className="text-xs px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 uppercase">
                        {p.transport}
                      </span>
                      {p.tags.map((t) => (
                        <span key={t} className="text-xs px-1.5 py-0.5 rounded bg-sky-900/40 text-sky-300">
                          {t}
                        </span>
                      ))}
                    </div>
                    <button
                      onClick={() => copyProfileLink(p.link)}
                      className="text-sky-400 hover:underline text-sm"
                    >
                      Copy
                    </button>
                  </div>
                  <input
                    readOnly
                    value={p.link}
                    className="w-full px-2 py-1.5 rounded bg-slate-800 border border-slate-700 text-xs font-mono"
                  />
                </div>
              ))}
          </div>
        </section>
      )}
    </div>
  );
}
