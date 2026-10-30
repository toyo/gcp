// Package secretenv は、サービスのシークレットを 1 つにまとめた値から、環境変数を入れる。
//
// Secret Manager は有効・無効の版ごとに料金がかかり、無料なのは請求先アカウント全体で 6 版なので、
// シークレットはサービスごとに 1 つ（名前はサービス名）にまとめ、service.yaml で環境変数 SECRETS に渡す。
// 値は「名前=値」の行で、空の行と # で始まる行は飛ばす。すでにある環境変数は上書きしない。
//
// 環境変数を読むパッケージで _ "github.com/toyo/gcp/secretenv" と読み込む。
// 読み込んだパッケージの初期化より先に入るので、パッケージの変数の初期化で読むところにも使える。
package secretenv

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Name は、まとめた値を渡す環境変数の名前。
const Name = `SECRETS`

func init() {
	if err := load(); err != nil {
		slog.Error(`Cannot set the secrets to the environment`, `err`, err)
		os.Exit(1)
	}
}

// load は Name の値を環境変数に入れ、Name を消す。エラーには値を入れない（行の番号だけ）。
func load() error {
	v, ok := os.LookupEnv(Name)
	if !ok {
		return nil
	}
	for i, line := range strings.Split(v, "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) == `` || strings.HasPrefix(line, `#`) {
			continue
		}
		key, value, found := strings.Cut(line, `=`)
		key = strings.TrimSpace(key)
		if !found || key == `` {
			return fmt.Errorf(`line %d of %s is not name=value`, i+1, Name)
		}
		if _, set := os.LookupEnv(key); set {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf(`line %d of %s: %w`, i+1, Name, err)
		}
	}
	return os.Unsetenv(Name)
}
