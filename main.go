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

	"github.com/disintegration/imaging"
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

			// Open image for compression
			img, err := imaging.Open(savePath)
			if err != nil {
				log.Println("Failed to open uploaded image for compression:", err)
			} else {
				width := img.Bounds().Dx()
				height := img.Bounds().Dy()
				log.Printf("Original image width: %dx%d\n", width, height)

				// Compress if width > 800px
				if width > 800 {
					img = imaging.Resize(img, 800, 0, imaging.Lanczos)
					if err := imaging.Save(img, savePath); err != nil {
						log.Println("Failed to save compressed image:", err)
					} else {
						newWidth := img.Bounds().Dx()
						newHeight := img.Bounds().Dy()
						log.Printf("Image compressed to: %dx%d\n", newWidth, newHeight)
					}
				}
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

		// Open image for compression
		img, err := imaging.Open(savePath)
		if err != nil {
			log.Println("Failed to open uploaded image for compression:", err)
		} else {
			width := img.Bounds().Dx()
			log.Printf("Original image width: %dpx\n", width)

			if width > 800 {
				img = imaging.Resize(img, 800, 0, imaging.Lanczos)
				if err := imaging.Save(img, savePath); err != nil {
					log.Println("Failed to save compressed embedded image:", err)
				} else {
					newWidth := img.Bounds().Dx()
					newHeight := img.Bounds().Dy()
					log.Printf("Embedded image compressed to: %dx%d\n", newWidth, newHeight)
				}
			}
		}

		// Return URL to client
		return c.JSON(fiber.Map{"url": "/" + savePath})
	})

	// start Fatal
	log.Fatal(app.Listen(":3030"))
}
