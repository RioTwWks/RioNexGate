export interface TransportPreset {
  enabled: boolean;
  port: number;
}

export interface RealitySettings {
  dest: string;
  server_names: string[];
  fingerprint: string;
  short_ids: string[];
  private_key?: string;
}

export interface FragmentationSettings {
  enabled: boolean;
  strategy?: string;
}

export interface StealthSettings {
  presets: {
    xhttp_reality: TransportPreset;
    vision_reality: TransportPreset;
    tls: TransportPreset;
    amneziawg: TransportPreset;
  };
  reality: RealitySettings;
  fragmentation: FragmentationSettings;
}

export interface DestTestResult {
  reachable: boolean;
  status_code?: number;
  latency_ms?: number;
  error?: string;
}

export interface ProfileLink {
  id: string;
  name: string;
  transport: 'xhttp' | 'vision' | 'tls' | 'awg';
  tags: string[];
  priority: number;
  link: string;
}

export const FINGERPRINT_OPTIONS = ['firefox', 'edge', 'chrome', 'safari', 'random'] as const;
export type Fingerprint = (typeof FINGERPRINT_OPTIONS)[number];

export const UNSAFE_DESTS = ['yahoo.com', 'vk.com', 'google.com', 'facebook.com', 'twitter.com'];
export const UNSAFE_FINGERPRINTS = ['chrome', 'safari'];
