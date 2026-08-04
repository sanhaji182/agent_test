// Package report menghasilkan laporan HTML dari hasil test run.
// Laporan tersedia dalam dua bahasa (Indonesia default, Inggris opsional)
// dengan bahasa yang sederhana agar mudah dipahami orang teknis maupun nonteknis.
package report

import (
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// uiStrings menyimpan seluruh teks antarmuka laporan untuk satu bahasa.
type uiStrings struct {
	HTMLLang string

	// Header
	Title       string
	RunLabel    string
	GeneratedAt string
	StartedAt   string
	FinishedAt  string
	ModeLabel   string
	TypeLabel   string

	// State
	StateCompleted string
	StateFailed    string
	StateRunning   string

	// Grade labels
	GradeExcellent string
	GradeGood      string
	GradeFair      string
	GradePoor      string
	GradeCritical  string
	GradeNoData    string

	// Sections
	HowItWorks     string
	ExecSummary    string
	TestObjective  string
	TargetLabel    string
	ResultsTitle   string
	ResultsNote    string
	FailureTitle   string
	ByCategory     string
	TestPlanTitle  string
	TestPlanIntro  string
	TestFilesTitle string
	Screenshots    string
	RecommendTitle string
	ExecError      string
	NoResults      string

	// Metrics
	MPassed   string
	MFailed   string
	MTotal    string
	MPassRate string
	MDuration string
	MFixes    string
	MHealed   string
	MRetried  string

	// Test plan
	StepsLabel      string
	PrioHigh        string
	PrioMedium      string
	PrioLow         string
	PrioNote        string
	CharsLabel      string
	TechDetailLabel string
	TechNote        string
	DurationLabel   string

	// How it works steps (5 langkah)
	Step1, Step2, Step3, Step4, Step5 string

	// Category labels
	CatTimeout  string
	CatNotFound string
	CatAssert   string
	CatNetwork  string
	CatFile     string
	CatOther    string

	Footer string
}

var uiID = uiStrings{
	HTMLLang: "id",

	Title:       "Laporan Hasil Test",
	RunLabel:    "Kode Run",
	GeneratedAt: "Dibuat tanggal",
	StartedAt:   "Mulai",
	FinishedAt:  "Selesai",
	ModeLabel:   "mode",
	TypeLabel:   "jenis",

	StateCompleted: "SELESAI",
	StateFailed:    "GAGAL",
	StateRunning:   "SEDANG BERJALAN",

	GradeExcellent: "Sangat Baik",
	GradeGood:      "Baik",
	GradeFair:      "Cukup",
	GradePoor:      "Kurang",
	GradeCritical:  "Buruk",
	GradeNoData:    "Belum Ada Data",

	HowItWorks:     "Cara Kerja Test Ini",
	ExecSummary:    "Ringkasan Singkat",
	TestObjective:  "Apa yang Dicek",
	TargetLabel:    "Yang dites",
	ResultsTitle:   "Hasilnya",
	ResultsNote:    "pengecekan berhasil",
	FailureTitle:   "Yang Perlu Diperbaiki",
	ByCategory:     "Dikelompokkan menurut jenisnya",
	TestPlanTitle:  "Daftar yang Dicek",
	TestPlanIntro:  "Ini daftar hal yang diperiksa oleh sistem, diurutkan dari yang paling penting. Setiap nomor adalah satu hal yang dicek.",
	TestFilesTitle: "File Test yang Dibuat Sistem",
	Screenshots:    "Tangkapan Layar",
	RecommendTitle: "Saran untuk Anda",
	ExecError:      "Terjadi Kesalahan",
	NoResults:      "Belum ada hasil untuk ditampilkan.",

	MPassed:   "Berhasil",
	MFailed:   "Gagal",
	MTotal:    "Total",
	MPassRate: "Tingkat Berhasil",
	MDuration: "Waktu",
	MFixes:    "Perbaikan Otomatis",
	MHealed:   "Diperbaiki Sendiri",
	MRetried:  "Dicoba Ulang",

	StepsLabel:      "langkah",
	PrioHigh:        "penting",
	PrioMedium:      "sedang",
	PrioLow:         "tambahan",
	PrioNote:        "Tanda warna menunjukkan tingkat kepentingan: merah = paling penting untuk dicek.",
	CharsLabel:      "karakter",
	TechDetailLabel: "Detail teknis (untuk tim teknis)",
	TechNote:        "Bagian ini berisi detail teknis untuk tim developer. Anda bisa melewatkannya.",
	DurationLabel:   "Durasi",

	Step1: "Sistem membuka website Anda, seperti orang membuka browser.",
	Step2: "Sistem menyusun daftar hal-hal yang perlu dicek.",
	Step3: "Sistem mengecek daftar itu satu per satu secara otomatis.",
	Step4: "Kalau ada yang gagal, sistem mencoba memperbaikinya sendiri lalu mengecek ulang.",
	Step5: "Sistem membuat laporan ini supaya Anda tahu hasilnya.",

	CatTimeout:  "Menunggu terlalu lama",
	CatNotFound: "Tidak menemukan yang dicari",
	CatAssert:   "Hasil tidak sesuai harapan",
	CatNetwork:  "Masalah jaringan",
	CatFile:     "File tidak ada",
	CatOther:    "Lainnya",

	Footer: "Dibuat oleh GoTest Agent — Platform Testing Bertenaga AI",
}

var uiEN = uiStrings{
	HTMLLang: "en",

	Title:       "Test Results Report",
	RunLabel:    "Run ID",
	GeneratedAt: "Generated",
	StartedAt:   "Started",
	FinishedAt:  "Finished",
	ModeLabel:   "mode",
	TypeLabel:   "type",

	StateCompleted: "COMPLETED",
	StateFailed:    "FAILED",
	StateRunning:   "RUNNING",

	GradeExcellent: "Excellent",
	GradeGood:      "Good",
	GradeFair:      "Fair",
	GradePoor:      "Poor",
	GradeCritical:  "Critical",
	GradeNoData:    "No Data",

	HowItWorks:     "How This Test Works",
	ExecSummary:    "Quick Summary",
	TestObjective:  "What Was Checked",
	TargetLabel:    "Target",
	ResultsTitle:   "The Results",
	ResultsNote:    "checks passed",
	FailureTitle:   "What Needs Fixing",
	ByCategory:     "Grouped by type",
	TestPlanTitle:  "What Was Checked",
	TestPlanIntro:  "This is the list of things the system checked, ordered from most important. Each number is one thing that was checked.",
	TestFilesTitle: "Test Files the System Created",
	Screenshots:    "Screenshots",
	RecommendTitle: "Suggestions for You",
	ExecError:      "Something Went Wrong",
	NoResults:      "No results to show yet.",

	MPassed:   "Passed",
	MFailed:   "Failed",
	MTotal:    "Total",
	MPassRate: "Success Rate",
	MDuration: "Duration",
	MFixes:    "Auto-Fixes",
	MHealed:   "Self-Healed",
	MRetried:  "Retried",

	StepsLabel:      "steps",
	PrioHigh:        "important",
	PrioMedium:      "medium",
	PrioLow:         "optional",
	PrioNote:        "The color shows how important each check is: red = most important.",
	CharsLabel:      "chars",
	TechDetailLabel: "Technical details (for technical team)",
	TechNote:        "This section contains technical details for the developer team. You can skip it.",
	DurationLabel:   "Duration",

	Step1: "The system opens your website, like a person opening a browser.",
	Step2: "The system makes a list of things to check.",
	Step3: "The system checks that list one by one, automatically.",
	Step4: "If something fails, the system tries to fix it and checks again.",
	Step5: "The system creates this report so you know the results.",

	CatTimeout:  "Waited too long",
	CatNotFound: "Couldn't find what it looked for",
	CatAssert:   "Result didn't match expectation",
	CatNetwork:  "Network problem",
	CatFile:     "File missing",
	CatOther:    "Other",

	Footer: "Generated by GoTest Agent — AI-Powered Testing Platform",
}

func stringsForLang(lang string) uiStrings {
	if strings.HasPrefix(strings.ToLower(lang), "en") {
		return uiEN
	}
	return uiID
}

// FailureCategory mengelompokkan kegagalan berdasarkan jenisnya.
type FailureCategory struct {
	Key   string // kunci internal: timeout/notfound/assert/network/file/other
	Label string // label sesuai bahasa
	Count int
	Items []humanFailure
}

type humanFailure struct {
	RawTest     string // nama teknis asli
	RawMessage  string // pesan error mentah
	HumanName   string // nama ramah bahasa manusia
	Title       string // judul singkat masalahnya
	Desc        string // deskripsi rinci apa yang terjadi
	Screenshot  string
	DurationSec string
}

// recommendation adalah satu saran untuk pengguna.
type recommendation struct {
	Icon  string
	Title string
	Body  string
}

// ReportData adalah data yang dikirim ke template HTML.
type ReportData struct {
	S            uiStrings
	Lang         string
	ID           string
	State        string
	Requirements string
	ProjectPath  string
	Mode         string
	TestType     string
	RunResult    *agent.RunResult
	TestPlan     *agent.TestPlan
	TestFiles    []agent.TestFile
	Screenshots  []string
	GeneratedAt  string
	CreatedAt    string
	FinishedAt   string
	PassRate     float64
	FailRate     float64
	DurationSec  float64
	FixAttempts  int
	RunError     string
	Grade        string
	GradeLabel   string
	Summary      string
	Categories   []FailureCategory
	Recs         []recommendation
	HasFailures  bool
}

// categoryKey menentukan kunci kategori dari pesan kegagalan.
func categoryKey(msg string) string {
	lower := strings.ToLower(msg)
	switch {
	case strings.Contains(lower, "timeout"):
		return "timeout"
	case strings.Contains(lower, "not found"):
		return "notfound"
	case strings.Contains(lower, "assert"):
		return "assert"
	case strings.Contains(lower, "egress") || strings.Contains(lower, "blocked"):
		return "network"
	case strings.Contains(lower, "no such file"):
		return "file"
	default:
		return "other"
	}
}

// humanizeTestName mengubah nama teknis menjadi ramah bahasa manusia
func humanizeTestName(raw, lang string) (humanName, actionNum string) {
	name := raw
	// Extract action number if present (format: "name.json:action_N")
	if idx := strings.LastIndex(name, ":action_"); idx >= 0 {
		actionNum = name[idx+len(":action_"):]
		name = name[:idx]
	}
	// Strip .json suffix
	name = strings.TrimSuffix(name, ".json")
	// Strip leading numeric prefix like "2_" (e.g. "2_main..." -> "main...")
	for len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = name[1:]
	}
	name = strings.TrimPrefix(name, "_")
	// Replace underscores with spaces
	name = strings.TrimSpace(strings.ReplaceAll(name, "_", " "))

	if lang == "id" {
		humanName = name
		if actionNum != "" {
			humanName = name + " — langkah ke-" + actionNum
		}
	} else {
		humanName = name
		if actionNum != "" {
			humanName = name + " (step " + actionNum + ")"
		}
	}
	return humanName, actionNum
}

