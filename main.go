package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/starwalkn/gotenberg-go-client/v8"
	"github.com/starwalkn/gotenberg-go-client/v8/document"
)

func main() {
	client, _ := gotenberg.NewClient("http://localhost:3001", http.DefaultClient)

	e := echo.New()
	e.GET("/", showForm)
	e.POST("/generate-pdf", generatePDFHandler(client))

	e.Logger.Fatal(e.Start(":8080"))
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

	req.PaperSize(gotenberg.A4)
	req.Scale(0.75)
	req.Margins(gotenberg.PageMargins{
		Top:    20,
		Right:  20,
		Bottom: 20,
		Left:   20,
	})
	req.SkipNetworkIdleEvent(true)

	resp, err := client.Send(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("gagal mengirim permintaan ke Gotenberg: %w", err)
	}
	defer resp.Body.Close()

	pdfBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca response body: %w", err)
	}

	return pdfBytes, nil
}
