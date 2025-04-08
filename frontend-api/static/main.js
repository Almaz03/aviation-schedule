const token = localStorage.getItem("token");
const API_FLIGHTS = "http://localhost:8000/flights";
const API_AUTH = "http://localhost:8082/me";
const API_GATEWAY = "http://localhost:8084/parse";

// Убедитесь, что переменные API_KEY и API_ADMIN_KEY настроены в вашем окружении или передаются через переменные.
const API_KEY = "TNUr9MZK3Sgmy5hswnJGCExvH7VacRbpFDP6YA4Wuf8dj"; // Ваш API ключ для работы с обычными пользователями
const API_ADMIN_KEY = "khyWYbSHGjxUd98J2BwR4fNPrpgXv6ztZVmDAELqCs7Kc"; // API ключ для работы с админами

let isAdmin = false;
let sortKey = null;
let sortDirection = 1;
let flights = [];
let editId = null;

if (!token) window.location.href = "login.html";

// Выход из аккаунта
document.getElementById("logoutBtn").onclick = () => {
    localStorage.removeItem("token");
    window.location.href = "login.html";
};

// Переключение темы
document.getElementById("toggleTheme").onclick = () => {
    const html = document.documentElement;
    const current = html.getAttribute("data-theme");
    const next = current === "dark" ? "light" : "dark";
    html.setAttribute("data-theme", next);
    localStorage.setItem("theme", next);
    applyTheme();
};

// Применение темы
function applyTheme() {
    const theme = localStorage.getItem("theme") || "light";
    document.documentElement.setAttribute("data-theme", theme);
    const isDark = theme === "dark";

    // body
    document.body.className = isDark ? "bg-gray-900 text-white" : "bg-gray-50 text-gray-800";

    // инпуты, селекты
    document.querySelectorAll("input, select, textarea").forEach(el => {
        el.classList.remove("bg-white", "text-gray-800", "border-gray-300", "bg-gray-800", "text-white", "border-gray-600");
        if (isDark) {
            el.classList.add("bg-gray-800", "text-white", "border-gray-600");
        } else {
            el.classList.add("bg-white", "text-gray-800", "border-gray-300");
        }
    });

    // иконка 🌙/☀️
    const icon = document.getElementById("themeIcon");
    if (icon) {
        icon.textContent = isDark ? "🌙" : "☀️";
    }
}

// Часы
function startClock() {
    setInterval(() => {
        const now = new Date();
        const months = [
            "января", "февраля", "марта", "апреля", "мая", "июня",
            "июля", "августа", "сентября", "октября", "ноября", "декабря"
        ];
        const dateStr = `${now.getDate()} ${months[now.getMonth()]} ${now.getFullYear()} года`;
        const timeStr = now.toLocaleTimeString("ru-RU", { hour12: false });
        document.getElementById("dateBox").textContent = dateStr;
        document.getElementById("timeBox").textContent = timeStr;
    }, 1000);
}

document.getElementById("statusFilter").onchange = renderFlights;
document.getElementById("sortDeparture").onclick = () => toggleSort("departure_time");
document.getElementById("sortArrival").onclick = () => toggleSort("arrival_time");
document.getElementById("sortNumber").onclick = () => toggleSort("number");
document.getElementById("sortOrigin").onclick = () => toggleSort("origin");
document.getElementById("sortDestination").onclick = () => toggleSort("destination");

// Сортировка по выбранному ключу
function toggleSort(key) {
    if (sortKey === key) sortDirection *= -1;
    else {
        sortKey = key;
        sortDirection = 1;
    }
    renderFlights();
}

// Форматирование времени
function formatDateTime(isoString) {
    const date = new Date(isoString);
    const options = {
        day: '2-digit',
        month: 'long',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false,
    };
    return date.toLocaleString("ru-RU", options);
}

