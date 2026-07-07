const terminal = document.getElementById('terminalOutput');

function appendLog(message, timeString) {
    if (!terminal) return;
    
    // Si no se proporciona hora, usar la actual
    let time = timeString;
    if (!time) {
        time = new Date().toLocaleTimeString([], { hour12: false });
    }
    
    const line = document.createElement('div');
    let content = `[${time}] ${message}`;
    
    // Aplicar colores básicos a la salida
    if (message.includes("[+]") || message.includes("ONLINE")) {
        line.style.color = "#34d399"; // verde
    } else if (message.includes("[-]") || message.includes("OFFLINE")) {
        line.style.color = "#f87171"; // rojo
    } else if (message.toLowerCase().includes("error") || message.toLowerCase().includes("falló")) {
        line.style.color = "#f87171"; // rojo
    } else {
        line.style.color = "#e2e8f0"; // gris claro por defecto
    }

    line.innerText = content;
    terminal.appendChild(line);
    
    // Limitar a 200 líneas en consola para no sobrecargar
    if (terminal.childElementCount > 200) {
        terminal.removeChild(terminal.firstChild);
    }

    // Auto-scroll al fondo
    terminal.scrollTop = terminal.scrollHeight;
}

function updateBadge(status) {
    const badge = document.getElementById('espStatusIndicator');
    if (badge) {
        if (status === 'online') {
            badge.className = 'status-badge online';
            badge.innerText = 'ONLINE';
        } else {
            badge.className = 'status-badge offline';
            badge.innerText = 'OFFLINE';
        }
    }
}

async function cargarHistorial() {
    try {
        const res = await fetch('/api/history');
        if (res.ok) {
            const history = await res.json();
            // Limpiar si hay algo
            terminal.innerHTML = '';
            appendLog("--- Cargando historial de eventos ---");
            
            if (history) {
                history.forEach(ev => {
                    const d = new Date(ev.Timestamp);
                    const timeStr = d.toLocaleTimeString([], { hour12: false });
                    
                    if (ev.Type === 'log') {
                        appendLog(ev.Message, timeStr);
                    } else if (ev.Type === 'status') {
                        if (ev.Message === 'online') {
                            appendLog("[+] ESP32 Conectado (ONLINE) - Histórico", timeStr);
                        } else {
                            appendLog("[-] ESP32 Desconectado (OFFLINE) - Histórico", timeStr);
                        }
                    }
                });
            }
            appendLog("--- Fin del historial ---");
        }
    } catch (e) {
        appendLog("[-] Error cargando el historial.");
    }
}

function conectarLogsEnVivo() {
    const evtSource = new EventSource('/api/logs');

    evtSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            if (data.type === 'log') {
                appendLog(data.message);
            } else if (data.type === 'init_status') {
                updateBadge(data.value);
            } else if (data.type === 'status') {
                updateBadge(data.value);
                if (data.value === 'online') {
                    appendLog("[+] ESP32 Conectado (ONLINE)");
                } else {
                    appendLog("[-] ESP32 Desconectado (OFFLINE)");
                }
            }
        } catch (e) {
            appendLog(event.data); // Fallback
        }
    };

    evtSource.onerror = function(err) {
        appendLog("[-] Conexión con el servidor perdida. Reconectando en 5s...");
        evtSource.close();
        setTimeout(conectarLogsEnVivo, 5000);
    };
}

window.addEventListener('load', async () => {
    await cargarHistorial();
    conectarLogsEnVivo();
});
