let currentShortUrl = '';

// Форма создания короткой ссылки
document.getElementById('shortenForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const fullUrl = document.getElementById('fullUrl').value;
    const customShort = document.getElementById('customShort').value;
    
    try {
        const response = await fetch('/shorten', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                url: fullUrl,
                custom_short: customShort || undefined
            })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            currentShortUrl = data.short_url;
            const shortUrlLink = document.getElementById('shortUrlLink');
            const fullShortUrl = `${window.location.origin}/${data.short_url}`;
            shortUrlLink.href = fullShortUrl;
            shortUrlLink.textContent = fullShortUrl;
            
            document.getElementById('result').classList.remove('hidden');
            document.getElementById('result').classList.remove('error');
        } else {
            showError(data.error || 'Ошибка при создании ссылки');
        }
    } catch (error) {
        showError('Ошибка соединения с сервером');
    }
});

// Форма аналитики
document.getElementById('analyticsForm').addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const shortUrl = document.getElementById('analyticsShortUrl').value;
    const startDate = document.getElementById('startDate').value;
    const endDate = document.getElementById('endDate').value;
    
    let url = `/analytics/${shortUrl}`;
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    if (params.toString()) url += `?${params.toString()}`;
    
    try {
        const response = await fetch(url);
        const data = await response.json();
        
        if (response.ok) {
            displayAnalytics(data);
        } else {
            showAnalyticsError(data.error || 'Ошибка при получении аналитики');
        }
    } catch (error) {
        showAnalyticsError('Ошибка соединения с сервером');
    }
});

function displayAnalytics(data) {
    const statsContent = document.getElementById('statsContent');
    
    let html = `
        <div class="stats-grid">
            <div class="stat-item">
                <div class="stat-label">Короткая ссылка</div>
                <div class="stat-value" style="font-size: 18px;">${data.short_url}</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Оригинальный URL</div>
                <div class="stat-value" style="font-size: 14px; word-break: break-all;">${data.full_url}</div>
            </div>
            <div class="stat-item">
                <div class="stat-label">Всего переходов</div>
                <div class="stat-value">${data.total_clicks}</div>
            </div>
        </div>
    `;
    
    if (data.clicks_by_day && data.clicks_by_day.length > 0) {
        html += `
            <h4 style="margin-top: 20px; color: #333;">Переходы по дням</h4>
            <table class="clicks-table">
                <thead>
                    <tr>
                        <th>Дата</th>
                        <th>Количество</th>
                    </tr>
                </thead>
                <tbody>
        `;
        
        data.clicks_by_day.forEach(item => {
            const date = new Date(item.date).toLocaleDateString('ru-RU');
            html += `
                <tr>
                    <td>${date}</td>
                    <td>${item.count}</td>
                </tr>
            `;
        });
        
        html += `
                </tbody>
            </table>
        `;
    }
    
    statsContent.innerHTML = html;
    document.getElementById('analyticsResult').classList.remove('hidden');
    document.getElementById('analyticsResult').classList.remove('error');
}

function showError(message) {
    const result = document.getElementById('result');
    result.innerHTML = `<h3>Ошибка ❌</h3><p>${message}</p>`;
    result.classList.remove('hidden');
    result.classList.add('error');
}

function showAnalyticsError(message) {
    const result = document.getElementById('analyticsResult');
    result.innerHTML = `<h3>Ошибка ❌</h3><p>${message}</p>`;
    result.classList.remove('hidden');
    result.classList.add('error');
}

function copyToClipboard() {
    const link = document.getElementById('shortUrlLink').textContent;
    navigator.clipboard.writeText(link).then(() => {
        alert('Ссылка скопирована в буфер обмена!');
    });
}

function showAnalytics() {
    document.getElementById('analyticsShortUrl').value = currentShortUrl;
    document.getElementById('analyticsShortUrl').focus();
    window.scrollTo({
        top: document.getElementById('analyticsForm').offsetTop - 20,
        behavior: 'smooth'
    });
}

// Установка дат по умолчанию
const today = new Date();
const monthAgo = new Date();
monthAgo.setMonth(monthAgo.getMonth() - 1);

document.getElementById('startDate').valueAsDate = monthAgo;
document.getElementById('endDate').valueAsDate = today;
