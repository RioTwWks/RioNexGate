import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { PageHeader, Spinner, StatCard } from '../components/ui';
import api from '../services/api';
import type { TotalStats, User } from '../types/user';

interface ClientMetrics {
  active_devices: number;
  sync_requests: number;
  registration_fails: number;
}

interface HealthInfo {
  status: string;
  core: string;
}

export function Dashboard() {
  const [stats, setStats] = useState<TotalStats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [metrics, setMetrics] = useState<ClientMetrics | null>(null);
  const [health, setHealth] = useState<HealthInfo | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    Promise.all([
      api.get<TotalStats>('/stats/total'),
      api.get<User[]>('/users'),
      api.get<ClientMetrics>('/stats/client').catch(() => null),
      api.get<HealthInfo>('/health').catch(() => null),
    ])
      .then(([statsRes, usersRes, metricsRes, healthRes]) => {
        if (cancelled) return;
        setStats(statsRes.data);
        setUsers(usersRes.data);
        setMetrics(metricsRes?.data ?? null);
        setHealth(healthRes?.data ?? null);
      })
      .catch(() => {
        if (!cancelled) setError('Failed to load stats');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const now = Date.now();
  const activeUsers = users.filter((u) => u.active && new Date(u.expires_at).getTime() > now).length;
  const expiredUsers = users.filter((u) => new Date(u.expires_at).getTime() <= now).length;

  const chartData =
    stats?.points.map((p) => ({
      time: new Date(p.time).toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
      gb: Number(((p.bytes_up + p.bytes_down) / (1024 * 1024 * 1024)).toFixed(3)),
      up: Number((p.bytes_up / (1024 * 1024 * 1024)).toFixed(3)),
      down: Number((p.bytes_down / (1024 * 1024 * 1024)).toFixed(3)),
    })) ?? [];

  return (
    <div className="page">
      <PageHeader
        title="Dashboard"
        subtitle="Traffic, users, and core health at a glance"
        actions={
          <>
            <Link to="/users" className="btn-secondary">
              Manage users
            </Link>
            <Link to="/settings" className="btn-primary">
              Core settings
            </Link>
          </>
        }
      />

      {error && <p className="alert-error">{error}</p>}
      {loading ? (
        <Spinner />
      ) : (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
            <StatCard
              label="Total traffic"
              value={`${stats ? stats.total_used_gb.toFixed(2) : '—'} GB`}
              hint="All-time recorded usage"
            />
            <StatCard
              label="Users"
              value={users.length}
              hint={`${activeUsers} active · ${expiredUsers} expired`}
              accent="emerald"
            />
            <StatCard
              label="Devices (24h)"
              value={metrics?.active_devices ?? '—'}
              hint={
                metrics
                  ? `${metrics.sync_requests} syncs · ${metrics.registration_fails} reg fails`
                  : 'Client metrics unavailable'
              }
              accent="amber"
            />
            <StatCard
              label="Core"
              value={health?.core ?? '—'}
              hint={health?.status === 'ok' ? 'Healthy' : 'Unreachable'}
              accent={health?.status === 'ok' ? 'sky' : 'rose'}
            />
          </div>

          <div className="card-pad h-[22rem]">
            <div className="flex items-center justify-between mb-4">
              <div>
                <h2 className="text-sm font-medium text-slate-200">Traffic (last 7 days)</h2>
                <p className="text-xs text-slate-500 mt-0.5">Upload + download volume</p>
              </div>
            </div>
            {chartData.length === 0 ? (
              <div className="h-full flex items-center justify-center text-slate-500 text-sm">
                No traffic data yet
              </div>
            ) : (
              <ResponsiveContainer width="100%" height="85%">
                <AreaChart data={chartData}>
                  <defs>
                    <linearGradient id="trafficFill" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="#38bdf8" stopOpacity={0.35} />
                      <stop offset="100%" stopColor="#38bdf8" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid stroke="#243049" strokeDasharray="3 3" vertical={false} />
                  <XAxis dataKey="time" stroke="#8b9bb8" fontSize={12} tickLine={false} axisLine={false} />
                  <YAxis
                    stroke="#8b9bb8"
                    fontSize={12}
                    tickLine={false}
                    axisLine={false}
                    unit=" GB"
                    width={56}
                  />
                  <Tooltip
                    contentStyle={{
                      background: '#121a2b',
                      border: '1px solid #243049',
                      borderRadius: 12,
                      fontSize: 12,
                    }}
                    labelStyle={{ color: '#8b9bb8' }}
                  />
                  <Area
                    type="monotone"
                    dataKey="gb"
                    stroke="#38bdf8"
                    fill="url(#trafficFill)"
                    strokeWidth={2}
                    dot={false}
                    name="Total GB"
                  />
                </AreaChart>
              </ResponsiveContainer>
            )}
          </div>
        </>
      )}
    </div>
  );
}
