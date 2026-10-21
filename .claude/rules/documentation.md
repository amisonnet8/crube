# ドキュメントの書き方

<!-- TODO: 書き方の細かい決まりと、更新のルール（仕様と実装がずれたらどちらを直すか）を書く -->

## 言語と置き場

- `docs/`は日本語で書く。目次は`docs/README.md`で、ページを足したら目次にも足す
- `docs/design/`：なぜこの形なのか（今の設計とその理由）。決めたことは`docs/design/decisions.md`の表にも1行で足す
- `docs/reference/`：正確な仕様（CLI・MCP・ファイル形式）。実装はこれに合わせる
- `docs/examples/`：使い方の例と、サンプルの`.crube/`（`docs/examples/sample-crube/`）
- docs は実装より先に書かれている。「未定」「実装時に確定」と書いたもの（`docs/reference/cli.md`の`set`・`rm`・`mv`の`--json`の形など）は、決めたら docs を直す。利用者の目に触れるものは、決める前に了承を得る（`.claude/rules/working-with-human.md` 4章）
- structure ファイルの例は、インデントと区切りが**タブ**。docs の例を書き換えるときに、エディタがタブをスペースに置き換えないよう気をつける（スペースのインデントは読み込みエラーになる仕様）

## 例の値

- **docs の例の値（ハッシュ・出力・件数など）は、推測で書かず、道具で求めるか、実際に動かして確かめて書く**（`.claude/rules/testing.md`「期待値は、テストを書く前に決める」と同じ考え方）。実装より先に docs を書くときも同じ
- 例はコードブロックで書く。**`e2e/`のテストが、docs の Markdown を直接読んで実際の出力と突き合わせる**（`qsoku check`に入っている）。docs の例を書く・直したら、`qsoku check`で確かめる
  - コマンドの例は、1行目から`$ `で始まるブロックにする。`$ `の行が実際に動かされ、その下の行と比べられる（`...`の行は「何行でも」）。使えるのは`crube …`・`cd`・`export`・`unset`・行末の`; echo $?`・`  # コメント`だけ
  - cli.md と README の例は、1つの架空のプロジェクト（`src/`・`data/`。`e2e/helpers_test.go`の`cliProject`）を前提にする。statusの例だけは別のプロジェクト（`statusProject`）。例を足すときは、このプロジェクトで実際に出る値を書く
  - JSON の例（```json）は、どのコマンドの出力かを`e2e/docs_test.go`の`TestJSONExamples`の表に足す。足さないと「知らない例」としてテストが落ちる。コマンドの出力でない JSON（クライアントの設定など）は、同じテストの`notOutput`に足す
  - cli.md「出力の共通の決まり」のエラー文は、1行ずつ`TestErrorExamples`の表に、それを出す場面を足す
  - README.md と docs/README.md の`$ `の例も、`cliProject`で動かして突き合わせる（`TestREADMEExample`）。どの OS でも同じ出力になる例だけを書く（`cliProject`のシンボリックリンク`current`は、Windows で作れないことがあるので、ルートの`ls`や`status`は README の`$ `の例にしない）
  - README のデモの画像（`docs/assets/demo.svg`）は、`e2e/demo_test.go`の`TestDemoSVG`が、実際の出力から作る。出力が変わると、このテストが落ちる。`qsoku demo`で作り直し、画像を見てからコミットする。手で書き換えない
  - ヘッダーの画像（`docs/assets/header.svg`）は手で書いた SVG。中の例は飾りで、テストしない

## `docs/design/`に書く経緯

`docs/design/`には**今の設計**を書く。経緯は、**前のやり方に戻すと、同じ問題がまた起きるもの**だけを書く。それ以外の経緯は、mtqgに記録する（`.claude/rules/mtqg-usage.md`）。

- **書く**：今の設計を選んだ理由で、捨てたやり方に戻すと問題が再び起きるもの。「何を選び、何を捨て、なぜか」を数行で書く。詳しいやり取りは、mtqgのID（`q`・`m`）で指す
  - 例：「XXXアルゴリズムを使う。以前はYYYアルゴリズムを使っていたが、zzzzの問題が起きたため」→ 書く。YYYに戻すと、zzzzがまた起きる
- **書かない**：今の設計の働きに関わらない変更の経緯（名前・表記・置き場所を変えた理由など）、日付ごとの流れ、誰が何を言ったか、途中で直した履歴、todo・bugの進み。これらはmtqgに記録する
  - 例：「ツールの名前はaaaとする。以前はbbbだったが、同じ名前のツールがすでにあったため変えた」→ 書かない。`docs/design/`には「名前はaaa」だけを書く
