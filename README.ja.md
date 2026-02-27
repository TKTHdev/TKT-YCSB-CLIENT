[English](README.md) | **日本語**

# ycsb-client

YCSB (Yahoo! Cloud Serving Benchmark) ワークロードを TCP テキストプロトコルでキーバリューサーバーに送信するベンチマーククライアント。

## サーバー実装要件

このセクションでは、ycsb-client でベンチマークするためにサーバーが実装しなければならない仕様を定義します。

### 接続モデル

- サーバーは指定アドレスで TCP 接続を受け付ける必要があります。
- 各 worker goroutine はベンチマーク開始時に **TCP コネクションを 1 本だけ開き**、10 秒間の計測中はそれを使い回します。サーバーはリクエストとリクエストの間でコネクションを **閉じてはいけません**。
- クライアントは **1 リクエストずつ逐次送信** します（パイプライニングなし）。リクエストを送ってレスポンスを受け取ってから次のリクエストを送ります。サーバーは 1 コネクション内のリクエストを順番に処理すれば十分です。
- 複数の worker が同時に動くため、サーバーは **N 本の同時接続** を処理できなければなりません（N = `--workers` の値）。

### ワイヤーフォーマット

すべてのメッセージは **UTF-8 プレーンテキスト**、1 行 1 メッセージ、`\n`（LF のみ、CRLF 不可）で終端します。

**SET リクエスト**（書き込み）:

```
SET <key> <value>\n
```

- `<key>`: 英数字のみ、スペースなし（例: `k0`, `k42`）
- `<value>`: 100 バイトの英数字文字列、スペースなし

サーバーはキーバリューペアを保存し、**正確に以下を返します**:

```
OK\n
```

**GET リクエスト**（読み込み）:

```
GET <key>\n
```

サーバーは**以下のいずれか 1 行を正確に返します**:

```
OK <value>\n    ← キーが存在する場合: "OK" + 半角スペース + 保存されている value
ERR\n           ← キーが存在しない場合、またはエラー
```

> **重要:** `OK` を受け取ったリクエストのみがスループット・レイテンシの計測対象になります。`ERR` は無視されます。YCSB-C（読み込み専用ワークロード）では、ベンチマーク前にストアにデータを投入しておかないと、全 GET が `ERR` を返してスループットがゼロになります。

### まとめ

| ケース | クライアントが送る | サーバーが返すべき値 |
|---|---|---|
| 書き込み | `SET k value\n` | `OK\n` |
| 読み込み（キーあり） | `GET k\n` | `OK value\n` |
| 読み込み（キーなし） | `GET k\n` | `ERR\n` |
| その他エラー | — | `ERR\n` |

### リファレンス実装

`test-server/` にこのプロトコルを満たす最小構成のシングルノードインメモリ KV サーバーがあります。自分の分散システムにサーバー側を実装する際の参考にしてください。

## ワークロード

| 名前   | 書き込み比率 | 読み込み比率 |
|--------|------------|------------|
| ycsb-a | 50%        | 50%        |
| ycsb-b | 5%         | 95%        |
| ycsb-c | 0%         | 100%       |

- キー: `k0` 〜 `k{N-1}` から一様ランダムに選択
- バリュー: 100 バイトのランダム英数字文字列
- 計測時間: 10 秒

## ビルド

```bash
# クライアント
cd ycsb-client
go build -o ycsb-client .

# テスト用インメモリ KV サーバー
cd test-server
go build -o test-server .
```

## 使い方

```bash
# テストサーバーを起動
./test-server --addr localhost:7000

# ベンチマーク実行
./ycsb-client \
  --addr     localhost:7000 \  # サーバーアドレス（複数可）
  --workload ycsb-a \          # ワークロード種別
  --workers  4 \               # 並列 worker 数
  --keys     100 \             # キー数
  --csv      result.csv        # CSV 出力先（省略可）
```

### 複数サーバーへの分散

`--addr` を複数指定すると worker がラウンドロビンで接続先を選びます。

```bash
./ycsb-client \
  --addr localhost:7000 \
  --addr localhost:7001 \
  --workers 8 \
  --workload ycsb-b
```

## 出力

### 標準出力

```
[ycsb-client] addrs=[localhost:7000] workload=ycsb-a workers=4 keys=100
Benchmark completed
Total ops: 120345
Throughput: 12034.50 ops/sec
Avg latency: 0.33 ms
RESULT:ycsb-a,4,100,12034.50,0.33
```

### CSV ファイル

ファイルが存在しない場合はヘッダー行を自動付与し、実行するたびに追記されます。

```csv
workload,workers,keys,throughput_ops_sec,avg_latency_ms
ycsb-a,4,100,12034.50,0.33
ycsb-b,4,100,15200.00,0.26
ycsb-c,4,100,18500.00,0.21
```

## ファイル構成

```
ycsb-client/
  main.go       CLI エントリーポイント
  benchmark.go  ワークロード生成・計測・CSV 出力
  client.go     TCP クライアント (GET/SET)
  go.mod
  test-server/
    main.go     テスト用インメモリ KV サーバー
    go.mod
```
