package loglint

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type RuleID string

const (
	RuleLowercase RuleID = "lowercase"
	RuleEnglish   RuleID = "english-only"
	RuleNoEmoji   RuleID = "no-emoji"
	RuleSensitive RuleID = "no-sensitive"
)

var emojiTable = &unicode.RangeTable{
	R16: []unicode.Range16{
		{Lo: 0x2600, Hi: 0x26FF, Stride: 1},
		{Lo: 0x2700, Hi: 0x27BF, Stride: 1},
	},
	R32: []unicode.Range32{
		{Lo: 0x1F300, Hi: 0x1F5FF, Stride: 1},
		{Lo: 0x1F600, Hi: 0x1F64F, Stride: 1},
		{Lo: 0x1F680, Hi: 0x1F6FF, Stride: 1},
		{Lo: 0x1F1E6, Hi: 0x1F1FF, Stride: 1},
		{Lo: 0x1F900, Hi: 0x1F9FF, Stride: 1},
		{Lo: 0x1FA70, Hi: 0x1FAFF, Stride: 1},
	},
}

type Violation struct {
	Rule    RuleID
	Message string
}

func validateMessage(msg string, cfg Config) []Violation {
	var out []Violation

	if cfg.Rules.Lowercase {
		if v := checkLowercase(msg); v != nil {
			out = append(out, *v)
		}
	}
	if cfg.Rules.NoEmoji {
		if v := checkNoEmojiOrSpecial(msg); v != nil {
			out = append(out, *v)
		}
	}
	if cfg.Rules.EnglishOnly {
		if v := checkEnglishOnly(msg); v != nil {
			out = append(out, *v)
		}
	}

	return out
}

func checkLowercase(msg string) *Violation {
	s := strings.TrimLeft(msg, " \t")
	if s == "" {
		return nil
	}

	r, _ := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return nil
	}

	if unicode.IsLetter(r) && !unicode.IsLower(r) {
		return &Violation{
			Rule:    RuleLowercase,
			Message: "message must start with a lowercase letter",
		}
	}

	return nil
}

func checkEnglishOnly(msg string) *Violation {
	for _, r := range msg {
		if r == '\t' || r == '\n' || r == '\r' {
			continue
		}
		if r < 0x20 || r > 0x7E {
			return &Violation{
				Rule:    RuleEnglish,
				Message: "message must be english-only",
			}
		}
	}
	return nil
}

func checkNoEmojiOrSpecial(msg string) *Violation {
	violation := func(what string) *Violation {
		return &Violation{
			Rule:    RuleNoEmoji,
			Message: "message must not contain emoji or special characters: " + what,
		}
	}

	if strings.Contains(msg, "...") {
		return violation(`found "..."`)
	}
	if strings.ContainsRune(msg, '\u2026') {
		return violation(`found "…"`)
	}

	for _, r := range msg {
		switch r {
		case '!', '?':
			return violation("found " + strconv.QuoteRune(r))
		}

		if r <= 0x7E {
			continue
		}

		if isEmojiRune(r) {
			return violation("found emoji " + strconv.QuoteRune(r))
		}
		if r == '\u200d' {
			return violation(`found emoji joiner "\u200d"`)
		}
		if r == '\ufe0f' {
			return violation(`found emoji variation selector "\ufe0f"`)
		}
	}

	return nil
}

func isEmojiRune(r rune) bool {
	return unicode.Is(emojiTable, r)
}

func checkNoSensitive(expr ast.Expr, cfg Config) (token.Pos, string, bool) {
	switch e := expr.(type) {

	case *ast.Ident:
		if containsSensitiveWord(e.Name, cfg) {
			return e.Pos(), "log contains sensitive identifier", true
		}

	case *ast.SelectorExpr:
		if containsSensitiveWord(e.Sel.Name, cfg) {
			return e.Sel.Pos(), "log contains sensitive field", true
		}

	case *ast.BinaryExpr:
		if e.Op != token.ADD {
			return token.NoPos, "", false
		}

		if pos, why, ok := checkNoSensitive(e.X, cfg); ok {
			return pos, why, true
		}
		if pos, why, ok := checkNoSensitive(e.Y, cfg); ok {
			return pos, why, true
		}

	case *ast.CallExpr:
		for _, a := range e.Args {
			if pos, why, ok := checkNoSensitive(a, cfg); ok {
				return pos, why, true
			}
		}
	}

	return token.NoPos, "", false
}

func containsSensitiveWord(name string, cfg Config) bool {
	s := strings.ToLower(name)

	for _, a := range cfg.Sensitive.Allowlist {
		if a == "" {
			continue
		}
		if strings.Contains(s, a) {
			return false
		}
	}

	for _, p := range cfg.Sensitive.Patterns {
		if p == "" {
			continue
		}
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
