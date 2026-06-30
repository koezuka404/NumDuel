# NumDuel バックエンド実装解説

本ドキュメントは、NumDuel バックエンド（`backend/`）の**現時点で実装済みの機能**を、設計意図と動作の流れで説明する。

仕様書: `NumDuel.pdf`（実装仕様書）に沿った構成。フロントエンド・CI/E2E は未実装。

---

## 1. 全体像

### 1.1 アーキテクチャ

レイヤーは次の順で依存する（内側ほどビジネスロジックに近い）。

```
HTTP / WebSocket
    ↓
controller / websocket/handler
    ↓
usecase          ← ビジネスルール・トランザクション境界
    ↓
model            ← エンティティ・ドメインエラー・インターフェース
    ↓
repository / redis / crypto   ← インフラ実装
    ↓
PostgreSQL / Redis
```

**原則**

- Controller / WS Handler は Repository を直接触らない（UseCase 経過）
- DB 更新後に WebSocket 通知（`EventNotifier`）
- ゲーム判定（桁合わせ・勝敗）は `model` の純粋関数

### 1.2 起動順（`main.go`）

1. 環境変数読み込み（`config.Load`）
2. PostgreSQL 接続・マイグレーション（primary / backup）
3. Redis 接続（任意。未接続時はロック・ターン・JWT 失効などが無効）
4. DI 組み立て（Auth / Game / Admin / WS など）
5. **起動時 1 回**: `RecoverActiveGames`（進行中ゲームのターン復元）
6. **起動時 1 回**: `SeedMaster`（master が 0 件なら作成）
7. Echo HTTP サーバー + 7 種 Worker 起動
8. `SIGINT` / `SIGTERM` で graceful shutdown

### 1.3 ディレクトリ

| パス | 役割 |
|------|------|
| `main.go` | エントリ・DI・Worker 起動 |
| `config/` | 環境変数 |
| `model/` | エンティティ・エラーコード・外部サービス IF |
| `usecase/` | ユースケース（1 操作 = 1 ファイルが基本） |
| `controller/` | HTTP ハンドラ |
| `router/` | ルート・Middleware 適用 |
| `middleware/` | CORS / Auth / RateLimit 等 |
| `repository/` | GORM Repository・BackupSyncer |
| `redis/` | Redis キー操作 |
| `crypto/` | bcrypt / JWT / refresh / 秘密数字 HMAC |
| `websocket/` | Hub・WS Handler |
| `worker/` | バックグラウンド Worker |
| `db/` | DB 接続・マイグレーション |
| `dto/` | HTTP レスポンス形式 |

---

## 2. 認証・セッション

### 2.1 方式

- **JWT**: HttpOnly Cookie `access_token`（Bearer ヘッダは使わない）
- **Refresh**: HttpOnly Cookie `refresh_token`、DB には SHA-256 ハッシュのみ保存
- **検証順**: 署名 → `exp` → `jwt:revoked:{jti}` → `force_logout_before` → 削除済みユーザー

### 2.2 主要 UseCase

| UseCase | ファイル | 概要 |
|---------|----------|------|
| RegisterUser | `register_user.go` | ユーザー登録（bcrypt cost=12） |
| Login | `login.go` | JWT 発行 + refresh 作成 + login_log |
| RefreshToken | `refresh_token.go` | ローテーション。revoked 再使用で family 一括失効 |
| Logout | `logout.go` | jti 失効・refresh 全失効・WS 切断 |
| GetMe | `get_me.go` | 自分の基本情報 |

### 2.3 自動ログアウト

- `ActivityUpdateMiddleware` が protected API 呼び出しごとに `users.last_activity_at` を更新
- `AutoLogoutWorker` が `last_activity_at < now - SESSION_TIMEOUT_MINUTES` のユーザーを処理
  - Redis `user:{userId}:force_logout_before` を SET
  - WS 接続があれば `ERROR(unauthorized)` で切断
  - refresh トークン失効

---

## 3. HTTP API

