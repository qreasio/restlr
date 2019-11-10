package model

import "time"

// User contains the compact data for user
type User struct {
	ID          uint64   `json:"ID"`
	DisplayName string   `json:"name"`
	NiceName    string   `json:"slug"`
	URL         string   `json:"url"`
	Link        string   `json:"link"`
	Description *string  `json:"description,omitempty"`
	Links       UserLink `json:"_links"`
}

// UserDetail contains the longer and complete user data
type UserDetail struct {
	User
	Email         *string            `json:"email,omitempty"`
	Login         *string            `json:"user_login,omitempty"`
	Pass          *string            `json:"user_pass,omitempty"`
	Registered    time.Time          `json:"user_registered,omitempty"`
	ActivationKey string             `json:"user_activation_key,omitempty"`
	Status        int                `json:"user_status,omitempty"`
	AvatarUrls    *map[string]string `json:"avatar_urls,omitempty"`
}

// UserLink represents _links in EmbedUserResponse
type UserLink struct {
	SelfLink   []map[string]string `json:"self"`
	Collection []map[string]string `json:"collection"`
}
