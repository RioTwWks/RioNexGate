import type { StealthSettings } from '../types/stealth';
import { UNSAFE_DESTS, UNSAFE_FINGERPRINTS } from '../types/stealth';

interface Props {
  settings: StealthSettings;
}

interface Warning {
  message: string;
  severity: 'warning' | 'error';
}

function collectWarnings(settings: StealthSettings): Warning[] {
  const warnings: Warning[] = [];
  const { presets, reality, fragmentation } = settings;

  if (!presets.xhttp_reality.enabled && !presets.vision_reality.enabled) {
    warnings.push({
      severity: 'warning',
      message: 'No Reality transports enabled. TCP-only or TLS-only setups are easier to fingerprint.',
    });
  }

  if (!presets.xhttp_reality.enabled && presets.vision_reality.enabled) {
    warnings.push({
      severity: 'warning',
      message: 'Only TCP+Vision enabled without XHTTP. Consider enabling XHTTP+Reality for better DPI resistance.',
    });
  }

  const destHost = reality.dest.replace(/:\d+$/, '').toLowerCase();
  if (UNSAFE_DESTS.some((d) => destHost.includes(d))) {
    warnings.push({
      severity: 'error',
      message: `Reality dest "${reality.dest}" is a popular target under increased DPI scrutiny. Use a lesser-known CDN or your own site.`,
    });
  }

  if (UNSAFE_FINGERPRINTS.includes(reality.fingerprint)) {
    warnings.push({
      severity: 'warning',
      message: `Fingerprint "${reality.fingerprint}" is commonly flagged. Prefer firefox or edge.`,
    });
  }

  if (fragmentation.enabled && fragmentation.strategy && fragmentation.strategy !== 'serverhello') {
    warnings.push({
      severity: 'warning',
      message: 'Aggressive fragmentation of all packets increases PPS anomalies. Fragment only ServerHello when possible.',
    });
  }

  if (!reality.dest) {
    warnings.push({
      severity: 'error',
      message: 'Reality dest is not configured. Clients cannot connect without a valid dest.',
    });
  }

  return warnings;
}

export function StealthWarnings({ settings }: Props) {
  const warnings = collectWarnings(settings);

  if (warnings.length === 0) {
    return (
      <div className="rounded-lg border border-emerald-800/50 bg-emerald-900/20 p-4 text-sm text-emerald-300">
        No security warnings detected for current settings.
      </div>
    );
  }

  return (
    <div className="space-y-2">
      {warnings.map((w, i) => (
        <div
          key={i}
          className={`rounded-lg border p-3 text-sm ${
            w.severity === 'error'
              ? 'border-red-800/50 bg-red-900/20 text-red-300'
              : 'border-amber-800/50 bg-amber-900/20 text-amber-300'
          }`}
        >
          {w.message}
        </div>
      ))}
    </div>
  );
}
