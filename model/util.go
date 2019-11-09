package model

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// GetMD5Hash returns md5 hash from string (it is used to generate gravatar url)
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateExcerpt is function to pull short excerpt from substring of post content
func GenerateExcerpt(content string) string {
	words := strings.Split(content, " ")
	if len(words) < 55 {
		return content
	}

	return strings.Join(words[:55], " ")
}

// GetBaseURL is function to get full base API path include version
func GetBaseURL(ctx context.Context) string {
	apiConfig := ctx.Value(APICONFIGKEY).(APIConfig)
	return apiConfig.APIHost + apiConfig.APIPath + apiConfig.Version
}

// StrStrMap if function to easily create new map with initial value string key with string value from parameters
func StrStrMap(key string, val string) map[string]string {
	return map[string]string{key: val}
}

// HrefMap if function to easily create new StrStrMap with specific map key with value 'href'
func HrefMap(url string) map[string]string {
	return StrStrMap("href", url)
}

// GetEmbeddableLink is function to easily create new EmbeddableLink struct
func GetEmbeddableLink(url string) EmbeddableLink {
	return EmbeddableLink{Href: url, Embeddable: true}
}
