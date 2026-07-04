import { describe, expect, it } from 'vitest';
import {
  validateFourDigits,
  validateLoginEmail,
  validatePassword,
  validateRegisterEmail,
  validateUsername,
} from './validation';

describe('validateLoginEmail', () => {
  it('accepts valid email', () => {
    expect(validateLoginEmail('user@test.local')).toBeNull();
  });

  it('rejects email without @', () => {
    expect(validateLoginEmail('invalid')).toBe('有効なメールアドレスを入力してください');
  });

  it('rejects email longer than 50 chars', () => {
    expect(validateLoginEmail(`${'a'.repeat(42)}@test.local`)).toBe(
      '有効なメールアドレスを入力してください',
    );
  });
});

describe('validateRegisterEmail', () => {
  it('accepts email up to 255 chars', () => {
    expect(validateRegisterEmail(`${'a'.repeat(240)}@test.local`)).toBeNull();
  });

  it('rejects email longer than 255 chars', () => {
    expect(validateRegisterEmail(`${'a'.repeat(250)}@test.local`)).toBe(
      '有効なメールアドレスを入力してください',
    );
  });
});

describe('validatePassword', () => {
  it('accepts 8+ chars', () => {
    expect(validatePassword('password')).toBeNull();
  });

  it('rejects short password', () => {
    expect(validatePassword('short')).toBe('パスワードは8文字以上必要です');
  });
});

describe('validateUsername', () => {
  it('accepts alphanumeric username', () => {
    expect(validateUsername('user_01')).toBeNull();
  });

  it('rejects too short username', () => {
    expect(validateUsername('ab')).toContain('3〜50文字');
  });

  it('rejects invalid characters', () => {
    expect(validateUsername('bad-name')).toContain('3〜50文字');
  });
});

describe('validateFourDigits', () => {
  it('accepts unique 4 digits', () => {
    expect(validateFourDigits('1234')).toBeNull();
  });

  it('rejects non-numeric input', () => {
    expect(validateFourDigits('12ab')).toBe('4桁の数字を入力してください');
  });

  it('rejects duplicate digits', () => {
    expect(validateFourDigits('1123')).toBe('重複しない4桁を入力してください');
  });
});
