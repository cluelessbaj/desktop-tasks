// State
let appData = {
    tasks: [],
    notes: ""
};

let currentTab = 'tasks';
let noteMode = 'preview'; // 'preview' | 'edit'
let saveTimeout = null;

// Date formatter
function updateDateDisplay() {
    const now = new Date();
    const options = { weekday: 'long', month: 'short', day: 'numeric' };
    document.getElementById('header-date').textContent = now.toLocaleDateString('en-US', options);
}

// Draggable window support
document.getElementById('header-box').addEventListener('mousedown', (e) => {
    if (e.button === 0 && !e.target.closest('button')) {
        if (window.backend_startDrag) {
            window.backend_startDrag();
        }
    }
});

// Tab Switching
function switchTab(tab) {
    currentTab = tab;
    document.getElementById('tab-btn-tasks').classList.toggle('active', tab === 'tasks');
    document.getElementById('tab-btn-notes').classList.toggle('active', tab === 'notes');
    document.getElementById('pane-tasks').classList.toggle('active', tab === 'tasks');
    document.getElementById('pane-notes').classList.toggle('active', tab === 'notes');

    if (tab === 'notes') {
        renderNotes();
        if (noteMode === 'edit') {
            document.getElementById('notes-editor').focus();
        }
    } else {
        document.getElementById('task-input').focus();
    }
}

// Tasks Management
function renderTasks() {
    const list = document.getElementById('task-list');
    list.innerHTML = '';

    let pendingCount = 0;

    appData.tasks.forEach((task) => {
        if (!task.completed) pendingCount++;

        const row = document.createElement('div');
        row.className = `task-row ${task.completed ? 'completed' : ''}`;

        const checkbox = document.createElement('input');
        checkbox.type = 'checkbox';
        checkbox.className = 'task-checkbox';
        checkbox.checked = !!task.completed;
        checkbox.onchange = () => toggleTask(task.id);

        const label = document.createElement('div');
        label.className = 'task-label';
        label.textContent = task.text;

        row.appendChild(checkbox);
        row.appendChild(label);

        if (task.due_time) {
            const dueBadge = document.createElement('div');
            dueBadge.className = 'task-due';
            dueBadge.textContent = task.due_time;
            row.appendChild(dueBadge);
        }

        const delBtn = document.createElement('button');
        delBtn.className = 'task-delete-btn';
        delBtn.textContent = '✕';
        delBtn.title = 'Delete task';
        delBtn.onclick = (e) => {
            e.stopPropagation();
            deleteTask(task.id);
        };
        row.appendChild(delBtn);

        list.appendChild(row);
    });

    document.getElementById('pending-badge').textContent = `${pendingCount} pending`;
}

function parseDueTime(text) {
    const match = text.match(/@(\d{1,2}:\d{2})/);
    if (match) {
        const timeStr = match[1];
        const cleanText = text.replace(match[0], '').trim();
        return { text: cleanText, due: timeStr };
    }
    return { text: text.trim(), due: "" };
}

function addTask(rawText) {
    if (!rawText || !rawText.trim()) return;
    const { text, due } = parseDueTime(rawText);
    if (!text) return;

    const maxId = appData.tasks.reduce((m, t) => Math.max(m, t.id || 0), 0);
    const newTask = {
        id: maxId + 1,
        text: text,
        completed: false,
        due_time: due,
        reminded: false,
        created_at: Math.floor(Date.now() / 1000)
    };

    appData.tasks.push(newTask);
    renderTasks();
    saveData();
}

function toggleTask(id) {
    const task = appData.tasks.find(t => t.id === id);
    if (task) {
        task.completed = !task.completed;
        renderTasks();
        saveData();
    }
}

function deleteTask(id) {
    appData.tasks = appData.tasks.filter(t => t.id !== id);
    renderTasks();
    saveData();
}

// Add task input listeners
const taskInput = document.getElementById('task-input');
taskInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') {
        addTask(taskInput.value);
        taskInput.value = '';
    }
});

document.getElementById('task-add-btn').addEventListener('click', () => {
    addTask(taskInput.value);
    taskInput.value = '';
    taskInput.focus();
});

// Notes Management
const notesEditor = document.getElementById('notes-editor');
const notesPreview = document.getElementById('notes-preview');

function toggleNoteMode() {
    setNoteMode(noteMode === 'preview' ? 'edit' : 'preview');
}

function setNoteMode(mode) {
    noteMode = mode;
    const toggleBtn = document.getElementById('btn-mode-toggle');

    if (mode === 'preview') {
        toggleBtn.textContent = '✎ Edit';
        toggleBtn.classList.remove('active');
        notesEditor.style.display = 'none';
        notesPreview.style.display = 'block';
        renderNotes();
    } else {
        toggleBtn.textContent = '👁 Preview';
        toggleBtn.classList.add('active');
        notesPreview.style.display = 'none';
        notesEditor.style.display = 'block';
        notesEditor.value = appData.notes || '';
        notesEditor.focus();
        updateNoteMetrics();
    }
}

