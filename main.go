package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/starwalkn/gotenberg-go-client/v8"
	"github.com/starwalkn/gotenberg-go-client/v8/document"
)

type Document struct {
	Header          string `json:"header" form:"header"`
	Body            string `json:"body" form:"body"`
	Footer          string `json:"footer" form:"footer"`
	BackgroundImage string `json:"-"`
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    []byte `json:"data,omitempty"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Static("/static", "static")

	e.GET("/", handleHome)
	e.POST("/generate-pdf", handleGeneratePDF)
	e.POST("/preview-pdf", handlePreviewPDF)

	port := os.Getenv("PORT")
	log.Printf("Starting server at port %s...", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func loadBackgroundImage() (string, error) {
	imgFile, err := os.ReadFile("static/kop.png")
	if err != nil {
		return "", fmt.Errorf("failed to read background image: %v", err)
	}

	return base64.StdEncoding.EncodeToString(imgFile), nil
}

func handleHome(c echo.Context) error {
	return c.File("templates/index.html")
}

func handleGeneratePDF(c echo.Context) error {
	doc := new(Document)
	if err := c.Bind(doc); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Invalid input data",
		})
	}

	// Validate input
	if doc.Header == "" || doc.Body == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Header and Body are required",
		})
	}

	// Load background image
	backgroundImage, err := loadBackgroundImage()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Status:  "error",
			Message: "Failed to load background image: " + err.Error(),
		})
	}
	doc.BackgroundImage = backgroundImage

	pdfBytes, err := generatePDF(doc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Status:  "error",
			Message: "Failed to generate PDF: " + err.Error(),
		})
	}

	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func handlePreviewPDF(c echo.Context) error {
	doc := new(Document)
	if err := c.Bind(doc); err != nil {
		return c.JSON(http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Invalid input data",
		})
	}

	// Validate input
	if doc.Header == "" || doc.Body == "" {
		return c.JSON(http.StatusBadRequest, Response{
			Status:  "error",
			Message: "Header and Body are required",
		})
	}

	// Load background image
	backgroundImage, err := loadBackgroundImage()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Status:  "error",
			Message: "Failed to load background image: " + err.Error(),
		})
	}
	doc.BackgroundImage = backgroundImage

	pdfBytes, err := generatePDF(doc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Response{
			Status:  "error",
			Message: "Failed to generate PDF preview: " + err.Error(),
		})
	}

	return c.Blob(http.StatusOK, "application/pdf", pdfBytes)
}

func generatePDF(doc *Document) ([]byte, error) {
	// Create HTML content
	tmpl, err := template.ParseFiles("templates/document.html")
	if err != nil {
		return nil, fmt.Errorf("template parsing error: %v", err)
	}

	var htmlContent bytes.Buffer
	if err := tmpl.Execute(&htmlContent, doc); err != nil {
		return nil, fmt.Errorf("template execution error: %v", err)
	}

	gotenbergURL := os.Getenv("GOTENBERG_URL")
	if gotenbergURL == "" {
		log.Fatal("GOTENBERG_URL or PORT not set in .env")
	}

	client, _ := gotenberg.NewClient(gotenbergURL, http.DefaultClient)

	gotenbergIsAuth := os.Getenv("GOTENBERG_IS_AUTH") == "true"
	gotenbergUsername := os.Getenv("GOTENBERG_USERNAMES")
	gotenbergPassword := os.Getenv("GOTENBERG_PASSWORD")

	html, err := document.FromString("index.html", htmlContent.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %v", err)
	}

	req := gotenberg.NewHTMLRequest(html)

	if gotenbergIsAuth {
		req.UseBasicAuth(gotenbergUsername, gotenbergPassword)
	}

	req.PaperSize(gotenberg.A4)
	req.Margins(gotenberg.NoMargins)

	// Generate PDF
	resp, err := client.Send(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("PDF generation error: %v", err)
	}

	pdfBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response body: %w", err)
	}
	defer resp.Body.Close()

	return pdfBytes, nil
}
