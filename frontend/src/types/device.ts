export interface Device {
  id: number;
  token: string;
  label: string;
  last_seen_at: string | null;
  created_at: string;
}

export type SyncStatus = 'synced' | 'stale' | 'never';

export function getSyncStatus(lastSeenAt: string | null): SyncStatus {
  if (!lastSeenAt) return 'never';
  const lastSeen = new Date(lastSeenAt).getTime();
  const hoursAgo = (Date.now() - lastSeen) / (1000 * 60 * 60);
  return hoursAgo <= 24 ? 'synced' : 'stale';
}
