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
        appendLog(event.data);
    };

    evtSource.onerror = function(err) {
        appendLog("[-] Conexión con el servidor perdida. Reconectando en 5s...");
        evtSource.close();
        setTimeout(conectarLogs, 5000);
    };
}

// Iniciar la conexión cuando carga la página
window.addEventListener('load', conectarLogs);
