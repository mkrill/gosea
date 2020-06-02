package api

type Post struct {
	UserName string `json:"userName"`
	Title    string `json:"title,omitempty"`
	Body     string `json:"body,omitempty"`
	Text     string `json:"text,omitempty"`
}
