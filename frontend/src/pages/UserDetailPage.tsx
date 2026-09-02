import { useParams } from 'react-router-dom';
import { UserDetail } from './UserDetail';

export function UserDetailPage() {
  const { id } = useParams<{ id: string }>();
  const userId = Number(id);

  if (!userId || Number.isNaN(userId)) {
    return <p className="text-red-400">Invalid user ID</p>;
  }

  return <UserDetail userId={userId} />;
}
