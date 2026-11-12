package main 

import (
	"log"
	//"github.com/disintegration/imaging"
	//"github.com/spf13/cobra"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)
func main () {
	log.Println("application kenshi")

imagePath := flag.String("image", "", "Path to the input image (.jpg, .png)")
	flag.Parse()

	if err := validateImageFile(*imagePath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	fmt.Printf("Valid image ready for processing: %s\n", *imagePath)
	// Proceed with your image processing logic here...
}

func validateImageFile(path string) error {
	// Check that the flag was actually provided
	if strings.TrimSpace(path) == "" {
		return errors.New("missing required --image flag")
	}

	// 2. Validate supported file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		// Supported
	default:
		return fmt.Errorf("unsupported file extension %q: must be .jpg or .png", ext)
	}

	// 3. Check existence and ensure it's a regular file (not a directory)
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a file, got a directory: %s", path)
	}
	if info.Size() == 0 {
		return fmt.Errorf("image file is empty: %s", path)
	}

	// 4. Verify the file can be opened and read
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 5. Verify actual image content (magic bytes) to prevent spoofed extensions
	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	mimeType := http.DetectContentType(header[:n])
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		return fmt.Errorf("invalid image content (%s): file must be a valid JPEG or PNG", mimeType)
	}

	return nil

}
