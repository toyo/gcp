package googlecloudlogging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

// TestReportName: 電文の URL から、ファイル名（拡張子なし）を取り出す。
func TestReportName(t *testing.T) {
	for uri, want := range map[string]string{
		`https://www.data.jma.go.jp/developer/xml/data/20261026171037_0_VPZU53_010000.xml`: `20261026171037_0_VPZU53_010000`,
		`file:///jmaxml_20260917_Samples/12_03_01_180117_VPZU53.xml`:                       `12_03_01_180117_VPZU53`,
		``: ``,
	} {
		if got := ReportName(uri); got != want {
			t.Errorf("%q: got %q, want %q", uri, got, want)
		}
	}
}

// TestWithReport: WithReport を付けた ctx で出したログにだけ `report` が付く。
func TestWithReport(t *testing.T) {
	var b bytes.Buffer
	logger := slog.New(handlerWithSpanContext(slog.NewJSONHandler(&b, nil)))

	ctx := WithReport(context.Background(), `https://www.data.jma.go.jp/developer/xml/data/20261026171037_0_VPZU53_010000.xml`)
	logger.InfoContext(ctx, `with`)
	logger.InfoContext(WithReport(context.Background(), ``), `without`)

	lines := strings.Split(strings.TrimSpace(b.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines: %q", lines)
	}
	if !strings.Contains(lines[0], `"report":"20261026171037_0_VPZU53_010000"`) {
		t.Errorf("with: %s", lines[0])
	}
	if strings.Contains(lines[1], `"report"`) {
		t.Errorf("without: %s", lines[1])
	}
}
