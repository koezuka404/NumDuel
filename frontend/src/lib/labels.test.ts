import { describe, expect, it } from 'vitest';
import { apiErrorMessage, backupStatusLabel, logTypeLabel, userRoleLabel } from './labels';

describe('logTypeLabel', () => {
  it('returns Japanese label for known types', () => {
    expect(logTypeLabel('guess')).toBe('予想');
  });

  it('returns original for unknown types', () => {
    expect(logTypeLabel('custom')).toBe('custom');
  });
});

describe('backupStatusLabel', () => {
  it('returns Japanese label for known statuses', () => {
    expect(backupStatusLabel('ok')).toBe('正常');
  });

  it('returns original for unknown statuses', () => {
    expect(backupStatusLabel('pending')).toBe('pending');
  });
});

describe('userRoleLabel', () => {
  it('returns Japanese labels for roles', () => {
    expect(userRoleLabel('user')).toBe('一般');
    expect(userRoleLabel('master')).toBe('管理者');
  });

  it('returns original for unknown roles', () => {
    expect(userRoleLabel('guest')).toBe('guest');
  });
});

describe('apiErrorMessage', () => {
  it('returns human-readable server message when provided', () => {
    expect(apiErrorMessage('validation_error', 'カスタムエラー')).toBe('カスタムエラー');
  });

  it.each([
    ['unauthorized', '認証に失敗しました'],
    ['forbidden', 'アクセスが禁止されています'],
    ['not_found', '見つかりません'],
    ['rate_limit_exceeded', '操作が多すぎます。しばらく待ってください。'],
    ['internal_error', 'サーバー内部エラーが発生しました'],
    ['duplicate_user', 'ユーザー名またはメールアドレスが既に使用されています'],
    ['user_in_active_game', '既に進行中のゲームがあります'],
    ['already_in_matching', '既にマッチング待機中です'],
    ['game_not_started', 'ゲームが開始されていません'],
    ['game_already_finished', 'ゲームは既に終了しています'],
    ['game_already_started', 'ゲームは既に開始されています'],
    ['not_your_turn', 'あなたのターンではありません'],
    ['user_already_deleted', 'ユーザーは既に削除されています'],
    ['cannot_delete_self', '自分自身は削除できません'],
    ['cannot_delete_master', '管理者ユーザーは削除できません'],
    ['token_expired', 'アクセストークンの有効期限が切れています'],
  ] as const)('maps %s', (code, expected) => {
    expect(apiErrorMessage(code, '')).toBe(expected);
  });

  it('uses validation fallback when message is empty', () => {
    expect(apiErrorMessage('validation_error', '')).toBe('入力内容が不正です');
  });

  it('uses default fallback for unknown codes', () => {
    expect(apiErrorMessage('unknown_code', '')).toBe('エラーが発生しました');
    expect(apiErrorMessage('unknown_code', '詳細')).toBe('詳細');
  });
});