// humanizeFailureMessage mengubah pesan error mentah menjadi kalimat mudah dipahami
func humanizeFailureMessage(f agent.Failure, lang string) (title, desc string) {
	key := categoryKey(f.Message)
	durationSec := float64(f.DurationMs) / 1000.0

	switch key {
	case "timeout":
		if lang == "id" {
			title = "Halaman terlalu lama dibuka"
			desc = fmt.Sprintf("Sistem menunggu lebih dari %.0f detik tapi halaman tidak kunjung terbuka. Website Anda mungkin sedang lambat dimuat. Coba cek kecepatan website Anda atau koneksi internet.", durationSec)
		} else {
			title = "Page load too slow"
			desc = fmt.Sprintf("System waited more than %.0f seconds but page didn't open. Your website may be loading slowly. Try checking your website speed or internet connection.", durationSec)
		}
	case "notfound":
		if lang == "id" {
			title = "Tidak menemukan yang dicari"
			desc = "Sistem mencari suatu tombol atau teks di halaman tapi tidak menemukannya. Mungkin tampilan website sudah berubah atau tombol/teks yang dicari memang belum ada."
		} else {
			title = "Couldn't find what it looked for"
			desc = "System looked for a button or text on the page but didn't find it. The layout may have changed or the element doesn't exist yet."
		}
	case "assert":
		if lang == "id" {
			title = "Isi tidak sesuai harapan"
			desc = "Hasil pengecekan tidak sesuai dengan yang diharapkan. Pastikan isi website sudah seperti yang seharusnya."
		} else {
			title = "Result didn't match expectation"
			desc = "The check result didn't match expectations. Verify that the website content is correct."
		}
	case "network":
		if lang == "id" {
			title = "Akses diblokir jaringan"
			desc = "Sistem tidak diizinkan mengakses alamat ini oleh pengaturan keamanan website atau firewall Anda."
		} else {
			title = "Access blocked by network"
			desc = "System is not allowed to access this address due to security settings or firewall."
		}
	case "file":
		if lang == "id" {
			title = "Bukti gambar tidak tersedia"
			desc = "Screenshot tidak berhasil disimpan. Cek apakah folder screenshot sudah dibuat dan writable."
		} else {
			title = "Screenshot not available"
			desc = "Screenshot failed to save. Check if screenshot folder exists and is writable."
		}
	default:
		if lang == "id" {
			title = "Pengecekan gagal"
			desc = "Ada masalah saat menjalankan pengecekan ini. Lihat pesan detail di bawah."
		} else {
			title = "Check failed"
			desc = "There was an issue running this check. See detailed message below."
		}
	}
	return title, desc
}

