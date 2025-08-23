package httpapi

import (
	"net/http"

	"github.com/kxddry/wbf/ginext"
)

// ServeStatic registers a handler that serves only the root index.html from the given directory.
func ServeStatic(r *ginext.Engine, routePrefix string, dir string) {
	indexPath := dir + "/index.html"
	r.GET(routePrefix, func(c *ginext.Context) {
		http.ServeFile(c.Writer, c.Request, indexPath)
	})
}
