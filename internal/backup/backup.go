package backup

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Backup creates a zip archive of the source directory and saves it to the backup directory
func Backup(sourceDir, backupDir string) (string, error) {
	sourceInfo, err := os.Stat(sourceDir)
	if err != nil {
		return "", fmt.Errorf("source directory not found: %w", err)
	}
	if !sourceInfo.IsDir() {
		return "", fmt.Errorf("source path is not a directory: %s", sourceDir)
	}

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupFile := filepath.Join(backupDir, fmt.Sprintf("backup_%s.zip", timestamp))

	zipFile, err := os.Create(backupFile)
	if err != nil {
		return "", fmt.Errorf("failed to create backup file: %w", err)
	}

	zipWriter := zip.NewWriter(zipFile)

	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filePath == backupFile {
			return nil
		}
		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", filePath, err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", filePath, err)
		}

		header.Name = filepath.ToSlash(relPath)
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create zip entry for %s: %w", relPath, err)
		}

		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %w", filePath, err)
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("failed to write file %s to zip: %w", filePath, err)
		}

		return nil
	})

	if err != nil {
		zipWriter.Close()
		zipFile.Close()
		os.Remove(backupFile)
		return "", fmt.Errorf("error walking source directory: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		zipFile.Close()
		os.Remove(backupFile)
		return "", fmt.Errorf("failed to close zip writer: %w", err)
	}

	if err := zipFile.Close(); err != nil {
		os.Remove(backupFile)
		return "", fmt.Errorf("failed to close backup file: %w", err)
	}

	if info, err := os.Stat(backupFile); err != nil || info.Size() == 0 {
		if err == nil {
			os.Remove(backupFile)
		}
		return "", fmt.Errorf("backup file was not created successfully")
	}

	return backupFile, nil
}

// Restore extracts a backup zip file to the target directory
func Restore(backupFile, targetDir string) error {
	fileInfo, err := os.Stat(backupFile)
	if err != nil {
		return fmt.Errorf("backup file not found: %w", err)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("backup file is empty")
	}

	fmt.Printf("Attempting to restore backup: %s (Size: %d bytes)\n", backupFile, fileInfo.Size())

	file, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file for reading: %w", err)
	}
	defer file.Close()

	header := make([]byte, 4)
	if _, err := file.Read(header); err != nil {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	if string(header) != "PK\x03\x04" {
		return fmt.Errorf("not a valid zip file (invalid header: %x)", header)
	}

	reader, err := zip.OpenReader(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open zip reader: %w", err)
	}
	defer reader.Close()

	fmt.Printf("Found %d files in backup\n", len(reader.File))

	if len(reader.File) == 0 {
		return fmt.Errorf("backup file is empty or corrupted (no files found in archive)")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	for _, file := range reader.File {
		extractPath := filepath.Join(targetDir, file.Name)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(extractPath, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(extractPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}

		fileWriter, err := os.OpenFile(extractPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		fileReader, err := file.Open()
		if err != nil {
			fileWriter.Close()
			return fmt.Errorf("failed to open file in backup: %w", err)
		}

		if _, err := io.Copy(fileWriter, fileReader); err != nil {
			fileWriter.Close()
			fileReader.Close()
			return fmt.Errorf("failed to extract file: %w", err)
		}

		fileWriter.Close()
		fileReader.Close()
	}

	return nil
}

// BackupInfo contains information about a backup file
type BackupInfo struct {
	Path    string //
	Name    string
	Size    int64
	ModTime time.Time
}

// ListBackups returns a list of all backup files in the backup directory with their metadata
func ListBackups(backupDir string) ([]BackupInfo, error) {
	files, err := filepath.Glob(filepath.Join(backupDir, "backup_*.zip"))
	if err != nil {
		return nil, fmt.Errorf("failed to list backup files: %w", err)
	}

	var backups []BackupInfo
	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		backups = append(backups, BackupInfo{
			Path:    file,
			Name:    filepath.Base(file),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime.After(backups[j].ModTime)
	})

	return backups, nil
}

// GetFileChecksum calculates the SHA-256 checksum of a file
func GetFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
