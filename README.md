# Num Duel

ログイン済みユーザー向け **1 対 1 リアルタイム数字当て対戦** Web アプリケーション。

4 桁の秘密の数字（0〜9・重複不可）を交互に予想し、先に全桁一致したプレイヤーが勝利します。マッチング・対戦・ランキング・管理機能を備え、PostgreSQL を正本、Redis を補助ストアとして Clean Architecture で構成しています。

---

## 技術スタック

| 区分 | 採用 |
|------|------|
| フロントエンド | React 18 + TypeScript + Vite |
| 状態管理 | React Context + useReducer（対戦画面） |
| ルーティング | React Router v6 |
| バックエンド | Go 1.25 + Echo v4 |
| WebSocket | gorilla/websocket |
| ORM | GORM |
| DB（本番） | PostgreSQL 15 |
| DB（テスト） | SQLite |
| キャッシュ | Redis 7（go-redis v9） |
| 認証 | JWT（HttpOnly Cookie `access_token` / `refresh_token`） |
| コンテナ | Docker / Docker Compose |
| CI/CD | GitHub Actions（`.github/workflows/cicd.yml`） |
| フロント本番 | Vercel |

---

## リポジトリ構成

```
NumDuel/
├── backend/          # Go API + WebSocket + Worker
├── frontend/         # React SPA
├── k6/               # 負荷試験シナリオ
├── docker-compose.yml
├── .env.example
└── .github/workflows/cicd.yml
```

### backend

```
backend/
├── main.go
├── config/           # 環境変数
├── controller/       # HTTP ハンドラ
├── usecase/          # ビジネスロジック
├── model/            # エンティティ・ドメインルール
├── repository/       # PostgreSQL（GORM）
├── redis/            # Redis Store
├── websocket/        # WS Hub / Handler
├── worker/           # タイムアウト・バックアップ等
├── migrations/       # DB マイグレーション CLI
└── router/           # ルート定義
```

### frontend

```
frontend/src/
├── pages/            # Register, Login, Matching, Game, Ranking, Profile, Admin
├── hooks/            # useAuth, useWebSocket, useGameState 等
├── components/       # UI・ゲーム・管理タブ
├── lib/              # apiBase, validation, labels
├── api/client.ts     # fetch ラッパー（credentials: include）
└── router/           # AppRouter, guards
```

---

## 前提条件

- Docker / Docker Compose
- Go 1.25+
- Node.js 22+
- npm

---

## ローカル開発

### 1. 環境変数

```bash
cp .env.example .env
# JWT_SECRET, GAME_SECRET_PEPPER 等を本番相当の長さで設定
```

### 2. インフラ + バックエンド（Docker Compose）

```bash
docker compose up -d
```

| サービス | URL / ポート |
|----------|-------------|
| Backend API | http://localhost:8090 |
| Health | http://localhost:8090/health |
| WebSocket | ws://localhost:8090/ws |
| PostgreSQL（primary） | localhost:5434 |
| PostgreSQL（backup） | localhost:5433 |
| Redis | localhost:6379 |

`migrate` サービスがスキーマ適用後、`backend` が起動します。

### 3. フロントエンド

```bash
cd frontend
npm ci
npm run dev
```

http://localhost:5173 で起動。Vite が `/api` と `/ws` を `localhost:8090` にプロキシします。

---

## 管理者アカウントの使用方法

管理者（`role=master`）は **対戦・マッチングは行えず**、管理画面（`/admin`）からユーザー・ログ・ランキング・バックアップを運用します。

### アカウントの作成（初回のみ）

管理者は一般ユーザーの `/register` では作成できません。バックエンド起動時に、**有効な master が 1 件も存在しない場合のみ** 環境変数から自動 seed されます。

| 環境変数 | 説明 | ローカル既定値（`.env.example`） |
|----------|------|----------------------------------|
| `NUMDUEL_MASTER_EMAIL` | 管理者ログイン用メール | `admin@local.test` |
| `NUMDUEL_MASTER_PASSWORD` | 管理者パスワード | `ChangeMeOnFirstLogin!` |

seed される管理者の **ユーザー名は常に `admin`** です（メールアドレスは上記 env の値）。

> **注意**
> - 既に master が DB に存在する場合、env を変更しても再 seed されません。
> - 本番デプロイ前に、必ず強力なパスワードへ変更してください。
> - master は複数 seed されません（`ExistsActiveMaster` が true のときスキップ）。

### ログイン手順

1. `docker compose up -d` で backend を起動（初回 seed 実行）
2. `npm run dev` で frontend を起動
3. http://localhost:5173/login を開く
4. 上記 **メールアドレス** と **パスワード** でログイン
5. ログイン成功後、自動的に **管理画面**（`/admin`）へ遷移

一般ユーザー（`role=user`）は `/matching` へ、管理者は `/admin` へリダイレクトされます。

### 管理画面（`/admin`）の機能

| タブ | 操作 | 説明 |
|------|------|------|
| **ユーザー** | 一覧表示 | `GET /api/admin/users` — 登録ユーザーの一覧 |
| | 検索 | ユーザー名・メールの部分一致検索 |
| | 削除 | 確認ダイアログ後に論理削除（`DELETE /api/admin/users/:id`） |
| **ログ** | 検索 | ログ種別で activity / login 等を絞り込み |
| | CSV ダウンロード | 検索条件に応じたログをエクスポート |
| **ランキング** | 再集計 | `win_count` からランキングを再構築 |
| **バックアップ** | 状況確認 | 最終同期日時・ステータス（`ok` / `error`）を表示 |

