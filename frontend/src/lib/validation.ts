export const USERNAME_PATTERN = /^[a-zA-Z0-9_]{3,50}$/;
const FOUR_DIGIT_PATTERN = /^[0-9]{4}$/;

export function validateLoginEmail(email: string): string | null {
  if (!email.includes('@') || email.length > 50) {
    return '有効なメールアドレスを入力してください';
  }
  return null;
}

export function validateRegisterEmail(email: string): string | null {
  if (!email.includes('@') || email.length > 255) {
    return '有効なメールアドレスを入力してください';
  }
  return null;
}

export function validatePassword(password: string): string | null {
  if (password.length < 8) {
    return 'パスワードは8文字以上必要です';
  }
  return null;
}

export function validateUsername(username: string): string | null {
  if (!USERNAME_PATTERN.test(username)) {
    return 'ユーザー名は3〜50文字の英数字とアンダースコアのみ使用できます';
  }
  return null;
}

export function validateFourDigits(value: string): string | null {
  if (!FOUR_DIGIT_PATTERN.test(value)) {
    return '4桁の数字を入力してください';
  }
  if (new Set(value).size !== 4) {
    return '重複しない4桁を入力してください';
  }
  return null;
}
