package googlecloudlogging

import (
	"log/slog"
	"testing"
)

func TestSeverity(t *testing.T) {
	for _, c := range []struct {
		level slog.Level
		want  string
	}{
		{slog.LevelWarn, `WARNING`},
		{slog.LevelError, `ERROR`},
		{LevelCritical, `CRITICAL`},
	} {
		a := replacer(nil, slog.Any(slog.LevelKey, c.level))
		if a.Key != `severity` || a.Value.String() != c.want {
			t.Errorf("%v: got %s=%s, want %s", c.level, a.Key, a.Value.String(), c.want)
		}
	}
}
