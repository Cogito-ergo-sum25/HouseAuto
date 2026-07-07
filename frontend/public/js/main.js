// -- Lógica de Apertura del Portón --
async function abrirPorton() {
    const pass = document.getElementById('password').value;
    const statusDiv = document.getElementById('status');
    const btn = document.getElementById('abrirBtn');

    if(!pass) { 
        statusDiv.innerText = "⚠️ Por favor, escribe la contraseña"; 
        statusDiv.style.color = "#f87171"; // red
        return; 
    }
    
    statusDiv.innerText = "⏳ Enviando comando...";
    statusDiv.style.color = "#fbbf24"; // yellow
    btn.disabled = true;

    const formData = new FormData();
    formData.append('password', pass);

    try {
        const response = await fetch('/api/abrir', { method: 'POST', body: formData });
        const result = await response.text();
        
        if (response.ok) {
            statusDiv.innerText = "✅ " + result;
            statusDiv.style.color = "#34d399"; // green
        } else {
            statusDiv.innerText = "❌ " + result;
            statusDiv.style.color = "#f87171"; // red
        }
    } catch (error) {
        statusDiv.innerText = "❌ Error al conectar con el servidor";
        statusDiv.style.color = "#f87171";
    } finally {
        setTimeout(() => { 
            btn.disabled = false; 
            statusDiv.innerText = ""; 
        }, 3000); // Limpiar después de 3 segundos
    }
}

function toggleTerminal() {
    const term = document.getElementById('terminalContainer');
    const btn = document.getElementById('toggleTerminalBtn');
    if (term.style.display === 'none') {
        term.style.display = 'flex';
        btn.innerText = 'Ocultar Consola Serial';
        term.querySelector('.terminal-output').scrollTop = term.querySelector('.terminal-output').scrollHeight;
    } else {
        term.style.display = 'none';
        btn.innerText = 'Mostrar Consola Serial';
    }
}

// -- Lógica de Consola (Server-Sent Events) --
const terminal = document.getElementById('terminalOutput');

function appendLog(message) {
    if (!terminal) return;
    const time = new Date().toLocaleTimeString([], { hour12: false });
    const line = document.createElement('div');
    
    let content = `[${time}] ${message}`;
    
    // Aplicar colores básicos a la salida
    if (message.includes("[+]")) {
        line.style.color = "#34d399"; // verde
    } else if (message.includes("[-]")) {
        line.style.color = "#f87171"; // rojo
    } else if (message.toLowerCase().includes("error") || message.toLowerCase().includes("falló")) {
        line.style.color = "#f87171"; // rojo
    } else {
        line.style.color = "#e2e8f0"; // gris claro por defecto
    }

    line.innerText = content;
    terminal.appendChild(line);
    
    // Mantener un máximo de 50 líneas para no sobrecargar el DOM
    if (terminal.childElementCount > 50) {
        terminal.removeChild(terminal.firstChild);
    }

    // Auto-scroll al fondo
    terminal.scrollTop = terminal.scrollHeight;
}

function conectarLogs() {
    const evtSource = new EventSource('/api/logs');

    evtSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            if (data.type === 'log') {
                appendLog(data.message);
            } else if (data.type === 'status') {
                const badge = document.getElementById('espStatusIndicator');
                if (badge) {
                    if (data.value === 'online') {
                        badge.className = 'status-badge online';
                        badge.innerText = 'ONLINE';
                        appendLog("[+] ESP32 Conectado (ONLINE)");
                    } else {
                        badge.className = 'status-badge offline';
                        badge.innerText = 'OFFLINE';
                        appendLog("[-] ESP32 Desconectado (OFFLINE)");
                    }
                }
            }
        } catch (e) {
            appendLog(event.data); // Fallback
        }
    };

    evtSource.onerror = function(err) {
        appendLog("[-] Conexión con el servidor perdida. Reconectando en 5s...");
        evtSource.close();
        setTimeout(conectarLogs, 5000);
    };
}

// Iniciar la conexión cuando carga la página
window.addEventListener('load', conectarLogs);
