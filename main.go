package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

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

			shortID := uuid.New().String()[:8]

			// Clean the original filename
			ext := filepath.Ext(file.Filename)
			name := strings.TrimSuffix(file.Filename, ext)

			// New filename with UUID
			newFilename := name + "_" + shortID + ext
			savePath := filepath.Join("uploads", newFilename)

			if err := c.SaveFile(file, savePath); err != nil {
				return c.Status(500).SendString("Failed to save file")
			}

			imagePath = "/" + savePath
		}

		post := Post{
			Title:   title,
			Content: template.HTML(content),
			Image:   imagePath,
		}

		db.Create(&post)
		return c.Redirect("/")
	})

	// handle embeded images (images inside quil txt editor)
	app.Post("/upload", func(c *fiber.Ctx) error {
		file, err := c.FormFile("image")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "No img file uploaded"})
		}

		os.MkdirAll("./uploads", os.ModePerm)

		// generate 8 character UUID
		shortID := uuid.New().String()[:8]

		// Clean the original filename (remove path)
		ext := filepath.Ext(file.Filename)
		name := strings.TrimSuffix(file.Filename, ext)

		// New filename with UUID
		newFilename := name + "_" + shortID + ext
		savePath := filepath.Join("uploads", newFilename)

		// Save the file
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
		}

		// Return URL to client
		return c.JSON(fiber.Map{"url": "/" + savePath})
	})

	// start Fatal
	log.Fatal(app.Listen(":3030"))
}
