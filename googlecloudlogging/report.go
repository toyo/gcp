package googlecloudlogging

import (
	"context"
	"net/url"
	"path"
	"strings"
)

// reportKey は、電文のファイル名を context に入れるときのキー。
type reportKey struct{}

// WithReport は、ctx に電文（気象庁防災情報 XML）の URL を付ける。この ctx で出したログには、電文のファイル名（拡張子なし。
// 例 20261026171037_0_VPZU53_010000）が `report` として付く。電文は xmlpull → jmx → chatbot・firebase → 配信サービスと
// Cloud Tasks をはさんで進み、サービスごとにトレース ID が変わるので、この値で 1 つの電文のログをまとめて検索する
// （jsonPayload.report="…"）。uri が空なら ctx のまま。
func WithReport(ctx context.Context, uri string) context.Context {
	if name := ReportName(uri); name != `` {
		return context.WithValue(ctx, reportKey{}, name)
	}
	return ctx
}

// ReportName は、電文の URL のファイル名（拡張子なし）。空の URL なら空。
func ReportName(uri string) string {
	if u, err := url.Parse(uri); err == nil {
		uri = u.Path
	}
	name := strings.TrimSuffix(path.Base(uri), `.xml`)
	if name == `.` || name == `/` {
		return ``
	}
	return name
}
