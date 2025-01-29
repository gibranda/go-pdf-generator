package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/starwalkn/gotenberg-go-client/v8"
	"github.com/starwalkn/gotenberg-go-client/v8/document"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	gotenbergURL := os.Getenv("GOTENBERG_URL")
	port := os.Getenv("PORT")
	if gotenbergURL == "" || port == "" {
		log.Fatal("GOTENBERG_URL or PORT not set in .env")
	}

	client, _ := gotenberg.NewClient(gotenbergURL, http.DefaultClient)

	e := echo.New()
	e.GET("/", showForm)
	e.POST("/generate-pdf", generatePDFHandler(client))

	log.Printf("Starting server at port %s...", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func showForm(c echo.Context) error {
	formHTML := `
    <!DOCTYPE html>
    <html>
    <head>
        <title>Generate PDF from HTML</title>
        <script src="https://cdn.ckeditor.com/ckeditor5/37.0.1/classic/ckeditor.js"></script>
        <style>
            body {
                font-family: Arial, sans-serif;
                background-color: #f4f4f9;
                padding: 20px;
                text-align: center;
            }
            h1 {
                color: #007BFF;
            }
            textarea {
                width: 80%;
                height: 300px;
                margin-bottom: 20px;
            }
            button {
                padding: 10px 20px;
                background-color: #28a745;
                color: white;
                border: none;
                border-radius: 5px;
                cursor: pointer;
            }
            button:hover {
                background-color: #218838;
            }
        </style>
    </head>
    <body>
        <h1>Generate PDF from HTML</h1>
        <form method="POST" action="/generate-pdf">
            <label for="htmlContent">Enter HTML Content:</label><br><br>
            <textarea id="htmlContent" name="htmlContent"></textarea><br><br>
            <button type="submit">Generate PDF</button>
        </form>
        <script>
            ClassicEditor
                .create(document.querySelector('#htmlContent'))
                .catch(error => {
                    console.error(error);
                });
        </script>
    </body>
    </html>
    `
	return c.HTML(http.StatusOK, formHTML)
}

func generatePDFHandler(client *gotenberg.Client) echo.HandlerFunc {
	return func(c echo.Context) error {
		htmlContent := c.FormValue("htmlContent")

		pdfBytes, err := generatePDF(client, htmlContent)
		if err != nil {
			return fmt.Errorf("Gagal mengonversi HTML ke PDF: %v", err)
		}

		c.Response().Header().Set("Content-Type", "application/pdf")
		c.Response().Header().Set("Content-Disposition", "attachment; filename=output.pdf")
		c.Response().Write(pdfBytes)

		return nil
	}
}

func generatePDF(client *gotenberg.Client, htmlContent string) ([]byte, error) {
	doc, _ := document.FromString("doc.html", htmlContent)

	req := gotenberg.NewHTMLRequest(doc)

	gotenbergIsAuth := os.Getenv("GOTENBERG_IS_AUTH") == "true"
	gotenbergUsername := os.Getenv("GOTENBERG_USERNAMES")
	gotenbergPassword := os.Getenv("GOTENBERG_PASSWORD")

	if gotenbergIsAuth {
		// Setting up basic auth (if needed).
		req.UseBasicAuth(gotenbergUsername, gotenbergPassword)
	}

	req.PaperSize(gotenberg.A4)
	req.Scale(0.75)
	req.SkipNetworkIdleEvent(true)

	resp, err := client.Send(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengirim permintaan ke Gotenberg: %w", err)
	}
	defer resp.Body.Close()

	pdfBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response body: %w", err)
	}

	return pdfBytes, nil
}
