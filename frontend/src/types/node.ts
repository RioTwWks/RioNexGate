export type NodeRole = 'entry' | 'exit';
export interface Node { id: number; name: string; address: string; port: number; active: boolean; role: NodeRole; protocol: string; credentials?: string; region?: string; priority: number; }
export interface CreateNodeInput { name: string; address: string; port: number; active?: boolean; role: NodeRole; protocol: string; credentials?: string; region?: string; priority: number; }
export interface UpdateNodeInput { name?: string; address?: string; port?: number; active?: boolean; role?: NodeRole; protocol?: string; credentials?: string; region?: string; priority?: number; }
export interface NodeHealthResult { reachable: boolean; check_type: string; latency_ms?: number; error?: string; }
export function parseCredentials(raw?: string) { if (!raw) return {}; try { return JSON.parse(raw); } catch { return {}; } }
export function stringifyCredentials(creds: Record<string, unknown>) { const c = Object.fromEntries(Object.entries(creds).filter(([,v]) => v)); return Object.keys(c).length ? JSON.stringify(c) : ''; }
