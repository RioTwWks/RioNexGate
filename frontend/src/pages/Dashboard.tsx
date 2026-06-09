import { useEffect, useState } from 'react';
import { Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import api from '../services/api';
import type { TotalStats } from '../types/user';

export function Dashboard() {
  const [stats, setStats] = useState<TotalStats | null>(null);
  const [error, setError] = useState('');

  useEffect(() => {
    api
      .get<TotalStats>('/stats/total')
      .then((res) => setStats(res.data))
      .catch(() => setError('Failed to load stats'));
  }, []);

  const chartData =
    stats?.points.map((p) => ({
      time: new Date(p.time).toLocaleDateString(),
      gb: (p.bytes_up + p.bytes_down) / (1024 * 1024 * 1024),
    })) ?? [];

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Dashboard</h1>
      {error && <p className="text-red-400 mb-4">{error}</p>}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        <div className="bg-slate-900 border border-slate-800 rounded-lg p-4">
          <p className="text-slate-400 text-sm">Total traffic used</p>
          <p className="text-3xl font-bold text-sky-400">
            {stats ? stats.total_used_gb.toFixed(2) : '—'} GB
          </p>
        </div>
      </div>
      <div className="bg-slate-900 border border-slate-800 rounded-lg p-4 h-80">
        <h2 className="text-sm text-slate-400 mb-4">Traffic (last 7 days)</h2>
        {chartData.length === 0 ? (
          <p className="text-slate-500">No traffic data yet</p>
        ) : (
          <ResponsiveContainer width="100%" height="90%">
            <LineChart data={chartData}>
              <XAxis dataKey="time" stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip contentStyle={{ background: '#1e293b', border: 'none' }} />
              <Line type="monotone" dataKey="gb" stroke="#38bdf8" dot={false} />
            </LineChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  );
}
