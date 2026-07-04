import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import LoginPage from './LoginPage';
import { ApiError } from '../api/client';

const login = vi.fn();
const navigate = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

vi.mock('../hooks/useAuth', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    isLoading: false,
    login,
    logout: vi.fn(),
    refreshUser: vi.fn(),
  }),
}));

describe('LoginPage', () => {
  afterEach(() => {
    cleanup();
    login.mockReset();
    navigate.mockReset();
  });

  it('shows unauthorized message', async () => {
    login.mockRejectedValueOnce(new ApiError('unauthorized', 'bad credentials', 401));
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'user@test.local' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));
    await waitFor(() => {
      expect(screen.getByText('メールまたはパスワードが正しくありません')).toBeInTheDocument();
    });
  });

  it('navigates after successful login', async () => {
    login.mockResolvedValueOnce({ id: '1', username: 'alice', role: 'user' });
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'user@test.local' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/matching'));
  });
});
