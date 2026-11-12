# ディレクトリ構成

新しいファイルを追加する際は、どこに置くかをここで確かめる。**構成を変えたら、この文書も同じ変更の中で直す。**

**このリポジトリ自身の説明（`.crube/`）も保つ。** ファイル・ディレクトリを足す・動かす・消すときは、同じ変更の中で英語と日本語の説明を直し、コミットの前に両方の言語で`qsoku crube status --require all`を通す。やり方（MCPとCLI、`.crube/ignore`の保ち方）は`.claude/rules/crube-usage.md`

```
.
├── CLAUDE.md               ← プロジェクトルール（参照先の案内）
├── LICENSE                 （MIT）
├── README.md / README_ja.md（英語と日本語の対）
├── qsokufile               ← ビルド・テストの近道（qsoku）
├── go.mod / go.sum
├── .gitattributes          （`* text=auto eol=lf`）
├── .gitignore
├── .golangci.yaml          ← lintの設定
├── .goreleaser.yaml        ← 配布物の作り方（goreleaser。`release.yml`と`qsoku snapshot`が使う。`.claude/rules/distribution.md`）
├── trivy.yaml              ← 脆弱性・ライセンス検査の設定
├── .mcp.json               ← mtqg・このリポジトリ自身のcrube（1つ。言語は道具の引数`lang`）・fnazeのMCPサーバーの配線
├── .crube/                 ← このリポジトリ自身の説明（crube の dogfooding。`structure.txt`は英語、`structure_ja.txt`は日本語、`ignore`は`.gitignore`に合わせて手で保つ）。**コミットする**
├── .fnaze/changes/         ← コミットごとの説明（fnaze。`.claude/rules/fnaze-usage.md`）。**コミットする**
├── .githooks/              ← git のフック（`pre-commit`：説明の無いコミットを止める。`core.hooksPath`は`postCreate.sh`が設定する）
├── cmd/crube/              ← CLIの`main.go`（引数・環境を`internal/cli`に渡すだけ。リリースの版を受け取る`version`）
├── e2e/                    ← docs の例と実際の出力の突き合わせ、CLI と MCP の一致、補完スクリプトを本物のシェルで動かす（テストだけのパッケージ。`qsoku unit`に入る）
├── internal/               ← このモジュールの中だけで使うコード（コア）
│   ├── cli/                ← CLI（引数の解析、`core`の呼び出し、出力・エラー文・終了コード、シェル補完）。判定は`core`に任せる。`completions/`は埋め込む補完スクリプト
│   ├── core/               ← CLIとMCPが呼ぶ操作（init・set・rm・mv、参照の操作）。ほかのパッケージを組み合わせる
│   ├── ignore/             ← `.crube/ignore`の除外判定（gitignore と同じ書式）と、init が書く既定の中身
│   ├── lang/               ← `--lang`（MCP は`lang`）か`CRUBE_LANG`から、使う structure ファイルを1つ決める
│   ├── mcpserver/          ← MCPサーバー（`crube mcp`。公式のGo SDK）。各ツールは`core`を呼ぶだけ
│   ├── match/              ← 照合（実在のパスに当たるエントリ。完全一致→`path.Match`を書いた順）
│   ├── project/            ← `.crube/`の探索と、`.crube/`への書き込み（書き込みガード）。**`.crube/`に書くのはここだけ**
│   ├── search/             ← 検索（find。パス・desc・detail、インデックスなし）
│   ├── status/             ← 走査と差分エンジン（未説明・孤立・上限超え・充足度）。読むだけ
│   ├── structure/          ← structure ファイルの読み込みと正規形での書き出し、desc・detail の上限の検証
│   └── textwidth/          ← 表示幅（全角2・半角1・曖昧1・結合文字0）
├── tools/licenses/         ← 配布物に入れる依存のライセンス文を集める道具（goreleaser が`go run`で動かす。crube には組み込まれない）
├── docs/                   ← 設計と仕様（各ページが英語`X.md`と日本語`X_ja.md`の対。目次は`docs/README.md`）
│   ├── design/             ← なぜこの形なのか
│   ├── reference/          ← 正確な仕様（CLI・MCP・ファイル形式）
│   ├── examples/           ← 使い方の例。`sample-crube/`はサンプルの`.crube/`
│   └── assets/             ← README の画像（`header.svg`は手で書いたもの、`demo.svg`・`demo_ja.svg`は`qsoku demo`がテストで作るもの）
├── .devcontainer/          ← 開発環境（devcontainer.json、postCreate.sh）
├── .claude/                ← rules/（ルール）、skills/（mutation-check：壊して確かめる手順と、変異の道具`mutants/`（別の Go モジュール）・crube の仕様の変異の一覧）、hooks/（build.sh、crube-untracked.sh：Write のあとに未説明を知らせる。Claude Code のフックで、git のフックは`.githooks/`）、settings.json（人間が管理する）
├── .github/workflows/      ← CI（`ci.yml`）と、タグで動くリリース（`release.yml`。`ci.yml`を呼んでから Release を作って公開する）
├── .github/release-notes/  ← 版ごとのリリースノート（`<タグ>.md`。英語と日本語を1つに。`release.yml`が Release の本文の先頭に載せる。`.claude/rules/distribution.md`）
└── .mtqg/                  ← mtqgの記録
```