function renderNotes() {
    const raw = appData.notes || '';
    if (window.marked) {
        marked.setOptions({
            breaks: true,
            gfm: true
        });
        notesPreview.innerHTML = marked.parse(raw);

        // Make rendered checkboxes interactive
        const checkboxes = notesPreview.querySelectorAll('input[type="checkbox"]');
        checkboxes.forEach((cb, index) => {
            cb.removeAttribute('disabled');
            cb.style.cursor = 'pointer';
            cb.onchange = () => toggleMarkdownCheckbox(index, cb.checked);
        });
    } else {
        notesPreview.textContent = raw;
    }
    updateNoteMetrics();
}

function toggleMarkdownCheckbox(checkboxIndex, checked) {
    let currentIdx = 0;
    const lines = appData.notes.split('\n');
    for (let i = 0; i < lines.length; i++) {
        if (/^\s*-\s*\[([ xX])\]/.test(lines[i])) {
            if (currentIdx === checkboxIndex) {
                lines[i] = lines[i].replace(/^(\s*-\s*\[)([ xX])(\])/, `$1${checked ? 'x' : ' '}$3`);
                break;
            }
            currentIdx++;
        }
    }
    appData.notes = lines.join('\n');
    notesEditor.value = appData.notes;
    saveData();
    renderNotes();
}

notesEditor.addEventListener('input', () => {
    appData.notes = notesEditor.value;
    updateNoteMetrics();
    document.getElementById('notes-status').textContent = 'Saving...';

    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(() => {
        saveData();
        document.getElementById('notes-status').textContent = 'Auto-saved';
    }, 300);
});

function updateNoteMetrics() {
    const text = appData.notes || '';
    const chars = text.length;
    const words = text.trim() ? text.trim().split(/\s+/).length : 0;
    document.getElementById('notes-metrics').textContent = `${words} words • ${chars} chars`;
}

// Markdown formatting helpers
function insertMarkdown(prefix) {
    if (noteMode !== 'edit') setNoteMode('edit');
    const start = notesEditor.selectionStart;
    const end = notesEditor.selectionEnd;
    const val = notesEditor.value;
    notesEditor.value = val.substring(0, start) + prefix + val.substring(end);
    notesEditor.selectionStart = notesEditor.selectionEnd = start + prefix.length;
    notesEditor.focus();
    notesEditor.dispatchEvent(new Event('input'));
}

function wrapSelection(prefix, suffix) {
    if (noteMode !== 'edit') setNoteMode('edit');
    const start = notesEditor.selectionStart;
    const end = notesEditor.selectionEnd;
    const val = notesEditor.value;
    const selected = val.substring(start, end) || 'text';
    notesEditor.value = val.substring(0, start) + prefix + selected + suffix + val.substring(end);
    notesEditor.selectionStart = start + prefix.length;
    notesEditor.selectionEnd = start + prefix.length + selected.length;
    notesEditor.focus();
    notesEditor.dispatchEvent(new Event('input'));
}

function copyNotes() {
    navigator.clipboard.writeText(appData.notes || '').then(() => {
        const status = document.getElementById('notes-status');
        const oldText = status.textContent;
        status.textContent = 'Copied to clipboard!';
        setTimeout(() => { status.textContent = oldText; }, 2000);
    });
}

// Global Keyboard Shortcuts
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        if (window.backend_sinkDesktop) {
            window.backend_sinkDesktop();
        }
    } else if (e.ctrlKey && e.key === '1') {
        e.preventDefault();
        switchTab('tasks');
    } else if (e.ctrlKey && e.key === '2') {
        e.preventDefault();
        switchTab('notes');
    } else if (e.ctrlKey && e.key.toLowerCase() === 'e') {
        if (currentTab === 'notes') {
            e.preventDefault();
            toggleNoteMode();
        }
    }
});

// Backend Communication
function loadData() {
    if (window.backend_loadData) {
        window.backend_loadData().then((jsonStr) => {
            try {
                if (jsonStr) {
                    appData = JSON.parse(jsonStr);
                }
            } catch (err) {
                console.error("Failed to parse data:", err);
            }
            renderTasks();
            renderNotes();
        });
    }
}

function saveData() {
    if (window.backend_saveData) {
        window.backend_saveData(JSON.stringify(appData));
    }
}

// External IPC triggers from Go
window.onExternalToggle = function() {
    if (currentTab === 'tasks') {
        taskInput.focus();
    } else if (noteMode === 'edit') {
        notesEditor.focus();
    }
};

window.onExternalSwitchTab = function(tab) {
    switchTab(tab);
};

window.onExternalAddTask = function(text) {
    addTask(text);
    switchTab('tasks');
};

// Initialize
updateDateDisplay();
setInterval(updateDateDisplay, 60000);

window.addEventListener('DOMContentLoaded', () => {
    loadData();
});
