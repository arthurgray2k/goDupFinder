document.getElementById('scanForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const dirsInput = document.getElementById('directories').value;
    const workers = parseInt(document.getElementById('workers').value, 10);
    const minSize = parseInt(document.getElementById('minSize').value, 10);
    const maxDepth = parseInt(document.getElementById('maxDepth').value, 10);
    const algorithm = document.getElementById('algorithm').value;

    const directories = dirsInput.split(',').map(d => d.trim()).filter(d => d.length > 0);
    
    if (directories.length === 0) {
        alert("Please enter at least one directory.");
        return;
    }

    const btn = document.getElementById('scanBtn');
    const btnText = btn.querySelector('span');
    const loader = btn.querySelector('.loader');
    const resultsPanel = document.getElementById('resultsPanel');
    const resultsContent = document.getElementById('resultsContent');

    // UI Loading state
    btn.disabled = true;
    btnText.textContent = "Scanning...";
    loader.classList.remove('hidden');
    resultsPanel.classList.add('hidden');
    resultsContent.innerHTML = '';

    try {
        const response = await fetch('/api/scan', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                directories,
                workers,
                min_size: minSize,
                max_depth: maxDepth,
                algorithm
            })
        });

        const data = await response.json();

        if (!response.ok || data.error) {
            throw new Error(data.error || 'Failed to scan directories');
        }

        renderResults(data);

    } catch (error) {
        alert(`Error: ${error.message}`);
    } finally {
        // Reset UI
        btn.disabled = false;
        btnText.textContent = "Start Scan";
        loader.classList.add('hidden');
    }
});

function renderResults(data) {
    const resultsPanel = document.getElementById('resultsPanel');
    const resultsContent = document.getElementById('resultsContent');
    
    document.getElementById('timeBadge').textContent = `Time: ${data.elapsed}`;
    document.getElementById('countBadge').textContent = `${data.duplicates ? data.duplicates.length : 0} Groups`;
    
    resultsPanel.classList.remove('hidden');

    if (!data.duplicates || data.duplicates.length === 0) {
        resultsContent.innerHTML = `<div class="empty-state">
            <h3>No duplicates found!</h3>
            <p>Your directories are perfectly clean.</p>
        </div>`;
        return;
    }

    // Sort groups by number of files (descending)
    const groups = data.duplicates.sort((a, b) => b.files.length - a.files.length);

    let html = '';
    groups.forEach((group, index) => {
        html += `
            <div class="duplicate-group" style="animation: slideIn 0.3s ease forwards; animation-delay: ${index * 0.05}s; opacity: 0; transform: translateY(10px);">
                <div class="group-hash">Hash: ${group.hash}</div>
                <ul class="file-list">
                    ${group.files.map(f => `<li class="file-item">${escapeHtml(f)}</li>`).join('')}
                </ul>
            </div>
        `;
    });

    // Add keyframe animation dynamically if not present
    if (!document.getElementById('dynamic-animations')) {
        const style = document.createElement('style');
        style.id = 'dynamic-animations';
        style.innerHTML = `
            @keyframes slideIn {
                to { opacity: 1; transform: translateY(0); }
            }
        `;
        document.head.appendChild(style);
    }

    resultsContent.innerHTML = html;
}

function escapeHtml(unsafe) {
    return unsafe
         .replace(/&/g, "&amp;")
         .replace(/</g, "&lt;")
         .replace(/>/g, "&gt;")
         .replace(/"/g, "&quot;")
         .replace(/'/g, "&#039;");
}
