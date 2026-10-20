# ディレクトリ構成

新しいファイルを追加する際は、どこに置くかをここで確かめる。**構成を変えたら、この文書も同じ変更の中で直す。**

```
.
├── CLAUDE.md               ← プロジェクトルール（参照先の案内）
├── LICENSE                 （MIT）
├── README.md
├── qsokufile               ← ビルド・テストの近道（qsoku）
├── go.mod / go.sum
├── .gitattributes          （`* text=auto eol=lf`）
├── .gitignore
├── .golangci.yaml          ← lintの設定
├── trivy.yaml              ← 脆弱性・ライセンス検査の設定
├── .mcp.json               ← mtqgとsrwrのMCPサーバーの配線
├── .srwrignore             ← srwrがテープに記録しないファイル（`.mtqg/`・`.bcuz/`など）
├── .bcuz/                  ← bcuzの記録（`history.jsonl`）。`postCreate.sh`が`bcuz --init`で作る。**コミットする**（`merge=union`）
├── .srwr/                  ← srwrの鍵・ロック・テープ（最初のhook・MCPの呼び出しでできる。**閉じた`tapes/*.gz`だけコミットする**）
├── internal/               ← このモジュールの中だけで使うコード（コア）
│   ├── ignore/             ← `.crube/ignore`の除外判定（gitignore と同じ書式）と、init が書く既定の中身
│   ├── lang/               ← `CRUBE_LANG`から使う structure ファイルを1つ決める
│   ├── match/              ← 照合（実在のパスに当たるエントリ。完全一致→`path.Match`を書いた順）
│   ├── project/            ← `.crube/`の探索と、`.crube/`への書き込み（書き込みガード）。**`.crube/`に書くのはここだけ**
│   ├── structure/          ← structure ファイルの読み込みと正規形での書き出し、desc・detail の上限の検証
│   └── textwidth/          ← 表示幅（全角2・半角1・曖昧1・結合文字0）
├── docs/                   ← 設計と仕様（日本語。目次は`docs/README.md`）
│   ├── design/             ← なぜこの形なのか
│   ├── reference/          ← 正確な仕様（CLI・MCP・ファイル形式）
│   └── examples/           ← 使い方の例。`sample-crube/`はサンプルの`.crube/`
├── .devcontainer/          ← 開発環境（devcontainer.json、postCreate.sh）
├── .claude/                ← rules/（ルール）、skills/（commit-with-tape：テープを閉じてコミットする手順）、hooks/（build.sh、require-bcuz.sh とそのテスト）、settings.json（人間が管理する）
├── .github/workflows/      ← CI
└── .mtqg/                  ← mtqgの記録
```

## 配置の判断基準

<!-- TODO: パッケージの具体的な分け方を、決まったら書く -->

`docs/design/architecture.md`の構成に従う。

- **ロジックはすべてコア**（`internal/`の下）に置く。コアのモジュールは、ストア・言語選択・照合・除外判定・差分エンジン・検索。コアはAIもGitも呼び出さず、プロジェクトのファイルは読むだけで、書くのは`.crube/`だけ
- **CLI（`cmd/crube/`）とMCPサーバー（`crube mcp`）は、コアの薄いラッパー**にする。引数・出力の形の変換だけを置き、判定や検証を書かない（CLIとMCPで挙動がずれないように）

## 各ファイル・ディレクトリの補足

- **`cmd/`**: まだ無い。CLIの`crube`を実装するときに`cmd/crube/main.go`を作る。`main.go`は引数を`internal/`に渡すだけにする
- **`internal/`**: Goの仕組みとして、モジュールの外からimportできない。外へ公開するパッケージは`internal/`に置かない
- **`internal/structure/`**: structure ファイルのバイト列と木（`Tree`・`Entry`）の変換だけを持つ。ファイルの探索・ディスクへの書き込み・照合は持たない（ほかのコアのパッケージに置く）
- **`go.sum`**: コミットする（改ざん検知・再現可能なビルドのため）
- **`.mtqg/`**: コミットされる。中のファイルは直接編集せず、すべて`mtqg`のコマンド経由で書く（`.claude/rules/mtqg-usage.md`）。`.mtqg/.local/`はコミットされない
- **`.bcuz/`**: bcuzが`.bcuz/history.jsonl`に追記する、コマンドの実行の記録（`.claude/rules/bcuz-usage.md`）。**コミットする**（`.gitignore`の`!/.bcuz/`）。行を足すだけなので、`.gitattributes`の`merge=union`で、マージしても両側の行が残る。コミットの前に`git add .bcuz/history.jsonl`する（`git commit`は`bcuz`を通さない）。秘密をコマンドに書かない（記録に残る）。中のファイルは直接編集しない。`.srwrignore`でテープにも載せない
- **`.claude/hooks/require-bcuz.sh`**: Bashツールのコマンドを、`bcuz`を通っていなければ実行前に断る`PreToolUse`フック。判定の表は`require-bcuz.test.sh`（`bash .claude/hooks/require-bcuz.test.sh`）。判定を変えたら、テストの表も同じ変更の中で直す
- **`.mcp.json`**: `mtqg mcp`と`srwr mcp`（MCPサーバー）の配線。`mtqg init --agent claude-code`・`srwr init`が書くものと同じ
- **`.srwr/`**: 最初のhook・MCPの呼び出しで、srwrが作る（`srwr init`でも作られる。鍵`key`・`lock`・`active`・`tapes/`・`init-backup/`）。**コミットするのは閉じて圧縮したテープ`tapes/*.tape.jsonl.gz`だけ**。書き込み中の生のテープ`*.tape.jsonl`は`.gitignore`で対象外（閉じると`.gz`に置き換わり、生のままコミットすると「削除＋追加」の差分が出るため）。鍵は共有しない。テンプレートには置いていない（鍵をプロジェクトごとに別にするため）
- **`.srwrignore`**: テープに記録しないファイル（`.gitignore`と同じ書き方）。生成物や巨大なファイルはここに足す。`.env`・`*.pem`・`*.key`などはsrwrが元から外す
- **`.devcontainer/devcontainer-lock.json`**: コンテナのビルドで自動で作られる（このテンプレートには置いていない）。**できたらコミットする**（この版のfeatureでテストなどが通った証明になる）。featureを新しくしたいときは、lockを消してビルドし直し、新しいlockをコミットする
- **VSCodeの設定**: `.devcontainer/devcontainer.json`の`customizations.vscode`に置く（拡張と設定を1か所にまとめる）。`.vscode/`は作らない
- **`.gitattributes`**: `* text=auto eol=lf`。Windowsランナーでの改行コード変換による誤検知を防ぐ（`.claude/rules/testing.md`）。`.mtqg/`の中の`.gitattributes`は`mtqg init`が置くもので、別物
- **`.gitignore`のドットファイル**: ルート直下のドットファイル・ドットディレクトリは**許可リスト方式**（`/.*`を無視し、追跡するものだけ`!`で戻す。`.claude/`の中は`/.claude/*`を無視して`settings.json`・`rules/`・`skills/`・`hooks/`だけ戻す）。サンドボックスが作業ディレクトリ直下に出すダミーファイルを`git add -A`で混ぜないため（`.claude/rules/testing.md`）。**追跡するドットファイルを新しく足すときは、`.gitignore`に`!`の行を足すこと**。足さないと`git add`が無視する
- **配布物（ビルド済みバイナリ）**: リポジトリにコミットしない。`.gitignore`には**ルート直下に限定して**書く（`/名前`）。名前だけを書くと、同名のソースディレクトリまで無視される
- **`docs/`**: 設計と仕様。書き方は`.claude/rules/documentation.md`
- **`docs/examples/sample-crube/`**: 説明用のサンプル。このリポジトリ自身の`.crube/`ではない（ドットで始まらないので`.gitignore`の許可リストの対象外）。このリポジトリ自身に`.crube/`を置くときは、`.gitignore`に`!/.crube/`を足す
- **シェルスクリプト**: 最小限にする。追跡中のものは`qsoku shellcheck`の対象になる（`.claude/rules/testing.md`）
