const SEMAPHORE = {
  idle:       { sym: '\u26AA', color: '#5c5c6e' },
  pending:    { sym: '\U0001F535', color: '#3b82f6' },
  in_progress: { sym: '\U0001F7E1', color: '#dcb43c' },
  completed:  { sym: '\U0001F7E2', color: '#50c878' },
  failed:     { sym: '\U0001F534', color: '#ff2830' },
}

function el(id) { return document.getElementById(id) }

function updateDashboard(data) {
  const sem = SEMAPHORE[data.status] || SEMAPHORE.idle

  el('version-badge').textContent = 'v' + data.version
  el('status-badge').textContent = sem.sym + ' ' + data.status.toUpperCase()
  el('status-badge').className = 'badge status status-' + data.status
  el('location-badge').textContent = '\U0001F4C1 ' + data.project

  el('metric-version').textContent = data.version
  el('metric-status').textContent = sem.sym
  el('metric-phase').textContent = 'PHASE ' + data.phase
  el('metric-atenea').textContent = (data.atenea_score || 0) + '/80'

  const health = Math.min(100, Math.max(0, data.health))
  const circumference = 100
  const offset = circumference - (health / 100) * circumference
  const ring = el('health-ring-fill')
  ring.setAttribute('stroke-dasharray', offset + ', ' + circumference)
  el('health-text').textContent = health + '%'

  let ringColor = '#50c878'
  if (health < 30) ringColor = '#ff2830'
  else if (health < 60) ringColor = '#dcb43c'
  ring.setAttribute('stroke', ringColor)

  el('task-title').textContent = 'CURRENT TASK'
  el('task-id').textContent = data.task_id || '--'
  el('task-desc').textContent = data.description || 'Waiting for data...'
  el('progress-pct').textContent = health + '%'
  el('progress-fill').style.width = health + '%'

  let progressColor = 'linear-gradient(90deg, #50c878, #50c878)'
  if (health < 30) progressColor = 'linear-gradient(90deg, #ff2830, #ff6b6b)'
  else if (health < 60) progressColor = 'linear-gradient(90deg, #dcb43c, #e8c84a)'
  el('progress-fill').style.background = progressColor

  const agents = data.agents && data.agents.length
    ? data.agents.map(a => '\u25CF ' + a).join(' ')
    : '--'
  el('task-agents').textContent = agents

  const artifactCount = data.artifacts || 0
  el('task-artifacts').textContent = artifactCount + ' generated'

  renderTimeline(data)

  el('next-action').textContent = data.next_action || '--'
  el('last-update').textContent = data.timestamp || '--:--:--'
  el('poll-info').textContent = 'Polling every ' + (data.interval || 5) + 's'
}

function renderTimeline(data) {
  const tl = el('timeline')
  tl.innerHTML = ''

  const events = []

  if (data.agents && data.agents.length > 0) {
    for (const agent of data.agents) {
      events.push({
        type: 'agent',
        label: agent,
        desc: 'completed'
      })
    }
  }

  if (data.telemetry_events && data.telemetry_events.length > 0) {
    for (const ev of data.telemetry_events) {
      let label = ev.source || ev.type
      let desc = ev.data || ev.type
      events.push({ type: 'telemetry', label, desc })
    }
  }

  events.reverse()

  if (events.length === 0) {
    tl.innerHTML = '<div class="timeline-empty">No events yet</div>'
    return
  }

  const max = Math.min(events.length, 20)
  for (let i = 0; i < max; i++) {
    const ev = events[i]
    const item = document.createElement('div')
    item.className = 'timeline-item'
    const dotClass = ev.type === 'agent' ? 'timeline-dot agent' : 'timeline-dot telemetry'
    item.innerHTML = '<span class="' + dotClass + '"></span><span class="timeline-agent">' + ev.label + '</span><span class="timeline-action">\u2014 ' + escHtml(ev.desc) + '</span>'
    tl.appendChild(item)
  }
}

function escHtml(s) {
  if (!s) return ''
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
}

function connectSSE() {
  const es = new EventSource('/api/events')
  es.onmessage = function(e) {
    try {
      const data = JSON.parse(e.data)
      updateDashboard(data)
    } catch (err) {
      console.error('SSE parse error:', err)
    }
  }
  es.onerror = function() {
    es.close()
    setTimeout(connectSSE, 2000)
  }
}

connectSSE()