### 3.1 公開

| Method | Path | 説明 |
|--------|------|------|
| GET | `/health` | DB 疎通確認 |
| POST | `/api/auth/register` | 登録 |
| POST | `/api/auth/login` | ログイン |
| POST | `/api/auth/refresh` | トークン更新 |

### 3.2 認証必須（JWT Cookie）

| Method | Path | 説明 |
|--------|------|------|
| POST | `/api/auth/logout` | ログアウト |
| GET | `/api/me` | 自分 |
| GET | `/api/me/profile` | プロフィール |
| GET | `/api/me/match-history` | 対戦履歴 |
| GET | `/api/me/login-history` | ログイン履歴 |
| GET | `/api/me/ws-history` | WS 接続履歴 |
| POST | `/api/matching/start` | マッチング開始 |
| POST | `/api/matching/cancel` | マッチング取消 |
| GET | `/api/matching/status` | マッチング状態 |
| GET | `/api/games/:id` | ゲーム状態 |
| GET | `/api/ranking` | ランキング上位 3 名 |

### 3.3 管理 API（role=master）

| Method | Path | 説明 |
|--------|------|------|
| GET | `/api/admin/users` | ユーザー一覧 |
| GET | `/api/admin/users/search` | ユーザー検索 |
| DELETE | `/api/admin/users/:id` | 論理削除 |
| POST | `/api/admin/ranking/rebuild` | ランキング手動再集計 |
| GET | `/api/admin/logs/types` | ログ種別一覧 |
| GET | `/api/admin/logs` | 操作ログ検索 |
| GET | `/api/admin/logs/download` | 操作ログ CSV |
| GET | `/api/admin/backup/status` | バックアップ状態 |

### 3.4 Middleware 適用順

```
Recover → CORS → RequestLog（全体）
/api:
  RateLimit → Auth → ActivityUpdate（protected）
  Admin（/api/admin/*）
```

---

## 4. WebSocket（`/ws`）

### 4.1 認証

- ハンドシェイク時の Cookie `access_token` から JWT を読む
- 接続後 5 秒以内に `{ "type": "AUTH" }` 必須
- 認証成功で Redis `ws:user:{userId}` に connectionId を SET（後勝ち切断）

### 4.2 イベント

| type | 方向 | 概要 |
|------|------|------|
| AUTH | C→S | 認証完了 |
| SET_SECRET | C→S | 秘密数字登録 |
| GUESS | C→S | 予想 |
| PING | C→S | 生存確認・activity 更新 |
| GAME_STATE_SYNC | S→C | 状態同期 |
| TURN_CHANGED | S→C | ターン交代 |
| GAME_OVER | S→C | 終了 |
| ERROR | S→C | エラー |

### 4.3 設計上の注意

- WS Handler は UseCase のみ呼ぶ（`ws_connection.go` で接続ログ記録）
- 秘密数字の平文・ハッシュは WS/HTTP レスポンスに含めない

---

## 5. ゲームフロー

### 5.1 マッチング

1. `StartMatching` → `matching_queue` に INSERT
2. 同 TX 内で `MatchPlayers`（先頭 2 人で `games` 作成、`WAITING_SECRET`）
3. 両者に WS `MATCHED`

### 5.2 対戦

```
WAITING_SECRET
  → 両者 SET_SECRET → IN_PROGRESS
  → GUESS ループ（ターン制）
  → guess_win で FINISHED
```

- 秘密数字: HMAC + pepper（`GameSecretPepper`）、DB 非保存
- 桁判定: `model.JudgeDigits` / `model.IsWin`
- 連打防止: Redis `game:{id}:player:{id}:secret_lock` / `guess_lock`

### 5.3 タイムアウト

| 条件 | 処理 | Worker |
|------|------|--------|
| 秘密数字未登録 | ゲーム終了（勝者なし） | SecretSetupTimeoutWorker |
| ターン期限切れ | 自動 GUESS（`isAuto=true`） | TurnTimeoutWorker |
| サーバー再起動 | ターン復元 or 秘密数字期限切れ処理 | RecoverActiveGames（起動時） |

