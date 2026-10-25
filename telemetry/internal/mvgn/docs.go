package mvgn

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"time"
)

// cacheEntry holds a cached value with its expiry time.
// DocsReader reuses the same cache entry shape as ContextReader.

// DocsReader reads MVGN project documentation files (PRD, tasks, changelog)
// with TTL-based caching (default 5s). All methods degrade gracefully on
// missing files by returning empty/nil defaults.
type DocsReader struct {
	projectPath string
	cache       map[string]cacheEntry
	ttl         time.Duration
}

// PRDSummary contains a count and the raw lines of functional requirements
// extracted from docs/01_prd.md.
type PRDSummary struct {
	RFCount      int
	Requirements []string
}

// ADREntry represents a single architecture decision record.
type ADREntry struct {
	Number string
	Title  string
	Status string
}

// TaskSummary holds counters parsed from docs/03_tasks.md.
type TaskSummary struct {
	Completed int
	Pending   int
	Total     int
	Blocked   int
}

// LastTask holds info about the most recently completed task.
type LastTask struct {
	ID          string // T-XXX
	Description string // human-readable title
	Author      string // who completed it (e.g., "Ncape")
}

// ChangelogEntry represents a single version entry from docs/04_changelog.md.
type ChangelogEntry struct {
	Version string
	Date    string
	Message string
}

// NewDocsReader creates a DocsReader rooted at projectPath.
func NewDocsReader(projectPath string) *DocsReader {
	return &DocsReader{
		projectPath: projectPath,
		cache:       make(map[string]cacheEntry),
		ttl:         5 * time.Second,
	}
}

// ReadPRD extracts functional requirement references from docs/01_prd.md.
// Returns an empty summary on missing or unparseable file.
func (r *DocsReader) ReadPRD() (*PRDSummary, error) {
	f, err := os.Open(r.projectPath + "/docs/01_prd.md")
	if err != nil {
		return &PRDSummary{}, nil
	}
	defer f.Close()

	summary := &PRDSummary{}
	rfRe := regexp.MustCompile(`\*\*RF-\d+`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if rfRe.MatchString(line) {
			summary.RFCount++
			summary.Requirements = append(summary.Requirements, strings.TrimSpace(line))
		}
	}

	return summary, nil
}

// ReadTasks parses docs/03_tasks.md and counts completed, pending, and blocked tasks.
// Counts task-level markers only:
//   - Lines starting with "- [x]" or "[x]" (standard + backtick completion)
//   - Ignores acceptance criteria inside task bodies
// Returns an empty summary on missing file.
func (r *DocsReader) ReadTasks() (*TaskSummary, error) {
	f, err := os.Open(r.projectPath + "/docs/03_tasks.md")
	if err != nil {
		return &TaskSummary{}, nil
	}
	defer f.Close()

	ts := &TaskSummary{}
	inCriteria := false    // are we inside a criteria block?

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detect task header: ### T-XXX — resets criteria flag
		if strings.HasPrefix(line, "### T-") {
			inCriteria = false
			continue
		}

		// Detect criteria block: lines starting with **Criterios:**
		if strings.HasPrefix(line, "**Criterios:") || line == "**Criterios:**" {
			inCriteria = true
			continue
		}

		// Section header ends criteria block
		if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "---") {
			inCriteria = false
			continue
		}

		// Always count completion markers regardless of criteria blocks
		// Standard:   - [x] Some task description
		// Backtick:   [x] Completada — T-XXX (common in MVGN docs)
		trimmed := strings.TrimLeft(line, "`")
		if strings.HasPrefix(trimmed, "[x]") || strings.HasPrefix(trimmed, "[X]") ||
		   strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			ts.Completed++
			continue
		}

		// Inside criteria: each "- [ ]" is an acceptance criterion, NOT a task
		if inCriteria {
			continue
		}

		// Pending/blocked markers only count outside criteria blocks
		if strings.HasPrefix(trimmed, "- [ ]") {
			ts.Pending++
		} else if strings.HasPrefix(trimmed, "- [!]") {
			ts.Blocked++
		}
	}

	ts.Total = ts.Completed + ts.Pending + ts.Blocked
	return ts, nil
}

// ReadLastCompletedTask finds the most recently completed task from docs/03_tasks.md.
// It scans for the last task header (### T-XXX — Description) that has a
// completion marker [x] in its body, and returns the task ID, description,
// and author (from the changelog if available).
func (r *DocsReader) ReadLastCompletedTask() (*LastTask, error) {
	f, err := os.Open(r.projectPath + "/docs/03_tasks.md")
	if err != nil {
		return nil, nil
	}
	defer f.Close()

	var lastTask *LastTask
	var currentID, currentDesc string
	foundCompletion := false

	taskHeaderRe := regexp.MustCompile(`^### (T-\d+) — (.+)$`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for task header
		if m := taskHeaderRe.FindStringSubmatch(line); len(m) >= 3 {
			// If previous task had a completion, save it
			if foundCompletion && currentID != "" {
				lastTask = &LastTask{
					ID:          currentID,
					Description: currentDesc,
					Author:      "Ncape", // default human author
				}
			}
			currentID = m[1]
			currentDesc = strings.TrimSpace(m[2])
			foundCompletion = false
			continue
		}

		// Check for completion marker in the task body
		// Format: [x] Completada — T-XXX description
		// Also: - [x], `[x]`, etc.
		trimmed := strings.TrimLeft(line, " `-")
		if strings.HasPrefix(trimmed, "[x]") || strings.HasPrefix(trimmed, "[X]") {
			foundCompletion = true
		}
	}

	// Last task in file
	if foundCompletion && currentID != "" {
		lastTask = &LastTask{
			ID:          currentID,
			Description: currentDesc,
			Author:      "Ncape",
		}
	}

	return lastTask, nil
}

// ReadChangelog reads the last N entries from docs/04_changelog.md.
// Returns nil on missing file.
func (r *DocsReader) ReadChangelog(limit int) ([]ChangelogEntry, error) {
	f, err := os.Open(r.projectPath + "/docs/04_changelog.md")
	if err != nil {
		return nil, nil
	}
	defer f.Close()

	var entries []ChangelogEntry
	versionRe := regexp.MustCompile(`## \[(\d+\.\d+\.\d+)\]`)
	dateRe := regexp.MustCompile(`- (\d{4}-\d{2}-\d{2})`)

	var current ChangelogEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if m := versionRe.FindStringSubmatch(line); len(m) > 1 {
			if current.Version != "" {
				entries = append(entries, current)
			}
			current = ChangelogEntry{Version: m[1]}
		} else if m := dateRe.FindStringSubmatch(line); len(m) > 1 {
			current.Date = m[1]
		} else if strings.HasPrefix(line, "-") && current.Version != "" {
			current.Message = strings.TrimPrefix(line, "- ")
		}
	}
	if current.Version != "" {
		entries = append(entries, current)
	}

	if len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}
