package google

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"path"
	"strings"
)

const defaultImageMIMEType = "image/jpeg"

func inferImageMIMEFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return defaultImageMIMEType
	}

	ext := strings.ToLower(path.Ext(parsed.Path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return defaultImageMIMEType
	}
}

func parseBase64ImageDataURI(uri string) (string, []byte, bool) {
	if !strings.HasPrefix(uri, "data:image/") {
		return "", nil, false
	}

	idx := strings.Index(uri, ";base64,")
	if idx <= len("data:") {
		return "", nil, false
	}

	mediaType := uri[len("data:"):idx]
	base64Data := uri[idx+len(";base64,"):]
	if mediaType == "" || base64Data == "" {
		return "", nil, false
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", nil, false
	}
	return mediaType, data, true
}

func rawMessageIsJSONArray(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == '['
}
