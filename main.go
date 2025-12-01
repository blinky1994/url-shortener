package main

import (
	"strings"

	// The web framework used to define routes and handlers.
	"github.com/gin-gonic/gin"
	// UUID generator used to produce reasonably unique short IDs.
	"github.com/google/uuid"
)

// main is the program entrypoint. It wires up the HTTP server and routes.
// High-level flow:
//   - Load HTML template used for the UI
//   - Define handlers:
//     GET  /        -> show the form
//     POST /shorten -> create a short id and store mapping in the DB
//     GET  /:id     -> lookup id in DB and redirect to stored target
//   - Start the server on the configured port (defaults to :8080)
func main() {
	// gin.Default() returns a router with logging and recovery middleware
	// already attached — useful for development.
	r := gin.Default()

	// Load the single HTML template used by the web UI.
	// Templates are looked up relative to the current working directory.
	r.LoadHTMLFiles("templates/index.html")

	// GET / - show the page with the form to shorten a URL.
	r.GET("/", func(c *gin.Context) {
		// Render the template and pass data (a title) to it.
		c.HTML(200, "index.html", gin.H{"title": "Rate-limited URL Shortener"})
	})

	// POST /shorten - receive a URL and create a short id for it.
	r.POST("/shorten", func(c *gin.Context) {
		// Read the "url" field from the submitted form.
		target := c.PostForm("url")
		if target == "" {
			// If the user didn't provide a URL, return HTTP 400 (Bad Request).
			c.String(400, "no url")
			return
		}

		// Generate a short identifier. We use a UUID, remove dashes and take
		// the first 8 characters. This is simple and avoids collisions in small
		// projects — in production you'd use a more robust approach.
		id := strings.ReplaceAll(uuid.New().String(), "-", "")[:8]

		// Insert the new mapping into the database. DB is a global *sql.DB
		// initialized in `db.go`'s init() function. Using parameter placeholders
		// like `?` prevents SQL injection by keeping SQL and data separate.
		_, err := DB.Exec("INSERT INTO links(id, target) VALUES(?, ?)", id, target)
		if err != nil {
			// On DB errors, return HTTP 500 (Internal Server Error).
			c.String(500, "db, error")
			return
		}

		// Commit is largely unnecessary with database/sql + SQLite in this
		// simple usage, but the call is harmless here. Errors are ignored for
		// brevity; a real app should handle them.
		DB.Exec("COMMIT")

		// Build a full short URL for the developer environment and render it
		// back to the user. In production you'd use your domain name instead
		// of localhost.
		short := "http://localhost:8080/" + id
		c.HTML(200, "index.html", gin.H{
			"title": "done!",
			"short": short,
		})
	})

	// GET /:id - redirect a short id to the stored target URL.
	r.GET("/:id", func(c *gin.Context) {
		// Extract the id path parameter from the URL.
		id := c.Param("id")

		// Query the database for the original target URL. QueryRow returns
		// a single row; Scan will return sql.ErrNoRows if the id doesn't
		// exist, which we treat as a 404 (link is dead).
		var target string
		err := DB.QueryRow("SELECT target FROM links WHERE id = ?", id).Scan(&target)
		if err != nil {
			c.String(404, "link is dead")
			return
		}

		// Increment a click counter for basic stats. We intentionally ignore
		// errors from this statement — in a real app you'd handle/log them.
		DB.Exec("UPDATE links SET clicks = clicks + 1 WHERE id = ?", id)

		// Redirect the client to the original URL.
		c.Redirect(302, target)
	})

	// Start the HTTP server. r.Run() without arguments listens on :8080 by
	// default. You can pass an address like ":8081" to change the port or
	// set the PORT environment variable before running the program.
	r.Run()
}
