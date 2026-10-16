import { useCallback, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import api from '../services/api';
import { UserChainSection } from '../components/UserChainSection';
import { UserForm } from '../components/UserForm';
import { PageHeader, Spinner } from '../components/ui';
import type { User } from '../types/user';

interface Props {
  userId: number;
}

export function UserEdit({ userId }: Props) {
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState('');

  const load = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      const res = await api.get<User>(`/users/${userId}`);
      setUser(res.data);
    } catch {
      setError('Failed to load user');
      setUser(null);
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) {
    return (
      <div className="page">
        <Spinner />
      </div>
    );
  }

  if (!user) {
    return (
      <div className="page">
        <PageHeader
          title="Edit user"
          actions={
            <Link to="/users" className="btn-secondary">
              Back to users
            </Link>
          }
        />
        <p className="alert-error">{error || 'User not found'}</p>
      </div>
    );
  }

  return (
    <div className="page">
      <PageHeader
        title="Edit user"
        subtitle={user.email}
        actions={
          <>
            <Link to={`/users/${user.id}`} className="btn-secondary">
              Details
            </Link>
            <Link to="/users" className="btn-secondary">
              Back to users
            </Link>
          </>
        }
      />

      {error && <p className="alert-error">{error}</p>}
      {saved && <p className="alert-ok">{saved}</p>}

      <div className="grid gap-6 lg:grid-cols-[minmax(0,24rem)_minmax(0,1fr)] items-start">
        <section className="card-pad space-y-4">
          <div>
            <h2 className="text-lg font-medium">Account</h2>
            <p className="text-xs text-slate-500 mt-1">Email, traffic quota, and active flag</p>
          </div>
          <UserForm
            initial={{
              email: user.email,
              traffic_gb: user.traffic_gb,
              expire_days: 30,
              active: user.active,
            }}
            showExpireDays={false}
            onCancel={() => navigate('/users')}
            onSubmit={async (data) => {
              setError('');
              setSaved('');
              try {
                const res = await api.put<User>(`/users/${user.id}`, {
                  email: data.email,
                  traffic_gb: data.traffic_gb,
                  active: data.active,
                });
                setUser(res.data);
                setSaved('Account saved');
              } catch {
                setError('Failed to save user');
                throw new Error('Failed to save user');
              }
            }}
          />
        </section>

        <UserChainSection
          user={user}
          onUpdated={(u) => {
            setUser(u);
            setSaved('');
          }}
        />
      </div>
    </div>
  );
}
