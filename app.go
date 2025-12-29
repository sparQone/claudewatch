package main

import (
	"context"
	"os/exec"
	"runtime"
	"sync"
	"time"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	monitor      *Monitor
	alertTracker map[string]map[int]bool // sessionID -> threshold -> alerted
	mu           sync.Mutex
	stopChan     chan struct{}
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		alertTracker: make(map[string]map[int]bool),
		stopChan:     make(chan struct{}),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.monitor = NewMonitor()

	// Start background monitoring for alerts
	go a.backgroundMonitor()
}

// domReady is called when the DOM is ready, positions window in top right
func (a *App) domReady(ctx context.Context) {
	// Get screen dimensions
	screens, err := wailsRuntime.ScreenGetAll(ctx)
	if err != nil || len(screens) == 0 {
		return
	}

	// Use primary screen
	screen := screens[0]
	for _, s := range screens {
		if s.IsPrimary {
			screen = s
			break
		}
	}

	// Window width (must match main.go)
	windowWidth := 240

	// Calculate top right position with small margin
	margin := 20
	x := screen.Size.Width - windowWidth - margin
	y := margin

	wailsRuntime.WindowSetPosition(ctx, x, y)
}

// shutdown is called when the app closes
func (a *App) shutdown(ctx context.Context) {
	close(a.stopChan)
}

// GetSessions returns all active Claude sessions with their context usage
func (a *App) GetSessions() []SessionInfo {
	return a.monitor.GetActiveSessions()
}

// backgroundMonitor checks for threshold crossings and triggers alerts
func (a *App) backgroundMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.stopChan:
			return
		case <-ticker.C:
			sessions := a.monitor.GetActiveSessions()
			for _, session := range sessions {
				a.checkThresholds(session)
			}
		}
	}
}

// checkThresholds checks if a session has crossed alert thresholds
func (a *App) checkThresholds(session SessionInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.alertTracker[session.ID]; !exists {
		a.alertTracker[session.ID] = make(map[int]bool)
	}

	thresholds := []int{75, 90}
	for _, threshold := range thresholds {
		if session.Percentage >= threshold && !a.alertTracker[session.ID][threshold] {
			a.alertTracker[session.ID][threshold] = true
			go a.triggerAlert(session, threshold)
		}
	}
}

// triggerAlert sends OS-specific notifications and sounds
func (a *App) triggerAlert(session SessionInfo, threshold int) {
	title := "Claude Watch"
	message := session.ProjectName + " - " + string(rune(threshold)) + "% context used"

	if threshold >= 90 {
		message = "⚠️ " + session.ProjectName + " - Context almost full!"
	} else {
		message = session.ProjectName + " - Save context soon!"
	}

	switch runtime.GOOS {
	case "darwin":
		a.macAlert(title, message, threshold)
	case "windows":
		a.windowsAlert(title, message, threshold)
	case "linux":
		a.linuxAlert(title, message, threshold)
	}
}

// macAlert triggers macOS notification and sound
func (a *App) macAlert(title, message string, threshold int) {
	// Play sound
	soundFile := "/System/Library/Sounds/Glass.aiff"
	if threshold >= 90 {
		soundFile = "/System/Library/Sounds/Sosumi.aiff"
	}
	exec.Command("afplay", soundFile).Start()

	// Show notification
	script := `display notification "` + message + `" with title "` + title + `" sound name "Glass"`
	exec.Command("osascript", "-e", script).Run()

	// Voice alert at 90%
	if threshold >= 90 {
		exec.Command("say", "Claude context almost full").Start()
	}
}

// windowsAlert triggers Windows notification and sound
func (a *App) windowsAlert(title, message string, threshold int) {
	// PowerShell toast notification
	ps := `
	[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
	[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null
	$template = @"
	<toast>
		<visual>
			<binding template="ToastText02">
				<text id="1">` + title + `</text>
				<text id="2">` + message + `</text>
			</binding>
		</visual>
		<audio src="ms-winsoundevent:Notification.Default"/>
	</toast>
"@
	$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
	$xml.LoadXml($template)
	$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)
	[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier("Claude Watch").Show($toast)
	`
	exec.Command("powershell", "-Command", ps).Run()

	// Play system sound
	exec.Command("powershell", "-Command", "[console]::beep(800,300)").Start()
}

// linuxAlert triggers Linux notification and sound
func (a *App) linuxAlert(title, message string, threshold int) {
	// Try notify-send (most common)
	exec.Command("notify-send", title, message).Run()

	// Try to play a sound
	sounds := []string{
		"/usr/share/sounds/freedesktop/stereo/complete.oga",
		"/usr/share/sounds/gnome/default/alerts/glass.ogg",
		"/usr/share/sounds/ubuntu/stereo/message.ogg",
	}

	for _, sound := range sounds {
		if err := exec.Command("paplay", sound).Start(); err == nil {
			break
		}
	}

	// Fallback to terminal bell
	exec.Command("echo", "-e", "\a").Run()
}

// ResetAlerts clears alert history (useful when starting new sessions)
func (a *App) ResetAlerts() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.alertTracker = make(map[string]map[int]bool)
}
