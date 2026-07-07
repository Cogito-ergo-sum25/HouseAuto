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

// -- Lógica del Hub (Estatus en vivo) --
function conectarStatusEnVivo() {
    const evtSource = new EventSource('/api/logs');

    evtSource.onmessage = function(event) {
        try {
            const data = JSON.parse(event.data);
            if (data.type === 'status' || data.type === 'init_status') {
                const badge = document.getElementById('espStatusIndicator');
                if (badge) {
                    if (data.value === 'online') {
                        badge.className = 'status-badge online';
                        badge.innerText = 'ONLINE';
                    } else {
                        badge.className = 'status-badge offline';
                        badge.innerText = 'OFFLINE';
                    }
                }
            }
        } catch (e) {
            // Ignorar los logs normales en el Hub
        }
    };

    evtSource.onerror = function(err) {
        const badge = document.getElementById('espStatusIndicator');
        if (badge) {
            badge.className = 'status-badge offline';
            badge.innerText = 'OFFLINE';
        }
        evtSource.close();
        setTimeout(conectarStatusEnVivo, 5000);
    };
}

// -- Lógica del Reloj Hub (CDMX GMT-6) --
function actualizarReloj() {
    const timeDiv = document.getElementById('hub-time');
    const dateDiv = document.getElementById('hub-date');
    if (!timeDiv || !dateDiv) return;

    // Configurar para usar siempre el timezone de America/Mexico_City
    const optionsTime = { 
        timeZone: 'America/Mexico_City', 
        hour: '2-digit', 
        minute: '2-digit', 
        second: '2-digit',
        hour12: false
    };
    
    const optionsDate = { 
        timeZone: 'America/Mexico_City',
        weekday: 'long', 
        day: 'numeric', 
        month: 'long' 
    };

    const formatterTime = new Intl.DateTimeFormat('es-MX', optionsTime);
    const formatterDate = new Intl.DateTimeFormat('es-MX', optionsDate);
    
    const now = new Date();
    timeDiv.innerText = formatterTime.format(now);
    dateDiv.innerText = formatterDate.format(now);
}

// Iniciar la conexión y el reloj cuando carga el hub
window.addEventListener('load', () => {
    conectarStatusEnVivo();
    actualizarReloj();
    setInterval(actualizarReloj, 1000);
});
