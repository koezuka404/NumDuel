import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from './useAuth';

export function useLogout() {
  const navigate = useNavigate();
  const { logout } = useAuth();

  return useCallback(async () => {
    await logout();
    navigate('/login');
  }, [logout, navigate]);
}
