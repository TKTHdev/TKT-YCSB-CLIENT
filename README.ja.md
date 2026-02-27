[English](README.md) | **日本語**

# ycsb-client

YCSB (Yahoo! Cloud Serving Benchmark) ワークロードを TCP テキストプロトコルでキーバリューサーバーに送信するベンチマーククライアント。

## プロトコル

サーバーは以下の行ベーステキストプロトコルを実装する必要があります。

```
# 書き込み
Client → Server:  SET key value\n
Server → Client:  OK\n

# 読み込み
Client → Server:  GET key\n
Server → Client:  OK value\n   (キーが存在する場合)
                  ERR\n         (キーが存在しない、またはエラー)
```

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
