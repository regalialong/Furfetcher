package main

import (
	"encoding/json"
	"time"
)

// For whatever fucking reason, Akkoma returns ISO 8601 datetime
// with and *without* milliseconds timezone
// Someone needs life insurance in the present future.

// AkkomaTime parses dates returned by Akkoma with inclusivity for timezone and millisecond data.
// Passed data must be ISO 8601.
type AkkomaTime time.Time

func (ct *AkkomaTime) UnmarshalJSON(b []byte) error {
	var strTime string
	err := json.Unmarshal(b, &strTime)
	if err != nil {
		return err
	}

	layoutTzMs := "2006-01-02T15:04:05.000Z"
	layoutRegular := "2006-01-02T15:04:05"
	parsedTime := time.Time{}

	if len(strTime) == len(layoutTzMs) { // With ms+tz
		parsedTime, err = time.Parse(time.RFC3339, strTime)
	} else if len(strTime) == len(layoutRegular) { // Without
		parsedTime, err = time.Parse(layoutRegular, strTime)
	} else {
		panic("Datetime has unexpected length!")
	}
	if err != nil {
		return err
	}

	*ct = AkkomaTime(parsedTime)
	return nil
}

type Status struct {
	Id                 string      `json:"id"`
	CreatedAt          AkkomaTime  `json:"created_at"`
	InReplyToId        interface{} `json:"in_reply_to_id"`
	InReplyToAccountId interface{} `json:"in_reply_to_account_id"`
	Sensitive          bool        `json:"sensitive"`
	SpoilerText        string      `json:"spoiler_text"`
	Visibility         string      `json:"visibility"`
	Language           string      `json:"language"`
	Uri                string      `json:"uri"`
	Url                string      `json:"url"`
	RepliesCount       int         `json:"replies_count"`
	ReblogsCount       int         `json:"reblogs_count"`
	FavouritesCount    int         `json:"favourites_count"`
	Favourited         bool        `json:"favourited"`
	Reblogged          bool        `json:"reblogged"`
	Muted              bool        `json:"muted"`
	Bookmarked         bool        `json:"bookmarked"`
	Content            string      `json:"content"`
	Reblog             interface{} `json:"reblog"`
	Application        struct {
		Name    string      `json:"name"`
		Website interface{} `json:"website"`
	} `json:"application"`
	Account struct {
		Id             string        `json:"id"`
		Username       string        `json:"username"`
		Acct           string        `json:"acct"`
		DisplayName    string        `json:"display_name"`
		Locked         bool          `json:"locked"`
		Bot            bool          `json:"bot"`
		Discoverable   bool          `json:"discoverable"`
		Group          bool          `json:"group"`
		CreatedAt      AkkomaTime    `json:"created_at"`
		Note           string        `json:"note"`
		Url            string        `json:"url"`
		Avatar         string        `json:"avatar"`
		AvatarStatic   string        `json:"avatar_static"`
		Header         string        `json:"header"`
		HeaderStatic   string        `json:"header_static"`
		FollowersCount int           `json:"followers_count"`
		FollowingCount int           `json:"following_count"`
		StatusesCount  int           `json:"statuses_count"`
		LastStatusAt   AkkomaTime    `json:"last_status_at"`
		Emojis         []interface{} `json:"emojis"`
		Fields         []struct {
			Name       string      `json:"name"`
			Value      string      `json:"value"`
			VerifiedAt *AkkomaTime `json:"verified_at"`
		} `json:"fields"`
	} `json:"account"`
	MediaAttachments []interface{} `json:"media_attachments"`
	Mentions         []interface{} `json:"mentions"`
	Tags             []interface{} `json:"tags"`
	Emojis           []interface{} `json:"emojis"`
	Card             struct {
		Url          string      `json:"url"`
		Title        string      `json:"title"`
		Description  string      `json:"description"`
		Type         string      `json:"type"`
		AuthorName   string      `json:"author_name"`
		AuthorUrl    string      `json:"author_url"`
		ProviderName string      `json:"provider_name"`
		ProviderUrl  string      `json:"provider_url"`
		Html         string      `json:"html"`
		Width        int         `json:"width"`
		Height       int         `json:"height"`
		Image        interface{} `json:"image"`
		EmbedUrl     string      `json:"embed_url"`
	} `json:"card"`
	Poll interface{} `json:"poll"`
}

type StatusParameters struct {
	Status      string   `json:"status"`
	ContentType string   `json:"content_type"`
	MediaIds    []string ` json:"media_ids,omitempty"`
	Visibility  string   `json:"visibility,omitempty"`
}

type MediaResponse struct {
	Id         string      `json:"id"`
	Type       string      `json:"type"`
	Url        *string     `json:"url"`
	PreviewUrl string      `json:"preview_url"`
	RemoteUrl  interface{} `json:"remote_url"`
	TextUrl    string      `json:"text_url"`
	Meta       struct {
		Focus struct {
			X float64 `json:"x"`
			Y float64 `json:"y"`
		} `json:"focus"`
		Original struct {
			Width  int     `json:"width"`
			Height int     `json:"height"`
			Size   string  `json:"size"`
			Aspect float64 `json:"aspect"`
		} `json:"original"`
		Small struct {
			Width  int     `json:"width"`
			Height int     `json:"height"`
			Size   string  `json:"size"`
			Aspect float64 `json:"aspect"`
		} `json:"small"`
	} `json:"meta"`
	Description string `json:"description"`
	Blurhash    string `json:"blurhash"`
}
