package util

import (
	"fmt"
	copy2 "github.com/otiai10/copy"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

// Will delete all files/dirs in a given directory, except ignored files
func RemoveDirContents(dirname string, ignoreFiles []string) error {
	dir, err := ioutil.ReadDir(dirname)
	if err != nil {
		return fmt.Errorf("failed to read target repo dir: %s", err)
	}

	// 1. Remove everything in target except ignored files
	for _, file := range dir {
		shouldDelete := true
		for _, ignoreFile := range ignoreFiles {
			if file.Name() == ignoreFile {
				shouldDelete = false
			}
		}

		if shouldDelete {
			fmt.Printf("Deleting file in target repo: %s\n", file.Name())

			filePath := path.Join(dirname, file.Name())
			if file.IsDir() {
				err := ClearDir(filePath)
				if err != nil {
					return fmt.Errorf("failed to delete files in dir %s: %s", filePath, err)
				}
				err = os.Remove(path.Join(dirname, file.Name()))
				if err != nil {
					return fmt.Errorf("failed to delete dir %s: %s", file.Name(), err)
				}
			}

		}
	}

	return nil
}

// Moves contents from a directory, to another directory
func CopyDirContents(sourceDir string, targetDir string) error {
	// Read all files/dirs in source dir
	dir, err := ioutil.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to source: %s", err)
	}

	// For each file/dir source, copy to target dir
	for _, f := range dir {
		sourcePath := path.Join(sourceDir, f.Name())
		targetPath := path.Join(targetDir, f.Name())
		err := copy2.Copy(sourcePath, targetPath)
		if err != nil {
			return fmt.Errorf("failed to copy %s to %s: %s", sourcePath, targetPath, err)
		}
	}

	return nil
}

func ClearDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}
