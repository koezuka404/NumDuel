import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
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

  it('shows client validation errors', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );
    fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));
    await waitFor(() => {
      expect(screen.getByText('有効なメールアドレスを入力してください')).toBeInTheDocument();
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

  it('shows api and generic errors', async () => {
    login.mockRejectedValueOnce(new ApiError('internal_error', 'server down', 500));
    const { unmount } = render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'user@test.local' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));
    await waitFor(() => expect(screen.getByText('server down')).toBeInTheDocument());
    unmount();

    login.mockRejectedValueOnce(new Error('network'));
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    );
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'user@test.local' } });
    fireEvent.change(screen.getByLabelText('パスワード'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: 'ログイン' }));
    await waitFor(() => expect(screen.getByText('ログインに失敗しました')).toBeInTheDocument());
  });
});
