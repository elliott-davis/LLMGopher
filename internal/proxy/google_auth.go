package proxy

import (
	"context"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const googleCloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

// NewGoogleCloudTokenSource returns a reusable ADC-backed token source for GCP APIs.
func NewGoogleCloudTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	src, err := google.DefaultTokenSource(ctx, googleCloudPlatformScope)
	if err != nil {
		return nil, err
	}
	// ReuseTokenSource caches/refreshes tokens safely for concurrent callers.
	return oauth2.ReuseTokenSource(nil, src), nil
}