func categoryLabel(key string, s uiStrings) string {
	switch key {
	case "timeout":
		return "⏱ " + s.CatTimeout
	case "notfound":
		return "🔍 " + s.CatNotFound
	case "assert":
		return "❌ " + s.CatAssert
	case "network":
		return "🚫 " + s.CatNetwork
	case "file":
		return "📁 " + s.CatFile
	default:
		return "⚠️ " + s.CatOther
	}
}

// categorizeFailures membangun daftar kategori kegagalan terurut.
func categorizeFailures(failures []agent.Failure, s uiStrings) []FailureCategory {
	buckets := map[string][]agent.Failure{}
	var order []string
	for _, f := range failures {
		key := categoryKey(f.Message)
		if _, ok := buckets[key]; !ok {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], f)
	}
	out := make([]FailureCategory, 0, len(order))
	for _, key := range order {
		items := make([]humanFailure, 0, len(buckets[key]))
		for _, f := range buckets[key] {
			humanName, _ := humanizeTestName(f.Test, s.HTMLLang)
			title, desc := humanizeFailureMessage(f, s.HTMLLang)
			items = append(items, humanFailure{
				RawTest:     f.Test,
				RawMessage:  f.Message,
				HumanName:   humanName,
				Title:       title,
				Desc:        desc,
				Screenshot:  f.Screenshot,
				DurationSec: fmt.Sprintf("%.1f", float64(f.DurationMs)/1000.0),
			})
		}
		out = append(out, FailureCategory{
			Key:   key,
			Label: categoryLabel(key, s),
			Count: len(buckets[key]),
			Items: items,
		})
	}
	return out
}

// gradeForPassRate menghasilkan nilai huruf + label sesuai bahasa.
func gradeForPassRate(rate float64, s uiStrings) (string, string) {
	switch {
	case rate >= 95:
		return "A", s.GradeExcellent
	case rate >= 85:
		return "B", s.GradeGood
	case rate >= 70:
		return "C", s.GradeFair
	case rate >= 50:
		return "D", s.GradePoor
	default:
		return "F", s.GradeCritical
	}
}

