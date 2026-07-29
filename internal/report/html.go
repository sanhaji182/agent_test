// Package report menghasilkan laporan HTML dari hasil test run
package report

import (
	"html/template"
	"io"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// Template HTML untuk laporan test run — enhanced with screenshots, timeline, diagnostics
var htmlTmpl = template.Must(template.New("report").Funcs(template.FuncMap{
	"inc": func(i int) int { return i + 1 },
}).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GoTest Agent Report - {{.ID}}</title>
<style>
:root { --pass: #16a34a; --fail: #dc2626; --warn: #d97706; --bg: #f9fafb; --border: #e5e7eb; }
body { font-family: system-ui, -apple-system, sans-serif; max-width: 960px; margin: 0 auto; padding: 2rem; color: #1f2937; }
h1 { border-bottom: 2px solid var(--border); padding-bottom: 0.5rem; }
.badge { display: inline-block; padding: 0.25rem 0.75rem; border-radius: 9999px; font-size: 0.875rem; font-weight: 600; }
.badge-pass { background: #dcfce7; color: var(--pass); }
.badge-fail { background: #fee2e2; color: var(--fail); }
.badge-running { background: #fef3c7; color: var(--warn); }
.summary { display: flex; gap: 2rem; margin: 1.5rem 0; flex-wrap: wrap; }
.stat-card { background: var(--bg); border: 1px solid var(--border); border-radius: 0.5rem; padding: 1rem 1.5rem; text-align: center; min-width: 100px; }
.stat { font-size: 2.5rem; font-weight: bold; display: block; }
.stat.pass { color: var(--pass); } .stat.fail { color: var(--fail); }
table { width: 100%; border-collapse: collapse; margin: 1rem 0; }
th, td { border: 1px solid var(--border); padding: 0.75rem; text-align: left; }
th { background: var(--bg); font-weight: 600; }
tr:hover { background: #f3f4f6; }
.screenshot { max-width: 100%; border: 1px solid var(--border); border-radius: 0.5rem; margin: 0.5rem 0; }
.meta { color: #6b7280; font-size: 0.875rem; }
.section { margin: 2rem 0; }
.progress-bar { height: 8px; background: #e5e7eb; border-radius: 4px; overflow: hidden; margin: 0.5rem 0; }
.progress-fill { height: 100%; background: var(--pass); border-radius: 4px; }
code { background: var(--bg); padding: 0.125rem 0.375rem; border-radius: 0.25rem; font-size: 0.875rem; }
</style>
</head>
<body>
<h1>🧪 Test Run Report</h1>
<p class="meta">Run ID: <code>{{.ID}}</code> | Generated: {{.GeneratedAt}}</p>
<p>
  State:
  {{if eq .State "done"}}<span class="badge badge-pass">✓ DONE</span>
  {{else if eq .State "failed"}}<span class="badge badge-fail">✗ FAILED</span>
  {{else}}<span class="badge badge-running">● {{.State}}</span>{{end}}
</p>

{{if .RunResult}}
<div class="section">
<h2>Results Summary</h2>
<div class="summary">
  <div class="stat-card"><span class="stat pass">{{.RunResult.Passed}}</span>Passed</div>
  <div class="stat-card"><span class="stat fail">{{.RunResult.Failed}}</span>Failed</div>
  <div class="stat-card"><span class="stat">{{.RunResult.Total}}</span>Total</div>
  {{if .RunResult.Total}}
  <div class="stat-card"><span class="stat">{{printf "%.0f" .PassRate}}%</span>Pass Rate</div>
  {{end}}
  {{if .RunResult.DurationMs}}
  <div class="stat-card"><span class="stat">{{printf "%.1f" .DurationSec}}s</span>Duration</div>
  {{end}}
  {{if .RunResult.Healed}}
  <div class="stat-card"><span class="stat pass">{{.RunResult.Healed}}</span>Self-Healed</div>
  {{end}}
  {{if .RunResult.Retried}}
  <div class="stat-card"><span class="stat">{{.RunResult.Retried}}</span>Auto-Retried</div>
  {{end}}
</div>
{{if .RunResult.Total}}
<div class="progress-bar"><div class="progress-fill" style="width: {{printf "%.0f" .PassRate}}%"></div></div>
{{end}}
</div>

{{if .RunResult.VideoPath}}
<div class="section">
<h2>📹 Video Recording</h2>
<p><a href="{{.RunResult.VideoPath}}">Watch execution video</a></p>
</div>
{{end}}

{{if .RunResult.Failures}}
<div class="section">
<h2>❌ Failures ({{len .RunResult.Failures}})</h2>
<table>
<tr><th>#</th><th>Test</th><th>Error</th><th>Screenshot</th></tr>
{{range $i, $f := .RunResult.Failures}}
<tr>
  <td>{{inc $i}}</td>
  <td><code>{{$f.Test}}</code></td>
  <td>{{$f.Message}}</td>
  <td>{{if $f.Screenshot}}<img class="screenshot" src="{{$f.Screenshot}}" alt="failure screenshot" style="max-width:200px">{{else}}—{{end}}</td>
</tr>
{{end}}
</table>
</div>
{{end}}
{{else}}
<p>No results available.</p>
{{end}}

{{if .TestPlan}}
<div class="section">
<h2>📋 Test Plan</h2>
<p>{{.TestPlan.Summary}}</p>
<table>
<tr><th>Scenario</th><th>Priority</th><th>Steps</th></tr>
{{range .TestPlan.Scenarios}}
<tr><td><strong>{{.Name}}</strong></td><td>{{.Priority}}</td><td>{{len .Steps}} steps</td></tr>
{{end}}
</table>
</div>
{{end}}

{{if .TestFiles}}
<div class="section">
<h2>📄 Generated Test Files ({{len .TestFiles}})</h2>
<table>
<tr><th>#</th><th>File</th><th>Actions</th></tr>
{{range $i, $tf := .TestFiles}}
<tr><td>{{inc $i}}</td><td><code>{{$tf.Name}}</code></td><td>{{len $tf.Content}} chars</td></tr>
{{end}}
</table>
</div>
{{end}}

<footer class="meta" style="margin-top: 3rem; border-top: 1px solid var(--border); padding-top: 1rem;">
  Generated by GoTest Agent | AI-Powered Testing Platform
</footer>
</body>
</html>`))

// ReportData adalah data yang dikirim ke template HTML
type ReportData struct {
	ID          string
	State       string
	RunResult   *agent.RunResult
	TestPlan    *agent.TestPlan
	TestFiles   []agent.TestFile
	GeneratedAt string
	PassRate    float64
	DurationSec float64
}

// GenerateHTML menulis laporan HTML ke writer dari data test run
func GenerateHTML(w io.Writer, run *agent.TestRun) error {
	passRate := 0.0
	if run.RunResult != nil && run.RunResult.Total > 0 {
		passRate = float64(run.RunResult.Passed) / float64(run.RunResult.Total) * 100
	}
	data := ReportData{
		ID:          run.ID,
		State:       string(run.State),
		RunResult:   run.RunResult,
		TestPlan:    run.TestPlan,
		TestFiles:   run.TestFiles,
		GeneratedAt: time.Now().Format(time.RFC3339),
		PassRate:    passRate,
	}
	if run.RunResult != nil {
		data.DurationSec = float64(run.RunResult.DurationMs) / 1000.0
	}
	return htmlTmpl.Execute(w, data)
}

// GenerateHTMLString menghasilkan laporan HTML sebagai string
func GenerateHTMLString(run *agent.TestRun) (string, error) {
	w := &stringWriter{}
	if err := GenerateHTML(w, run); err != nil {
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
