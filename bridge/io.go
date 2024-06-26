package bridge

import (
	"archive/zip"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	Binary = "Binary"
	Text   = "Text"
)

func (a *App) Writefile(path string, content string, options IOOptions) FlagResult {
	log.Printf("Writefile [%s]: %s", options.Mode, path)

	path = GetPath(path)

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	b := []byte{}

	switch options.Mode {
	case Text:
		b = []byte(content)
	case Binary:
		b, err = base64.StdEncoding.DecodeString(content)
		if err != nil {
			return FlagResult{false, err.Error()}
		}
	}

	err = os.WriteFile(path, b, 0644)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) Readfile(path string, options IOOptions) FlagResult {
	log.Printf("Readfile [%s]: %s", options.Mode, path)

	path = GetPath(path)

	b, err := os.ReadFile(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	content := ""
	switch options.Mode {
	case Text:
		content = string(b)
	case Binary:
		content = base64.StdEncoding.EncodeToString(b)
	}

	return FlagResult{true, content}
}

func (a *App) Movefile(source string, target string) FlagResult {
	log.Printf("Movefile: %s -> %s", source, target)

	source = GetPath(source)
	target = GetPath(target)

	err := os.Rename(source, target)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) Removefile(path string) FlagResult {
	log.Printf("RemoveFile: %s", path)

	path = GetPath(path)

	err := os.RemoveAll(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) Copyfile(src string, dst string) FlagResult {
	log.Printf("Copyfile: %s -> %s", src, dst)

	src = GetPath(src)
	dst = GetPath(dst)

	srcFile, err := os.Open(src)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) Makedir(path string) FlagResult {
	log.Printf("Makedir: %s", path)

	path = GetPath(path)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	return FlagResult{true, "Success"}
}

func (a *App) Readdir(path string) FlagResult {
	log.Printf("Readdir: %s", path)

	path = GetPath(path)

	files, err := os.ReadDir(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	var result []string

	for _, file := range files {
		info, err := file.Info()
		if err == nil {
			result = append(result, fmt.Sprintf("%v,%v,%v", info.Name(), info.Size(), info.IsDir()))
		}
	}

	return FlagResult{true, strings.Join(result, "|")}
}

func (a *App) AbsolutePath(path string) FlagResult {
	log.Printf("AbsolutePath: %s", path)

	path = GetPath(path)

	return FlagResult{true, path}
}

func (a *App) UnzipZIPFile(path string, output string) FlagResult {
	log.Printf("UnzipZIPFile: %s -> %s", path, output)

	path = GetPath(path)
	output = GetPath(output)

	archive, err := zip.OpenReader(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(output, f.Name)

		if !strings.HasPrefix(filePath, filepath.Clean(output)+string(os.PathSeparator)) {
			log.Println("UnzipZIPFile: invalid file path")
			return FlagResult{false, "invalid file path"}
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return FlagResult{false, err.Error()}
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return FlagResult{false, err.Error()}
		}
		defer dstFile.Close()

		fileInArchive, err := f.Open()
		if err != nil {
			return FlagResult{false, err.Error()}
		}
		defer fileInArchive.Close()

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			return FlagResult{false, err.Error()}
		}
	}
	return FlagResult{true, "Success"}
}

func (a *App) UnzipGZFile(path string, output string) FlagResult {
	log.Printf("UnzipGZFile: %s -> %s", path, output)

	path = GetPath(path)
	output = GetPath(output)

	gzipFile, err := os.Open(path)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer gzipFile.Close()

	outputFile, err := os.Create(output)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer outputFile.Close()

	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		return FlagResult{false, err.Error()}
	}
	defer gzipReader.Close()

	_, err = io.Copy(outputFile, gzipReader)
	if err != nil {
		return FlagResult{false, err.Error()}
	}

	return FlagResult{true, "Success"}
}

func (a *App) FileExists(path string) FlagResult {
	log.Printf("FileExists: %s", path)

	path = GetPath(path)

	_, err := os.Stat(path)
	if err == nil {
		return FlagResult{true, "true"}
	} else if os.IsNotExist(err) {
		return FlagResult{true, "false"}
	}
	return FlagResult{false, err.Error()}
}
