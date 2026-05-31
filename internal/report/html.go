// Package report menghasilkan laporan HTML dari hasil test run
package report

import (
	"html/template"
	"io"
	"time"

	"github.com/go-go-golems/gotest-agent/internal/agent"
)

// Template HTML untuk laporan test run
var htmlTmpl = template.Must(template.New("report").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>GoTest Agent Report - {{.ID}}</title>
<style>
body { font-family: system-ui, sans-serif; max-width: 800px; margin: 0 auto; padding: 2rem; }
.pass { color: #16a34a; } .fail { color: #dc2626; }
table { width: 100%; border-collapse: collapse; margin: 1rem 0; }
th, td { border: 1px solid #e5e7eb; padding: 0.5rem; text-align: left; }
th { background: #f9fafb; }
.summary { display: flex; gap: 2rem; margin: 1rem 0; }
.stat { font-size: 2rem; font-weight: bold; }
</style>
</head>
<body>
<h1>Test Run Report</h1>
<p>Run ID: <code>{{.ID}}</code></p>
<p>State: <strong>{{.State}}</strong> | Generated: {{.GeneratedAt}}</p>
{{if .RunResult}}
<div class="summary">
  <div><span class="stat pass">{{.RunResult.Passed}}</span><br>Passed</div>
  <div><span class="stat fail">{{.RunResult.Failed}}</span><br>Failed</div>
  <div><span class="stat">{{.RunResult.Total}}</span><br>Total</div>
</div>
{{if .RunResult.Failures}}
<h2>Failures</h2>
<table>
<tr><th>Test</th><th>Error</th></tr>
{{range .RunResult.Failures}}
<tr><td>{{.Test}}</td><td>{{.Message}}</td></tr>
{{end}}
</table>
{{end}}
{{else}}
<p>No results available.</p>
{{end}}
{{if .TestPlan}}
<h2>Test Plan</h2>
<p>{{.TestPlan.Summary}}</p>
<ul>
{{range .TestPlan.Scenarios}}
<li><strong>{{.Name}}</strong> ({{.Priority}})</li>
{{end}}
</ul>
{{end}}
</body>
</html>`))

// ReportData adalah data yang dikirim ke template HTML
type ReportData struct {
	ID          string
	State       string
	RunResult   *agent.RunResult
	TestPlan    *agent.TestPlan
	GeneratedAt string
}

// GenerateHTML menulis laporan HTML ke writer dari data test run
func GenerateHTML(w io.Writer, run *agent.TestRun) error {
	data := ReportData{
		ID:          run.ID,
		State:       string(run.State),
		RunResult:   run.RunResult,
		TestPlan:    run.TestPlan,
		GeneratedAt: time.Now().Format(time.RFC3339),
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
