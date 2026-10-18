package main

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
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

// findValuesFiles returns the names of the values-*.yaml files in the chart directory dir; none if dir doesn't exist.
func findValuesFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if name := e.Name(); !e.IsDir() && strings.HasPrefix(name, "values-") && strings.HasSuffix(name, ".yaml") {
			files = append(files, name)
		}
	}
	return files, nil
}

// chartValuesFiles returns the values-*.yaml files of chart on either checkout, or a single "" for the
// chart's defaults if it has none.
func chartValuesFiles(base, head, chart string) ([]string, error) {
	baseValues, err := findValuesFiles(filepath.Join(base, chart))
	if err != nil {
		return nil, err
	}
	headValues, err := findValuesFiles(filepath.Join(head, chart))
	if err != nil {
		return nil, err
	}
	if files := merge(baseValues, headValues); len(files) > 0 {
		return files, nil
	}
	return []string{""}, nil
}

// merge returns the sorted union of a and b, without duplicates.
func merge(a, b []string) []string {
	merged := slices.Concat(a, b)
	slices.Sort(merged)
	return slices.Compact(merged)
}
