package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRandomNoteNestedImage(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "bass_clef_A1-B3")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "bass_G3.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 114, 182))); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	// Invalid overview data must never be selected or decoded.
	if err := os.WriteFile(filepath.Join(root, "overview.png"), []byte("overview"), 0644); err != nil {
		t.Fatal(err)
	}

	img, name, err := loadRandomNote(os.DirFS(root), ".", difficulties[4], "")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join("bass_clef_A1-B3", "bass_G3.png"); name != want {
		t.Fatalf("name = %q, want %q", name, want)
	}
	if img.Bounds().Dx() != 114 || img.Bounds().Dy() != 182 {
		t.Fatalf("unexpected image dimensions: %v", img.Bounds())
	}
	if _, _, err := loadRandomNote(os.DirFS(root), ".", difficulties[4], name); err == nil || !strings.Contains(err.Error(), "no alternative note PNG images") {
		t.Fatalf("error = %v, want no alternative notes", err)
	}
}

func TestLoadRandomNoteErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		file string
		want string
	}{
		{name: "missing directory", want: "scan note images"},
		{name: "empty directory", want: "no note PNG images"},
		{name: "overview only", file: "overview.png", want: "no note PNG images"},
		{name: "invalid note", file: "bass_G3.png", want: "decode note image"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.name == "missing directory" {
				root = filepath.Join(root, "missing")
			}
			if tc.file != "" {
				if err := os.WriteFile(filepath.Join(root, tc.file), []byte("invalid PNG"), 0644); err != nil {
					t.Fatal(err)
				}
			}
			img, name, err := loadRandomNote(os.DirFS(root), ".", difficulties[4], "")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
			if img != nil || name != "" {
				t.Fatal("failed load returned an image or name")
			}
		})
	}
}

func TestParsePitchOctaveOrder(t *testing.T) {
	sequence := []string{"F3", "G3", "A3", "B3", "C4", "D4", "E4", "F4", "G4"}
	previous := -1
	for _, name := range sequence {
		pitch, err := parsePitch(name)
		if err != nil {
			t.Fatal(err)
		}
		if previous >= 0 && pitch != previous+1 {
			t.Fatalf("%s has pitch %d, want %d", name, pitch, previous+1)
		}
		previous = pitch
	}
	b4, _ := parsePitch("B4")
	c5, _ := parsePitch("C5")
	if c5 != b4+1 {
		t.Fatal("B4 must be followed by C5")
	}
	for _, name := range []string{"", "C", "H4", "C-1", "C10", "C#4"} {
		if _, err := parsePitch(name); err == nil {
			t.Fatalf("accepted invalid pitch %q", name)
		}
	}
}

func TestDifficultyDatasetRanges(t *testing.T) {
	for i, tc := range []struct {
		count int
		min   int
		max   int
	}{
		{count: 9, min: 41, max: 55},
		{count: 15, min: 36, max: 60},
		{count: 23, min: 29, max: 67},
		{count: 29, min: 24, max: 72},
		{count: 33, min: 21, max: 76},
	} {
		t.Run(difficulties[i].name, func(t *testing.T) {
			paths, err := notePaths(noteAssets, "violin_and_bass_notes", difficulties[i])
			if err != nil {
				t.Fatal(err)
			}
			if len(paths) != tc.count {
				t.Fatalf("got %d images, want %d", len(paths), tc.count)
			}
			seen := make(map[int]bool)
			for _, path := range paths {
				stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
				pitch := stem[strings.LastIndexByte(stem, '_')+1:]
				// Independent chromatic ordering verifies the natural-note filter.
				offsets := map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'B': 11}
				value := int(pitch[1]-'0')*12 + offsets[pitch[0]]
				if value < tc.min || value > tc.max {
					t.Fatalf("out-of-range image: %s", path)
				}
				seen[value] = true
			}
			if !seen[tc.min] || !seen[tc.max] {
				t.Fatal("inclusive range endpoints are missing")
			}
		})
	}
}

func TestDifficultyWithNoMatchingImages(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bass_A1.png"), []byte("not decoded"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadRandomNote(os.DirFS(root), ".", difficulties[0], ""); err == nil || !strings.Contains(err.Error(), "no note PNG images") {
		t.Fatalf("error = %v, want no matching notes", err)
	}
}

func TestGameRetainsSelectedDifficulty(t *testing.T) {
	for _, level := range difficulties {
		t.Run(level.name, func(t *testing.T) {
			g := newGame()
			if g.difficulty != nil || g.note.image != nil {
				t.Fatal("game should start on the difficulty screen")
			}
			if err := g.sampleNote(level); err != nil {
				t.Fatal(err)
			}
			defer func() { g.note.image.Deallocate() }()
			for i := 0; i < 25; i++ {
				previous := g.note.name
				if err := g.sampleNote(*g.difficulty); err != nil {
					t.Fatal(err)
				}
				if g.note.name == previous {
					t.Fatalf("repeated note %q", previous)
				}
			}
			if *g.difficulty != level {
				t.Fatal("resampling changed the selected difficulty")
			}
			paths, err := notePaths(noteAssets, "violin_and_bass_notes", level)
			if err != nil {
				t.Fatal(err)
			}
			for _, path := range paths {
				name, err := filepath.Rel("violin_and_bass_notes", path)
				if err != nil {
					t.Fatal(err)
				}
				if g.note.name == name {
					return
				}
			}
			t.Fatalf("displayed image %q is not eligible for %s", g.note.name, level.name)
		})
	}
}

