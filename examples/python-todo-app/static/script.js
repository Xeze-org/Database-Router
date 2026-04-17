const API_BASE = '/api/todos';
const todoForm = document.getElementById('todo-form');
const taskInput = document.getElementById('task-input');
const todoList = document.getElementById('todo-list');

// Fetch and render todos immediately
fetchTodos();

async function fetchTodos() {
    try {
        const res = await fetch(API_BASE);
        const todos = await res.json();
        renderTodos(todos);
    } catch (err) {
        console.error("Failed to fetch tasks", err);
    }
}

function renderTodos(todos) {
    todoList.innerHTML = '';
    
    if (todos.length === 0) {
        todoList.innerHTML = `
            <div class="empty-state">
                <p>Everything is done! Relax.</p>
            </div>
        `;
        return;
    }

    todos.forEach(todo => {
        const li = document.createElement('li');
        if (todo.completed) li.classList.add('completed');
        
        li.innerHTML = `
            <div class="checkbox-container" onclick="toggleTodo(${todo.id}, ${!todo.completed})">
                <div class="custom-checkbox">
                    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                </div>
            </div>
            <span class="task-text">${escapeHtml(todo.task)}</span>
            <button class="delete-btn" onclick="deleteTodo(${todo.id})">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18"></path><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
            </button>
        `;
        todoList.appendChild(li);
    });
}

todoForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const task = taskInput.value.trim();
    if (!task) return;
    
    taskInput.value = ''; // clear input immediately for snappy UX
    
    try {
        await fetch(API_BASE, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ task })
        });
        fetchTodos();
    } catch (err) {
        console.error("Failed to add task", err);
    }
});

async function toggleTodo(id, completed) {
    try {
        await fetch(`${API_BASE}/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ completed })
        });
        fetchTodos();
    } catch (err) {
        console.error("Failed to toggle task", err);
    }
}

async function deleteTodo(id) {
    try {
        await fetch(`${API_BASE}/${id}`, {
            method: 'DELETE'
        });
        fetchTodos();
    } catch (err) {
        console.error("Failed to delete task", err);
    }
}

// XSS Prevention helper
function escapeHtml(unsafe) {
    return unsafe
         .replace(/&/g, "&amp;")
         .replace(/</g, "&lt;")
         .replace(/>/g, "&gt;")
         .replace(/"/g, "&quot;")
         .replace(/'/g, "&#039;");
}