// buildSummary membuat ringkasan dengan bahasa sederhana.
func buildSummary(run *agent.TestRun, passRate float64, s uiStrings) string {
	if run.RunResult == nil || run.RunResult.Total == 0 {
		if s.HTMLLang == "id" {
			return "Test sudah dijalankan, tapi belum ada pengecekan yang selesai. Lihat bagian kesalahan di atas untuk tahu penyebabnya."
		}
		return "The test ran, but no checks completed. See the error section above to find out why."
	}
	r := run.RunResult
	total := fmt.Sprintf("%d", r.Total)
	passed := fmt.Sprintf("%d", r.Passed)
	failed := fmt.Sprintf("%d", r.Failed)
	rate := fmt.Sprintf("%.0f", passRate)

	if s.HTMLLang == "id" {
		switch {
		case r.Failed == 0:
			return "Kabar baik! Semua " + total + " pengecekan berhasil. Website Anda berfungsi dengan baik."
		case passRate >= 80:
			return "Dari " + total + " pengecekan, " + passed + " berhasil (" + rate + "%). Secara umum website Anda berfungsi baik, tapi ada " + failed + " bagian yang perlu diperhatikan. Lihat bagian 'Yang Perlu Diperbaiki' di bawah."
		default:
			return "Hanya " + passed + " dari " + total + " pengecekan yang berhasil (" + rate + "%). Ada beberapa masalah yang perlu segera diperbaiki. Mulai dari bagian yang ditandai di bawah ini."
		}
	}
	// English
	switch {
	case r.Failed == 0:
		return "Good news! All " + total + " checks passed. Your website is working well."
	case passRate >= 80:
		return "Out of " + total + " checks, " + passed + " passed (" + rate + "%). Overall your website works well, but " + failed + " part(s) need attention. See 'What Needs Fixing' below."
	default:
		return "Only " + passed + " of " + total + " checks passed (" + rate + "%). There are issues that need fixing soon. Start with the items highlighted below."
	}
}

// buildRecommendations menyusun saran sesuai kategori kegagalan dan bahasa.
func buildRecommendations(cats []FailureCategory, hasFailures bool, s uiStrings) []recommendation {
	var recs []recommendation
	id := s.HTMLLang == "id"

	if hasFailures {
		for _, c := range cats {
			n := fmt.Sprintf("%d", c.Count)
			switch c.Key {
			case "timeout":
				if id {
					recs = append(recs, recommendation{"⏱", "Ada yang terlalu lambat", "Ada " + n + " pengecekan yang menunggu terlalu lama. Mungkin website Anda lambat dimuat. Coba periksa kecepatan website Anda."})
				} else {
					recs = append(recs, recommendation{"⏱", "Something is too slow", n + " check(s) waited too long. Your website may be slow to load. Try checking your website's speed."})
				}
			case "notfound":
				if id {
					recs = append(recs, recommendation{"🔍", "Ada yang tidak ketemu", "Ada " + n + " pengecekan yang tidak menemukan tombol atau teks yang dicari. Mungkin tampilan website Anda sudah berubah."})
				} else {
					recs = append(recs, recommendation{"🔍", "Something wasn't found", n + " check(s) couldn't find the button or text they looked for. Your website's layout may have changed."})
				}
			case "assert":
				if id {
					recs = append(recs, recommendation{"❌", "Ada hasil yang tidak cocok", "Ada " + n + " pengecekan yang hasilnya tidak sesuai harapan. Pastikan isi website sudah seperti yang seharusnya."})
				} else {
					recs = append(recs, recommendation{"❌", "Some results didn't match", n + " check(s) didn't match expectations. Make sure your website content is as it should be."})
				}
			case "network":
				if id {
					recs = append(recs, recommendation{"🚫", "Ada yang diblokir jaringan", "Ada " + n + " pengecekan yang diblokir oleh pengaturan jaringan. Pastikan alamat website diizinkan."})
				} else {
					recs = append(recs, recommendation{"🚫", "Blocked by network", n + " check(s) were blocked by network settings. Make sure the website address is allowed."})
				}
			default:
				if id {
					recs = append(recs, recommendation{"⚠️", "Ada kegagalan lain", "Ada " + n + " pengecekan yang gagal karena alasan lain. Lihat detail di atas."})
				} else {
					recs = append(recs, recommendation{"⚠️", "Other failures", n + " check(s) failed for other reasons. See the details above."})
				}
			}
		}
	} else {
		if id {
			recs = append(recs, recommendation{"✅", "Semua beres", "Tidak ada masalah yang ditemukan. Terus pantau dengan menjalankan test secara berkala."})
		} else {
			recs = append(recs, recommendation{"✅", "All clear", "No problems found. Keep monitoring by running tests regularly."})
		}
	}

	// Saran umum
	if id {
		recs = append(recs, recommendation{"🔁", "Jalankan test secara rutin", "Atur jadwal test otomatis supaya Anda segera tahu kalau ada yang rusak di website Anda."})
		recs = append(recs, recommendation{"📤", "Hubungkan ke sistem otomatis Anda", "Unduh hasil test dalam format standar (JUnit XML) supaya bisa dibaca oleh sistem otomatis tim Anda (CI/CD)."})
	} else {
		recs = append(recs, recommendation{"🔁", "Run tests regularly", "Set up an automatic test schedule so you know right away if something breaks on your website."})
		recs = append(recs, recommendation{"📤", "Connect to your automated system", "Download test results in a standard format (JUnit XML) so your team's automated system (CI/CD) can read them."})
	}
	return recs
}

