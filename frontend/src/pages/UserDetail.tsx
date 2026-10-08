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
import { PageHeader, Spinner, StatusBadge, TrafficBar } from '../components/ui';
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
  const [copiedProfile, setCopiedProfile] = useState<string | null>(null);
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

  const copyProfileLink = async (id: string, link: string) => {
    await navigator.clipboard.writeText(link);
    setCopiedProfile(id);
    setTimeout(() => setCopiedProfile(null), 2000);
  };

  if (loading) {
    return (
      <div className="page">
        <Spinner />
      </div>
    );
  }

  if (error && !user) {
    return <p className="alert-error">{error}</p>;
  }

  if (!user) return null;

  const expired = new Date(user.expires_at).getTime() <= Date.now();
  const overallSync =
    devices.length === 0
      ? ('never' as const)
      : devices.some((d) => getSyncStatus(d.last_seen_at) === 'synced')
        ? ('synced' as const)
        : ('stale' as const);

  return (
    <div className="page">
      <div>
        <Link to="/users" className="text-slate-400 hover:text-white text-sm">
          ← Users
        </Link>
      </div>

      <PageHeader
        title={user.email}
        subtitle={`Expires ${new Date(user.expires_at).toLocaleDateString()}`}
        actions={
          <>
            <StatusBadge active={user.active} expired={expired} />
            <SyncStatusBadge status={overallSync} />
          </>
        }
      />

      <div className="card-pad max-w-md">
        <TrafficBar used={user.used_gb} limit={user.traffic_gb} />
      </div>

      {error && <p className="alert-error">{error}</p>}

      <UserChainSection user={user} onUpdated={setUser} />

      <section className="card-pad">
        <h2 className="text-lg font-medium mb-3">Subscription</h2>
        {user.subscription_url ? (
          <div className="space-y-3">
            <div className="flex flex-col sm:flex-row gap-2">
              <input
                readOnly
                value={user.subscription_url}
                className="input font-mono"
                data-testid="subscription-url"
              />
              <button
                type="button"
                onClick={copySubscription}
                className="btn-primary whitespace-nowrap"
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

      <section className="card-pad">
        <h2 className="text-lg font-medium mb-3">Registered devices</h2>
        {devices.length === 0 ? (
          <p className="text-slate-500 text-sm">No devices registered via RioNexTunnel yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="table">
              <thead>
                <tr>
                  <th>Label</th>
                  <th>Token</th>
                  <th>Last seen</th>
                  <th>Status</th>
                  <th className="text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {devices.map((d) => (
                  <tr key={d.id}>
                    <td>{d.label || '—'}</td>
                    <td className="font-mono text-xs">{maskToken(d.token)}</td>
                    <td>
                      {d.last_seen_at ? new Date(d.last_seen_at).toLocaleString() : 'Never'}
                    </td>
                    <td>
                      <SyncStatusBadge
                        status={getSyncStatus(d.last_seen_at)}
                        lastSeenAt={d.last_seen_at}
                      />
                    </td>
                    <td className="text-right">
                      <button
                        type="button"
                        onClick={() => handleRevoke(d.id)}
                        disabled={revoking === d.id}
                        className="text-red-400 hover:underline disabled:opacity-50 text-sm"
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

      {profiles.length > 0 && (
        <section className="card-pad">
          <h2 className="text-lg font-medium mb-3">Transport profiles</h2>
          <div className="space-y-3">
            {profiles
              .slice()
              .sort((a, b) => a.priority - b.priority)
              .map((p) => (
                <div key={p.id} className="rounded-xl border border-surface-border bg-slate-950/40 p-3">
                  <div className="flex items-center justify-between gap-3 mb-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <span className="font-medium">{p.name}</span>
                      <span className="badge-neutral uppercase">{p.transport}</span>
                      {p.tags.map((t) => (
                        <span key={t} className="badge-info">
                          {t}
                        </span>
                      ))}
                    </div>
                    <button
                      type="button"
                      onClick={() => copyProfileLink(p.id, p.link)}
                      className="link-action"
                    >
                      {copiedProfile === p.id ? 'Copied!' : 'Copy'}
                    </button>
                  </div>
                  <input readOnly value={p.link} className="input font-mono text-xs" />
                </div>
              ))}
          </div>
        </section>
      )}
    </div>
  );
}
