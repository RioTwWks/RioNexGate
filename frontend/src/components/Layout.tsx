import { useEffect, useState } from 'react';
import { Link, Outlet, useLocation } from 'react-router-dom';
import api, { clearApiKey } from '../services/api';

const nav = [
  { to: '/', label: 'Dashboard', icon: '▣' },
  { to: '/users', label: 'Users', icon: '◉' },
  { to: '/nodes', label: 'Nodes', icon: '⬡' },
  { to: '/stealth', label: 'Stealth', icon: '◈' },
  { to: '/settings', label: 'Settings', icon: '⚙' },
];

function isActive(pathname: string, to: string) {
  if (to === '/') return pathname === '/';
  return pathname === to || pathname.startsWith(`${to}/`);
}

export function Layout() {
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [core, setCore] = useState<string>('—');
  const [healthy, setHealthy] = useState(false);

  useEffect(() => {
    setOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    let cancelled = false;
    const load = () => {
      api
        .get<{ status: string; core: string }>('/health')
        .then((res) => {
          if (cancelled) return;
          setCore(res.data.core || '—');
          setHealthy(res.data.status === 'ok');
        })
        .catch(() => {
          if (!cancelled) {
            setHealthy(false);
            setCore('—');
          }
        });
    };
    load();
    const id = window.setInterval(load, 30000);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const logout = () => {
    clearApiKey();
    window.location.replace('/login');
  };

  const sidebar = (
    <div className="flex h-full flex-col">
      <div className="px-4 pt-5 pb-4">
        <Link to="/" className="block group">
          <div className="text-xs uppercase tracking-[0.2em] text-sky-400/80">Panel</div>
          <h1 className="text-xl font-semibold text-sky-300 group-hover:text-sky-200 transition-colors">
            RioNexGate
          </h1>
        </Link>
        <div className="mt-4 flex items-center gap-2 rounded-lg border border-surface-border bg-slate-950/40 px-3 py-2">
          <span
            className={`w-2 h-2 rounded-full ${healthy ? 'bg-emerald-400 animate-pulse-soft' : 'bg-rose-400'}`}
          />
          <div className="min-w-0">
            <p className="text-[11px] uppercase tracking-wide text-slate-500">Core</p>
            <p className="text-sm font-medium truncate">{core}</p>
          </div>
        </div>
      </div>

      <nav className="flex-1 px-3 space-y-1">
        {nav.map((item) => (
          <Link
            key={item.to}
            to={item.to}
            className={isActive(location.pathname, item.to) ? 'nav-link-active' : 'nav-link'}
          >
            <span className="opacity-70 w-4 text-center text-xs" aria-hidden>
              {item.icon}
            </span>
            {item.label}
          </Link>
        ))}
      </nav>

      <div className="p-3 border-t border-surface-border">
        <button type="button" onClick={logout} className="nav-link w-full text-left">
          <span className="opacity-70 w-4 text-center text-xs" aria-hidden>
            ⎋
          </span>
          Logout
        </button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen lg:flex">
      {/* Desktop sidebar */}
      <aside className="hidden lg:flex lg:w-60 xl:w-64 shrink-0 border-r border-surface-border bg-surface-raised/80 backdrop-blur-md">
        {sidebar}
      </aside>

      {/* Mobile top bar */}
      <div className="lg:hidden sticky top-0 z-40 border-b border-surface-border bg-surface-raised/90 backdrop-blur-md">
        <div className="flex items-center justify-between px-4 py-3">
          <div>
            <p className="text-[10px] uppercase tracking-[0.18em] text-sky-400/80">RioNexGate</p>
            <p className="text-sm font-medium">{nav.find((n) => isActive(location.pathname, n.to))?.label ?? 'Panel'}</p>
          </div>
          <button
            type="button"
            className="btn-secondary px-3"
            onClick={() => setOpen((v) => !v)}
            aria-label="Toggle menu"
          >
            {open ? 'Close' : 'Menu'}
          </button>
        </div>
        {open && (
          <div className="border-t border-surface-border bg-surface-raised px-2 pb-3">
            {sidebar}
          </div>
        )}
      </div>

      <main className="flex-1 min-w-0 p-4 sm:p-6 lg:p-8 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}
