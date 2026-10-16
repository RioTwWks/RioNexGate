import { useParams } from 'react-router-dom';
import { UserEdit } from './UserEdit';

export function UserEditPage() {
  const { id } = useParams<{ id: string }>();
  const userId = Number(id);

  if (!userId || Number.isNaN(userId)) {
    return <p className="text-red-400">Invalid user ID</p>;
  }

  return <UserEdit userId={userId} />;
}
