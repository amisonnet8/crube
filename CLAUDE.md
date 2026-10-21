# CLAUDE.md

このファイルは、このプロジェクトで作業する際、Claude Codeが常に踏まえるべき前提を示す。

## このプロジェクトについて

crubeは、プロジェクト内のファイルとディレクトリに説明を付けて`.crube/`に保存し、説明付きの`ls`／`tree`／`find`で目的の場所へすぐたどり着けるようにするツール。説明の作成・更新にはAIを使えるが、参照するときにAIは要らない。実装はGo（単一バイナリ）、データはタブのインデントで書く独自のツリー形式のテキスト（`.crube/structure.txt`）、ライセンスはMIT。

設計と仕様は`docs/`にある（`docs/README.md`が目次）。**実装の前に、関係するページを読む。**

- `docs/design/`：なぜこの形なのか（概要と基本方針、アーキテクチャ、データモデル、鮮度と充足度、決定事項）
- `docs/reference/`：正確な仕様（CLI、MCP、ファイル形式）
- `docs/examples/`：使い方の例と、サンプルの`.crube/`

## 守ること（設計の芯）

実装の細部より優先する。これを崩す変更は、実装する前に必ず相談すること。出所は`docs/design/overview_ja.md`「基本方針」、`docs/design/architecture_ja.md`、`docs/design/decisions_ja.md`。

- **書き込むのは`.crube/`の中だけ**（不変条件）：プロジェクトのファイルは読むだけで、作成・変更・削除・移動をしない。`mv`・`rm`も説明を動かす・消すだけ。書き込みの直前に、書き込み先が`.crube/`の中であることをコアで検証する（シンボリックリンク経由で外に出るものも拒否）
- **生成と参照の分離**：crube本体はAIを呼び出さない。参照（`ls`・`tree`・`find`・`show`・`status`）は保存済みのデータだけで動く
- **ロジックはコアに集め、CLIとMCPはその薄いラッパーにする**：両者の間に機能差・挙動のズレを出さない
- **シンプルさを優先する**：`.crube/`は1つだけ（入れ子にしない）。設定ファイルを持たない（言語は`CRUBE_LANG`、ほかはオプション）。推測（移動の検出など）や自動修正をせず、事実の報告にとどめる。検索インデックスを持たない
- **独立している**：Git（`.gitignore`も含む）、特定のAIプロバイダー、ほかのツールに依存しない。structureファイルは標準ライブラリだけで読み書きする
- **役割を絞る**：扱うのは「何がどこにあるか」の説明だけ。フィールドは desc と detail だけ。目的別の参照順序、ツール本体からのAI呼び出し、他ツールからの情報の取り込みは足さない
- **判断は使う側に委ねる**：充足度の既定値・推奨の長さなどのポリシーを持たない。事実を報告し、合否は使う側が決める（`status --require`）

## ドキュメントの言語

`CLAUDE.md`・`.claude/rules/`は日本語。README と`docs/`は、英語（`X.md`）と日本語（`X_ja.md`）の対で、片方を直したら同じ変更でもう片方も直す（`.claude/rules/documentation.md`）。

<!-- TODO: コード・コメント・コミットメッセージの言語を決めて書く（CLI出力は英語だけと決まった。`docs/reference/cli_ja.md`「出力の共通の決まり」）。詳しい書き方は`.claude/rules/documentation.md` -->

## 参照すべきファイル

作業を始める前に、以下を確認すること。

