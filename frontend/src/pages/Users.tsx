import { useCallback, useEffect, useState } from 'react';
import api from '../services/api';
import { LinkModal } from '../components/LinkModal';
import { UserForm } from '../components/UserForm';
import type { User } from '../types/user';

export function Users() {
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
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
    await api.delete(`/users/${id}`);
    fetchUsers();
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold">Users</h1>
        <button
          onClick={() => setShowAdd(true)}
          className="px-4 py-2 rounded bg-sky-600 hover:bg-sky-500"
        >
          Add user
        </button>
      </div>

      {error && <p className="text-red-400 mb-4">{error}</p>}
      {loading ? (
        <p className="text-slate-400">Loading...</p>
      ) : (
        <div className="overflow-x-auto rounded-lg border border-slate-800">
          <table className="w-full text-sm">
            <thead className="bg-slate-900 text-slate-400">
              <tr>
                <th className="text-left p-3">Email</th>
                <th className="text-left p-3">Used / Limit</th>
                <th className="text-left p-3">Expires</th>
                <th className="text-left p-3">Active</th>
                <th className="text-right p-3">Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((u) => (
                <tr key={u.id} className="border-t border-slate-800">
                  <td className="p-3">{u.email}</td>
                  <td className="p-3">
                    {u.used_gb.toFixed(2)} / {u.traffic_gb} GB
                  </td>
                  <td className="p-3">{new Date(u.expires_at).toLocaleDateString()}</td>
                  <td className="p-3">{u.active ? 'Yes' : 'No'}</td>
                  <td className="p-3 text-right space-x-2">
                    <button
                      onClick={() => setLinkUserId(u.id)}
                      className="text-sky-400 hover:underline"
                    >
                      Link
                    </button>
                    <button
                      onClick={() => setEditUser(u)}
                      className="text-slate-300 hover:underline"
                    >
                      Edit
                    </button>
                    <button
                      onClick={() => deleteUser(u.id)}
                      className="text-red-400 hover:underline"
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
              {users.length === 0 && (
                <tr>
                  <td colSpan={5} className="p-6 text-center text-slate-500">
                    No users yet
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}

      {showAdd && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-slate-900 border border-slate-700 rounded-lg p-6 w-full max-w-md mx-4">
            <h2 className="text-lg font-semibold mb-4">Add user</h2>
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
          </div>
        </div>
      )}

      {editUser && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="bg-slate-900 border border-slate-700 rounded-lg p-6 w-full max-w-md mx-4">
            <h2 className="text-lg font-semibold mb-4">Edit user</h2>
            <UserForm
              initial={{
                email: editUser.email,
                traffic_gb: editUser.traffic_gb,
                expire_days: 30,
                active: editUser.active,
              }}
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
          </div>
        </div>
      )}

      {linkUserId !== null && (
        <LinkModal userId={linkUserId} onClose={() => setLinkUserId(null)} />
      )}
    </div>
  );
}
