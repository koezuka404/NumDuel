import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError } from '../api/client';
import { ToastProvider } from '../hooks/useToast';
import RegisterPage from './RegisterPage';

const navigate = vi.fn();
const registerMock = vi.fn();

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom');
  return { ...actual, useNavigate: () => navigate };
});

vi.mock('../hooks/useAuth', async () => {
  const actual = await vi.importActual<typeof import('../hooks/useAuth')>('../hooks/useAuth');
  return {
    ...actual,
    useAuth: () => ({
      user: null,
      isAuthenticated: false,
      isLoading: false,
      login: vi.fn(),
      register: registerMock,
      logout: vi.fn(),
      refreshUser: vi.fn(),
    }),
  };
});

function renderPage() {
  return render(
    <MemoryRouter>
      <ToastProvider>
        <RegisterPage />
      </ToastProvider>
    </MemoryRouter>,
  );
}

function fillValidForm() {
  fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'alice' } });
  fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'a@test.local' } });
  fireEvent.change(screen.getAllByLabelText('パスワード')[0], { target: { value: 'password123' } });
  fireEvent.change(screen.getByLabelText('パスワード確認'), { target: { value: 'password123' } });
}

describe('RegisterPage', () => {
  afterEach(() => {
    cleanup();
    navigate.mockReset();
    registerMock.mockReset();
  });

  it('shows validation error for short username', async () => {
    renderPage();
    fillValidForm();
    fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'ab' } });
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(screen.getByText(/3〜50文字/)).toBeInTheDocument());
  });

  it('registers and navigates to matching', async () => {
    registerMock.mockResolvedValueOnce({ id: '1', username: 'alice', role: 'user' });
    renderPage();
    fillValidForm();
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(navigate).toHaveBeenCalledWith('/matching'));
    expect(registerMock).toHaveBeenCalledWith('alice', 'a@test.local', 'password123');
  });

  it('shows api error', async () => {
    registerMock.mockRejectedValueOnce(new ApiError('validation_error', 'duplicate', 400));
    renderPage();
    fillValidForm();
    fireEvent.click(screen.getByRole('button', { name: '登録' }));
    await waitFor(() => expect(screen.getByText('duplicate')).toBeInTheDocument());
  });
});