// Инициализация данных
async function init() {
    applyTheme();
    startClock();

    // Получаем данные пользователя
    const res = await fetch(API_AUTH, {
        headers: { Authorization: `Bearer ${token}` },
    });
    const user = await res.json();
    isAdmin = user.role === "admin";
    isUser = user.role === "user";

    if (isAdmin) {
        document.getElementById("adminPanel")?.classList.remove("hidden");
    }
    if (isUser) {
        document.getElementById("th-actions")?.classList.add("hidden"); // Скрываем действия для user
    } else {
        document.getElementById("th-actions")?.classList.remove("hidden"); // Отображаем действия для admin
    }

    // Запрос на получение рейсов
    const flightRes = await fetch(API_FLIGHTS, {
        headers: {
            "Content-Type": "application/json",
            "API-Key": API_KEY // Добавлен API_KEY в заголовок
        }
    });
    flights = await flightRes.json();
    renderFlights();
}

// Отображение рейсов
function renderFlights() {
    const filter = document.getElementById("statusFilter").value;
    let filtered = [...flights];

    if (filter !== "all") {
        filtered = filtered.filter(f => f.status === filter);
    }

    if (sortKey) {
        filtered.sort((a, b) => (a[sortKey] > b[sortKey] ? sortDirection : -sortDirection));
    }

    const tbody = document.getElementById("flightTableBody");
    tbody.innerHTML = "";

    // Получаем текущую тему
    const isDark = document.documentElement.getAttribute("data-theme") === "dark";

    filtered.forEach(f => {
        const row = document.createElement("tr");

        // Определяем стиль для каждой ячейки в зависимости от текущей темы
        const baseTdClass = isDark
            ? "p-3 bg-gray-800 text-gray-100 border-b border-gray-700"  // Темная тема
            : "p-3 bg-white text-gray-800 border-b border-gray-200";    // Светлая тема

        const statusColor = {
            "scheduled": isDark ? "bg-gray-700 text-gray-100" : "bg-gray-100 text-gray-800",
            "active": isDark ? "bg-blue-700 text-blue-100" : "bg-blue-100 text-blue-800",
            "landed": isDark ? "bg-green-700 text-green-100" : "bg-green-100 text-green-800",
            "delayed": isDark ? "bg-yellow-700 text-yellow-100" : "bg-yellow-100 text-yellow-800",
            "cancelled": isDark ? "bg-red-700 text-red-100" : "bg-red-100 text-red-800",
        }[f.status.toLowerCase()] || (isDark ? "bg-gray-700 text-gray-100" : "bg-gray-100 text-gray-800");

        row.innerHTML = `
          <td class="${baseTdClass}">${f.number}</td>
          <td class="${baseTdClass}">${f.origin}</td>
          <td class="${baseTdClass}">${f.destination}</td>
          <td class="${baseTdClass}">${formatDateTime(f.departure_time)}</td>
          <td class="${baseTdClass}">${formatDateTime(f.arrival_time)}</td>
          <td class="${baseTdClass} status-overlay"><span class="px-3 py-1 rounded-full ${statusColor} text-sm font-medium">${f.status}</span></td>
          <td class="${baseTdClass} text-center actions">
            ${isAdmin ? `
              <button onclick="editFlight(${f.id})" class="text-blue-500 hover:underline">✏️</button>
              <button onclick="deleteFlight(${f.id})" class="text-red-500 hover:underline ml-2">🗑️</button>
            ` : ""}
          </td>
        `;

        // Если не админ, скрываем действия
        if (!isAdmin) {
            row.classList.add("user-row");
        }

        tbody.appendChild(row);
    });
}

// Создание нового рейса
// Создание нового рейса
async function createFlight() {
    const body = {
        number: document.getElementById("f_number").value,
        origin: document.getElementById("f_origin").value,
        destination: document.getElementById("f_dest").value,
        departure_time: document.getElementById("f_departure").value,
        arrival_time: document.getElementById("f_arrival").value,
        status: document.getElementById("f_status").value,
    };

    const res = await fetch(API_FLIGHTS, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
            "API-Key": API_KEY
        },
        body: JSON.stringify(body),
    });

    if (res.ok) {
        await init();
    } else {
        alert("Ошибка при добавлении рейса");
    }
}

