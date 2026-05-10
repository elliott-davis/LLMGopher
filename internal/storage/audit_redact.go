package storage

import (
	"regexp"
	"strings"
)

var (
	reBearerValue  = regexp.MustCompile(`(?i)(Bearer\s+)\S+`)
	reAuthHeader   = regexp.MustCompile(`(?i)(Authorization:\s*)\S+`)
	reAPIKeyPrefix = regexp.MustCompile(`(?i)\bsk-[A-Za-z0-9_\-]{6,}`)
	reCookieHeader = regexp.MustCompile(`(?i)(Cookie:\s*)\S+`)
	reLongToken    = regexp.MustCompile(`[A-Za-z0-9+/=_\-]{20,}`)
	// keyword followed by separator and a value of ≥4 non-whitespace chars
	reKeywordValue = regexp.MustCompile(`(?i)\b(key|secret|token|password|credential)([\s:=]+)(\S{4,})`)
)

// RedactErrorMessage replaces credential-like substrings in msg with [REDACTED].
// Surrounding human-readable context (including audit-relevant words like "budget")
// is preserved. The stored audit row is never mutated; call this only when building
// the HTTP response.
func RedactErrorMessage(msg string) string {
	if msg == "" {
		return msg
	}
	msg = reBearerValue.ReplaceAllString(msg, "${1}[REDACTED]")
	msg = reAuthHeader.ReplaceAllString(msg, "${1}[REDACTED]")
	msg = reAPIKeyPrefix.ReplaceAllString(msg, "[REDACTED]")
	msg = reCookieHeader.ReplaceAllString(msg, "${1}[REDACTED]")
	// keyword-value rule before long-token so the keyword label is preserved
	msg = reKeywordValue.ReplaceAllString(msg, "${1}${2}[REDACTED]")
	msg = reLongToken.ReplaceAllStringFunc(msg, redactIfMixedToken)
	return msg
}

// redactIfMixedToken returns [REDACTED] for strings that look like encoded tokens
// (contain ≥2 character classes: digits, uppercase, lowercase, or special chars).
// Pure-alpha strings that somehow reach ≥20 chars are also redacted since they
// are unlikely to be human-readable audit context.
func redactIfMixedToken(s string) string {
	if strings.Contains(s, "[REDACTED]") {
		return s
	}
	classes := 0
	hasDigit, hasUpper, hasLower, hasSpecial := false, false, false, false
	for _, c := range s {
		switch {
		case c >= '0' && c <= '9' && !hasDigit:
			hasDigit = true
			classes++
		case c >= 'A' && c <= 'Z' && !hasUpper:
			hasUpper = true
			classes++
		case c >= 'a' && c <= 'z' && !hasLower:
			hasLower = true
			classes++
		case (c == '+' || c == '/' || c == '=' || c == '_' || c == '-') && !hasSpecial:
			hasSpecial = true
			classes++
		}
	}
	if classes >= 2 {
		return "[REDACTED]"
	}
	return s
}
