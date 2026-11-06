# Notomat

A small desktop app for practicing musical notes, built with Go and Ebitengine.
All note images and the UI font are bundled into the executable.

## Download and run

Download `notomat-windows-amd64.zip` from this repository's GitHub Releases,
extract it, and double-click `notomat.exe`. No Go installation or separate
image folder is needed. The executable can run from any directory.
The ZIP also includes third-party license notices and this guide.

Release executables are currently unsigned; Windows may show a publisher warning.

## Development

With Go 1.26.5 or newer installed, run from the project directory:

```powershell
go run .
```

Note PNGs from `violin_and_bass_notes` are embedded at build time, including
`bass_clef_A1-B3` and `violin_clef_C4-E6`, but excluding `overview.png`.
Rebuild after changing these assets. Images are centered and scaled without
changing their aspect ratio; filenames and debug overlays are not displayed.

The app starts with difficulty buttons. Click one to begin displaying notes:

| Difficulty | Inclusive range |
| --- | --- |
| Easy | F3-G4 |
| Medium | C3-C5 |
| Hard | F2-G5 |
| Ultra | C2-C6 |
| Extreme | All notes |

Pitches follow C, D, E, F, G, A, B; the octave changes after B (B4 then C5).
Both clef collections are sampled within the selected range.

**Endless Mode** is checked by default on the difficulty screen. Leave it
checked to keep sampling indefinitely. Uncheck it to display every note in
the chosen range exactly once, in a shuffled order. After viewing the last
note, click **New image** or press **Space** once more to see **good job** and
a **Back to difficulty selection** button. Starting another session resets
the shuffled collection; the checkbox setting is kept when going back.

Click **New image** or press **Space** to sample another image. The current
image is excluded from the random selection, so notes never repeat consecutively.
The chosen difficulty remains active for each sample.
Holding Space samples only once per press.
Click **Back** to return to difficulty selection and choose a different range.
Close the window or press Escape to exit.



## Structure

- `main.go`: window setup, layout, and application loop.
- `notes.go`: embedded assets, difficulty filtering, and note loading.
- `text.go`: normal UI text rendering with a bundled font.
- `scripts\build-release.ps1`: Windows packaging and dependency licenses.
- `.github\workflows\build.yml`: CI and tag-triggered releases.
