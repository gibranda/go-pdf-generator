package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/starwalkn/gotenberg-go-client/v8"
	"github.com/starwalkn/gotenberg-go-client/v8/document"
)

func main() {
	// URL server Gotenberg (sesuaikan dengan konfigurasi Anda)
	client, _ := gotenberg.NewClient("http://localhost:3001", http.DefaultClient)

	// Kode HTML yang akan dikonversi
	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Sample PDF</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				text-align: center;
				margin: 20px;
			}
			h1 {
				color: blue;
			}
		</style>
	</head>
	<body>
		<h1>Hello, Gotenberg v8!</h1>
		<p>This is a simple example of HTML to PDF conversion using Gotenberg v8.</p>
	</body>
	</html>`

	// Nama file keluaran
	outputFilename := "output.pdf"

	// Konversi HTML ke PDF
	if err := generatePDF(client, outputFilename, htmlContent); err != nil {
		log.Fatalf("Gagal mengonversi HTML ke PDF: %v", err)
	}

	fmt.Printf("PDF berhasil dibuat: %s\n", outputFilename)
}

// generatePDF mengonversi HTML ke PDF menggunakan Gotenberg
func generatePDF(client *gotenberg.Client, outputFilename, htmlContent string) error {
	doc, _ := document.FromString("doc.html", htmlContent)

	// Buat permintaan HTML ke PDF
	req := gotenberg.NewHTMLRequest(doc)

	// (Opsional) Tambahkan parameter tambahan
	req.PaperSize(gotenberg.A4)          // Ukuran kertas A4
	req.Margins(gotenberg.PageMargins{}) // Margin atas, kanan, bawah, kiri dalam milimeter

	// Skips the IDLE events for faster PDF conversion.
	req.SkipNetworkIdleEvent(true)

	// Buffer untuk menampung PDF yang dihasilkan
	var buf bytes.Buffer

	// Kirim permintaan ke server Gotenberg
	if err := client.Store(context.Background(), req, "doc.pdf"); err != nil {
		return fmt.Errorf("gagal mengirim permintaan ke Gotenberg: %w", err)
	}

	// Simpan hasil PDF ke file
	if err := saveToFile(outputFilename, buf.Bytes()); err != nil {
		return fmt.Errorf("gagal menyimpan PDF: %w", err)
	}

	return nil
}

// saveToFile menyimpan data byte ke dalam file
func saveToFile(filename string, data []byte) error {
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("gagal menulis file %s: %w", filename, err)
	}
	return nil
}
