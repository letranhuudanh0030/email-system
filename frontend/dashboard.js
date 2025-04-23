let totalEmails = 0;

async function uploadCSV() {
    const file = document.getElementById('csvFile').files[0];
    if (!file) return;

    const formData = new FormData();
    formData.append('emails', file);

    try {
        const response = await fetch('http://localhost:8080/upload', {
            method: 'POST',
            body: formData
        });
        const result = await response.text();
        alert(result);
        totalEmails = (await fetchStats()).sent + (await fetchStats()).failed;
        updateDashboard();
    } catch (error) {
        alert('Upload failed: ' + error.message);
    }
}

async function fetchQueue() {
    const response = await fetch('http://localhost:8080/queue');
    return await response.json();
}

async function fetchStats() {
    const response = await fetch('http://localhost:8080/stats');
    return await response.json();
}

function updateDashboard() {
    fetchQueue().then(data => {
        document.getElementById('queueLength').textContent = data.queue_length;
        const progress = ((totalEmails - data.queue_length) / totalEmails * 100).toFixed(1);
        document.getElementById('queueProgress').style.width = progress + '%';
    });

    fetchStats().then(data => {
        document.getElementById('sentCount').textContent = data.sent;
        document.getElementById('failedCount').textContent = data.failed;
    });
}

// Auto-refresh every 2 seconds
setInterval(updateDashboard, 2000);