## 配置の判断基準

<!-- TODO: パッケージの具体的な分け方を、決まったら書く -->

`docs/design/architecture.md`の構成に従う。

- **ロジックはすべてコア**（`internal/`の下）に置く。コアのモジュールは、ストア・言語選択・照合・除外判定・差分エンジン・検索。コアはAIもGitも呼び出さず、プロジェクトのファイルは読むだけで、書くのは`.crube/`だけ
- **CLI（`cmd/crube/`）とMCPサーバー（`crube mcp`）は、コアの薄いラッパー**にする。引数・出力の形の変換だけを置き、判定や検証を書かない（CLIとMCPで挙動がずれないように）

## 各ファイル・ディレクトリの補足

- **`cmd/crube/`**: `main.go`は、引数・環境変数・標準出力を`internal/cli`の`Run`に渡し、その終了コードで終わるだけにする。テストは`internal/cli`に書く
- **`internal/cli/`**: 引数をどう読むか、結果をどう見せるか（JSON・テキスト）、エラーを`crube: …`の1行と終了コードにすることだけを持つ。値の検証・合否の判定（`status --require`など）は`core`にあり、ここには書かない。`mcp`コマンドは`internal/mcpserver`を起動するだけ（SDK を import しない）
- **`internal/mcpserver/`**: ツールの名前・引数・説明文と、`core`の結果をそのまま返すことだけを持つ。絞り込み（`kinds`）やルートからのパスの読み方も`core`にある（`FilterKinds`・`AtRoot`）。結果の型が再帰すると SDK が出力のスキーマを作れないので（`tree`）、そのときはスキーマを手で渡す
- **`internal/`**: Goの仕組みとして、モジュールの外からimportできない。外へ公開するパッケージは`internal/`に置かない
- **`tools/licenses/`**: 配布物に同梱する依存のライセンス文を集める（`.claude/rules/distribution.md`）。`go run ./tools/licenses <出力先> ./cmd/crube`。依存を足す・外したら、`TestRun`の一覧を直す
- **`.claude/skills/mutation-check/mutants/`**: 変異の道具。`go.mod`を持つ別のモジュールなので、`qsoku check`（`./...`）の対象に入らない。変えたら、そのディレクトリで整形・vet・テストを通す（Skill の「気をつけること」）
- **`internal/core/`**: CLI・MCPから呼ばれる操作の入口。引数の検証・判定はここか、ここが呼ぶパッケージに置き、CLI・MCPには書かない。書き込みは`internal/project`の関数だけを通す
- **`internal/structure/`**: structure ファイルのバイト列と木（`Tree`・`Entry`）の変換だけを持つ。ファイルの探索・ディスクへの書き込み・照合は持たない（ほかのコアのパッケージに置く）
- **`go.sum`**: コミットする（改ざん検知・再現可能なビルドのため）
- **`.mtqg/`**: コミットされる。中のファイルは直接編集せず、すべて`mtqg`のコマンド経由で書く（`.claude/rules/mtqg-usage.md`）。`.mtqg/.local/`はコミットされない
- **`.mcp.json`**: `mtqg mcp`（MCPサーバー）の配線。`mtqg init --agent claude-code`が書くものと同じ。その後ろに、このリポジトリ自身のcrubeのMCPサーバー（`qsoku crube mcp --allow-write`。1つだけで、言語は道具の引数`lang`で渡す）を手で足している（`.claude/rules/crube-usage.md`）。最後に、fnazeのMCPサーバー（`fnaze mcp --allow-write`。1つだけ）を足している（`.claude/rules/fnaze-usage.md`）
- **`.fnaze/`**: コミットごとの説明を fnaze で書き、そのコミットに入れる（`.claude/rules/fnaze-usage.md`）。**コミットする**（`.gitignore`の`!/.fnaze/`）。`.devcontainer/postCreate.sh`が無ければ`fnaze init`で作る。説明のファイルは、crube の説明を付けず`.crube/ignore`で外す（パターンのエントリにすると、最初の`fnaze new`まで何にも当たらず`orphan`になるため）
- **`.githooks/`**: 追跡する git のフック。`.git/hooks/`は git に乗らないので、ここに置いて`core.hooksPath`で指す（`.devcontainer/postCreate.sh`）。`pre-commit`は fnaze の`docs/examples/pre-commit`と、先頭のコメントのほかは同じ中身
- **`.devcontainer/devcontainer-lock.json`**: コンテナのビルドで自動で作られる（このテンプレートには置いていない）。**できたらコミットする**（この版のfeatureでテストなどが通った証明になる）。featureを新しくしたいときは、lockを消してビルドし直し、新しいlockをコミットする
- **VSCodeの設定**: `.devcontainer/devcontainer.json`の`customizations.vscode`に置く（拡張と設定を1か所にまとめる）。`.vscode/`は作らない
- **`.gitattributes`**: `* text=auto eol=lf`。Windowsランナーでの改行コード変換による誤検知を防ぐ（`.claude/rules/testing.md`）。`.mtqg/`の中の`.gitattributes`は`mtqg init`が置くもので、別物
- **`.gitignore`のドットファイル**: ルート直下のドットファイル・ドットディレクトリは**許可リスト方式**（`/.*`を無視し、追跡するものだけ`!`で戻す。`.claude/`の中は`/.claude/*`を無視して`settings.json`・`rules/`・`skills/`・`hooks/`だけ戻す）。サンドボックスが作業ディレクトリ直下に出すダミーファイルを`git add -A`で混ぜないため（`.claude/rules/testing.md`）。**追跡するドットファイルを新しく足すときは、`.gitignore`に`!`の行を足すこと**。足さないと`git add`が無視する
- **配布物（ビルド済みバイナリ）**: リポジトリにコミットしない。`.gitignore`には**ルート直下に限定して**書く（`/名前`）。名前だけを書くと、同名のソースディレクトリまで無視される
- **`docs/`**: 設計と仕様。書き方は`.claude/rules/documentation.md`
- **`docs/examples/sample-crube/`**: 説明用のサンプル。このリポジトリ自身の`.crube/`（ルート直下。`.gitignore`の`!/.crube/`で戻している）とは別物。ドットで始まらないので`.gitignore`の許可リストの対象外
- **シェルスクリプト**: 最小限にする。追跡中のものは`qsoku shellcheck`の対象になる（`.claude/rules/testing.md`）
