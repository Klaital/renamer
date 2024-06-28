package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func leftPad(a string, pad string, desiredLen int) string {
	if len(a) >= desiredLen {
		return a
	}

	ret := strings.Builder{}
	for i := 0; i < desiredLen-len(a); i++ {
		ret.WriteString(pad)
	}
	ret.WriteString(a)
	return ret.String()
}

func extractFirstNumber(p string) string {
	re := regexp.MustCompile("[0-9]+")
	return re.FindString(p)
}
func extractFileExtension(p string) string {
	re := regexp.MustCompile("\\..+$")
	extension := re.FindString(p)
	return strings.Trim(extension, "\n")
}

func stringInSet(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}

	return false
}

func lastString(a []string) string {
	if len(a) == 0 {
		return ""
	}
	return a[len(a)-1]
}

// buildRenameSet looks into a directory and calculates how to rename any media files found.
func buildRenameSet(dir string, allowedExtensions []string, titleOverride string, season string) (map[string]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading media files: %w", err)
	}
	// Extract the series title from the directory name if not given
	if len(titleOverride) == 0 {
		d, err := filepath.Abs(dir)
		if err != nil {
			return nil, fmt.Errorf("calculating absolute path: %w", err)
		}
		_, filename := filepath.Split(d)
		fmt.Printf("Series name guess: '%s'\n", filename)
		titleOverride = filename
	}

	renameSet := make(map[string]string, len(entries))
	for _, f := range entries {
		epNum := extractFirstNumber(f.Name())
		epNum = leftPad(epNum, "0", 2)
		ext := extractFileExtension(f.Name())
		if !stringInSet(allowedExtensions, ext) {
			continue
		}
		newName := fmt.Sprintf("%s S%sE%s%s", titleOverride, season, epNum, ext)
		if oldName, ok := renameSet[newName]; ok {
			return nil, fmt.Errorf("filename collision: both '%s' and '%s' map to new filename '%s'", oldName, f.Name(), newName)
		}
		if _, err := os.Stat(filepath.Join(dir, newName)); !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("file already exists: '%s'", newName)
		}
		renameSet[newName] = f.Name()
	}

	return renameSet, nil
}

func main() {
	var title string
	var season string
	extensions := []string{
		".avi", ".mp4", ".m4v", ".mkv", ".asf",
	}
	var directory string
	var actuallyRenameFiles bool
	flag.StringVar(&title, "title", "", "Name to use for the series")
	flag.StringVar(&season, "season", "01", "What season. Manually add zero padding")
	flag.StringVar(&directory, "d", ".", "The directory to rename files in")
	flag.BoolVar(&actuallyRenameFiles, "y", false, "Actually execute the rename")
	flag.Parse()

	renames, err := buildRenameSet(directory, extensions, title, season)
	if err != nil {
		fmt.Printf("Failed to build rename set: %s\n", err.Error())
		os.Exit(1)
	}
	for newName, oldName := range renames {
		fmt.Printf("%s ->\t%s\n", oldName, newName)
		if actuallyRenameFiles {
			err = os.Rename(filepath.Join(directory, oldName), filepath.Join(directory, newName))
			if err != nil {
				fmt.Printf("Failed to rename file %s: %+v\n", oldName, err)
				os.Exit(1)
			}
		}
	}
}
