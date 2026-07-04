import { Navigate } from 'react-router-dom';
import AppShell from '../components/layout/AppShell';
import { Spinner } from '../components/ui/FormField';
import { homePathForRole } from '../lib/routes';
import { useAuth } from '../hooks/useAuth';

function LoadingScreen() {
  return (
    <AppShell>
      <Spinner />
    </AppShell>
  );
}

type GuardProps = {
  children: JSX.Element;
};

export function RequireAuth({ children }: GuardProps) {
  const { isAuthenticated, isLoading } = useAuth();
  if (isLoading) {
    return <LoadingScreen />;
  }
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  return children;
}

export function RequireUser({ children }: GuardProps) {
  const { user, isLoading } = useAuth();
  if (isLoading) {
    return <LoadingScreen />;
  }
  if (user?.role === 'master') {
    return <Navigate to="/admin" replace />;
  }
  return children;
}

export function RequireMaster({ children }: GuardProps) {
  const { user, isLoading } = useAuth();
  if (isLoading) {
    return <LoadingScreen />;
  }
  if (user?.role !== 'master') {
    return <Navigate to="/matching" replace />;
  }
  return children;
}

export function GuestOnly({ children }: GuardProps) {
  const { isAuthenticated, isLoading, user } = useAuth();
  if (isLoading) {
    return <LoadingScreen />;
  }
  if (isAuthenticated && user) {
    return <Navigate to={homePathForRole(user.role)} replace />;
  }
  return children;
}
