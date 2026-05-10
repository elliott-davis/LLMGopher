package storage

import (
	"strings"
	"testing"
)

func TestRedactErrorMessage_Empty(t *testing.T) {
	if got := RedactErrorMessage(""); got != "" {
		t.Fatalf("empty string: got %q, want %q", got, "")
	}
}

func TestRedactErrorMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRedact  []string // substrings that MUST be gone
		wantPresent []string // substrings that MUST survive
	}{
		{
			name:        "bearer token",
			input:       "upstream error: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.abc",
			wantRedact:  []string{"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"},
			wantPresent: []string{"Bearer", "upstream error"},
		},
		{
			name:        "authorization header",
			input:       "Authorization: Bearer sk-abc123xyz456def789ghi012",
			wantRedact:  []string{"sk-abc123xyz456def789ghi012"},
			wantPresent: []string{"Authorization"},
		},
		{
			name:        "sk- api key prefix lowercase",
			input:       "invalid key sk-testkey123456789012345 rejected",
			wantRedact:  []string{"sk-testkey123456789012345"},
			wantPresent: []string{"invalid key", "rejected"},
		},
		{
			name:        "sk- api key prefix uppercase SK-",
			input:       "request failed SK-MySecretKeyABCDEF1234 not authorised",
			wantRedact:  []string{"SK-MySecretKeyABCDEF1234"},
			wantPresent: []string{"request failed", "not authorised"},
		},
		{
			name:        "cookie header",
			input:       "Cookie: sid=sessiontoken123456789012",
			wantRedact:  []string{"sessiontoken123456789012"},
			wantPresent: []string{"Cookie"},
		},
		{
			name:        "base64-like long substring",
			input:       "error decoding payload dGVzdHRva2VuMTIzNDU2Nzg5MA== from body",
			wantRedact:  []string{"dGVzdHRva2VuMTIzNDU2Nzg5MA=="},
			wantPresent: []string{"error decoding payload", "from body"},
		},
		{
			name:        "keyword token with value",
			input:       "token: abc1234xyz rejected",
			wantRedact:  []string{"abc1234xyz"},
			wantPresent: []string{"token"},
		},
		{
			name:        "keyword password equals",
			input:       "password=hunter2xyz failed authentication",
			wantRedact:  []string{"hunter2xyz"},
			wantPresent: []string{"password", "failed authentication"},
		},
		{
			name:        "keyword secret",
			input:       "secret: mySecret1234 exposed",
			wantRedact:  []string{"mySecret1234"},
			wantPresent: []string{"secret", "exposed"},
		},
		{
			name:        "keyword credential",
			input:       "credential: APIKEY123ABCDEF exposed in log",
			wantRedact:  []string{"APIKEY123ABCDEF"},
			wantPresent: []string{"credential", "exposed in log"},
		},
		{
			name:        "budget keyword preserved for outcome derivation",
			input:       "budget exceeded for key sk-abc123def456ghi789jkl012, token: mySecret123 denied",
			wantRedact:  []string{"sk-abc123def456ghi789jkl012"},
			wantPresent: []string{"budget exceeded", "denied"},
		},
		{
			name:        "no sensitive content unchanged",
			input:       "upstream timeout after 30s",
			wantRedact:  nil,
			wantPresent: []string{"upstream timeout after 30s"},
		},
		{
			name:        "case-insensitive keyword matching",
			input:       "Token: ABCD1234xyz and Key: EFG567890abc pass denied",
			wantRedact:  []string{"ABCD1234xyz", "EFG567890abc"},
			wantPresent: []string{"Token", "Key", "pass denied"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := RedactErrorMessage(tc.input)
			for _, secret := range tc.wantRedact {
				if strings.Contains(got, secret) {
					t.Errorf("secret %q still present in %q", secret, got)
				}
			}
			for _, keep := range tc.wantPresent {
				if !strings.Contains(got, keep) {
					t.Errorf("expected %q to be present in %q", keep, got)
				}
			}
			if len(tc.wantRedact) > 0 && !strings.Contains(got, "[REDACTED]") {
				t.Errorf("expected [REDACTED] marker in %q", got)
			}
		})
	}
}