// Template HTML untuk laporan test run — bilingual, bahasa sederhana.
var htmlTmpl = template.Must(template.New("report").Funcs(template.FuncMap{
	"inc":  func(i int) int { return i + 1 },
	"secs": func(ms int64) string { return fmt.Sprintf("%.1f", float64(ms)/1000.0) },
}).Parse(`<!DOCTYPE html>
<html lang="{{.S.HTMLLang}}">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.S.Title}} — {{.ID}}</title>
<style>
:root {
  --pass: #16a34a; --pass-bg: #dcfce7;
  --fail: #dc2626; --fail-bg: #fee2e2;
  --warn: #d97706; --warn-bg: #fef3c7;
  --info: #2563eb; --info-bg: #dbeafe;
  --bg: #f8fafc; --surface: #ffffff; --border: #e2e8f0;
  --text: #1e293b; --muted: #64748b;
}
* { box-sizing: border-box; }
body { font-family: system-ui, -apple-system, 'Segoe UI', sans-serif; margin: 0; background: var(--bg); color: var(--text); line-height: 1.7; }
.container { max-width: 1000px; margin: 0 auto; padding: 0 1.5rem 3rem; }
.banner { background: linear-gradient(135deg, #1e293b, #334155); color: #fff; padding: 2rem 0 1.5rem; margin-bottom: 2rem; }
.banner .container { display: flex; justify-content: space-between; align-items: flex-start; flex-wrap: wrap; gap: 1rem; padding-bottom: 0; }
.banner h1 { margin: 0 0 0.25rem; font-size: 1.5rem; }
.banner .sub { color: #94a3b8; font-size: 0.85rem; }
.banner .sub code { background: rgba(255,255,255,0.12); color: #e2e8f0; padding: 0.1rem 0.4rem; border-radius: 4px; }
.grade-box { text-align: center; background: rgba(255,255,255,0.1); border-radius: 12px; padding: 0.75rem 1.5rem; }
.grade-box .grade { font-size: 2.5rem; font-weight: 800; line-height: 1; }
.grade-box .label { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: #cbd5e1; }
.banner-right { display: flex; flex-direction: column; gap: 0.75rem; align-items: flex-end; }
.lang-switch { display: flex; gap: 0.25rem; background: rgba(255,255,255,0.12); border-radius: 8px; padding: 0.25rem; }
.lang-switch button { border: none; background: transparent; color: #cbd5e1; padding: 0.3rem 0.8rem; border-radius: 6px; cursor: pointer; font-size: 0.85rem; font-weight: 600; transition: all 0.15s; }
.lang-switch button.active { background: rgba(255,255,255,0.92); color: #1e293b; }
.lang-switch button:hover:not(.active) { background: rgba(255,255,255,0.2); color: #fff; }
.badge { display: inline-block; padding: 0.2rem 0.7rem; border-radius: 9999px; font-size: 0.8rem; font-weight: 600; }
.badge-pass { background: var(--pass-bg); color: var(--pass); }
.badge-fail { background: var(--fail-bg); color: var(--fail); }
.badge-running { background: var(--warn-bg); color: var(--warn); }
.metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(130px, 1fr)); gap: 1rem; margin: 1.5rem 0; }
.metric { background: var(--surface); border: 1px solid var(--border); border-radius: 12px; padding: 1.25rem; text-align: center; }
.metric .val { font-size: 2rem; font-weight: 800; display: block; line-height: 1.1; }
.metric .lbl { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: var(--muted); margin-top: 0.25rem; }
.metric.pass .val { color: var(--pass); }
.metric.fail .val { color: var(--fail); }
.metric.info .val { color: var(--info); }
.section { background: var(--surface); border: 1px solid var(--border); border-radius: 12px; padding: 1.5rem; margin: 1.5rem 0; }
.section h2 { margin: 0 0 1rem; font-size: 1.15rem; display: flex; align-items: center; gap: 0.5rem; }
.section h3 { font-size: 0.95rem; margin: 1.25rem 0 0.5rem; color: var(--muted); }
.summary-text { font-size: 1.05rem; background: var(--bg); border-left: 4px solid var(--info); padding: 1rem 1.25rem; border-radius: 0 8px 8px 0; }
.progress { height: 14px; background: #e5e7eb; border-radius: 7px; overflow: hidden; margin: 0.75rem 0; display: flex; }
.progress .fill-pass { background: var(--pass); }
.progress .fill-fail { background: var(--fail); }
table { width: 100%; border-collapse: collapse; margin: 0.75rem 0; font-size: 0.9rem; }
th, td { border: 1px solid var(--border); padding: 0.6rem 0.75rem; text-align: left; vertical-align: top; }
th { background: var(--bg); font-weight: 600; font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.03em; color: var(--muted); }
code { background: var(--bg); padding: 0.1rem 0.4rem; border-radius: 4px; font-size: 0.85em; font-family: ui-monospace, 'SF Mono', monospace; }
.cat-chip { display: inline-flex; align-items: center; gap: 0.4rem; padding: 0.3rem 0.75rem; border-radius: 9999px; font-size: 0.8rem; font-weight: 600; margin: 0.2rem; background: var(--fail-bg); color: var(--fail); }
.failure-card { border: 1px solid var(--border); border-left: 4px solid var(--fail); border-radius: 8px; padding: 1rem; margin: 0.75rem 0; background: #fffbfb; }
.failure-card .test-name { font-weight: 600; font-family: ui-monospace, monospace; font-size: 0.85rem; color: var(--fail); }
.failure-card .msg { margin: 0.5rem 0; font-size: 0.9rem; }
.tech-detail { margin-top: 0.75rem; border: 1px dashed var(--border); }
.tech-detail summary { font-size: 0.8rem; color: var(--muted); background: transparent; }
.tech-detail pre { font-size: 0.75rem; }
.failure-card img { max-width: 320px; border: 1px solid var(--border); border-radius: 8px; margin-top: 0.5rem; }
.steps-flow { counter-reset: step; list-style: none; padding: 0; margin: 0; }
.steps-flow li { position: relative; padding: 0.75rem 0 0.75rem 3rem; border-left: 2px solid var(--border); margin-left: 1rem; }
.steps-flow li:last-child { border-left-color: transparent; }
.steps-flow li::before { counter-increment: step; content: counter(step); position: absolute; left: -1rem; top: 0.6rem; width: 2rem; height: 2rem; background: var(--info); color: #fff; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 0.9rem; }
.scenario { border: 1px solid var(--border); border-radius: 8px; margin: 0.75rem 0; overflow: hidden; }
.scenario-head { padding: 0.75rem 1rem; background: var(--bg); display: flex; align-items: center; gap: 0.5rem; font-weight: 600; }
.scenario ol { margin: 0; padding: 1rem 1rem 1rem 2.5rem; }
.scenario li { margin: 0.35rem 0; font-size: 0.92rem; }
.prio { display: inline-block; padding: 0.1rem 0.5rem; border-radius: 4px; font-size: 0.7rem; font-weight: 700; text-transform: uppercase; }
.prio-high { background: var(--fail-bg); color: var(--fail); }
.prio-medium { background: var(--warn-bg); color: var(--warn); }
.prio-low { background: var(--info-bg); color: var(--info); }
details { border: 1px solid var(--border); border-radius: 8px; margin: 0.5rem 0; }
details summary { padding: 0.75rem 1rem; cursor: pointer; font-weight: 600; font-size: 0.9rem; background: var(--bg); border-radius: 8px; }
details[open] summary { border-bottom: 1px solid var(--border); border-radius: 8px 8px 0 0; }
details pre { margin: 0; padding: 1rem; overflow-x: auto; font-size: 0.8rem; background: #0f172a; color: #e2e8f0; border-radius: 0 0 8px 8px; }
.gallery { display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 1rem; }
.gallery img { width: 100%; border: 1px solid var(--border); border-radius: 8px; }
.rec { display: flex; gap: 0.75rem; padding: 0.85rem 0; border-bottom: 1px solid var(--border); }
.rec:last-child { border-bottom: none; }
.rec .icon { font-size: 1.4rem; }
.rec .body strong { display: block; font-size: 0.95rem; }
.rec .body span { font-size: 0.88rem; color: var(--muted); }
.note { font-size: 0.85rem; color: var(--muted); background: var(--bg); padding: 0.75rem 1rem; border-radius: 8px; margin-top: 0.75rem; }
.meta { color: var(--muted); font-size: 0.85rem; }
footer { margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid var(--border); color: var(--muted); font-size: 0.8rem; text-align: center; }
@media print { .banner { -webkit-print-color-adjust: exact; } }
</style>
</head>
<body>

<div class="banner">
  <div class="container">
    <div>
      <h1>🧪 {{.S.Title}}</h1>
      <div class="sub">
        {{.S.RunLabel}} <code>{{.ID}}</code><br>
        {{.S.GeneratedAt}} {{.GeneratedAt}}
        {{if .CreatedAt}} · {{.S.StartedAt}} {{.CreatedAt}}{{end}}
        {{if .FinishedAt}} · {{.S.FinishedAt}} {{.FinishedAt}}{{end}}
      </div>
      <div style="margin-top:0.75rem">
        {{if eq .State "done"}}<span class="badge badge-pass">✓ {{.S.StateCompleted}}</span>
        {{else if eq .State "failed"}}<span class="badge badge-fail">✗ {{.S.StateFailed}}</span>
        {{else}}<span class="badge badge-running">● {{.S.StateRunning}}</span>{{end}}
        {{if .Mode}}<span class="badge" style="background:rgba(255,255,255,0.12);color:#e2e8f0">{{.S.ModeLabel}}: {{.Mode}}</span>{{end}}
        {{if .TestType}}<span class="badge" style="background:rgba(255,255,255,0.12);color:#e2e8f0">{{.S.TypeLabel}}: {{.TestType}}</span>{{end}}
      </div>
    </div>
    <div class="banner-right">
      <div class="lang-switch">
        <button onclick="setLang('id')" class="{{if eq .Lang "id"}}active{{end}}">🇮🇩 ID</button>
        <button onclick="setLang('en')" class="{{if eq .Lang "en"}}active{{end}}">🇬🇧 EN</button>
      </div>
      <div class="grade-box">
        <div class="grade">{{.Grade}}</div>
        <div class="label">{{.GradeLabel}}</div>
      </div>
    </div>
  </div>
</div>

<div class="container">

{{if .RunError}}
<div class="section" style="border-left:4px solid var(--fail)">
  <h2>⚠️ {{.S.ExecError}}</h2>
  <pre style="background:var(--fail-bg);color:var(--fail);padding:1rem;border-radius:8px;white-space:pre-wrap">{{.RunError}}</pre>
</div>
{{end}}

<!-- Cara Kerja Test -->
<div class="section">
  <h2>🚀 {{.S.HowItWorks}}</h2>
  <ol class="steps-flow">
    <li>{{.S.Step1}}</li>
    <li>{{.S.Step2}}</li>
    <li>{{.S.Step3}}</li>
    <li>{{.S.Step4}}</li>
    <li>{{.S.Step5}}</li>
  </ol>
</div>

<!-- Ringkasan -->
<div class="section">
  <h2>📊 {{.S.ExecSummary}}</h2>
  <div class="summary-text">{{.Summary}}</div>
</div>

<!-- Apa yang dicek -->
{{if .Requirements}}
<div class="section">
  <h2>🎯 {{.S.TestObjective}}</h2>
  <p style="margin:0">{{.Requirements}}</p>
  {{if .ProjectPath}}<p class="meta" style="margin:0.5rem 0 0">{{.S.TargetLabel}}: <code>{{.ProjectPath}}</code></p>{{end}}
</div>
{{end}}

{{if .RunResult}}
<!-- Angka-angka hasil -->
<div class="metrics">
  <div class="metric pass"><span class="val">{{.RunResult.Passed}}</span><span class="lbl">{{.S.MPassed}}</span></div>
  <div class="metric fail"><span class="val">{{.RunResult.Failed}}</span><span class="lbl">{{.S.MFailed}}</span></div>
  <div class="metric"><span class="val">{{.RunResult.Total}}</span><span class="lbl">{{.S.MTotal}}</span></div>
  <div class="metric info"><span class="val">{{printf "%.0f" .PassRate}}%</span><span class="lbl">{{.S.MPassRate}}</span></div>
  {{if .RunResult.DurationMs}}<div class="metric"><span class="val">{{printf "%.1f" .DurationSec}}s</span><span class="lbl">{{.S.MDuration}}</span></div>{{end}}
  {{if .FixAttempts}}<div class="metric"><span class="val">{{.FixAttempts}}</span><span class="lbl">{{.S.MFixes}}</span></div>{{end}}
  {{if .RunResult.Healed}}<div class="metric pass"><span class="val">{{.RunResult.Healed}}</span><span class="lbl">{{.S.MHealed}}</span></div>{{end}}
  {{if .RunResult.Retried}}<div class="metric"><span class="val">{{.RunResult.Retried}}</span><span class="lbl">{{.S.MRetried}}</span></div>{{end}}
</div>

{{if .RunResult.Total}}
<div class="section">
  <h2>📈 {{.S.ResultsTitle}}</h2>
  <div class="progress">
    <div class="fill-pass" style="width: {{printf "%.1f" .PassRate}}%"></div>
    <div class="fill-fail" style="width: {{printf "%.1f" .FailRate}}%"></div>
  </div>
  <p class="meta">{{.RunResult.Passed}} {{.S.ResultsNote}} ({{printf "%.0f" .PassRate}}%).</p>
</div>
{{end}}

{{if .RunResult.VideoPath}}
<div class="section">
  <h2>📹 Video</h2>
  <p><a href="{{.RunResult.VideoPath}}">▶ {{.S.Screenshots}}</a></p>
</div>
{{end}}

<!-- Yang perlu diperbaiki -->
{{if .RunResult.Failures}}
<div class="section">
  <h2>❌ {{.S.FailureTitle}} ({{len .RunResult.Failures}})</h2>
  <h3>{{.S.ByCategory}}</h3>
  <div>
    {{range .Categories}}<span class="cat-chip">{{.Label}} · {{.Count}}</span>{{end}}
  </div>
  {{range .Categories}}
  <h3>{{.Label}} ({{.Count}})</h3>
  {{range $i, $f := .Items}}
  <div class="failure-card">
    <div class="test-name">{{$f.HumanName}}</div>
    <div class="msg"><strong>{{$f.Title}}</strong><br>{{$f.Desc}}</div>
    {{if $f.DurationSec}}<div class="meta">{{$.S.DurationLabel}}: {{$f.DurationSec}}s</div>{{end}}
    {{if $f.Screenshot}}<img src="{{$f.Screenshot}}" alt="screenshot">{{end}}
    <details class="tech-detail">
      <summary>{{$.S.TechDetailLabel}}</summary>
      <pre>{{$f.RawTest}}
{{$f.RawMessage}}</pre>
    </details>
  </div>
  {{end}}
  {{end}}
</div>
{{end}}
{{else}}
<div class="section"><p class="meta">{{.S.NoResults}}</p></div>
{{end}}

<!-- Daftar yang dicek -->
{{if .TestPlan}}
<div class="section">
  <h2>📋 {{.S.TestPlanTitle}}</h2>
  <p>{{.TestPlan.Summary}}</p>
  <p class="note">{{.S.TestPlanIntro}}</p>
  {{range $i, $sc := .TestPlan.Scenarios}}
  <div class="scenario">
    <div class="scenario-head">
      {{if eq $sc.Priority "high"}}<span class="prio prio-high">{{$.S.PrioHigh}}</span>
      {{else if eq $sc.Priority "medium"}}<span class="prio prio-medium">{{$.S.PrioMedium}}</span>
      {{else}}<span class="prio prio-low">{{$.S.PrioLow}}</span>{{end}}
      <span>{{$sc.Name}}</span>
      <span class="meta">({{len $sc.Steps}} {{$.S.StepsLabel}})</span>
    </div>
    <ol>
      {{range $sc.Steps}}<li>{{.}}</li>{{end}}
    </ol>
  </div>
  {{end}}
  <p class="note">{{.S.PrioNote}}</p>
</div>
{{end}}

<!-- File test -->
{{if .TestFiles}}
<div class="section">
  <h2>📄 {{.S.TestFilesTitle}} ({{len .TestFiles}})</h2>
  <p class="note">{{.S.TechNote}}</p>
  {{range $i, $tf := .TestFiles}}
  <details>
    <summary>{{inc $i}}. <code>{{$tf.Name}}</code> <span class="meta">({{len $tf.Content}} {{$.S.CharsLabel}})</span></summary>
    <pre>{{$tf.Content}}</pre>
  </details>
  {{end}}
</div>
{{end}}

<!-- Screenshots -->
{{if .Screenshots}}
<div class="section">
  <h2>🖼️ {{.S.Screenshots}} ({{len .Screenshots}})</h2>
  <div class="gallery">
    {{range .Screenshots}}<a href="{{.}}" target="_blank"><img src="{{.}}" alt="screenshot"></a>{{end}}
  </div>
</div>
{{end}}

<!-- Saran -->
<div class="section">
  <h2>💡 {{.S.RecommendTitle}}</h2>
  {{range .Recs}}
  <div class="rec"><div class="icon">{{.Icon}}</div><div class="body"><strong>{{.Title}}</strong><span>{{.Body}}</span></div></div>
  {{end}}
</div>

<footer>{{.S.Footer}} · {{.S.RunLabel}} {{.ID}}</footer>

</div>
<script>
function setLang(l){var p=new URLSearchParams(window.location.search);p.set('lang',l);window.location.search=p.toString();}
</script>
</body>
</html>`))

