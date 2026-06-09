import { Navigate } from 'react-router-dom';
import { getApiKey } from '../services/api';

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  if (!getApiKey()) {
    return <Navigate to="/login" replace />;
  }
  return <>{children}</>;
}
