import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';
import { ToastProvider } from '../hooks/useToast';
import RegisterPage from './RegisterPage';

describe('RegisterPage', () => {
  it('shows validation error for short username', async () => {
    render(
      <MemoryRouter>
        <ToastProvider>
          <RegisterPage />
        </ToastProvider>
      </MemoryRouter>,
    );

    fireEvent.change(screen.getByLabelText('ユーザー名'), { target: { value: 'ab' } });
    fireEvent.change(screen.getByLabelText('メールアドレス'), { target: { value: 'a@test.local' } });
    fireEvent.change(screen.getAllByLabelText('パスワード')[0], { target: { value: 'password123' } });
    fireEvent.change(screen.getByLabelText('パスワード確認'), { target: { value: 'password123' } });
    fireEvent.click(screen.getByRole('button', { name: '登録' }));

    await waitFor(() => {
      expect(screen.getByText(/3〜50文字/)).toBeInTheDocument();
    });
  });
});
