# AtCoder Shojin

`fken_prime_57` 向けに、未ACの AtCoder Beginner Contest D/E/F 問題を75分で解くトレーニングアプリです。

## v1 の動作

- D1/E3/F1の5問をStart時に初めて公開
- 終了済みABCのうち最新10回、7/8問構成以外、過去AC済みを除外
- 同一セッション内では全問題を異なるABCから選択
- Difficulty範囲を `指定 → ±100 → ±200 → indexのみ` の順で緩和
- 15秒間隔の提出同期、75分後の自動終了、Abort、再読込復帰
- 結果と履歴をMariaDBへ永続化

AtCoderの問題・Difficulty・提出情報には、非公式の [AtCoder Problems API](https://github.com/kenkoooo/AtCoderProblems/blob/master/doc/api.md) を利用します。

## 構成

```text
backend/             Go / Echo API、domain/usecase/infrastructure、MariaDB migration
react-sample-app/    React / TypeScript SPA
docs/openapi.yaml    API契約
Dockerfile           フロントとバックエンドの統合イメージ
```

## ローカル起動方法

### 前提環境

- Node.js 22
- Go 1.26
- Docker Desktop（MariaDB 11 の起動に使用）

以下は、すべてプロジェクトルート（このREADMEがあるディレクトリ）を起点に実行します。3つのターミナルを使用してください。

### 1. MariaDBを起動する（ターミナル1）

```powershell
cd backend
docker compose up -d
```

### 2. バックエンドを起動する（ターミナル2）

```powershell
cd backend
$env:DATABASE_URL='mariadb://shojin:replace-me@localhost:3306/atcoder_shojin'
go run .C:\Users\kenta\maimai_puzzle\SPECIFICATION_V2.md
```

起動後、`http://localhost:8080/healthz` でヘルスチェックを確認できます。APIのベースURLは `http://localhost:8080/apis` です。

### 3. フロントエンドを起動する（ターミナル3）

```powershell
cd react-sample-app
npm ci
$env:REACT_APP_API_BASE_URL='http://localhost:8080/apis'
npm start
```

ブラウザで `http://localhost:3000` を開いてください。`npm ci` は初回起動時、または依存関係が変わったときに実行すれば十分です。

環境変数の設定例は `backend/.env.example` と `react-sample-app/.env.example` にあります。

### 停止方法

バックエンドとフロントエンドは、それぞれのターミナルで `Ctrl+C` を押して停止します。MariaDBは次のコマンドで停止します。

```powershell
cd backend
docker compose down
```

## テスト

```powershell
cd backend
go test . ./internal/...

cd ../react-sample-app
npm test -- --watchAll=false --runInBand
npm run build
```

## 重要な実装上の保証

- セッションと5問は1トランザクションで保存されます。
- 生成列のUnique Indexにより、固定ユーザーのACTIVEセッションはDB上も最大1件です。
- `(session_id, contest_id)` のUnique制約により、同一ABC重複をDBでも拒否します。
- ACの有効区間は `started_at <= accepted_at < deadline` です。
- Abort後は元の75分期限まで新規Startを拒否し、問題ガチャを防止します。
- 外部API同期に失敗しても、作成済みセッションとタイマーは継続します。

## 外部API上の制約

AtCoder Problemsは非公式APIで、提出APIは一度に最大500件です。初回Startでは過去提出を差分キャッシュへ同期するため、提出数に応じて開始まで時間がかかる場合があります。またAtCoder Problems側の反映遅延により、AC表示が遅れる場合があります。
