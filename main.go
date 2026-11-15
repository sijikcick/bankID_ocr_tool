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
func preprocess(path string) (string, error) {
    // AutoOrientation fixes phone photos that are rotated via EXIF
    img, err := imaging.Open(path, imaging.AutoOrientation(true))
    if err != nil {
        return "", fmt.Errorf("decode image: %w", err)
    }

    // Tesseract works best when text is reasonably large
    if img.Bounds().Dx() < 1800 {
        img = imaging.Resize(img, 1800, 0, imaging.Lanczos)
    }

    img = imaging.Grayscale(img)
    img = imaging.AdjustContrast(img, 30)
    img = imaging.Sharpen(img, 1.0)

    tmp, err := os.CreateTemp("", "ocr-*.png")
    if err != nil {
        return "", err
    }
    tmp.Close()

    if err := imaging.Save(img, tmp.Name()); err != nil {
        os.Remove(tmp.Name())
        return "", fmt.Errorf("save processed image: %w", err)
    }
    return tmp.Name(), nil
}
