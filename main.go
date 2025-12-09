package main

import (
	"log"
	"os"
	"path/filepath"
	"html/template"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB model
type Post struct {
	ID      uint `gorm:"primaryKey"`
	Title   string
	Content template.HTML
	Image   string
}

func main() {
	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{Views: engine})

	// connect to SQLite DB
	db, err := gorm.Open(sqlite.Open("blog.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("DB connection failed")
	}

	// create table if not exist
	db.AutoMigrate(&Post{})

	// serve static files
	app.Static("/uploads", "./uploads")

	// route homepage
	app.Get("/", func(c *fiber.Ctx) error {
		var posts []Post
		db.Order("id desc").Find(&posts)

		return c.Render("index", fiber.Map{
			"Posts": posts,
		})
	})

	// create post
	app.Post("/create", func(c *fiber.Ctx) error {
		title := c.FormValue("title")
		content := c.FormValue("content")

		// handle uploaded file
		imagePath := ""

		file, err := c.FormFile("image")
		if err == nil {
			os.MkdirAll("./uploads", os.ModePerm)

			filename := filepath.Join("uploads", file.Filename)
			c.SaveFile(file, filename)

			imagePath = "/" + filename
		}

		post := Post{
			Title:   title,
			Content: template.HTML(content),
			Image:   imagePath,
		}

		db.Create(&post)
		return c.Redirect("/")
	})

	// start Fatal
	log.Fatal(app.Listen(":3030"))
}