func TestLoadRandomNoteExcludesPrevious(t *testing.T) {
	root := t.TempDir()
	names := []string{"bass_C3.png", "bass_D3.png"}
	for _, name := range names {
		file, err := os.Create(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
			file.Close()
			t.Fatal(err)
		}
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
	}
	for i, previous := range names {
		_, name, err := loadRandomNote(os.DirFS(root), ".", difficulties[1], previous)
		if err != nil {
			t.Fatal(err)
		}
		if want := names[1-i]; name != want {
			t.Fatalf("after %q got %q, want %q", previous, name, want)
		}
	}
	if _, _, err := loadRandomNote(os.DirFS(root), ".", difficulties[1], "bass_E3.png"); err != nil {
		t.Fatalf("unavailable previous note should not prevent sampling: %v", err)
	}
}

func TestGameBackToDifficultySelection(t *testing.T) {
	g := newGame()
	if err := g.sampleNote(difficulties[0]); err != nil {
		t.Fatal(err)
	}
	defer func() { g.showDifficultySelection() }()
	previous := g.note.name

	g.showDifficultySelection()
	if g.difficulty != nil || g.note.image != nil {
		t.Fatal("Back should return to difficulty selection and release the image")
	}
	if err := g.sampleNote(difficulties[1]); err != nil {
		t.Fatal(err)
	}
	if g.difficulty == nil || *g.difficulty != difficulties[1] || g.note.image == nil {
		t.Fatal("selecting a new difficulty should resume displaying notes")
	}
	if g.note.name == previous {
		t.Fatal("changing difficulty should not repeat the last displayed note")
	}
}

func TestFiniteSessionShowsEveryNoteOnce(t *testing.T) {
	for _, level := range difficulties {
		t.Run(level.name, func(t *testing.T) {
			g := newGame()
			g.endless = false
			defer g.showDifficultySelection()
			paths, err := notePaths(noteAssets, "violin_and_bass_notes", level)
			if err != nil {
				t.Fatal(err)
			}
			expected := make(map[string]bool)
			for _, path := range paths {
				name, err := filepath.Rel("violin_and_bass_notes", path)
				if err != nil {
					t.Fatal(err)
				}
				expected[name] = true
			}
			if err := g.startSession(level); err != nil {
				t.Fatal(err)
			}
			seen := make(map[string]bool)
			for i := 0; i < len(paths); i++ {
				if g.completed || g.note.image == nil {
					t.Fatalf("session completed before note %d was viewed", i+1)
				}
				if !expected[g.note.name] || seen[g.note.name] {
					t.Fatalf("unexpected or repeated note %q", g.note.name)
				}
				seen[g.note.name] = true
				if len(g.remaining) != len(paths)-i-1 {
					t.Fatal("remaining collection did not advance exactly once")
				}
				if i < len(paths)-1 {
					if err := g.sampleNote(level); err != nil {
						t.Fatal(err)
					}
				}
			}
			last := g.note.name
			if err := g.sampleNote(level); err != nil {
				t.Fatal(err)
			}
			if !g.completed || g.note.image != nil || g.difficulty == nil {
				t.Fatal("advancing past the final note should show completion")
			}

			g.showDifficultySelection()
			if g.completed || g.remaining != nil || g.difficulty != nil || g.endless {
				t.Fatal("Back should reset the session but preserve finite mode")
			}
			if err := g.startSession(level); err != nil {
				t.Fatal(err)
			}
			if g.completed || len(g.remaining) != len(paths)-1 || g.note.image == nil {
				t.Fatal("new session did not reset the collection")
			}
			if g.note.name == last {
				t.Fatal("new session repeated the last displayed note")
			}
		})
	}
}

func TestFiniteSessionLoadFailureDoesNotConsumeNote(t *testing.T) {
	g := newGame()
	g.endless = false
	if err := g.startSession(difficulties[0]); err != nil {
		t.Fatal(err)
	}
	defer g.showDifficultySelection()
	previous := g.note
	missing := filepath.Join(t.TempDir(), "missing.png")
	g.remaining = []string{missing}
	if err := g.sampleNote(difficulties[0]); err == nil {
		t.Fatal("expected image-loading error")
	}
	if g.note != previous || g.completed || len(g.remaining) != 1 || g.remaining[0] != missing {
		t.Fatal("failed load changed the displayed note or consumed the pending note")
	}
}

func TestEmbeddedAssetsWorkOutsideProjectDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	g := newGame()
	defer g.showDifficultySelection()
	if err := g.startSession(difficulties[0]); err != nil {
		t.Fatalf("load bundled image outside project: %v", err)
	}
	g.showDifficultySelection()
	g.endless = false
	if err := g.startSession(difficulties[1]); err != nil {
		t.Fatalf("start finite session outside project: %v", err)
	}
	if err := g.sampleNote(difficulties[1]); err != nil {
		t.Fatalf("load next bundled image outside project: %v", err)
	}
}
