import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import AppRouter from '../router/AppRouter';

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    isLoading: false,
    login: vi.fn(),
    logout: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

describe('AppRouter', () => {
  it('redirects unauthenticated users to login for protected routes', () => {
    render(
      <MemoryRouter initialEntries={['/matching']}>
        <AppRouter />
      </MemoryRouter>,
    );
    expect(screen.getByRole('heading', { name: 'ログイン' })).toBeInTheDocument();
  });
});
