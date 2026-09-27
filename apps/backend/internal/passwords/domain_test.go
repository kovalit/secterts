package passwords

import "testing"

func TestExtractDomain(t *testing.T) {
	cases := map[string]string{
		"https://mail.google.com/mail/u/0/#inbox": "mail.google.com",
		"https://github.com/settings/profile":     "github.com",
		"https://www.example.com":                 "example.com",
		"http://reg.ru":                            "reg.ru",
		"github.com":                               "github.com",
		"www.site.io/path":                         "site.io",
		"":                                         "",
		"ftp://files.example.com":                  "",
		"not a url with spaces":                    "",
	}
	for in, want := range cases {
		if got := ExtractDomain(in); got != want {
			t.Errorf("ExtractDomain(%q) = %q, want %q", in, got, want)
		}
	}
}
