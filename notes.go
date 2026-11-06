package main

import (
	"embed"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"math/rand/v2"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed violin_and_bass_notes/*/*.png
var noteAssets embed.FS

type difficulty struct {
	name       string
	rangeLabel string
	minPitch   int
	maxPitch   int
	allNotes   bool
}

// Natural pitches use C D E F G A B within each octave.
var difficulties = []difficulty{
	{name: "Easy", rangeLabel: "F3-G4", minPitch: 3*7 + 3, maxPitch: 4*7 + 4},
	{name: "Medium", rangeLabel: "C3-C5", minPitch: 3 * 7, maxPitch: 5 * 7},
	{name: "Hard", rangeLabel: "F2-G5", minPitch: 2*7 + 3, maxPitch: 5*7 + 4},
	{name: "Ultra", rangeLabel: "C2-C6", minPitch: 2 * 7, maxPitch: 6 * 7},
	{name: "Extreme", rangeLabel: "All notes", allNotes: true},
}

func parsePitch(value string) (int, error) {
	if len(value) < 2 {
		return 0, fmt.Errorf("invalid natural note %q", value)
	}
	step := strings.IndexByte("CDEFGAB", value[0])
	octave, err := strconv.Atoi(value[1:])
	if step < 0 || err != nil || octave < 0 || octave > 9 {
		return 0, fmt.Errorf("invalid natural note %q", value)
	}
	return octave*7 + step, nil
}

func notePaths(files fs.FS, root string, level difficulty) ([]string, error) {
	var paths []string
	err := fs.WalkDir(files, root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.EqualFold(filepath.Ext(path), ".png") && !strings.EqualFold(entry.Name(), "overview.png") {
			stem := strings.TrimSuffix(entry.Name(), filepath.Ext(path))
			pitch, err := parsePitch(stem[strings.LastIndexByte(stem, '_')+1:])
			if err != nil {
				return fmt.Errorf("parse note filename %q: %w", path, err)
			}
			if level.allNotes || (pitch >= level.minPitch && pitch <= level.maxPitch) {
				paths = append(paths, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan note images in %q: %w", root, err)
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no note PNG images found in %q for %s (%s)", root, level.name, level.rangeLabel)
	}
	return paths, nil
}

func loadRandomNote(files fs.FS, root string, level difficulty, previousName string) (image.Image, string, error) {
	paths, err := notePaths(files, root, level)
	if err != nil {
		return nil, "", err
	}
	if previousName != "" {
		previousPath := filepath.ToSlash(filepath.Join(root, previousName))
		for i, path := range paths {
			if path == previousPath {
				paths = append(paths[:i], paths[i+1:]...)
				break
			}
		}
	}
	if len(paths) == 0 {
		return nil, "", fmt.Errorf("no alternative note PNG images found in %q for %s (%s)", root, level.name, level.rangeLabel)
	}
	path := paths[rand.IntN(len(paths))]
	return loadNote(files, root, path)
}

func loadNote(files fs.FS, root, path string) (image.Image, string, error) {
	file, err := files.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open note image %q: %w", path, err)
	}
	defer file.Close()
	img, err := png.Decode(file)
	if err != nil {
		return nil, "", fmt.Errorf("decode note image %q: %w", path, err)
	}
	name, err := filepath.Rel(root, path)
	if err != nil {
		return nil, "", fmt.Errorf("get note image name %q: %w", path, err)
	}
	return img, name, nil
}
