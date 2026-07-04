package model

import "errors"

var (
	ErrBadUsername       = errors.New("ユーザー名は3〜50文字の半角英字・数字とアンダースコアのみ使用できます")
	ErrBadEmail          = errors.New("メールアドレスは1〜255文字で入力してください")
	ErrWeakPassword      = errors.New("パスワードは8文字以上必要です")
	ErrBadLoginEmail     = errors.New("有効なメールアドレスを入力してください")
	ErrBadRole           = errors.New("無効なロールです")
	ErrBadGameStatus     = errors.New("無効なゲーム状態です")
	ErrBadDigit          = errors.New("数字のみ入力できます")
	ErrBadDigitLength    = errors.New("4桁の数字を入力してください")
	ErrDuplicateDigit    = errors.New("重複しない4桁を入力してください")
	ErrBadMatchingStatus = errors.New("無効なマッチングキュー状態です")
	ErrBadLoginAction    = errors.New("無効なログイン操作です")
	ErrBadRefreshToken   = errors.New("無効なリフレッシュトークン状態です")

	ErrUnauthorized        = errors.New("認証に失敗しました")
	ErrForbidden           = errors.New("アクセスが禁止されています")
	ErrNotFound            = errors.New("見つかりません")
	ErrNotYourTurn         = errors.New("あなたのターンではありません")
	ErrGameNotStarted      = errors.New("ゲームが開始されていません")
	ErrGameAlreadyFinished = errors.New("ゲームは既に終了しています")
	ErrGameAlreadyStarted  = errors.New("ゲームは既に開始されています")
	ErrDuplicateUser       = errors.New("ユーザー名またはメールアドレスが既に使用されています")
	ErrUserInActiveGame    = errors.New("既に進行中のゲームがあります")
	ErrAlreadyInMatching   = errors.New("既にマッチング待機中です")
	ErrUserAlreadyDeleted  = errors.New("ユーザーは既に削除されています")
	ErrCannotDeleteSelf    = errors.New("自分自身は削除できません")
	ErrCannotDeleteMaster  = errors.New("管理者ユーザーは削除できません")
	ErrRateLimitExceeded   = errors.New("操作が多すぎます。しばらく待ってください。")
	ErrTokenExpired        = errors.New("アクセストークンの有効期限が切れています")

	ErrInternal = errors.New("サーバー内部エラーが発生しました")
)
