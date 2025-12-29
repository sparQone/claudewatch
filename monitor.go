package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	MaxContextTokens = 200000
	MaxProjects      = 10
	ActiveMinutes    = 60
)

// SessionInfo represents a Claude Code session's context usage
type SessionInfo struct {
	ID          string `json:"id"`
	ProjectName string `json:"projectName"`
	ProjectPath string `json:"projectPath"`
	UsedTokens  int    `json:"usedTokens"`
	FreeTokens  int    `json:"freeTokens"`
	Percentage  int    `json:"percentage"`
	LastUpdated string `json:"lastUpdated"`
}

// Monitor handles scanning and parsing Claude session files
type Monitor struct {
	claudeDir string
}

// Message structures for parsing JSONL
type AssistantMessage struct {
	Type    string `json:"type"`
	Cwd     string `json:"cwd"`
	Message struct {
		Usage *Usage `json:"usage"`
	} `json:"message"`
}

type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
}

// NewMonitor creates a new monitor instance
func NewMonitor() *Monitor {
	homeDir, _ := os.UserHomeDir()
	claudeDir := filepath.Join(homeDir, ".claude", "projects")

	// Handle Windows path if needed
	if runtime.GOOS == "windows" {
		claudeDir = filepath.Join(homeDir, ".claude", "projects")
	}

	return &Monitor{
		claudeDir: claudeDir,
	}
}

// GetActiveSessions returns all active Claude sessions
func (m *Monitor) GetActiveSessions() []SessionInfo {
	var sessions []SessionInfo

	// Find all project directories
	projectDirs, err := os.ReadDir(m.claudeDir)
	if err != nil {
		return sessions
	}

	type sessionFile struct {
		path    string
		modTime time.Time
		project string
	}

	var allSessions []sessionFile

	cutoff := time.Now().Add(-ActiveMinutes * time.Minute)

	for _, projectDir := range projectDirs {
		if !projectDir.IsDir() {
			continue
		}

		projectPath := filepath.Join(m.claudeDir, projectDir.Name())
		files, err := os.ReadDir(projectPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			// Skip agent files and non-jsonl files
			if !strings.HasSuffix(file.Name(), ".jsonl") || strings.HasPrefix(file.Name(), "agent-") {
				continue
			}

			filePath := filepath.Join(projectPath, file.Name())
			info, err := file.Info()
			if err != nil {
				continue
			}

			// Only include recently modified files
			if info.ModTime().After(cutoff) {
				allSessions = append(allSessions, sessionFile{
					path:    filePath,
					modTime: info.ModTime(),
					project: projectDir.Name(),
				})
			}
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(allSessions, func(i, j int) bool {
		return allSessions[i].modTime.After(allSessions[j].modTime)
	})

	// Keep only the most recent session per project
	seenProjects := make(map[string]bool)
	var uniqueSessions []sessionFile
	for _, sf := range allSessions {
		if !seenProjects[sf.project] {
			seenProjects[sf.project] = true
			uniqueSessions = append(uniqueSessions, sf)
		}
	}
	allSessions = uniqueSessions

	// Limit to MaxProjects
	if len(allSessions) > MaxProjects {
		allSessions = allSessions[:MaxProjects]
	}

	// Parse each session
	for _, sf := range allSessions {
		info := m.parseSession(sf.path, sf.project)
		if info != nil {
			info.LastUpdated = sf.modTime.Format("15:04:05")
			sessions = append(sessions, *info)
		}
	}

	return sessions
}

// parseSession parses a single session file and returns context stats
func (m *Monitor) parseSession(filePath, projectDir string) *SessionInfo {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var totalOutputTokens int
	var lastUsage *Usage
	var cwd string

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		var msg AssistantMessage
		if err := json.Unmarshal(scanner.Bytes(), &msg); err != nil {
			continue
		}

		// Capture cwd from the first message that has it
		if cwd == "" && msg.Cwd != "" {
			cwd = msg.Cwd
		}

		if msg.Type == "assistant" && msg.Message.Usage != nil {
			totalOutputTokens += msg.Message.Usage.OutputTokens
			lastUsage = msg.Message.Usage
		}
	}

	if lastUsage == nil {
		return nil
	}

	// Calculate total context usage
	currentInput := lastUsage.InputTokens + lastUsage.CacheReadInputTokens + lastUsage.CacheCreationInputTokens
	totalUsed := totalOutputTokens + currentInput
	percentage := (totalUsed * 100) / MaxContextTokens
	freeTokens := MaxContextTokens - totalUsed

	if freeTokens < 0 {
		freeTokens = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	// Use folder name from cwd if available, otherwise fall back to directory name
	projectName := filepath.Base(cwd)
	if projectName == "" || projectName == "." {
		projectName = projectDir
	}

	return &SessionInfo{
		ID:          filepath.Base(filePath),
		ProjectName: projectName,
		ProjectPath: projectDir,
		UsedTokens:  totalUsed,
		FreeTokens:  freeTokens,
		Percentage:  percentage,
	}
}

