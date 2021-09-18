package tachidesk

type Source struct {
	Name string `json:"name"`
}

type Manga struct {
	Id       int    `json:"id"`
	Url      string `json:"url"`
	SourceId string `json:"sourceId"`

	Author      string   `json:"author"`
	Artist      string   `json:"artist"`
	Description string   `json:"description"`
	Genre       []string `json:"genre"`
	Status      string   `json:"status"`
}

type Chapter struct {
	Index     int    `json:"index"`
	Name      string `json:"name"`
	PageCount int    `json:"pageCount"`
}