ターン期限は Redis `game:{gameId}:turn`（JSON: turn, playerId, expiresAt）。

---

## 6. Redis キー一覧

| キー | 用途 | TTL |
|------|------|-----|
| `jwt:revoked:{jti}` | ログアウト済み JWT | JWT 残寿命 |
| `user:{userId}:force_logout_before` | 強制ログアウト境界 | 30 日 |
| `ws:user:{userId}` | WS 接続 ID | JWT 有効期限 |
| `game:{gameId}:turn` | ターン期限 | ターン長 |
| `game:{gameId}:player:{playerId}:guess_lock` | 予想連打防止 | GAME_LOCK_SECONDS |
| `game:{gameId}:player:{playerId}:secret_lock` | 秘密数字連打防止 | GAME_LOCK_SECONDS |
| `admin:{adminId}:ranking_rebuild_lock` | ランキング再集計 | ADMIN_LOCK_SECONDS |
| `admin:{adminId}:log_download_lock` | ログ CSV DL | ADMIN_LOCK_SECONDS |
| `admin:{adminId}:user_delete_lock` | ユーザー削除 | ADMIN_LOCK_SECONDS |
| `backup:status` | バックアップ同期状態 | なし |

ロック取得失敗 → HTTP/WS は `429 rate_limit_exceeded`（管理 API）または処理スキップ（Worker・HandleTimeout）。

---

## 7. バックグラウンド Worker（7 種）

仕様 §12.1 の Worker はすべて実装済み。

### 7.1 ゲーム系（ポーリング）

| Worker | 間隔 env | UseCase | 起動条件 |
|--------|----------|---------|----------|
| TurnTimeoutWorker | `TURN_TIMEOUT_POLL_SECONDS`（1s） | HandleTimeout | Redis 必須 |
| SecretSetupTimeoutWorker | `SECRET_TIMEOUT_POLL_SECONDS`（1s） | CancelGameBySecretTimeout | 常時 |
| AutoLogoutWorker | `AUTO_LOGOUT_POLL_SECONDS`（60s） | AutoLogout | Redis 必須 |

### 7.2 運用系（cron / UTC）

| Worker | cron env | デフォルト | UseCase |
|--------|----------|------------|---------|
| RankingRebuildWorker | `RANKING_REBUILD_CRON` | `*/10 * * * *` | RebuildRanking |
| LogRetentionWorker | `LOG_RETENTION_CRON` | `30 3 * * 0` | RunLogRetention |
| BackupWorker | `BACKUP_CRON` | `0 3 * * *` | RunBackupSync |
| RefreshTokenCleanupWorker | `REFRESH_TOKEN_CLEANUP_CRON` | `0 4 * * *` | RunRefreshTokenCleanup |

#### RankingRebuildWorker

- `users.win_count` から `rankings` を全件再構築（削除済み・master 除外）
- Worker 用ロック: `admin:00000000-0000-0000-0000-000000000000:ranking_rebuild_lock`
- 管理 API `POST /api/admin/ranking/rebuild` は別ロック（`admin:{adminId}:...`）+ activity_log 記録

#### LogRetentionWorker

- バッチ DELETE + バッチ間スリープ（長時間ロック回避）
- 対象: `activity_logs` / `login_logs` / `ws_connection_logs`
- テーブル単位の失敗は警告ログのみ（アプリ継続）

#### BackupWorker

- primary DB → backup DB へ差分 UPSERT（`updated_at > lastSyncedAt`）
- 成功/失敗を Redis `backup:status` に記録（`ok` / `error`）
- 失敗時最大 3 回リトライ
- 要: `BACKUP_DATABASE_URL` + Redis

#### RefreshTokenCleanupWorker

- 猶予日数（`REFRESH_TOKEN_CLEANUP_GRACE_DAYS`、既定 7）経過後に物理 DELETE
- 対象: 期限切れ active / 失効済み revoked

