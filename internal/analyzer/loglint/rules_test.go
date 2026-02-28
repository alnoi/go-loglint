package loglint

import (
	"go/parser"
	"go/token"
	"testing"
)

// helpers

func cfgAllOn() *Config {
	var cfg Config
	cfg.Rules.Lowercase = true
	cfg.Rules.NoEmoji = true
	cfg.Rules.EnglishOnly = true
	cfg.Sensitive.Patterns = []string{"password", "token", "secret", "apikey"}
	cfg.Sensitive.Allowlist = nil
	return &cfg
}

func cfgSensitive(patterns, allow []string) *Config {
	var cfg Config
	cfg.Sensitive.Patterns = patterns
	cfg.Sensitive.Allowlist = allow
	return &cfg
}

type anyExpr struct {
	expr any
}

// tests

func TestValidateMessage_OrderAndPresence(t *testing.T) {
	cfg := cfgAllOn()

	msg := "Hello..."
	got := validateMessage(msg, cfg)

	if len(got) != 2 {
		t.Fatalf("expected 2 violations, got %d: %#v", len(got), got)
	}
	if got[0].Rule != RuleLowercase {
		t.Fatalf("expected first rule %q, got %q", RuleLowercase, got[0].Rule)
	}
	if got[1].Rule != RuleNoEmoji {
		t.Fatalf("expected second rule %q, got %q", RuleNoEmoji, got[1].Rule)
	}
}

func TestValidateMessage_EnglishOnlyAndNoEmojiBothPossible(t *testing.T) {
	cfg := cfgAllOn()

	msg := "started 🚀"
	got := validateMessage(msg, cfg)

	if len(got) != 2 {
		t.Fatalf("expected 2 violations, got %d: %#v", len(got), got)
	}
	if got[0].Rule != RuleNoEmoji {
		t.Fatalf("expected first rule %q, got %q", RuleNoEmoji, got[0].Rule)
	}
	if got[1].Rule != RuleEnglish {
		t.Fatalf("expected second rule %q, got %q", RuleEnglish, got[1].Rule)
	}
}

func TestCheckLowercase(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"empty", "", false},
		{"spaces_only", "   \t\t", false},
		{"already_lower", "hello", false},
		{"leading_spaces_then_lower", "   hello", false},
		{"starts_upper_letter", "Hello", true},
		{"leading_spaces_then_upper", "   Hello", true},
		{"starts_digit", "1hello", false},
		{"invalid_utf8_runeerror", string([]byte{0xff, 0xfe}) + "X", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := checkLowercase(tt.msg)
			if (v != nil) != tt.want {
				t.Fatalf("checkLowercase(%q) violation=%v, want %v; v=%#v", tt.msg, v != nil, tt.want, v)
			}
			if v != nil && v.Rule != RuleLowercase {
				t.Fatalf("expected rule %q, got %q", RuleLowercase, v.Rule)
			}
		})
	}
}

func TestCheckEnglishOnly(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"ascii_ok", "ok...", false},
		{"tabs_newlines_ok", "ok\tok\nok\rok", false},
		{"control_char_bad", "ok" + string([]byte{0x1f}) + "x", true},
		{"non_ascii_bad", "ok…", true},
		{"emoji_bad", "🚀", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := checkEnglishOnly(tt.msg)
			if (v != nil) != tt.want {
				t.Fatalf("checkEnglishOnly(%q) violation=%v, want %v; v=%#v", tt.msg, v != nil, tt.want, v)
			}
			if v != nil && v.Rule != RuleEnglish {
				t.Fatalf("expected rule %q, got %q", RuleEnglish, v.Rule)
			}
		})
	}
}

func TestCheckNoEmojiOrSpecial(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want bool
	}{
		{"ok", "all good", false},
		{"three_dots_ascii", "ok...", true},
		{"ellipsis_unicode", "ok…", true},
		{"question_mark", "why?!", true},
		{"emoji", "started 🚀", true},
		{"zwj_joiner", "A\u200dB", true},
		{"vs16_selector", "A\ufe0fB", true},
		{"non_ascii_non_emoji_not_special", "привет", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := checkNoEmojiOrSpecial(tt.msg)
			if (v != nil) != tt.want {
				t.Fatalf("checkNoEmojiOrSpecial(%q) violation=%v, want %v; v=%#v", tt.msg, v != nil, tt.want, v)
			}
			if v != nil && v.Rule != RuleNoEmoji {
				t.Fatalf("expected rule %q, got %q", RuleNoEmoji, v.Rule)
			}
		})
	}
}

func TestIsEmojiRune(t *testing.T) {
	if isEmojiRune('a') {
		t.Fatalf("expected 'a' to not be emoji")
	}
	if !isEmojiRune('🚀') {
		t.Fatalf("expected 🚀 to be emoji by table")
	}
}

func TestContainsSensitiveWord(t *testing.T) {
	cfg := cfgSensitive(
		[]string{"token", "pass"},
		[]string{"tokenized", "compassion"},
	)

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"empty", "", false},
		{"match_pattern_lower", "token", true},
		{"match_pattern_upper", "TOKEN", true},
		{"contains_pattern", "mytokenvalue", true},
		{"allowlist_blocks_tokenized", "tokenized", false},
		{"allowlist_blocks_compassion", "compassion", false},
		{"allowlist_blocks_contains", "xxcompassionyy", false},
		{"no_match", "session", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsSensitiveWord(tt.in, cfg)
			if got != tt.want {
				t.Fatalf("containsSensitiveWord(%q)=%v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCheckNoSensitive_AstShapes(t *testing.T) {
	cfg := cfgSensitive([]string{"password", "token"}, nil)

	type tc struct {
		name string
		expr string
		want bool
	}

	tests := []tc{
		{"ident_sensitive", "password", true},
		{"ident_not_sensitive", "username", false},
		{"selector_sensitive", "user.password", true},
		{"selector_not_sensitive", "user.name", false},
		{"binary_add_sensitive_left", "password + x", true},
		{"binary_add_sensitive_right", "x + token", true},
		{"binary_not_add_ignored", "x - password", false},
		{"call_args_sensitive", "foo(bar(password))", true},
		{"call_args_not_sensitive", "foo(bar(username))", false},
		{"nested_calls_sensitive", "zap.String(\"k\", token)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := parser.ParseExpr(tt.expr)
			if err != nil {
				t.Fatalf("ParseExpr(%q): %v", tt.expr, err)
			}

			pos, _, ok := checkNoSensitive(expr, cfg)
			if ok != tt.want {
				t.Fatalf("checkNoSensitive(%q) ok=%v want %v (pos=%v)", tt.expr, ok, tt.want, pos)
			}

			if ok && pos == token.NoPos {
				t.Fatalf("expected pos != NoPos for %q", tt.expr)
			}
		})
	}
}

func TestCheckNoSensitive_AllowlistWins(t *testing.T) {
	cfg := cfgSensitive([]string{"token"}, []string{"tokenized"})

	expr, err := parser.ParseExpr("tokenized")
	if err != nil {
		t.Fatalf("ParseExpr: %v", err)
	}

	pos, _, ok := checkNoSensitive(expr, cfg)
	if ok {
		t.Fatalf("expected allowlist to block detection, got ok=true (pos=%v)", pos)
	}
}
