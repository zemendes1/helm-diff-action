package main

import (
	"errors"
	"io/fs"
	"path/filepath"
)

func findCharts(root, dir string) ([]string, error) {
	var charts []string

	start := filepath.Join(root, dir)
	err := filepath.WalkDir(start, func(path string, d fs.DirEntry, err error) error {
		if path == start && errors.Is(err, fs.ErrNotExist) {
			return filepath.SkipAll // dir isn't in this checkout: no charts
		}
		if err != nil {
			return err // couldn't read something; stop the walk
		}
		if d.IsDir() && exists(filepath.Join(path, "Chart.yaml")) {
			relativePath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			charts = append(charts, relativePath)
			return filepath.SkipDir // don't look inside the directory
		}
		return nil
	})
	return charts, err
}
