package httpapi

import (
	"net/http"

	"github.com/kxddry/wbf/ginext"
)

func ServeStatic(r *ginext.Engine, routePrefix string, dir string) {
	fileServer := http.FileServer(http.Dir(dir))
	r.StaticFS(routePrefix, http.Dir(dir))
	// Ensure index.html is served at root
	r.GET("/", func(c *ginext.Context) { fileServer.ServeHTTP(c.Writer, c.Request) })
}
