package utiles

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func CloseFiles(files []*os.File) (err error) {
	errs := make([]error, 0)
	for _, f := range files {
		if err = f.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return errors.Join(errs...)
	}
	return nil
}

func OpenFiles(pathes ...string) (files []*os.File, err error) {
	files = make([]*os.File, len(pathes))
	for i, path := range pathes {
		if files[i], err = os.Open(path); err != nil {
			if err2 := CloseFiles(files[:i]); err2 != nil {
				err = fmt.Errorf(
					"ошибка открытия файла: %w, ошибка при экстренном закрытии файлов %w",
					err, err2,
				)
			}
			return nil, err
		}
	}
	return files, nil
}

func GetDirFiles(dirpath string) (pathes []string, err error) {
	if info, err := os.Stat(dirpath); err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, fmt.Errorf("path must be a dir")
	}

	pathes = make([]string, 0)
	err = filepath.WalkDir(
		dirpath,
		func(path string, dir fs.DirEntry, err error) error {
			if !dir.IsDir() {
				pathes = append(pathes, path)
			}
			return err
		},
	)
	if err != nil {
		return nil, err
	}
	return pathes, nil
}
