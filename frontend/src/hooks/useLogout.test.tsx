import { act, renderHook } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { useLogout } from './useLogout';

const logout = vi.fn();
const navigate = vi.fn();

vi.mock('./useAuth', () => ({
  useAuth: () => ({ logout }),
}));

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

describe('useLogout', () => {
  it('logs out and navigates to login', async () => {
    logout.mockResolvedValueOnce(undefined);
    const { result } = renderHook(() => useLogout());
    await act(async () => {
      await result.current();
    });
    expect(logout).toHaveBeenCalled();
    expect(navigate).toHaveBeenCalledWith('/login');
  });
});
