package agent

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeFFmpegShim membuat executable palsu bernama "ffmpeg" di sebuah direktori
// sementara. Shim ini mencatat argumen ke <dir>/args.txt dan membuat file
// output kosong pada argumen terakhir (posisi output ffmpeg). Dipakai untuk
// menguji BuildSlideshowVideo secara deterministik tanpa ffmpeg asli.
func fakeFFmpegShim(t *testing.T) (binDir string) {
	t.Helper()
	binDir = t.TempDir()
	shim := filepath.Join(binDir, "ffmpeg")
	script := `#!/bin/sh
echo "$@" >> "$(dirname "$0")/args.txt"
# output ffmpeg adalah argumen terakhir yang tidak berawalan "-"
for arg in "$@"; do
  case "$arg" in
    -*) ;;
    *) out="$arg" ;;
  esac
done
: > "$out"
`
	if err := os.WriteFile(shim, []byte(script), 0o755); err != nil {
		t.Fatalf("write ffmpeg shim: %v", err)
	}
	// Setelah shim dibuat, t.Setenv menempatkannya di depan PATH sehingga
	// exec.LookPath("ffmpeg") menemukannya.
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return binDir
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// PNG 1x1 piksel — ffmpeg shim tidak membaca isinya.
	data := []byte{
		0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 'I', 'D', 'A', 'T',
		0x08, 0xd7, 0x63, 0x78, 0x00, 0x00, 0x00, 0x02,
		0x00, 0x1f, 0x16, 0x9c, 0x5c, 0xca, 0x00, 0x00,
		0x00, 0x00, 'I', 'E', 'N', 'D', 0xae, 0x42, 0x60,
		0x82,
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write png: %v", err)
	}
}

func TestBuildSlideshowVideo_NoScreenshots(t *testing.T) {
	url, dur := BuildSlideshowVideo("run-1", "/tmp/screens", "/tmp/vids", nil)
	if url != "" || dur != 0 {
		t.Fatalf("expected empty result for no screenshots, got %q %v", url, dur)
	}
}

func TestBuildSlideshowVideo_MissingFiles(t *testing.T) {
	// Tidak ada file di disk → harus graceful (empty), tidak error.
	fakeFFmpegShim(t)
	url, dur := BuildSlideshowVideo("run-1", t.TempDir(), t.TempDir(), []string{"/screenshots/step_a.png"})
	if url != "" || dur != 0 {
		t.Fatalf("expected empty result for missing files, got %q %v", url, dur)
	}
}

func TestBuildSlideshowVideo_Success(t *testing.T) {
	binDir := fakeFFmpegShim(t)
	shotDir := t.TempDir()
	vidDir := t.TempDir()

	// Dua screenshot dengan urutan nama terbalik untuk memastikan sortir.
	writePNG(t, filepath.Join(shotDir, "step_b_2.png"))
	writePNG(t, filepath.Join(shotDir, "step_a_1.png"))

	url, dur := BuildSlideshowVideo("run-abc", shotDir, vidDir, []string{
		"/screenshots/step_b_2.png",
		"/screenshots/step_a_1.png",
	})
	if url != "/videos/run-abc/recording.mp4" {
		t.Fatalf("unexpected video url: %q", url)
	}
	if dur != 2 {
		t.Fatalf("expected duration 2 (2 frames at 1fps), got %v", dur)
	}
	out := filepath.Join(vidDir, "run-abc", "recording.mp4")
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
	// Argumen ffmpeg harus memakai urutan frame yang benar dan pola %04d.
	args, err := os.ReadFile(filepath.Join(binDir, "args.txt"))
	if err != nil {
		t.Fatalf("read shim args: %v", err)
	}
	got := string(args)
	for _, want := range []string{"-framerate", "1", "%04d.png", "libx264", "-pix_fmt", "yuv420p"} {
		if !contains(got, want) {
			t.Errorf("ffmpeg args missing %q: %s", want, got)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
