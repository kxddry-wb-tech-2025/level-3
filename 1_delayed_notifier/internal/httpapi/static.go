package httpapi

import (
	"net/http"

	"github.com/kxddry/wbf/ginext"
)

// ServeStatic registers a handler that serves only the root index.html from the given directory.
func ServeStatic(r *ginext.Engine, routePrefix string, dir string) {
	// Serve only index.html at root to avoid wildcard conflicts
	indexPath := dir + "/index.html"
	r.GET("/", func(c *ginext.Context) {
		http.ServeFile(c.Writer, c.Request, indexPath)
	})
}
