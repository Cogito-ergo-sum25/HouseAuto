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

// Iniciar la conexión cuando carga el hub
window.addEventListener('load', conectarStatusEnVivo);
