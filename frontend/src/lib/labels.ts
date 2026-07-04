import type { UserRole } from '../types/dto';

const LOG_TYPE_LABELS: Record<string, string> = {
  guess: '予想',
  game_over: 'ゲーム終了',
  timeout: 'タイムアウト',
  recover: '復旧',
  admin_delete_user: 'ユーザー削除',
  admin_rebuild_ranking: 'ランキング再集計',
  http_request: 'HTTPリクエスト',
};

const BACKUP_STATUS_LABELS: Record<string, string> = {
  ok: '正常',
  error: 'エラー',
};

const USER_ROLE_LABELS: Record<UserRole, string> = {
  user: '一般',
  master: '管理者',
};

export function logTypeLabel(logType: string): string {
  return LOG_TYPE_LABELS[logType] ?? logType;
}

export function backupStatusLabel(status: string): string {
  return BACKUP_STATUS_LABELS[status] ?? status;
}

export function userRoleLabel(role: string): string {
  if (role in USER_ROLE_LABELS) {
    return USER_ROLE_LABELS[role as UserRole];
  }
  return role;
}

export function apiErrorMessage(code: string, message: string): string {
  if (message && !/^[a-z_]+$/.test(message)) {
    return message;
  }
  switch (code) {
    case 'unauthorized':
      return '認証に失敗しました';
    case 'forbidden':
      return 'アクセスが禁止されています';
    case 'not_found':
      return '見つかりません';
    case 'validation_error':
      return message || '入力内容が不正です';
    case 'rate_limit_exceeded':
      return '操作が多すぎます。しばらく待ってください。';
    case 'internal_error':
      return 'サーバー内部エラーが発生しました';
    case 'duplicate_user':
      return 'ユーザー名またはメールアドレスが既に使用されています';
    case 'user_in_active_game':
      return '既に進行中のゲームがあります';
    case 'already_in_matching':
      return '既にマッチング待機中です';
    case 'game_not_started':
      return 'ゲームが開始されていません';
    case 'game_already_finished':
      return 'ゲームは既に終了しています';
    case 'game_already_started':
      return 'ゲームは既に開始されています';
    case 'not_your_turn':
      return 'あなたのターンではありません';
    case 'user_already_deleted':
      return 'ユーザーは既に削除されています';
    case 'cannot_delete_self':
      return '自分自身は削除できません';
    case 'cannot_delete_master':
      return '管理者ユーザーは削除できません';
    case 'token_expired':
      return 'アクセストークンの有効期限が切れています';
    default:
      return message || 'エラーが発生しました';
  }
}
