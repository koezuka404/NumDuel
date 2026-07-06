Num Duel

ログイン済みユーザー向け 1 対 1 リアルタイム数字当て対戦 Web アプリケーション。

4 桁の秘密の数字（0〜9・重複不可）を交互に予想し、先に全桁一致したプレイヤーが勝利します。マッチング・対戦・ランキング・管理機能を備え、PostgreSQL / Redis を補助ストアとして Clean Architecture で構成しています。


技術スタック

区分採用フロントエンドReact 18 + TypeScript + Vite状態管理React Context + useReducer（対戦画面）ルーティングReact Router v6バックエンドGo 1.25 + Echo v4WebSocketgorilla/websocketORMGORMDB（本番）PostgreSQL 15DB（テスト）SQLiteキャッシュRedis 7（go-redis v9）認証JWT（HttpOnly Cookie access_token / refresh_token）コンテナDocker / Docker ComposeCI/CDGitHub Actions（.github/workflows/cicd.yml）フロント本番Vercelバックエンド本番Render負荷試験k6


リポジトリ構成

NumDuel/
├── backend/          # Go API + WebSocket + Worker
├── frontend/         # React SPA
├── k6/               # 負荷試験シナリオ
├── docker-compose.yml
├── .env.example
└── .github/workflows/cicd.yml

backend

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

frontend

frontend/src/
├── pages/            # Register, Login, Matching, Game, Ranking, Profile, Admin
├── hooks/            # useAuth, useWebSocket, useGameState 等
├── components/       # UI・ゲーム・管理タブ
├── lib/              # apiBase, validation, labels
├── api/client.ts     # fetch ラッパー（credentials: include）
└── router/           # AppRouter, guards


前提条件


Docker / Docker Compose
Go 1.25+
Node.js 22+
npm



ローカル開発

1. 環境変数

bashcp .env.example .env
# JWT_SECRET, GAME_SECRET_PEPPER 等を本番相当の長さで設定

2. インフラ + バックエンド（Docker Compose）

bashdocker compose up -d

サービスURL / ポートBackend APIhttp://localhost:8090Healthhttp://localhost:8090/healthWebSocketws://localhost:8090/wsPostgreSQL（primary）localhost:5434PostgreSQL（backup）localhost:5433Redislocalhost:6379

migrate サービスがスキーマ適用後、backend が起動します。

3. フロントエンド

bashcd frontend
npm ci
npm run dev

http://localhost:5173 で起動。Vite が /api と /ws を localhost:8090 にプロキシします。


管理者アカウントの使用方法

管理者（role=master）は 対戦・マッチングは行えず、管理画面（/admin）からユーザー・ログ・ランキング・バックアップを運用します。

アカウントの作成（初回のみ）

管理者は一般ユーザーの /register では作成できません。バックエンド起動時に、有効な master が 1 件も存在しない場合のみ 環境変数から自動 seed されます。

環境変数説明ローカル既定値（.env.example）NUMDUEL_MASTER_EMAIL管理者ログイン用メールadmin@local.testNUMDUEL_MASTER_PASSWORD管理者パスワードChangeMeOnFirstLogin!

seed される管理者の ユーザー名は常に admin です（メールアドレスは上記 env の値）。


注意


既に master が DB に存在する場合、env を変更しても再 seed されません。
本番デプロイ前に、必ず強力なパスワードへ変更してください。
master は複数 seed されません（ExistsActiveMaster が true のときスキップ）。




ログイン手順


docker compose up -d で backend を起動（初回 seed 実行）
npm run dev で frontend を起動
http://localhost:5173/login を開く
上記 メールアドレス と パスワード でログイン
ログイン成功後、自動的に 管理画面（/admin）へ遷移


一般ユーザー（role=user）は /matching へ、管理者は /admin へリダイレクトされます。

管理画面（/admin）の機能

タブ操作説明ユーザー一覧表示GET /api/admin/users — 登録ユーザーの一覧検索ユーザー名・メールの部分一致検索削除確認ダイアログ後に論理削除（DELETE /api/admin/users/:id）ログ検索ログ種別で activity / login 等を絞り込みCSV ダウンロード検索条件に応じたログをエクスポートランキング再集計win_count からランキングを再構築バックアップ状況確認最終同期日時・ステータス（ok / error）を表示

ヘッダーの ログアウト でセッションを終了し、ログイン画面へ戻ります。

管理者の制限

操作結果マッチング開始（/matching）不可（API は 403 forbidden）対戦（/game/:id）ルートガードにより /admin へリダイレクト自分自身の削除不可（cannot_delete_self）他の master の削除不可（cannot_delete_master）対戦中ユーザーの削除不可（user_in_active_game）

管理者は WebSocket 対戦接続も行いません（role=user のみ WS 接続）。

本番環境での設定

bash# 例: 本番 backend の環境変数
NUMDUEL_MASTER_EMAIL=admin@your-domain.example
NUMDUEL_MASTER_PASSWORD=<強力なパスワード>

初回 migrate 後、backend を 1 回だけ 起動して master を seed してください。seed 後は env からパスワードを削除しても既存アカウントは残りますが、パスワード変更 API はないため、初回 seed 前に正しい値を設定することが重要です。


主要 API

メソッドパス説明POST/api/auth/registerユーザー登録POST/api/auth/loginログイン（Set-Cookie）POST/api/auth/refreshトークン更新POST/api/auth/logoutログアウトGET/api/me自分の情報POST/api/matching/startマッチング開始GET/api/games/:idゲーム状態GET/api/rankingランキングGET/wsWebSocket（接続後 AUTH）GET/healthヘルスチェック

