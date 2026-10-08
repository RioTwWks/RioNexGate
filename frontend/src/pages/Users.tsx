import { Link } from 'react-router-dom';
import { useCallback, useEffect, useMemo, useState } from 'react';
import api from '../services/api';
import { LinkModal } from '../components/LinkModal';
import { UserForm } from '../components/UserForm';
import { EmptyState, Modal, PageHeader, Spinner, StatusBadge, TrafficBar } from '../components/ui';
import type { User } from '../types/user';

export function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [query, setQuery] = useState('');
  const [showAdd, setShowAdd] = useState(false);
  const [editUser, setEditUser] = useState<User | null>(null);
  const [linkUserId, setLinkUserId] = useState<number | null>(null);

  const fetchUsers = useCallback(async () => {
    try {
      const res = await api.get<User[]>('/users');
      setUsers(res.data);
      setError('');
    } catch {
      setError('Failed to load users');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchUsers();
  }, [fetchUsers]);

  const deleteUser = async (id: number) => {
    if (!confirm('Delete this user?')) return;
    try {
      await api.delete(`/users/${id}`);
      fetchUsers();
    } catch {
      setError('Failed to delete user');
    }
  };

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return users;
    return users.filter((u) => u.email.toLowerCase().includes(q));
  }, [users, query]);

  const now = Date.now();

  return (
    <div className="page">
      <PageHeader
        title="Users"
        subtitle={`${users.length} accounts · subscription links and traffic limits`}
        actions={
          <button type="button" onClick={() => setShowAdd(true)} className="btn-primary">
            Add user
          </button>
        }
      />

      {error && <p className="alert-error">{error}</p>}

      <div className="flex flex-col sm:flex-row gap-3 sm:items-center">
        <input
          className="input sm:max-w-xs"
          placeholder="Filter by email…"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
        />
        <p className="text-xs text-slate-500">
          Showing {filtered.length} of {users.length}
        </p>
      </div>

      {loading ? (
        <Spinner />
      ) : (
        <div className="table-shell overflow-x-auto">
          <table className="table">
            <thead>
              <tr>
                <th>Email</th>
                <th>Traffic</th>
                <th>Expires</th>
                <th>Status</th>
                <th className="text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filtered.map((u) => {
                const expired = new Date(u.expires_at).getTime() <= now;
                return (
                  <tr key={u.id} className={expired ? 'opacity-80' : undefined}>
                    <td>
                      <Link to={`/users/${u.id}`} className="font-medium text-slate-100 hover:text-sky-300">
                        {u.email}
                      </Link>
                    </td>
                    <td>
                      <TrafficBar used={u.used_gb} limit={u.traffic_gb} />
                    </td>
                    <td className="whitespace-nowrap text-slate-300">
                      {new Date(u.expires_at).toLocaleDateString()}
                    </td>
                    <td>
                      <StatusBadge active={u.active} expired={expired} />
                    </td>
                    <td className="text-right whitespace-nowrap space-x-3">
                      <Link to={`/users/${u.id}`} className="text-emerald-400 hover:underline text-sm">
                        Details
                      </Link>
                      <button type="button" onClick={() => setLinkUserId(u.id)} className="link-action">
                        Link
                      </button>
                      <button
                        type="button"
                        onClick={() => setEditUser(u)}
                        className="text-slate-300 hover:underline text-sm"
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        onClick={() => deleteUser(u.id)}
                        className="text-red-400 hover:underline text-sm"
                      >
                        Delete
                      </button>
                    </td>
                  </tr>
                );
              })}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={5}>
                    <EmptyState
                      title={users.length === 0 ? 'No users yet' : 'No matches'}
                      description={
                        users.length === 0
                          ? 'Create a user to issue subscription and client links.'
                          : 'Try a different email filter.'
                      }
                    />
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {showAdd && (
        <Modal title="Add user" onClose={() => setShowAdd(false)}>
          <UserForm
            submitLabel="Create"
            onCancel={() => setShowAdd(false)}
            onSubmit={async (data) => {
              await api.post('/users', {
                email: data.email,
                traffic_gb: data.traffic_gb,
                expire_days: data.expire_days,
              });
              setShowAdd(false);
              fetchUsers();
            }}
          />
        </Modal>
      )}

      {editUser && (
        <Modal title="Edit user" onClose={() => setEditUser(null)}>
          <UserForm
            initial={{
              email: editUser.email,
              traffic_gb: editUser.traffic_gb,
              expire_days: 30,
              active: editUser.active,
            }}
            showExpireDays={false}
            onCancel={() => setEditUser(null)}
            onSubmit={async (data) => {
              await api.put(`/users/${editUser.id}`, {
                email: data.email,
                traffic_gb: data.traffic_gb,
                active: data.active,
              });
              setEditUser(null);
              fetchUsers();
            }}
          />
        </Modal>
      )}

      {linkUserId !== null && (
        <LinkModal userId={linkUserId} onClose={() => setLinkUserId(null)} />
      )}
    </div>
  );
}
