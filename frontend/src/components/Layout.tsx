import { Link, Outlet, useLocation } from 'react-router-dom';
import { clearApiKey } from '../services/api';

const nav = [
  { to: '/', label: 'Dashboard' },
  { to: '/users', label: 'Users' },
  { to: '/settings', label: 'Settings' },
];

export function Layout() {
  const location = useLocation();

  const logout = () => {
    clearApiKey();
    window.location.replace('/login');
  };

  return (
    <div className="min-h-screen flex">
      <aside className="w-56 bg-slate-900 border-r border-slate-800 p-4 flex flex-col">
        <h1 className="text-lg font-semibold mb-6 text-sky-400">proxy-mgr</h1>
        <nav className="flex flex-col gap-1">
          {nav.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              className={`px-3 py-2 rounded ${
                location.pathname === item.to
                  ? 'bg-sky-600 text-white'
                  : 'text-slate-300 hover:bg-slate-800'
              }`}
            >
              {item.label}
            </Link>
          ))}
        </nav>
        <button
          onClick={logout}
          className="mt-auto text-sm text-slate-400 hover:text-white text-left px-3 py-2"
        >
          Logout
        </button>
      </aside>
      <main className="flex-1 p-6 overflow-auto">
        <Outlet />
      </main>
    </div>
  );
}
