package tachiyomi

type LocalDetail struct {
	Title        string   `json:"title,omitempty"`
	Author       string   `json:"author,omitempty"`
	Artist       string   `json:"artist,omitempty"`
	Description  string   `json:"description,omitempty"`
	Genre        []string `json:"genre,omitempty"`
	Status       string   `json:"status,omitempty"`
	StatusValues []string `json:"_status values,omitempty"`
}
