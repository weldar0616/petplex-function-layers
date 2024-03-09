#!/bin/sh

# スクリプトが置いてあるディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# Layerが格納されているディレクトリへ移動
cd "$SCRIPT_DIR"

# 各Layerのトップレベルディレクトリでループ処理
for layerTop in */ ; do
    echo "Processing layer top directory: $layerTop"
    # フルパスから最後の'/'を削除し、トップレベルのディレクトリ名を取得
    layerTopDir=${layerTop%/}

    # 各レイヤーディレクトリでループ処理
    for layer in "$layerTop"*/ ; do
        if [ -d "$layer" ]; then  # ディレクトリであることを確認
            echo "Processing layer directory: $layer"
            # フルパスから最後の'/'を削除し、実際のレイヤー名を取得
            layerDir=${layer%/}
            layerName=${layerDir##*/}

            # Goモジュールが存在するか確認
            if [ -f "$layer/go.mod" ]; then
                # ビルドディレクトリへ移動
                cd "$layer"

                # Goのビルドコマンドを実行
                GOOS=linux GOARCH=amd64 go build -o bootstrap main.go

                # ビルドに成功したらZIPファイルに圧縮
                if [ -f "bootstrap" ]; then
                    zip "${layerTopDir}_${layerName}.zip" bootstrap
                else
                    echo "Build failed for $layer, skipping."
                fi

                # 実行ファイルとZIPファイルを削除
                rm -f bootstrap

                # 元のディレクトリに戻る
                cd - > /dev/null
            else
                echo "Skipping $layer, no Go module found."
            fi
        fi
    done
done

echo "All layers processed."