ヘッダーの **ログアウト** でセッションを終了し、ログイン画面へ戻ります。

### 管理者の制限

| 操作 | 結果 |
|------|------|
| マッチング開始（`/matching`） | **不可**（API は `403 forbidden`） |
| 対戦（`/game/:id`） | ルートガードにより `/admin` へリダイレクト |
| 自分自身の削除 | **不可**（`cannot_delete_self`） |
| 他の master の削除 | **不可**（`cannot_delete_master`） |
| 対戦中ユーザーの削除 | **不可**（`user_in_active_game`） |

管理者は WebSocket 対戦接続も行いません（`role=user` のみ WS 接続）。

### 本番環境での設定

```bash
# 例: 本番 backend の環境変数
NUMDUEL_MASTER_EMAIL=admin@your-domain.example
NUMDUEL_MASTER_PASSWORD=<強力なパスワード>
```

初回 migrate 後、backend を **1 回だけ** 起動して master を seed してください。seed 後は env からパスワードを削除しても既存アカウントは残りますが、パスワード変更 API はないため、**初回 seed 前に正しい値を設定する**ことが重要です。

---

## 主要 API

| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/api/auth/register` | ユーザー登録 |
| POST | `/api/auth/login` | ログイン（Set-Cookie） |
| POST | `/api/auth/refresh` | トークン更新 |
| POST | `/api/auth/logout` | ログアウト |
| GET | `/api/me` | 自分の情報 |
| POST | `/api/matching/start` | マッチング開始 |
| GET | `/api/games/:id` | ゲーム状態 |
| GET | `/api/ranking` | ランキング |
| GET | `/ws` | WebSocket（接続後 `AUTH`） |
| GET | `/health` | ヘルスチェック |

管理 API（`role=master`）: `/api/admin/*`

---

## 認証

- **HTTP**: Cookie `access_token`（Path `/`）。`Authorization` ヘッダーは使用しません。
- **リフレッシュ**: Cookie `refresh_token`（Path `/api/auth/refresh`）。
- **フロント**: JWT 文字列は保持せず、`GET /api/me` + `credentials: 'include'` でセッション管理。
- **WebSocket**: `{ "type": "AUTH" }` を送信（token は Cookie からサーバーが取得）。

本番では `COOKIE_SECURE=true`、`CORS_ALLOWED_ORIGINS` / `WS_ALLOWED_ORIGINS` にフロントのオリジンを設定してください。

---

## テスト

```bash
# バックエンド
cd backend
go vet ./...
go test ./... -count=1

# フロントエンド
cd frontend
npm test
npm run build
```

---

## CI/CD

`.github/workflows/cicd.yml` が PR / push（`main`, `master`, `develop`）で実行されます。

| ジョブ | 内容 |
|--------|------|
| frontend-ci | test + build |
| backend-ci | vet + test + build |
| integration-test | Postgres + Redis + migrate + `/health` |
| container-build | Docker イメージビルド |
| security-scan | gosec, npm audit, Trivy, gitleaks |
| publish-images | GHCR push（本番ブランチのみ） |
| production-migrate / smoke-test / k6-load / deploy-production | 本番ブランチ push 時 |

### 本番デプロイに必要な Secrets

| Secret | 用途 |
|--------|------|
| `VERCEL_TOKEN` | Vercel CLI |
| `VERCEL_ORG_ID` / `VERCEL_PROJECT_ID` | Vercel プロジェクト |
| `PRODUCTION_DATABASE_URL` | 本番 DB マイグレーション |
| `PRODUCTION_API_BASE_URL` | smoke / k6 |

### Vercel（フロント）

- Root Directory: `frontend`
- Build: `npm run build` / Output: `dist`
- 環境変数（本番）:
  - `VITE_API_BASE_URL=https://your-api.example.com/api`
  - `VITE_WS_BASE_URL=wss://your-api.example.com/ws`

バックエンド API / WS は Vercel 以外（Render, Railway 等）にデプロイする想定です。本番では `APP_ENV=production` と Redis が必須です。

---

## ゲーム概要

| 項目 | 内容 |
|------|------|
| 数字 | 0〜9、4 桁、重複不可 |
| 判定 | 各桁「位置も含め一致」のみ（○/×）。API は 0/1 |
| 秘密数字登録 | `SECRET_SETUP_SECONDS`（既定 60 秒） |
| 1 ターン | `TURN_DURATION_SECONDS`（既定 30 秒） |
| ターン未入力 | crypto/rand で自動予想 |
| 先攻 | マッチング待機キュー先着が player1 |

---

## 環境変数（主要）

`.env.example` を参照。特に重要な項目:

| 変数 | 説明 |
|------|------|
| `DATABASE_URL` | PostgreSQL 接続 URL |
| `REDIS_ADDR` / `REDIS_URL` | Redis（本番必須） |
| `JWT_SECRET` | 32 文字以上 |
| `GAME_SECRET_PEPPER` | 32 バイト以上 |
| `PORT` | API ポート（既定 8090） |
| `CORS_ALLOWED_ORIGINS` | 例: `http://localhost:5173` |

---

## ドキュメント

詳細な設計・API・WebSocket イベント・テスト要件は **Num Duel 実装仕様書** を参照してください。

---

## ライセンス

未設定（必要に応じて追記）。
