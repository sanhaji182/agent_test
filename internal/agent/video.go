package agent

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// videoCodecs adalah daftar codec yang dicoba berurutan untuk membangun video
// slideshow. libx264 (mp4) punya kompatibilitas browser terbaik; libvpx-vp9
// (webm) dipakai sebagai fallback bila ffmpeg di image tidak punya libx264.
var videoCodecs = []struct {
	codec string
	ext   string
}{
	{"libx264", "mp4"},
	{"libvpx-vp9", "webm"},
	{"mpeg4", "mp4"},
}

// BuildSlideshowVideo meng-compile daftar screenshot per-step menjadi satu
// video slideshow memakai ffmpeg (bila tersedia). Screenshot dirujuk sebagai
// URL publik seperti "/screenshots/step_xxx.png"; screenshotDir adalah root
// file di disk dan videoDir adalah root tempat video disimpan.
//
// Mengembalikan URL publik video yang dihasilkan (misal
// "/videos/<runID>/recording.mp4") dan durasi perkiraan dalam detik. Ketika
// ffmpeg tidak tersedia, screenshot tidak ada di disk, atau encoding gagal,
// fungsi ini mengembalikan ("", 0) tanpa error — video slideshow bersifat
// best-effort dan tidak boleh menggagalkan run.
func BuildSlideshowVideo(runID, screenshotDir, videoDir string, screenshotURLs []string) (videoURL string, duration float64) {
	if len(screenshotURLs) == 0 {
		return "", 0
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		slog.Warn("ffmpeg not available; skipping slideshow video", "run_id", runID, "error", err)
		return "", 0
	}

	// Kumpulkan file screenshot yang benar-benar ada di disk.
	var files []string
	for _, u := range screenshotURLs {
		rel := strings.TrimPrefix(strings.TrimPrefix(u, "/screenshots/"), "/")
		if rel == "" || strings.Contains(rel, "..") {
			continue
		}
		p := filepath.Join(screenshotDir, rel)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			files = append(files, p)
		}
	}
	if len(files) == 0 {
		slog.Warn("no screenshot files on disk; skipping slideshow video", "run_id", runID)
		return "", 0
	}
	sort.Strings(files)

	// Salin ke direktori sementara dengan nama zero-padded berurutan supaya
	// pola glob ffmpeg (%04d) menghasilkan urutan frame yang benar.
	tmpDir, err := os.MkdirTemp("", "slideshow-*")
	if err != nil {
		slog.Warn("slideshow temp dir failed", "run_id", runID, "error", err)
		return "", 0
	}
	defer os.RemoveAll(tmpDir)

	frames := 0
	for i, f := range files {
		name := fmt.Sprintf("%04d.png", i+1)
		if err := copyFile(f, filepath.Join(tmpDir, name)); err != nil {
			continue
		}
		frames++
	}
	if frames == 0 {
		return "", 0
	}

	outDir := filepath.Join(videoDir, runID)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		slog.Warn("slideshow output dir failed", "run_id", runID, "error", err)
		return "", 0
	}

	filter := "scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2:color=black"
	input := filepath.Join(tmpDir, "%04d.png")

	// Coba codec satu per satu; pakai hasil pertama yang berhasil.
	for _, vc := range videoCodecs {
		outPath := filepath.Join(outDir, "recording."+vc.ext)
		args := []string{
			"-y",
			"-framerate", "1",
			"-i", input,
			"-vf", filter,
			"-c:v", vc.codec,
			"-pix_fmt", "yuv420p",
			"-r", "1",
		}
		if vc.codec == "libvpx-vp9" {
			args = append(args, "-b:v", "1M")
		} else if vc.codec == "libx264" {
			args = append(args, "-movflags", "+faststart")
		}
		args = append(args, outPath)

		cmd := exec.Command("ffmpeg", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			slog.Warn("slideshow ffmpeg failed", "run_id", runID, "codec", vc.codec, "error", err, "output", string(out))
			_ = os.Remove(outPath)
			continue
		}
		return "/videos/" + runID + "/recording." + vc.ext, float64(frames)
	}

	slog.Warn("slideshow video: all codecs failed", "run_id", runID)
	return "", 0
}

// copyFile menyalin src ke dst dengan permission 0o644.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
