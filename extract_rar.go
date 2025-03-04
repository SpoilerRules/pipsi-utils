package main

import (
	"fmt"
	"github.com/nwaples/rardecode"
	"io"
	"os"
	"path/filepath"
)

func extractRar(source, destination string) error {
	rr, err := rardecode.OpenReader(source, "")
	if err != nil {
		return fmt.Errorf("failed to open RAR file: %w", err)
	}
	defer func() {
		if closeErr := rr.Close(); closeErr != nil {
			fmt.Printf("Error closing RAR reader: %v\n", closeErr)
		}
	}()

	for {
		header, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read header: %w", err)
		}

		destPath := filepath.Join(destination, header.Name)
		if header.IsDir {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		out, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		if err := out.Chmod(header.Mode()); err != nil {
			closeErr := out.Close()
			if closeErr != nil {
				return fmt.Errorf("failed to set file permissions and close file: %w", closeErr)
			}
			return fmt.Errorf("failed to set file permissions: %w", err)
		}

		if _, err := io.Copy(out, rr); err != nil {
			closeErr := out.Close()
			if closeErr != nil {
				return fmt.Errorf("failed to write file and close file: %w", closeErr)
			}
			return fmt.Errorf("failed to write file: %w", err)
		}

		if closeErr := out.Close(); closeErr != nil {
			return fmt.Errorf("error closing file %s: %w", destPath, closeErr)
		}
	}

	return nil
}
