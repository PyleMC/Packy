package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func zipPackFromFolder(folderPath string, outputPath string) (string, error) {
	info, err := os.Stat(folderPath)
	if err != nil {
		return "", fmt.Errorf("folder path error: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("folder path is not a directory")
	}

	if strings.TrimSpace(outputPath) == "" {
		base := filepath.Base(folderPath)
		exePath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("resolve executable path: %w", err)
		}
		packsDir := filepath.Join(filepath.Dir(exePath), "packs")
		if err := os.MkdirAll(packsDir, 0o755); err != nil {
			return "", fmt.Errorf("create packs folder: %w", err)
		}
		outputPath = filepath.Join(packsDir, base+".zip")
	}

	absFolder, err := filepath.Abs(folderPath)
	if err != nil {
		return "", fmt.Errorf("resolve folder path: %w", err)
	}
	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}
	folderWithSep := absFolder + string(os.PathSeparator)
	if strings.HasPrefix(absOutput, folderWithSep) {
		return "", fmt.Errorf("output zip cannot be inside the source folder")
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("create zip: %w", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	err = filepath.WalkDir(folderPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if absPath == absOutput {
			return nil
		}

		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		relPath = filepath.ToSlash(relPath)

		if entry.IsDir() {
			_, err := zipWriter.Create(relPath + "/")
			return err
		}

		fileInfo, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		inFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer inFile.Close()

		_, err = io.Copy(writer, inFile)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("zip folder: %w", err)
	}

	return outputPath, nil
}
