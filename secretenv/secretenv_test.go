package secretenv

import (
	"os"
	"strings"
	"testing"
)

// unset は、テストのあとで元に戻るようにして、環境変数を消す。
func unset(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, ``)
	os.Unsetenv(key)
}

func TestLoad(t *testing.T) {
	t.Setenv(`SECRETENV_A`, `old`)
	unset(t, `SECRETENV_B`)
	unset(t, `SECRETENV_C`)
	t.Setenv(Name, "SECRETENV_A=new\r\n\n# comment\nSECRETENV_B=x=y\r\n SECRETENV_C = spaced \n")

	if err := load(); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		`SECRETENV_A`: `old`, // すでにあるものは上書きしない
		`SECRETENV_B`: `x=y`,
		`SECRETENV_C`: ` spaced `,
	} {
		if got := os.Getenv(key); got != want {
			t.Errorf(`%s = %q, want %q`, key, got, want)
		}
	}
	if _, ok := os.LookupEnv(Name); ok {
		t.Errorf(`%s is not removed`, Name)
	}
}

func TestLoadWithoutSecrets(t *testing.T) {
	unset(t, Name)
	if err := load(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadError(t *testing.T) {
	unset(t, `SECRETENV_A`)
	for _, v := range []string{"SECRETENV_A=a\nxoxb-token-only", "=xoxb-token-only"} {
		t.Setenv(Name, v)
		err := load()
		if err == nil {
			t.Fatalf(`no error for %q`, v)
		}
		if strings.Contains(err.Error(), `xoxb`) {
			t.Errorf(`the error has the value: %v`, err)
		}
	}
}