// Удаление рейса
async function deleteFlight(id) {
    if (!confirm("Удалить рейс?")) return;

    await fetch(`${API_FLIGHTS}/${id}`, {
        method: "DELETE",
        headers: {
            Authorization: `Bearer ${token}`,
            "API-Key": API_KEY
        },
    });

    await init();
}

// Удалить прошедшие рейсы
async function deletePastFlights() {
    if (!confirm("Удалить все рейсы, которые уже прилетели?")) return;

    const res = await fetch("http://localhost:8000/flights/past", {
        method: "DELETE",
        headers: { Authorization: `Bearer ${token}`, "API-Key": API_KEY },
    });

    if (res.ok) {
        alert("Прошедшие рейсы удалены");
        await init();
    } else {
        alert("Ошибка при удалении");
    }
}

// Удалить все рейсы
async function deleteAllFlights() {
    if (!confirm("⚠️ Удалить все рейсы безвозвратно?")) return;

    const res = await fetch("http://localhost:8000/flights/all", {
        method: "DELETE",
        headers: {
            Authorization: `Bearer ${token}`,
            "API-Key": API_KEY
        }
    });

    if (res.ok) {
        alert("Все рейсы удалены");
        await init();  // Перезагружаем список рейсов
    } else {
        const error = await res.json();
        alert(`Ошибка при удалении всех рейсов: ${error.error || 'Неизвестная ошибка'}`);
    }
}
// Внешнее API - загрузка рейсов по ICAO коду
async function loadExternalFlights() {
    const icao = document.getElementById("icaoInput").value.trim().toUpperCase();
    if (!icao) return alert("Введите ICAO код");

    const res = await fetch(`${API_GATEWAY}?icao=${icao}`, {
        headers: {
            "Content-Type": "application/json",
            "API-Key": API_KEY
        }
    });
    if (res.ok) {
        const msg = await res.text();
        alert(msg);
        await init(); // обновим таблицу
    } else {
        alert("Ошибка при загрузке рейсов");
    }
}
async function addAdmin() {
    const username = document.getElementById("admin_username").value.trim();
    const password = document.getElementById("admin_password").value.trim();

    if (!username || !password) return alert("Введите логин и пароль");

    const res = await fetch("http://localhost:8082/register", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
            "API-Key": API_ADMIN_KEY
        },
        body: JSON.stringify({ username, password, role: "admin" }),
    });

    if (res.ok) {
        alert("Администратор добавлен");
        document.getElementById("admin_username").value = "";
        document.getElementById("admin_password").value = "";
    } else {
        alert("Ошибка при добавлении администратора");
    }
}
// Редактирование рейса
function editFlight(id) {
    const flight = flights.find(f => f.id === id);
    if (!flight) return;

    editId = id;
    document.getElementById("edit_number").value = flight.number;
    document.getElementById("edit_origin").value = flight.origin;
    document.getElementById("edit_dest").value = flight.destination;
    document.getElementById("edit_departure").value = flight.departure_time;
    document.getElementById("edit_arrival").value = flight.arrival_time;
    document.getElementById("edit_status").value = flight.status;

    document.getElementById("editModal").classList.remove("hidden");
}

function closeModal() {
    document.getElementById("editModal").classList.add("hidden");
}

async function saveEdit() {
    const body = {
        number: document.getElementById("edit_number").value,
        origin: document.getElementById("edit_origin").value,
        destination: document.getElementById("edit_dest").value,
        departure_time: document.getElementById("edit_departure").value,
        arrival_time: document.getElementById("edit_arrival").value,
        status: document.getElementById("edit_status").value,
    };

    const res = await fetch(`${API_FLIGHTS}/${editId}`, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${token}`,
            "API-Key": API_KEY
        },
        body: JSON.stringify(body),
    });

    if (res.ok) {
        closeModal();
        await init();
    } else {
        alert("Ошибка при сохранении");
    }
}
init();