- **`.claude/rules/bcuz-usage.md`** — bcuzの使い方。**Bashで実行するコマンドは、必ず`bcuz "理由" コマンド…`の形で打つ**（フックが、そうでないものを断る）。理由は「なぜ」を書く。**コマンドを打つ前に読む**
- **`.claude/rules/mtqg-usage.md`** — mtqgの使い方（記録者、種類の使い分け、こまめに記録する）
- **`.claude/rules/srwr-usage.md`** — srwrの使い方（**厳格モード**）。**`Edit`・`Write`は禁止されており、ファイルの変更は`look`／`edit`／`new`（理由`why`つき）で行う。作業のまとまりごとに`session`で題つきのテープを始め、コミットの前にテープを閉じる。ファイルを変える前に読む**
- **`.claude/rules/qsoku-usage.md`** — qsokuの使い方（`qsokufile`の名前の一覧、呼び方、書式、`//`、シェル連携）。**ビルド・テスト・lintは`go`などを直接打たず、ここの名前で実行する**
- **`.claude/rules/testing.md`** — テスト方針（開発環境の前提、動作確認の粒度、Trivy、ShellCheck、`-race`、Bashサンドボックス・CIの落とし穴）
- **`.claude/rules/directory-structure.md`** — ディレクトリ構成と、ファイルの置き場所の判断基準。**新しいファイルを追加するときに読む**
- **`.claude/rules/datetime.md`** — 日時の扱い（データは UTC で持ち、表示だけ利用者のタイムゾーンに直す。瞬間で比べる、時計を差し替える）。**日時を保存・表示・比較するコードを書くときに読む**
- **`.claude/rules/working-with-human.md`** — 人間とのやり取り（頼み方の3種類、小さなつまずきの頼み方、確認の自動化、見え方を変えるときの了承）。**人間に何かを頼む前に読む**
- **`.claude/rules/naming.md`** — 命名規則（未記入。決まったら書く）
- **`.claude/rules/distribution.md`** — 配布方法（リリースの流れ。そのほかは未記入で、決まったら書く）
- **`.claude/rules/documentation.md`** — ドキュメントの書き方（例の値は道具で求める、`docs/design/`に書く経緯の線引き。そのほかは未記入で、決まったら書く）

これらのルールファイルは、実装中に得た細かい気づき・教訓をClaude Code自身が育てていくものである。新しく気づいたルール・踏んだ落とし穴があれば、該当するファイルに追記すること。どのファイルにも当てはまらない新しい種類の気づきであれば、新しいルールファイルを作ってよい（作ったらこの一覧にも足すこと）。

## 開発の進め方

- **Bashのコマンドは、すべて`bcuz "理由" コマンド…`で実行する**（`bcuz "lintの確認" qsoku lint`）。理由は「何をするか」でなく「なぜするか」。`bcuz`で始まらない・`&&`や`;`でつなぐコマンドは、`.claude/settings.json`の`PreToolUse`フック（`.claude/hooks/require-bcuz.sh`）が実行前に断る。**`git commit`だけは`bcuz`を通さない**（通すとコミットの後に記録が増えるため。理由はコミットメッセージに書く）。実行の記録は`.bcuz/history.jsonl`に残り（コミットする）、`bcuz -H`で見られる。**トークンなどの秘密を、コマンドの引数・`env`の指定・理由に書かない**（記録に残り、コミットされる）（`.claude/rules/bcuz-usage.md`）
- **ファイルの変更は、srwrの道具（`look`で見て`edit`、新規は`new`）で行い、すべてに理由（`why`）を添える。** 組み込みの`Edit`・`Write`は`.claude/settings.json`で禁止されている（厳格モード）。Bashでの書き換え（`sed -i`・リダイレクト）もしない
- **作業のまとまり（todo・bug・人間の依頼・作業の種類の変わり目）ごとに、最初に`session`で題つきのテープを始める。** 1つの作業＝1本のテープ、コミットも同じ区切り
- **コミットの前に、テープを閉じる**（`srwr tapes check`のあと`srwr tapes new`）。コミットまでの手順はSkill`commit-with-tape`。書き込み中の生のテープは`.gitignore`で対象外で、閉じて圧縮した`.srwr/tapes/*.tape.jsonl.gz`だけがgitに入る（`.claude/rules/srwr-usage.md`）
- 実装後は`.claude/rules/testing.md`に従い、ビルド確認に加えて実際の動作確認（`bcuz "…" qsoku check`、必要に応じて`qsoku race`・`qsoku trivy`・`qsoku shellcheck`）を行うこと。判断に迷ったら実行する側に倒すこと
- 依存ライブラリは、標準ライブラリ → `golang.org/x` → 何かの公式のもの → デファクトの順で選ぶ（依存を足すこと自体が人間に聞く判断。`.claude/rules/working-with-human.md`）
- 依存ライブラリを足したら、Trivyでライセンス・脆弱性を確認すること（`qsoku trivy`）

## gitの扱い（Claude Codeの作業として）

- **pushは人間が行う。** Claude Codeはコミットまでで止めること。`git push`は`.claude/settings.json`で拒否（deny）されている
- 作業の区切りでこまめにコミットしてよい（消してしまったときに戻せるように）
- `git reset --hard`・`git clean`は、コミットしていない作業を消すので、使う前に必ず理由を説明すること
- **コミットの前に`git config user.email`／`user.name`を確かめる。** 正しい値は`65074207+amisonnet8@users.noreply.github.com`／`amisonnet8`（GitHubのプライバシー用noreplyアドレス）。違う値なら、`.git/config`の`[user]`（ローカルの上書き）を疑う。ローカルに設定が無いと、コンテナのglobalの設定（noreplyでないことがある）が使われる

