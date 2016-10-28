package chronograf

import "net/http"

// Assets returns a handler to serve the website.
type Assets interface {
	Handler() http.Handler
}
