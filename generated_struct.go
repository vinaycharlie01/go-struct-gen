package main

type EventData1 struct {
	ID int `json:"id,omitempty"`
	Owner Owner `json:"owner,omitempty"`
}
type Owner struct {
	Login string `json:"login,omitempty"`
	ID int `json:"id,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	GravatarID string `json:"gravatar_id,omitempty"`
	URL string `json:"url,omitempty"`
	HTMLURL string `json:"html_url,omitempty"`
	FollowersURL string `json:"followers_url,omitempty"`
	FollowingURL string `json:"following_url,omitempty"`
	GistsURL string `json:"gists_url,omitempty"`
	StarredURL string `json:"starred_url,omitempty"`
	SubscriptionsURL string `json:"subscriptions_url,omitempty"`
	OrganizationsURL string `json:"organizations_url,omitempty"`
	ReposURL string `json:"repos_url,omitempty"`
	EventsURL string `json:"events_url,omitempty"`
	ReceivedEventsURL string `json:"received_events_url,omitempty"`
	Type string `json:"type,omitempty"`
	SiteAdmin bool `json:"site_admin,omitempty"`
	Listdata []int `json:"listdata,omitempty"`
	Hello []any `json:"hello,omitempty"`
}