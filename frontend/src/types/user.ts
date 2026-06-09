export interface User {
  id: number;
  uuid: string;
  email: string;
  traffic_gb: number;
  used_gb: number;
  expires_at: string;
  active: boolean;
  created_at: string;
}

export interface StatsPoint {
  time: string;
  bytes_up: number;
  bytes_down: number;
}

export interface TotalStats {
  total_used_gb: number;
  points: StatsPoint[];
}