// GenerateHTML menulis laporan HTML ke writer dengan bahasa tertentu ("id" atau "en").
func GenerateHTML(w io.Writer, run *agent.TestRun, lang string) error {
	s := stringsForLang(lang)

	passRate := 0.0
	if run.RunResult != nil && run.RunResult.Total > 0 {
		passRate = float64(run.RunResult.Passed) / float64(run.RunResult.Total) * 100
	}
	grade, gradeLabel := gradeForPassRate(passRate, s)
	if run.RunResult == nil || run.RunResult.Total == 0 {
		grade, gradeLabel = "—", s.GradeNoData
	}

	var categories []FailureCategory
	hasFailures := false
	if run.RunResult != nil && len(run.RunResult.Failures) > 0 {
		categories = categorizeFailures(run.RunResult.Failures, s)
		hasFailures = true
	}

	createdAt, finishedAt := "", ""
	if !run.CreatedAt.IsZero() {
		createdAt = run.CreatedAt.Format("2006-01-02 15:04:05")
	}
	if run.FinishedAt != nil {
		finishedAt = run.FinishedAt.Format("2006-01-02 15:04:05")
	}

	data := ReportData{
		S:            s,
		Lang:         s.HTMLLang,
		ID:           run.ID,
		State:        string(run.State),
		Requirements: run.Requirements,
		ProjectPath:  run.ProjectPath,
		Mode:         run.Mode,
		TestType:     run.TestType,
		RunResult:    run.RunResult,
		TestPlan:     run.TestPlan,
		TestFiles:    run.TestFiles,
		Screenshots:  run.Screenshots,
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		CreatedAt:    createdAt,
		FinishedAt:   finishedAt,
		PassRate:     passRate,
		FailRate:     100.0 - passRate,
		FixAttempts:  run.FixAttempts,
		RunError:     run.Error,
		Grade:        grade,
		GradeLabel:   gradeLabel,
		Summary:      buildSummary(run, passRate, s),
		Categories:   categories,
		Recs:         buildRecommendations(categories, hasFailures, s),
		HasFailures:  hasFailures,
	}
	if run.RunResult != nil {
		data.DurationSec = float64(run.RunResult.DurationMs) / 1000.0
	}
	return htmlTmpl.Execute(w, data)
}

// GenerateHTMLString menghasilkan laporan HTML sebagai string (default bahasa Indonesia).
func GenerateHTMLString(run *agent.TestRun) (string, error) {
	return GenerateHTMLStringLang(run, "id")
}

// GenerateHTMLStringLang menghasilkan laporan HTML sebagai string dengan bahasa tertentu.
func GenerateHTMLStringLang(run *agent.TestRun, lang string) (string, error) {
	w := &stringWriter{}
	if err := GenerateHTML(w, run, lang); err != nil {
		return "", err
	}
	return w.String(), nil
}

type stringWriter struct {
	data []byte
}

func (s *stringWriter) Write(p []byte) (int, error) {
	s.data = append(s.data, p...)
	return len(p), nil
}

func (s *stringWriter) String() string {
	return string(s.data)
}