管理 API（role=master）: /api/admin/*


認証


HTTP: Cookie access_token（Path /）。Authorization ヘッダーは使用しません。
リフレッシュ: Cookie refresh_token（Path /api/auth/refresh）。
フロント: JWT 文字列は保持せず、GET /api/me + credentials: 'include' でセッション管理。
WebSocket: { "type": "AUTH" } を送信（token は Cookie からサーバーが取得）。

Cookie は HttpOnly 属性付きのため、curl で保存した cookie ファイルを他クライアント（Node の ws 等）で再利用する場合、行頭に付与される #HttpOnly_ プレフィックスの除去が必要です（CI の smoke test 参照）。





本番では COOKIE_SECURE=true、CORS_ALLOWED_ORIGINS / WS_ALLOWED_ORIGINS にフロントのオリジンを設定してください。


テスト

bash# バックエンド
cd backend
go vet ./...
go test ./... -count=1

# フロントエンド
cd frontend
npm test
npm run build


CI/CD

.github/workflows/cicd.yml が PR / push（main, master, develop）で実行されます。

ジョブ内容実行条件frontend-citest + build常時backend-civet + test + build常時integration-testPostgres + Redis + migrate + /health常時container-buildDocker イメージビルド常時security-scangosec, npm audit, Trivy, gitleaks常時publish-imagesGHCR pushpush（main/master/develop）production-migrate本番 DB マイグレーションpush（main/masterのみ）smoke-test本番 API smoke（register/login/me/ws）push（main/masterのみ）k6-load本番 /health 負荷試験push（main/masterのみ）deploy-productionVercel Production デプロイpush（main/masterのみ）deploy-previewVercel Preview デプロイpush（developのみ）

develop ブランチへの push は本番 DB マイグレーション・本番 API への smoke/k6 テストを実行せず、container-build と security-scan の完了後に Vercel の Preview 環境へデプロイされます。main / master への push のみ、本番マイグレーション → smoke test → k6 負荷試験 → 本番デプロイのフルパイプラインが実行されます。

本番デプロイに必要な Secrets

Secret用途登録先VERCEL_TOKENVercel CLIRepository（または production Environment）VERCEL_ORG_ID / VERCEL_PROJECT_IDVercel プロジェクトRepository（または production Environment）PRODUCTION_DATABASE_URL本番 DB マイグレーションRepository（または production Environment）PRODUCTION_API_BASE_URLsmoke / k6（末尾に /api を含める、例: https://xxx.example.com/api）Repository（または production Environment）


environment: production を指定しているジョブ（production-migrate / deploy-production）は、Secret が Environment secrets（Settings → Environments → production）に登録されていないと空文字として渡され、invalid argument エラー等の原因になります。environment: を指定していないジョブ（smoke-test / k6-load）からは Environment secrets は参照できないため、Repository secrets への登録を基本としてください。



k6 負荷試験のしきい値

k6/scenarios/health.js の thresholds は、実際のホスティング環境（Render 等の低〜中スペックプラン）での実測値に合わせて調整しています。極端に厳しい閾値（例: p95 500ms 未満、失敗率 5% 未満）は、コールドスタートや無料/低スペックプランのレイテンシで容易に超過するため、実測値にマージンを持たせた値を設定してください。

Vercel（フロント）


Root Directory: frontend
Build: npm run build / Output: dist
環境変数:

Production: VITE_API_BASE_URL=https://your-api.example.com/api, VITE_WS_BASE_URL=wss://your-api.example.com/ws
Preview（develop ブランチ用）: 開発用バックエンドを用意している場合はそのエンドポイントを、本番 API を共用する場合は Production と同値を設定してください。





バックエンド API / WS は Vercel 以外（Render, Railway 等）にデプロイする想定です。本番では APP_ENV=production と Redis が必須です。


ゲーム概要

項目内容数字0〜9、4 桁、重複不可判定各桁「位置も含め一致」のみ（○/×）。API は 0/1秘密数字登録SECRET_SETUP_SECONDS（既定 60 秒）1 ターンTURN_DURATION_SECONDS（既定 30 秒）ターン未入力crypto/rand で自動予想先攻マッチング待機キュー先着が player1


環境変数（主要）

.env.example を参照。特に重要な項目:

変数説明DATABASE_URLPostgreSQL 接続 URLREDIS_ADDR / REDIS_URLRedis（本番必須）JWT_SECRET32 文字以上GAME_SECRET_PEPPER32 バイト以上PORTAPI ポート（既定 8090）CORS_ALLOWED_ORIGINS例: http://localhost:5173


トラブルシューティング（CI/CD）

症状原因対処migrate: invalid argumentDATABASE_URL が空文字Secret 名・登録先（Repository / Environment）を確認curl: (3) URL rejected: No host part in the URLAPI_BASE_URL が空文字、または末尾スラッシュで ORIGIN 計算が崩れているSecret の値を確認（末尾スラッシュなし）register attemptが404API_BASE_URL に /api が含まれていないPRODUCTION_API_BASE_URL の末尾に /api を付与WS 認証が unauthorizedHttpOnly Cookie 行が #HttpOnly_ プレフィックス付きで、コメント行として除外されているCookie パース時に #HttpOnly_ を除去してから読み込むnpm audit で CI 失敗devDependencies（vite/vitest 等）の脆弱性を検出npm audit --omit=dev --audit-level=high に変更Trivy で Go 依存の CVE 検出依存ライブラリのバージョンが古いgo get <module>@<fixed-version> && go mod tidyk6 の thresholds crossed閾値が実インフラの性能に対して厳しすぎる実測値に基づき thresholds を緩和