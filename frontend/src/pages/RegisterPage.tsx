import { FormEvent, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ApiError } from '../api/client';
import AppShell from '../components/layout/AppShell';
import FormField, { FormError } from '../components/ui/FormField';
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../hooks/useToast';
import { homePathForRole } from '../lib/routes';
import { validatePassword, validateRegisterEmail, validateUsername } from '../lib/validation';

export default function RegisterPage() {
  const navigate = useNavigate();
  const { register } = useAuth();
  const { showToast } = useToast();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [passwordConfirm, setPasswordConfirm] = useState('');
  const [formError, setFormError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const validateForm = (): string | null => {
    return (
      validateUsername(username) ??
      validateRegisterEmail(email) ??
      validatePassword(password) ??
      (password !== passwordConfirm ? 'パスワード確認が一致しません' : null)
    );
  };

  const onSubmit = async (event: FormEvent) => {
    event.preventDefault();
    const clientError = validateForm();
    if (clientError) {
      setFormError(clientError);
      return;
    }

    setSubmitting(true);
    setFormError('');
    try {
      const user = await register(username, email, password);
      showToast('登録完了', 'success');
      navigate(homePathForRole(user.role));
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'rate_limit_exceeded') {
          showToast(err.message, 'error');
        } else {
          setFormError(err.message);
        }
      } else {
        setFormError('登録に失敗しました');
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AppShell narrow>
      <h1>新規登録</h1>
      {formError && <FormError message={formError} />}
      <form onSubmit={onSubmit} className="form-card">
        <FormField label="ユーザー名" value={username} onChange={setUsername} required />
        <FormField label="メールアドレス" type="email" value={email} onChange={setEmail} required />
        <FormField label="パスワード" type="password" value={password} onChange={setPassword} required />
        <FormField
          label="パスワード確認"
          type="password"
          value={passwordConfirm}
          onChange={setPasswordConfirm}
          required
        />
        <button type="submit" className="btn-primary btn-block" disabled={submitting}>
          {submitting ? '処理中…' : '登録'}
        </button>
      </form>
      <p className="auth-footer">
        <Link to="/login">ログインはこちら</Link>
      </p>
    </AppShell>
  );
}