---

## 8. 管理機能

### 8.1 ユーザー削除（DeleteUser）

1. Redis ロック `admin:{adminId}:user_delete_lock`
2. 自己削除・master 削除・対戦中ユーザーは拒否
3. `force_logout_before` SET → WS 切断 → refresh 失効 → matching キュー削除
4. `users.deleted_at` 更新 + activity_log

### 8.2 ログ閲覧

- `activity_logs` を検索・CSV ダウンロード
- CSV は formula injection 対策（`sanitizeCSVCell`）
- ダウンロード時 Redis ロック `admin:{adminId}:log_download_lock`
- 記録例: `guess` / `game_over` / `admin_delete_user` / `admin_rebuild_ranking`

### 8.3 バックアップ状態

- `GET /api/admin/backup/status` → Redis `backup:status` を参照
- Redis 未接続時は `{ status: "ok" }` を返す（開発用）

---

## 9. データベース

### 9.1 主要テーブル

users, games, guesses, match_histories, rankings, matching_queue, activity_logs, login_logs, ws_connection_logs, refresh_tokens

### 9.2 バックアップ

- `repository/BackupSyncer`: 7 テーブルを差分 UPSERT
- 対象外: `ws_connection_logs`, `matching_queue`
- backup DB は `BACKUP_DATABASE_URL` 指定時のみ接続

### 9.3 ランキング（CQRS Read Model）

- 対戦 TX からは更新しない
- `RankingRebuildWorker` または管理 API から `RebuildRanking` で再集計

---

## 10. 監査・ログ

| 種別 | テーブル | 記録タイミング |
|------|----------|----------------|
| 操作ログ | activity_logs | 重要操作（RequestLogMiddleware + UseCase） |
| ログイン | login_logs | login / logout / auto_logout |
| WS 接続 | ws_connection_logs | 接続・切断 |

RequestLogMiddleware は `/api/admin/logs*` をスキップ（ログのログを避ける）。

---

## 11. 環境変数（主要）

`.env.example` を参照。カテゴリ別:

| カテゴリ | 例 |
|----------|-----|
| DB | `DATABASE_URL`, `BACKUP_DATABASE_URL` |
| Redis | `REDIS_ADDR` |
| 認証 | `JWT_SECRET`, `REFRESH_TOKEN_EXPIRY_DAYS` |
| ゲーム | `TURN_DURATION_SECONDS`, `GAME_SECRET_PEPPER`, `GAME_LOCK_SECONDS` |
| セッション | `SESSION_TIMEOUT_MINUTES`, `AUTO_LOGOUT_POLL_SECONDS` |
| Worker cron | `BACKUP_CRON`, `RANKING_REBUILD_CRON`, `LOG_RETENTION_CRON`, `REFRESH_TOKEN_CLEANUP_CRON` |
| ログ保持 | `ACTIVITY_LOG_RETENTION_DAYS`, `RETENTION_BATCH_SIZE` 等 |
| 管理 | `ADMIN_LOCK_SECONDS`, `NUMDUEL_MASTER_*` |

---

## 12. 未実装・今後の作業

| 項目 | 状態 |
|------|------|
| フロントエンド（React/Vite） | 未着手 |
| 単体・E2E テスト | 未着手 |
| CI（GitHub Actions） | 未着手 |
| `REDIS_URL` 形式（現状 `REDIS_ADDR`） | 仕様との差 |
| 本番で Redis 必須化 | 任意（現状は開発時スキップ可） |
| HTTP RateLimit のユーザー単位制限 | IP 制限のみ実装 |

---

## 13. ローカル起動

```bash
# Docker Compose（postgres × 2, redis, backend）
docker compose up

# または backend のみ
cd backend
cp ../.env.example ../.env   # 必要に応じて編集
go run .
```

管理画面 API を試す場合は `NUMDUEL_MASTER_EMAIL` / `NUMDUEL_MASTER_PASSWORD` で seed された master でログインする。
