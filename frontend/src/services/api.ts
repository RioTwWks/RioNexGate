import axios from 'axios';
import type { Device } from '../types/device';
import type { DestTestResult, ProfileLink, StealthSettings } from '../types/stealth';
import type { User } from '../types/user';

const API_KEY_STORAGE = 'rionexgate_api_key';

const api = axios.create({
  baseURL: '/api',
});

api.interceptors.request.use((config) => {
  const key = localStorage.getItem(API_KEY_STORAGE);
  if (key) {
    config.headers['X-API-Key'] = key;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      clearApiKey();
      if (!window.location.pathname.startsWith('/login')) {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  },
);

export function setApiKey(key: string) {
  localStorage.setItem(API_KEY_STORAGE, key);
}

export function getApiKey(): string | null {
  return localStorage.getItem(API_KEY_STORAGE);
}

export function clearApiKey() {
  localStorage.removeItem(API_KEY_STORAGE);
}

// --- Users ---

export async function getUser(id: number): Promise<User> {
  const res = await api.get<User>(`/users/${id}`);
  const user = res.data;
  if (!user.subscription_url && user.subscription_token) {
    user.subscription_url = `${window.location.origin}/api/subscription/${user.subscription_token}`;
  }
  return user;
}

export async function getUserDevices(userId: number): Promise<Device[]> {
  const res = await api.get<Device[]>(`/users/${userId}/devices`);
  return res.data;
}

export async function revokeDevice(userId: number, deviceId: number): Promise<void> {
  await api.delete(`/users/${userId}/devices/${deviceId}`);
}

export async function getUserProfiles(userId: number): Promise<ProfileLink[]> {
  try {
    const res = await api.get<ProfileLink[]>(`/users/${userId}/profiles`);
    return res.data;
  } catch {
    const res = await api.get<{
      links?: Array<{
        profile?: string;
        transport?: string;
        priority?: number;
        tags?: string;
        link?: string;
      }>;
    }>(`/users/${userId}/link`, { params: { all: 'true', proto: 'vless' } });
    const links = res.data.links ?? [];
    return links
      .filter((l) => l.link)
      .map((l, i) => ({
        id: l.profile || `profile-${i}`,
        name: l.profile || l.transport || `Profile ${i + 1}`,
        transport: mapTransport(l.transport),
        tags: l.tags ? l.tags.split(',').map((t) => t.trim()).filter(Boolean) : [],
        priority: l.priority ?? i,
        link: l.link!,
      }));
  }
}

function mapTransport(t?: string): ProfileLink['transport'] {
  switch (t?.toLowerCase()) {
    case 'xhttp':
      return 'xhttp';
    case 'vision':
    case 'tcp':
      return 'vision';
    case 'tls':
      return 'tls';
    case 'awg':
    case 'amneziawg':
      return 'awg';
    default:
      return 'xhttp';
  }
}

// --- Stealth ---

export async function getStealthSettings(): Promise<StealthSettings> {
  const res = await api.get<StealthSettings>('/stealth/settings');
  return res.data;
}

export async function updateStealthSettings(settings: StealthSettings): Promise<StealthSettings> {
  const res = await api.put<StealthSettings>('/stealth/settings', settings);
  return res.data;
}

export async function testDestAvailability(dest: string): Promise<DestTestResult> {
  const res = await api.post<DestTestResult>('/stealth/test-dest', { dest });
  return res.data;
}

export default api;
