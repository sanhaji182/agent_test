package drift

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Detector analyzes changed files from a push and records drifts in the store.
type Detector struct {
	store *Store
}

func NewDetector(store *Store) *Detector { return &Detector{store: store} }

// DetectDrift compares changed files against test conventions and the cloned
// repository on disk (repoDir, may be empty if not cloned). Detected drifts
// are stored and returned.
func (d *Detector) DetectDrift(repository, repoDir string, added, modified, removed []string) []Drift {
	changed := map[string]bool{}
	for _, f := range added {
		changed[f] = true
	}
	for _, f := range modified {
		changed[f] = true
	}
	removedSet := map[string]bool{}
	for _, f := range removed {
		removedSet[f] = true
	}

	var found []Drift

	for f := range changed {
		if !isSourceFile(f) || IsTestFile(f) {
			continue
		}
		candidates := TestCandidates(f)
		if anyIn(candidates, changed) {
			continue // test updated together with source
		}
		if repoDir != "" && anyExists(repoDir, candidates) {
			found = append(found, Drift{
				Repository:  repository,
				Type:        TypeOutdatedTest,
				FilePath:    f,
				Description: fmt.Sprintf("source %s changed but its test was not updated", f),
				Severity:    SeverityMedium,
			})
			continue
		}
		found = append(found, Drift{
			Repository:  repository,
			Type:        TypeMissingTest,
			FilePath:    f,
			Description: fmt.Sprintf("no test found for changed source %s", f),
			Severity:    SeverityHigh,
		})
	}

	for _, f := range removed {
		if IsTestFile(f) {
			src := sourceCandidate(f)
			if src != "" && !removedSet[src] {
				found = append(found, Drift{
					Repository:  repository,
					Type:        TypeRemovedTest,
					FilePath:    f,
					Description: fmt.Sprintf("test %s was removed but source %s still exists", f, src),
					Severity:    SeverityHigh,
				})
			}
			continue
		}
		if isSourceFile(f) && repoDir != "" && anyExists(repoDir, TestCandidates(f)) {
			found = append(found, Drift{
				Repository:  repository,
				Type:        TypeRemovedTest,
				FilePath:    f,
				Description: fmt.Sprintf("source %s was removed but its test still exists (orphaned test)", f),
				Severity:    SeverityLow,
			})
		}
	}

	result := make([]Drift, 0, len(found))
	for _, dr := range found {
		if d.store.HasPending(dr.Repository, dr.Type, dr.FilePath) {
			continue
		}
		result = append(result, *d.store.Add(dr))
	}
	return result
}

func anyIn(paths []string, set map[string]bool) bool {
	for _, p := range paths {
		if set[p] {
			return true
		}
	}
	return false
}

func anyExists(root string, paths []string) bool {
	base := filepath.Clean(root)
	for _, p := range paths {
		full := filepath.Clean(filepath.Join(base, filepath.FromSlash(p)))
		if !strings.HasPrefix(full, base+string(filepath.Separator)) {
			continue // webhook-supplied path escaping the repo dir
		}
		if _, err := os.Stat(full); err == nil {
			return true
		}
	}
	return false
}

func isSourceFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go", ".js", ".jsx", ".ts", ".tsx", ".py", ".php":
		return true
	}
	return false
}

// IsTestFile reports whether path follows a known test-file convention.
func IsTestFile(path string) bool {
	base := filepath.Base(path)
	dir := filepath.ToSlash(filepath.Dir(path))
	switch strings.ToLower(filepath.Ext(path)) {
	case ".go":
		return strings.HasSuffix(base, "_test.go")
	case ".js", ".jsx", ".ts", ".tsx":
		return strings.Contains(base, ".test.") || strings.Contains(base, ".spec.") ||
			strings.Contains(dir+"/", "__tests__/")
	case ".py":
		return strings.HasPrefix(base, "test_") || strings.HasSuffix(trimExt(base), "_test")
	case ".php":
		return strings.HasSuffix(trimExt(base), "Test")
	}
	return false
}

// TestCandidates returns conventional test file paths for a source file.
func TestCandidates(path string) []string {
	path = filepath.ToSlash(path)
	dir := filepath.ToSlash(filepath.Dir(path))
	base := filepath.Base(path)
	name := trimExt(base)
	ext := filepath.Ext(base)
	join := func(d, f string) string {
		if d == "." || d == "" {
			return f
		}
		return d + "/" + f
	}
	switch strings.ToLower(ext) {
	case ".go":
		return []string{join(dir, name+"_test.go")}
	case ".js", ".jsx", ".ts", ".tsx":
		return []string{
			join(dir, name+".test"+ext),
			join(dir, name+".spec"+ext),
			join(dir, "__tests__/"+name+".test"+ext),
			join(dir, "__tests__/"+name+".spec"+ext),
		}
	case ".py":
		return []string{
			join(dir, "test_"+base),
			join(dir, name+"_test.py"),
			join(dir, "tests/test_"+base),
			"tests/test_" + base,
		}
	case ".php":
		return []string{
			join(dir, name+"Test.php"),
			"tests/" + name + "Test.php",
			"tests/Unit/" + name + "Test.php",
			"tests/Feature/" + name + "Test.php",
		}
	}
	return nil
}

// sourceCandidate maps a test file back to its conventional source path, or
// "" if the mapping is ambiguous.
func sourceCandidate(testPath string) string {
	testPath = filepath.ToSlash(testPath)
	dir := filepath.ToSlash(filepath.Dir(testPath))
	base := filepath.Base(testPath)
	name := trimExt(base)
	ext := filepath.Ext(base)
	join := func(d, f string) string {
		if d == "." || d == "" {
			return f
		}
		return d + "/" + f
	}
	switch strings.ToLower(ext) {
	case ".go":
		if strings.HasSuffix(base, "_test.go") {
			return join(dir, strings.TrimSuffix(base, "_test.go")+".go")
		}
	case ".js", ".jsx", ".ts", ".tsx":
		n := strings.TrimSuffix(strings.TrimSuffix(name, ".test"), ".spec")
		if n != name {
			d := strings.TrimSuffix(dir, "/__tests__")
			return join(d, n+ext)
		}
	case ".py":
		if strings.HasPrefix(base, "test_") {
			return join(dir, strings.TrimPrefix(base, "test_"))
		}
		if strings.HasSuffix(name, "_test") {
			return join(dir, strings.TrimSuffix(name, "_test")+".py")
		}
	case ".php":
		if strings.HasSuffix(name, "Test") {
			return join(dir, strings.TrimSuffix(name, "Test")+".php")
		}
	}
	return ""
}

func trimExt(base string) string {
	return strings.TrimSuffix(base, filepath.Ext(base))
}
