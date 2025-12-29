// Update interval in milliseconds
const UPDATE_INTERVAL = 2000;

// Format token count (e.g., 133000 -> "133k")
function formatTokens(tokens) {
    if (tokens >= 1000) {
        return Math.floor(tokens / 1000) + 'k';
    }
    return tokens.toString();
}

// Get state class based on percentage
function getStateClass(percentage) {
    if (percentage >= 90) return 'state-critical';
    if (percentage >= 75) return 'state-danger';
    if (percentage >= 60) return 'state-warning';
    return 'state-good';
}

// Render a single session
function renderSession(session) {
    const stateClass = getStateClass(session.percentage);
    
    return `
        <div class="session ${stateClass}">
            <div class="session-header">
                <span class="project-name" title="${session.projectPath}">${session.projectName}</span>
                <span class="percentage">${session.percentage}%</span>
            </div>
            <div class="progress-container">
                <div class="progress-bar" style="width: ${session.percentage}%"></div>
            </div>
            <div class="session-footer">
                <span>Used: ${formatTokens(session.usedTokens)}</span>
                <span>Free: ${formatTokens(session.freeTokens)}</span>
            </div>
        </div>
    `;
}

// Render empty state
function renderEmpty() {
    return `
        <div class="empty-state">
            <div class="icon">💤</div>
            <p>No active Claude sessions<br>Sessions modified in the last 60 min will appear here</p>
        </div>
    `;
}

// Update the UI with session data
async function updateSessions() {
    const container = document.getElementById('sessions');
    const status = document.getElementById('status');
    const lastUpdate = document.getElementById('lastUpdate');
    
    try {
        // Call Go backend
        const sessions = await window.go.main.App.GetSessions();
        
        if (!sessions || sessions.length === 0) {
            container.innerHTML = renderEmpty();
            status.textContent = 'No sessions';
        } else {
            container.innerHTML = sessions.map(renderSession).join('');
            status.textContent = `${sessions.length} session${sessions.length > 1 ? 's' : ''}`;
        }
        
        // Update timestamp
        const now = new Date();
        lastUpdate.textContent = `Updated ${now.toLocaleTimeString()}`;
        
    } catch (error) {
        console.error('Error fetching sessions:', error);
        status.textContent = 'Error';
    }
}

// Reset alert history
async function resetAlerts() {
    try {
        await window.go.main.App.ResetAlerts();
        // Visual feedback
        const btn = document.querySelector('footer button');
        btn.textContent = '✓';
        setTimeout(() => { btn.textContent = '🔔'; }, 1000);
    } catch (error) {
        console.error('Error resetting alerts:', error);
    }
}

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    // Initial update
    updateSessions();
    
    // Periodic updates
    setInterval(updateSessions, UPDATE_INTERVAL);
});