## 権限・自動化について

`.claude/settings.json`（人間が管理する）により、次の設定になっている。

- 拒否（deny）：`git push`、`gh pr merge`、`gh release create`
- **`bcuz`を通さないBashコマンドの拒否**：`PreToolUse`フック（`.claude/hooks/require-bcuz.sh`）が、`bcuz`で始まらないコマンド（`git commit`を除く。逆に`bcuz`を通した`git commit`を断る）と、後ろに別のコマンドをつなぐもの（`&&`・`||`・`;`・`&`・改行・単一引用符の外のコマンド置換）を、実行前に断る。直し方は断りのメッセージと`.claude/rules/bcuz-usage.md`に書いてある
- 実行前に確認（ask）：`git reset --hard`、`git clean`、`sudo`、`gh pr create`
- **Bashコマンドは既定でOSレベルのサンドボックス（Linux bubblewrap）内で実行される。** ファイルシステムの書き込み先とネットワーク接続先は許可リストで絞られている。この範囲内なら`curl`・`wget`・`go get`などは確認なしに実行できる。リストに無い宛先・書き込み先は`sandbox_violations`として拒否される。そのときは、`.claude/rules/working-with-human.md` 2章の形（原因と、`.claude/settings.json` のどこに何を足すか）で**短く頼む**（許可を広げるかは人間が判断する）。落とし穴は`.claude/rules/testing.md`「Bashサンドボックスの落とし穴」参照

確認を求められた場合、無理に実行しようとしない。`.claude/settings.json`を変更したいときは、変更案を差分で示すにとどめること（人間が貼るだけで済む形にする）。

**srwrの厳格モード**：`permissions.deny`に`Edit`・`Write`・`MultiEdit`・`NotebookEdit`があり、`.mcp.json`にsrwrのMCPサーバー、`PostToolUse`に`srwr hook`（読み取りと外部変更の記録）がある。テープ（`.srwr/tapes/`）はgitで共有する（鍵`.srwr/key`は入れない）。許可を広げる（緩いモードにする）のは人間の判断。詳細は`.claude/rules/srwr-usage.md`。

ビルドの自動フック（`PostToolUse`）が設定されている。srwrの`edit`・`new`で`.go`・`go.mod`・`go.sum`を変えると`.claude/hooks/build.sh`が`qsoku build`を実行し、失敗すると理由がClaude Codeに返る。複数ファイルにまたがる編集の途中は、まだ書いていないファイルを参照して一時的に失敗することがある。

## mtqgでの記録について

作業の中で生まれる「気づき・やること・質問と答え・不具合とそのやり取り・用語の合意・この先も効き続ける決まり事」は、コードにもコミットにも残らず、チャットの履歴やメモアプリに散って失われやすい。`mtqg`（`m`emo & rules・`t`odo・`q`a & bugs・`g`lossary）はそれを`.mtqg/journal.jsonl`に追記し、gitに乗せて共有する。使い方は`.claude/rules/mtqg-usage.md`。`.mtqg/`には、テンプレートから引き継いだ決まり事（rule）を9件登録済み——`mtqg context`で確認できる。AIエージェントからは、フック（`.claude/settings.json`の`mtqg hook`）とMCPサーバー（`.mcp.json`の`mtqg mcp`）の両方で使える。

## ルール・スキルの提案

作業を進める中で、以下に気づいたら、都度こちらから提案すること（提案するだけで、勝手に作成・適用はしない。判断はこちらが行う）。

- **新しいルールにした方がよさそうな知見**: 同じ種類の判断や落とし穴に複数回遭遇した、既存の`.claude/rules/`のどれにも当てはまらない、といった場合。どのファイルに追記すべきか、あるいは新規ファイルが必要かも合わせて提案する
- **Skill化した方が効率的そうな作業**: 「同じ手順を3回目繰り返している」など、パターン化できそうな一連の作業に気づいた場合。まだ実装が薄い段階では無理に提案する必要はない。パターンが実際に繰り返されてから提案すること

提案は気づいたタイミングで随時行ってよく、まとめて報告するために貯めておく必要はない。

This repository records its development with mtqg: run `mtqg context` at the start of a session, and see `.mtqg/SCHEMA.md` for the data format.
