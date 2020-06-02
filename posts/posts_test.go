package posts

import (
	"testing"
)

func TestPosts_loadPosts(t *testing.T) {
	p := NewWithSEA()

	posts, err := p.LoadPosts()

	t.Log(err)
	t.Log(posts)
}
