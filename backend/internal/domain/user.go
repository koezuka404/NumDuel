package domain

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// User はユーザーアカウント Entity（users テーブル）。
//
// 永続化は Repository 経由。パスワードは bcrypt ハッシュのみ保持（平文禁止）。
// 削除は deleted_at による論理削除。物理削除は行わない。
type User struct {
	ID             uuid.UUID  // PK
	Username       string     // UNIQUE, 3〜50 文字, ^[a-zA-Z0-9_]+$
	Email          string     // UNIQUE, RFC5322, 255 文字以下
	PasswordHash   string     // bcrypt cost=12（Infrastructure で生成）
	Role           Role       // user / master
	WinCount       int        // 累計勝利数。FinishGameService で +1
	DeletedAt      *time.Time // NULL = 有効ユーザー
	DeletedBy      *uuid.UUID // 削除した master の ID
	LastActivityAt time.Time  // 無操作自動ログアウト判定用（SESSION_TIMEOUT_MINUTES）
	CreatedAt      time.Time
	UpdatedAt      time.Time  // バックアップ差分同期用
}

// IsDeleted は論理削除済みか。Login / 対戦 / マッチング前に UseCase が確認する。
func (u *User) IsDeleted() bool {
	return u != nil && u.DeletedAt != nil
}

// IsMaster は master 権限か。AdminMiddleware および CanMatch 判定に使用。
func (u *User) IsMaster() bool {
	return u != nil && u.Role == RoleMaster
}

// CanMatch はマッチングキューに入れるか。
//
// 条件: 削除済みでない AND master でない。
// 対戦中チェック（user_in_active_game）は UseCase 側で GameRepository を参照する。
func (u *User) CanMatch() bool {
	return u != nil && !u.IsDeleted() && !u.IsMaster()
}

// ValidateUsername は RegisterUserUseCase の入力検証。
func ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 || !usernamePattern.MatchString(username) {
		return errValidation("username must be 3-50 alphanumeric/underscore characters")
	}
	return nil
}

// ValidateEmail はメールの長さ検証。RFC5322 形式の詳細検証は API 層で行う。
func ValidateEmail(email string) error {
	if len(email) == 0 || len(email) > 255 {
		return errValidation("email must be 1-255 characters")
	}
	return nil
}

// ValidatePassword は RegisterUserUseCase / LoginUseCase のパスワード検証（8 文字以上）。
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errValidation("password must be at least 8 characters")
	}
	return nil
}

// ValidateLoginEmail は LoginUseCase のメール検証（50 文字以下、簡易形式チェック）。
func ValidateLoginEmail(email string) error {
	if len(email) == 0 || len(email) > 50 {
		return errValidation("email must be 1-50 characters")
	}
	if !strings.Contains(email, "@") || strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return errValidation("email format is invalid")
	}
	return nil
}
