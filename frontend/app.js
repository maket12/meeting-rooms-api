// =================================================================
// 1. КОНФИГУРАЦИЯ И ГЛОБАЛЬНЫЙ СТЕЙТ
// =================================================================
const API_BASE_URL = 'http://localhost:8080';

const state = {
    token: localStorage.getItem('mr_token') || null,
    role: localStorage.getItem('mr_role') || null, // 'user' or 'admin'
    userId: localStorage.getItem('mr_user_id') || null,
    email: localStorage.getItem('mr_email') || null,
    rooms: [],
    selectedRoomId: null,
    selectedSlotId: null,
    slots: [],
    adminLogPage: 1,
    adminLogPageSize: 10
};

// =================================================================
// 2. ИНИЦИАЛИЗАЦИЯ И ОБРАБОТЧИКИ СОБЫТИЙ (DOM)
// =================================================================
document.addEventListener('DOMContentLoaded', () => {
    // Дата в календаре слотов — всегда текущая, а не захардкоженная
    const datePicker = document.getElementById('slots-date-picker');
    if (datePicker && !datePicker.value) {
        datePicker.value = new Date().toISOString().slice(0, 10);
    }

    setupEventListeners();
    initApp();
    initUtcClock();
});

// Проверка авторизации при старте приложения
function initApp() {
    if (state.token) {
        showScreen(state.role === 'admin' ? 'admin' : 'user');
        loadInitialData();
    } else {
        showScreen('auth');
    }
}

function setupEventListeners() {
    // Формы авторизации
    document.getElementById('form-signin').addEventListener('submit', handleSignIn);
    document.getElementById('form-signup').addEventListener('submit', handleSignUp);

    // Быстрый тестовый вход (Dummy Login) - Обязательная часть ТЗ
    document.getElementById('btn-dummy-user').addEventListener('click', () => handleDummyLogin('user'));
    document.getElementById('btn-dummy-admin').addEventListener('click', () => handleDummyLogin('admin'));

    // Выход из системы (привязка ко всем кнопкам Выйти)
    document.querySelectorAll('.btn-logout').forEach(btn => {
        btn.addEventListener('click', logout);
    });

    // Изменение даты в календаре слотов
    document.getElementById('slots-date-picker').addEventListener('change', () => {
        state.slots = [];
        if (state.selectedRoomId) loadSlots(state.selectedRoomId);
    });

    // Подтверждение бронирования в модалке
    document.getElementById('btn-confirm-booking').addEventListener('click', createBooking);

    // Панель админа: Создание комнаты
    document.getElementById('form-create-room').addEventListener('submit', createRoom);

    // Панель админа: Создание расписания
    document.getElementById('form-create-schedule').addEventListener('submit', createSchedule);

    // Панель админа: Пагинация логов
    document.getElementById('btn-log-prev').addEventListener('click', () => navigateAdminLogs(-1));
    document.getElementById('btn-log-next').addEventListener('click', () => navigateAdminLogs(1));

    // Set minimum date to today (only allow future dates)
    const datePicker = document.getElementById('slots-date-picker');
    const today = new Date().toISOString().slice(0, 10);
    const twoWeeksLater = new Date(new Date().getTime() + 14 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);

    datePicker.min = today;
    datePicker.max = twoWeeksLater;
}

// =================================================================
// 3. СЕТЕВЫЕ ЗАПРОСЫ К API (Fetch Wrapper + обработка ошибок по OpenAPI)
// =================================================================

// Класс ошибки API. Хранит HTTP-статус и код ошибки от бэкенда (поле `error`
// из ErrorResponse по спецификации), а также готовое сообщение для пользователя.
class ApiError extends Error {
    constructor(status, code, message) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
        this.code = code;
    }
}

// Карта соответствия кодов ошибок бэкенда (значения поля `error` из api.yaml)
// понятным пользователю сообщениям на русском языке.
const ERROR_MESSAGES = {
    // Общие / валидация
    'invalid json': 'Некорректный формат запроса. Попробуйте ещё раз.',
    'invalid input': 'Проверьте правильность заполнения полей формы.',
    'invalid identifier format': 'Некорректный идентификатор ресурса.',

    // Авторизация / доступ
    'missing auth header': 'Вы не авторизованы. Пожалуйста, войдите заново.',
    'invalid or expired token': 'Сессия истекла. Пожалуйста, войдите заново.',
    'insufficient permissions': 'У вас недостаточно прав для этого действия.',
    'invalid email or password': 'Неверный email или пароль.',

    // Сущности не найдены
    'user not found': 'Тестовый пользователь для этой роли не настроен на сервере.',
    'room not found': 'Переговорка не найдена. Возможно, она была удалена.',
    'schedule not found': 'Для этой переговорки ещё не задано расписание.',
    'slot not found': 'Выбранный временной слот не найден.',
    'booking not found': 'Бронирование не найдено.',

    // Конфликты
    'user with given email already exists': 'Пользователь с таким email уже зарегистрирован.',
    'schedule for this room already exists': 'Расписание для этой переговорки уже создано и не может быть изменено.',
    'booking for this slot already exists': 'Этот слот уже занят другим бронированием. Обновите список слотов.',
    'specified slot is in the past': 'Нельзя забронировать слот в прошлом.',

    // Сервер
    'internal error': 'Внутренняя ошибка сервера. Попробуйте позже.'
};

// Сообщения по умолчанию, если конкретный код от бэкенда не распознан.
const STATUS_FALLBACK_MESSAGES = {
    0: 'Не удалось подключиться к серверу. Убедитесь, что бэкенд запущен и доступен по адресу ' + API_BASE_URL + '.',
    400: 'Некорректный запрос.',
    401: 'Требуется авторизация.',
    403: 'Недостаточно прав для этого действия.',
    404: 'Запрашиваемый ресурс не найден.',
    409: 'Конфликт данных: операция невозможна в текущем состоянии.',
    500: 'Внутренняя ошибка сервера. Попробуйте позже.'
};

function resolveErrorMessage(status, code) {
    if (code && ERROR_MESSAGES[code]) return ERROR_MESSAGES[code];
    if (STATUS_FALLBACK_MESSAGES[status]) return STATUS_FALLBACK_MESSAGES[status];
    return 'Произошла непредвиденная ошибка. Попробуйте ещё раз.';
}

// Эндпоинты, для которых 401 не должен триггерить принудительный логаут
// (это сами эндпоинты аутентификации).
const AUTH_ENDPOINTS = ['/login', '/register', '/dummyLogin'];

async function apiRequest(endpoint, options = {}) {
    const url = `${API_BASE_URL}${endpoint}`;

    options.headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };

    if (state.token) {
        options.headers['Authorization'] = `Bearer ${state.token}`;
    }

    let response;
    try {
        response = await fetch(url, options);
    } catch (networkErr) {
        // Сервер недоступен / нет сети — это ОТДЕЛЬНЫЙ вид ошибки,
        // не должен подменяться выдуманными данными.
        throw new ApiError(0, 'network_error', resolveErrorMessage(0, 'network_error'));
    }

    let data = {};
    try {
        data = await response.json();
    } catch (parseErr) {
        data = {};
    }

    if (!response.ok) {
        const code = data && data.error ? data.error : null;
        const apiError = new ApiError(response.status, code, resolveErrorMessage(response.status, code));

        // Глобальная обработка истёкшей/невалидной сессии для уже авторизованных запросов.
        if (response.status === 401 && !AUTH_ENDPOINTS.includes(endpoint) && state.token) {
            forceLogoutWithMessage(apiError.message);
        }

        throw apiError;
    }

    return data;
}

// Принудительный разлогин при истёкшей сессии (401 на защищённом эндпоинте).
function forceLogoutWithMessage(message) {
    logout();
    showAuthError(message);
}

// =================================================================
// 4. УВЕДОМЛЕНИЯ (Toast) — отдельное "окно" для некритичных фоновых ошибок
// =================================================================

function showToast(message, type = 'error') {
    const container = document.getElementById('toast-container');
    if (!container) {
        console.error(`[${type}]`, message);
        return;
    }

    const alertClass = {
        error: 'alert-error',
        success: 'alert-success',
        warning: 'alert-warning',
        info: 'alert-info'
    }[type] || 'alert-error';

    const toast = document.createElement('div');
    toast.className = `alert ${alertClass} shadow-lg text-sm py-3 pr-2 max-w-sm`;
    toast.innerHTML = `
        <span class="flex-1">${escapeHtml(message)}</span>
        <button class="btn btn-ghost btn-xs" aria-label="Закрыть">✕</button>
    `;

    toast.querySelector('button').addEventListener('click', () => toast.remove());
    container.appendChild(toast);

    setTimeout(() => toast.remove(), 6000);
}

// =================================================================
// 5. ЛОГИКА АУТЕНТИФИКАЦИИ
// =================================================================

// Декодирует payload JWT без проверки подписи — только для чтения
// user_id/role на клиенте (сама проверка токена выполняется на бэкенде).
function decodeJwtPayload(token) {
    try {
        const payloadPart = token.split('.')[1];
        const base64 = payloadPart.replace(/-/g, '+').replace(/_/g, '/');
        const json = decodeURIComponent(
            atob(base64)
                .split('')
                .map(c => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
                .join('')
        );
        return JSON.parse(json);
    } catch (e) {
        return null;
    }
}

// Обязательный эндпоинт по ТЗ: /dummyLogin
async function handleDummyLogin(role) {
    try {
        showAuthError(null);
        const data = await apiRequest('/dummyLogin', {
            method: 'POST',
            body: JSON.stringify({ role: role })
        });

        applySession(data.token, role, `${role}-test@company.com`);
        initApp();
    } catch (err) {
        showAuthError(err.message);
    }
}

// Дополнительный эндпоинт: /login
async function handleSignIn(e) {
    e.preventDefault();
    const email = document.getElementById('signin-email').value;
    const password = document.getElementById('signin-password').value;

    try {
        showAuthError(null);
        const data = await apiRequest('/login', {
            method: 'POST',
            body: JSON.stringify({ email, password })
        });

        // Роль и user_id берём из самого JWT (как и предписывает спецификация),
        // а не угадываем по содержимому email.
        const payload = decodeJwtPayload(data.token);
        const role = (payload && (payload.role === 'admin' || payload.role === 'user')) ? payload.role : 'user';
        applySession(data.token, role, email);
        initApp();
    } catch (err) {
        showAuthError(err.message);
    }
}

// Дополнительный эндпоинт: /register
async function handleSignUp(e) {
    e.preventDefault();
    const email = document.getElementById('signup-email').value;
    const password = document.getElementById('signup-password').value;
    const role = document.getElementById('signup-role').value;

    try {
        showAuthError(null);
        await apiRequest('/register', {
            method: 'POST',
            body: JSON.stringify({ email, password, role })
        });

        showToast('Регистрация успешна! Теперь вы можете войти.', 'success');
        switchAuthTab('signin');
    } catch (err) {
        showAuthError(err.message);
    }
}

function applySession(token, role, email) {
    const payload = decodeJwtPayload(token);
    const userId = payload && payload.user_id ? payload.user_id : null;

    state.token = token;
    state.role = role;
    state.email = email;
    state.userId = userId;

    localStorage.setItem('mr_token', token);
    localStorage.setItem('mr_role', role);
    localStorage.setItem('mr_email', email);
    if (userId) localStorage.setItem('mr_user_id', userId);
}

function logout() {
    state.token = null;
    state.role = null;
    state.userId = null;
    state.email = null;
    state.rooms = [];
    state.selectedRoomId = null;
    state.selectedSlotId = null;
    state.slots = [];

    // Удаляем только ключи сессии. localStorage.clear() стирал ВСЁ хранилище,
    // включая mr_slot_cache — из-за этого кэш метаданных слотов (комната/время
    // брони) обнулялся при каждом логине через Dev UI Switcher, и "Мои бронирования"
    // всегда показывали "данные недоступны".
    localStorage.removeItem('mr_token');
    localStorage.removeItem('mr_role');
    localStorage.removeItem('mr_email');
    localStorage.removeItem('mr_user_id');

    showAuthError(null);
    showScreen('auth');
}

// =================================================================
// 6. ФУНКЦИОНАЛ ПОЛЬЗОВАТЕЛЯ (USER PANEL)
// =================================================================

async function loadInitialData() {
    const userEmailEl = document.getElementById('user-display-email');
    if (userEmailEl) userEmailEl.textContent = state.email || '—';
    const userIdEl = document.getElementById('user-display-id');
    if (userIdEl) userIdEl.textContent = state.userId ? `ID: ${state.userId.substring(0, 8)}...` : '';

    const adminEmailEl = document.getElementById('admin-display-email');
    if (adminEmailEl) adminEmailEl.textContent = state.email || '—';
    const adminIdEl = document.getElementById('admin-display-id');
    if (adminIdEl) adminIdEl.textContent = state.userId ? `ID: ${state.userId.substring(0, 8)}...` : 'Права: Полный доступ';

    if (state.role === 'user') {
        await loadRooms();
        await loadMyBookings();
    } else if (state.role === 'admin') {
        await loadRooms(); // Админу тоже нужен список для привязки расписания
        await loadAdminLogs();
    }
}

// Получить список комнат: GET /rooms/list
async function loadRooms() {
    try {
        const data = await apiRequest('/rooms/list');
        state.rooms = data.rooms || [];
    } catch (err) {
        // Ни в коем случае не подставляем выдуманные комнаты — просто пустой список
        // и явное сообщение об ошибке.
        state.rooms = [];
        showToast('Не удалось загрузить список переговорок: ' + err.message, 'error');
    }

    if (state.role === 'user') {
        renderUserRooms();
        if (state.rooms.length > 0) {
            // Make sure date is set before loading slots
            const datePicker = document.getElementById('slots-date-picker');
            if (!datePicker.value) {
                datePicker.value = new Date().toISOString().slice(0, 10);
            }
            selectRoom(state.rooms[0].id);
        } else {
            state.selectedRoomId = null;
            state.slots = [];
            renderSlotsGrid();
        }
    } else if (state.role === 'admin') {
        renderAdminRoomSelect();
    }
}

function renderUserRooms() {
    const container = document.getElementById('user-rooms-list');
    container.innerHTML = '';

    if (state.rooms.length === 0) {
        container.innerHTML = '<p class="text-xs text-base-content/40 italic text-center py-4">Переговорки пока не добавлены</p>';
        return;
    }

    state.rooms.forEach(room => {
        const isSelected = room.id === state.selectedRoomId;
        const div = document.createElement('div');
        div.className = `p-3 bg-base-200 rounded-xl border-2 transition-all cursor-pointer ${isSelected ? 'border-accent' : 'border-transparent hover:border-base-300'}`;
        div.onclick = () => selectRoom(room.id);

        div.innerHTML = `
            <div class="flex items-start justify-between">
                <h3 class="font-bold text-sm text-base-content">${escapeHtml(room.name)}</h3>
                <span class="badge badge-sm badge-neutral font-semibold">${room.capacity || 0} чел</span>
            </div>
            <p class="text-xs text-base-content/60 mt-1">${escapeHtml(room.description || 'Описание отсутствует')}</p>
        `;
        container.appendChild(div);
    });
}

function selectRoom(roomId) {
    state.selectedRoomId = roomId;
    renderUserRooms();
    loadSlots(roomId);
}

// Получить свободные слоты: GET /rooms/{id}/slots/list?date=YYYY-MM-DD
async function loadSlots(roomId) {
    const datePicker = document.getElementById('slots-date-picker');
    const dateInput = datePicker.value;

    // If date is not set, set it to today and retry
    if (!dateInput) {
        datePicker.value = new Date().toISOString().slice(0, 10);
        return loadSlots(roomId);
    }

    try {
        const data = await apiRequest(`/rooms/${roomId}/slots/list?date=${dateInput}`);
        state.slots = data.slots || [];
        // Бэкенд не отдаёт room_id/время для уже созданной брони (только slot_id),
        // поэтому запоминаем метаданные слотов, которые реально видели с сервера —
        // это единственный источник правды для последующего отображения "Моих броней".
        cacheSlotMeta(state.slots, roomId);
    } catch (err) {
        state.slots = [];
        showToast('Не удалось загрузить слоты: ' + err.message, 'error');
    }
    renderSlotsGrid();
}

// =================================================================
// Локальный кэш метаданных слотов (room_id/start/end по slot_id).
// Нужен, чтобы показывать комнату и время в "Моих бронированиях":
// API /bookings/my отдаёт только slot_id, без данных о комнате/времени,
// а /rooms/{id}/slots/list не возвращает уже забронированные слоты.
// Кэш пополняется РЕАЛЬНЫМИ данными сервера при каждой загрузке слотов
// и переживает перезагрузку страницы (localStorage) — никаких выдуманных
// значений сюда не попадает.
// =================================================================
const SLOT_CACHE_KEY = 'mr_slot_cache';
const SLOT_CACHE_LIMIT = 500;

function loadSlotCache() {
    try {
        return JSON.parse(localStorage.getItem(SLOT_CACHE_KEY)) || {};
    } catch (e) {
        return {};
    }
}

function cacheSlotMeta(slots, roomId) {
    if (!slots || slots.length === 0) return;
    const cache = loadSlotCache();

    slots.forEach(slot => {
        cache[slot.id] = {
            room_id: slot.room_id || roomId,
            start: slot.start,
            end: slot.end
        };
    });

    // Ограничиваем размер кэша, чтобы не разрастался бесконечно —
    // отбрасываем старые записи по порядку добавления.
    const keys = Object.keys(cache);
    if (keys.length > SLOT_CACHE_LIMIT) {
        keys.slice(0, keys.length - SLOT_CACHE_LIMIT).forEach(k => delete cache[k]);
    }

    try {
        localStorage.setItem(SLOT_CACHE_KEY, JSON.stringify(cache));
    } catch (e) {
        // localStorage переполнен/недоступен — не критично, просто не кэшируем.
    }
}

function getCachedSlotMeta(slotId) {
    const cache = loadSlotCache();
    return cache[slotId] || null;
}

function renderSlotsGrid() {
    const grid = document.getElementById('slots-grid');
    const emptyState = document.getElementById('slots-empty-state');
    const actionBtn = document.getElementById('btn-open-booking-modal');
    const selectedDate = document.getElementById('slots-date-picker').value;

    grid.innerHTML = '';
    state.selectedSlotId = null;
    actionBtn.disabled = true;
    actionBtn.textContent = 'Выберите временной слот';

    // Filter slots by selected date
    const filteredSlots = state.slots.filter(slot => {
        // Extract date from slot.start (format: "2026-07-06 09:00:00 +0000 UTC")
        const slotDate = slot.start.substring(0, 10); // Get YYYY-MM-DD
        return slotDate === selectedDate;
    });

    if (filteredSlots.length === 0) {
        emptyState.classList.remove('hidden');
        return;
    }
    emptyState.classList.add('hidden');

    filteredSlots.forEach(slot => {
        const btn = document.createElement('button');
        btn.className = 'btn btn-outline btn-sm h-12 text-xs flex flex-col gap-0.5 border-base-300 hover:bg-accent/10 hover:text-accent hover:border-accent';

        const startTime = formatIsoTime(slot.start);
        const endTime = formatIsoTime(slot.end);

        btn.innerHTML = `
            <span class="font-mono font-bold text-sm">${startTime} - ${endTime}</span>
            <span class="text-[10px] font-normal tracking-wide opacity-60">Свободно</span>
        `;

        btn.onclick = () => {
            Array.from(grid.children).forEach(child => child.classList.remove('btn-accent'));
            btn.classList.add('btn-accent');

            state.selectedSlotId = slot.id;
            actionBtn.disabled = false;
            actionBtn.textContent = `Забронировать выбранный слот (${startTime})`;

            const currentRoom = state.rooms.find(r => r.id === state.selectedRoomId);
            document.getElementById('modal-room-name').textContent = currentRoom ? currentRoom.name : 'Выбранная комната';
            document.getElementById('modal-booking-date').textContent = selectedDate;
            document.getElementById('modal-booking-time').textContent = `${startTime} - ${endTime}`;
            showBookingModalError(null);
        };

        grid.appendChild(btn);
    });
}

// Отдельное "окно" ошибки внутри модалки бронирования (409-конфликты и т.п.)
function showBookingModalError(message) {
    const alertEl = document.getElementById('booking-modal-alert');
    if (!alertEl) return;
    if (message) {
        alertEl.textContent = message;
        alertEl.classList.remove('hidden');
    } else {
        alertEl.classList.add('hidden');
    }
}

// Создать бронирование: POST /bookings/create
async function createBooking() {
    if (!state.selectedSlotId) return;

    const createConfLink = document.getElementById('modal-create-conf').checked;
    const payload = {
        slot_id: state.selectedSlotId,
        create_conference_link: createConfLink
    };

    const confirmBtn = document.getElementById('btn-confirm-booking');
    confirmBtn.disabled = true;

    try {
        showBookingModalError(null);
        await apiRequest('/bookings/create', {
            method: 'POST',
            body: JSON.stringify(payload)
        });

        document.getElementById('booking_modal').close();
        showToast('Бронирование успешно создано!', 'success');

        await loadSlots(state.selectedRoomId);
        await loadMyBookings();
    } catch (err) {
        // Ошибку бронирования (в т.ч. 409 — слот уже занят/в прошлом) показываем
        // прямо в модалке, не закрывая её и не подменяя данными.
        showBookingModalError(err.message);
        if (err.status === 409) {
            await loadSlots(state.selectedRoomId);
        }
    } finally {
        confirmBtn.disabled = false;
    }
}

// Получить бронирования пользователя: GET /bookings/my
async function loadMyBookings() {
    const container = document.getElementById('user-bookings-list');
    container.innerHTML = '';

    let bookings = [];
    try {
        const data = await apiRequest('/bookings/my');
        bookings = data.bookings || [];
    } catch (err) {
        showToast('Не удалось загрузить ваши бронирования: ' + err.message, 'error');
        container.innerHTML = '<p class="text-xs text-error/70 italic text-center py-4">Не удалось загрузить бронирования</p>';
        return;
    }

    if (bookings.length === 0) {
        container.innerHTML = '<p class="text-xs text-base-content/40 italic text-center py-4">У вас нет активных бронирований</p>';
        return;
    }

    // Обогащаем брони данными о комнате/времени из локального кэша слотов
    // (см. cacheSlotMeta) — сам /bookings/my отдаёт только slot_id.
    const enriched = bookings.map(booking => {
        const slotMeta = getCachedSlotMeta(booking.slot_id);
        const room = slotMeta ? state.rooms.find(r => r.id === slotMeta.room_id) : null;
        return {
            booking,
            roomName: room ? room.name : null,
            start: slotMeta ? slotMeta.start : null,
            end: slotMeta ? slotMeta.end : null
        };
    });

    // Сортировка: сначала активные брони, затем — по названию комнаты (по возрастанию).
    // Брони с неизвестным названием комнаты (нет данных в кэше) уходят в конец своей группы.
    enriched.sort((a, b) => {
        const statusRank = s => (s.booking.status === 'active' ? 0 : 1);
        const statusDiff = statusRank(a) - statusRank(b);
        if (statusDiff !== 0) return statusDiff;

        if (a.roomName && b.roomName) {
            return a.roomName.localeCompare(b.roomName, 'ru');
        }
        if (a.roomName && !b.roomName) return -1;
        if (!a.roomName && b.roomName) return 1;
        return 0;
    });

    enriched.forEach(({ booking, roomName, start, end }) => {
        const div = document.createElement('div');
        div.className = 'p-4 bg-base-200 rounded-xl border border-base-300 relative flex flex-col justify-between gap-3';

        const displayRoomName = roomName || 'Переговорная комната (данные недоступны)';
        const timeRange = (start && end) ? `${formatIsoTime(start)} - ${formatIsoTime(end)}` : null;
        const dateStr = start ? formatDateOnly(start) : null;
        const statusBadge = booking.status === 'active' ? 'badge-success' : 'badge-neutral opacity-50';

        let confLinkHtml = '';
        if (booking.conference_link) {
            confLinkHtml = `
                <div class="p-2 bg-primary/10 rounded-lg flex items-center justify-between border border-primary/20">
                    <div class="flex flex-col">
                        <span class="text-[10px] uppercase font-bold text-primary tracking-wider">Видеоконференция</span>
                        <span class="text-xs opacity-70 font-mono truncate max-w-[180px]">${escapeHtml(booking.conference_link)}</span>
                    </div>
                    <a href="${booking.conference_link}" target="_blank" class="btn btn-xs btn-primary font-bold px-3">Войти</a>
                </div>
            `;
        }

        div.innerHTML = `
            <div>
                <div class="flex items-center justify-between mb-1">
                    <span class="font-bold text-sm text-base-content">${escapeHtml(displayRoomName)}</span>
                    <span class="badge badge-sm ${statusBadge} font-bold text-[10px]">${booking.status.toUpperCase()}</span>
                </div>
                <div class="text-xs font-mono text-base-content/70">
                    ${timeRange ? `📅 ${dateStr} | 🕒 ${timeRange}` : 'Время недоступно'}
                </div>
                <div class="text-[10px] font-mono text-base-content/40 mt-0.5">
                    ID Брони: ${booking.id.substring(0, 8)}...
                </div>
            </div>
            ${confLinkHtml}
            ${booking.status === 'active' ? `<button class="btn btn-xs btn-outline btn-error w-full mt-1" onclick="cancelBooking('${booking.id}')">Отменить бронирование</button>` : ''}
        `;
        container.appendChild(div);
    });
}

// Отмена брони: POST /bookings/{id}/cancel
async function cancelBooking(bookingId) {
    if (!confirm('Вы уверены, что хотите отменить эту бронь?')) return;
    try {
        await apiRequest(`/bookings/${bookingId}/cancel`, { method: 'POST' });
        showToast('Бронирование отменено', 'success');
        await loadMyBookings();
        if (state.selectedRoomId) loadSlots(state.selectedRoomId);
    } catch (err) {
        showToast('Не удалось отменить бронирование: ' + err.message, 'error');
    }
}

// =================================================================
// 7. ФУНКЦИОНАЛ АДМИНИСТРАТОРА (ADMIN PANEL)
// =================================================================

// Создание комнаты: POST /rooms/create
async function createRoom(e) {
    e.preventDefault();
    const name = document.getElementById('admin-room-name').value;
    const capacityRaw = document.getElementById('admin-room-capacity').value;
    const capacity = capacityRaw ? parseInt(capacityRaw, 10) : null;
    const description = document.getElementById('admin-room-desc').value;

    try {
        await apiRequest('/rooms/create', {
            method: 'POST',
            body: JSON.stringify({ name, capacity, description })
        });
        showToast('Комната успешно создана!', 'success');
        document.getElementById('form-create-room').reset();
        await loadRooms(); // Перезагружаем список в выпадающем меню расписания
    } catch (err) {
        showToast('Ошибка добавления переговорки: ' + err.message, 'error');
    }
}

function renderAdminRoomSelect() {
    const select = document.getElementById('admin-sched-room-id');
    select.innerHTML = '';

    if (state.rooms.length === 0) {
        select.innerHTML = '<option value="" disabled selected>Сначала создайте переговорку</option>';
        return;
    }

    state.rooms.forEach(room => {
        const opt = document.createElement('option');
        opt.value = room.id;
        opt.textContent = `${room.name} (до ${room.capacity || '?'} чел.)`;
        select.appendChild(opt);
    });
}

// Создание расписания: POST /rooms/{id}/schedule/create
async function createSchedule(e) {
    e.preventDefault();
    const roomId = document.getElementById('admin-sched-room-id').value;
    if (!roomId) {
        showToast('Сначала создайте и выберите переговорку', 'warning');
        return;
    }

    const startTime = document.getElementById('admin-sched-start').value; // HH:MM
    const endTime = document.getElementById('admin-sched-end').value;     // HH:MM

    const checkedDays = Array.from(document.querySelectorAll('input[name="days"]:checked'))
        .map(cb => parseInt(cb.value, 10));

    if (checkedDays.length === 0) {
        showToast('Выберите хотя бы один день недели!', 'warning');
        return;
    }

    const payload = {
        days_of_week: checkedDays,
        start_time: startTime,
        end_time: endTime
    };

    try {
        await apiRequest(`/rooms/${roomId}/schedule/create`, {
            method: 'POST',
            body: JSON.stringify(payload)
        });
        showToast('Временная сетка расписания успешно активирована!', 'success');
    } catch (err) {
        showToast('Ошибка активации расписания: ' + err.message, 'error');
    }
}

// Аудит всех броней: GET /bookings/list?page=X&pageSize=Y
async function loadAdminLogs() {
    const tableBody = document.getElementById('admin-bookings-table-body');
    tableBody.innerHTML = '';

    let bookings = [];
    let pagination = { page: state.adminLogPage, page_size: state.adminLogPageSize, total: 0 };

    try {
        const data = await apiRequest(`/bookings/list?page=${state.adminLogPage}&pageSize=${state.adminLogPageSize}`);
        bookings = data.bookings || [];
        pagination = data.pagination || pagination;
    } catch (err) {
        showToast('Не удалось загрузить журнал броней: ' + err.message, 'error');
        tableBody.innerHTML = '<tr><td colspan="5" class="text-center text-error/70 py-4">Не удалось загрузить данные</td></tr>';
        return;
    }

    document.getElementById('btn-log-page').textContent = `Страница ${pagination.page || state.adminLogPage}`;

    // Управление доступностью кнопок пагинации на основе данных сервера
    const prevBtn = document.getElementById('btn-log-prev');
    const nextBtn = document.getElementById('btn-log-next');
    prevBtn.disabled = (pagination.page || state.adminLogPage) <= 1;
    if (typeof pagination.total === 'number' && typeof pagination.page_size === 'number' && pagination.page_size > 0) {
        nextBtn.disabled = (pagination.page * pagination.page_size) >= pagination.total;
    } else {
        nextBtn.disabled = bookings.length < state.adminLogPageSize;
    }

    if (bookings.length === 0) {
        tableBody.innerHTML = '<tr><td colspan="5" class="text-center opacity-50 py-4">Записи логов отсутствуют</td></tr>';
        return;
    }

    bookings.forEach(b => {
        const tr = document.createElement('tr');
        const statusClass = b.status === 'active' ? 'badge-success' : 'badge-neutral opacity-50';
        tr.innerHTML = `
            <td class="font-mono font-bold text-primary">${b.id.substring(0, 8)}...</td>
            <td class="font-mono">${b.slot_id.substring(0, 8)}...</td>
            <td class="font-mono opacity-60">${b.user_id.substring(0, 8)}...</td>
            <td class="opacity-70">${formatDateTime(b.created_at)}</td>
            <td><span class="badge ${statusClass} font-bold text-[9px] px-2 py-0.5">${b.status.toUpperCase()}</span></td>
        `;
        tableBody.appendChild(tr);
    });
}

function navigateAdminLogs(direction) {
    if (state.adminLogPage + direction < 1) return;
    state.adminLogPage += direction;
    loadAdminLogs();
}

// =================================================================
// 8. ВСПОМОГАТЕЛЬНЫЕ ФУНКЦИИ (Утилиты)
// =================================================================

function showScreen(screenName) {
    document.getElementById('screen-auth').classList.add('hidden');
    document.getElementById('screen-user').classList.add('hidden');
    document.getElementById('screen-admin').classList.add('hidden');

    document.getElementById('screen-' + screenName).classList.remove('hidden');
}

function showAuthError(msg) {
    const alertEl = document.getElementById('auth-alert');
    const textEl = document.getElementById('auth-alert-text');
    if (msg) {
        textEl.textContent = msg;
        alertEl.classList.remove('hidden');
    } else {
        alertEl.classList.add('hidden');
    }
}

// Извлекает время HH:MM из строки даты/времени.
// Поддерживает оба формата:
//   ISO 8601:        "2026-07-03T09:00:00Z"
//   Go time.String(): "2026-07-03 09:00:00 +0000 UTC"
function formatIsoTime(dateTimeString) {
    if (!dateTimeString) return '—';

    // Ищем первое вхождение времени вида HH:MM (после даты)
    const match = String(dateTimeString).match(/(\d{2}):(\d{2})(?::\d{2})?/);
    if (match) {
        return `${match[1]}:${match[2]}`;
    }

    // Запасной вариант — пробуем распарсить как Date
    const d = new Date(dateTimeString);
    if (!isNaN(d.getTime())) {
        return d.toISOString().substring(11, 16);
    }

    return '—';
}

function formatDateTime(dateTimeString) {
    if (!dateTimeString) return '—';

    // Нормализуем формат Go ("2026-07-03 09:00:00 +0000 UTC") в ISO
    const normalized = String(dateTimeString)
        .replace(' +0000 UTC', 'Z')
        .replace(' ', 'T');

    const d = new Date(normalized);
    if (isNaN(d.getTime())) return dateTimeString; // показываем как есть, если не распарсилось

    return d.toLocaleString('ru-RU', {
        timeZone: 'UTC',
        day: '2-digit', month: '2-digit', year: 'numeric',
        hour: '2-digit', minute: '2-digit'
    });
}

// Извлекает только дату (DD.MM.YYYY) из строки даты/времени, поддерживая
// как ISO 8601, так и формат Go time.String().
function formatDateOnly(dateTimeString) {
    if (!dateTimeString) return '—';

    const normalized = String(dateTimeString)
        .replace(' +0000 UTC', 'Z')
        .replace(' ', 'T');

    const d = new Date(normalized);
    if (isNaN(d.getTime())) return '—';

    return d.toLocaleDateString('ru-RU', {
        timeZone: 'UTC',
        day: '2-digit', month: '2-digit', year: 'numeric'
    });
}

function escapeHtml(str) {
    if (!str) return '';
    return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

// =================================================================
// 9. DEV UI SWITCHER — кнопки ролей выполняют НАСТОЯЩУЮ авторизацию
// через обязательный эндпоинт /dummyLogin. Без валидного токена от
// бэкенда ни один рабочий экран не открывается. Кнопка Auth — разлогин.
// =================================================================

async function devLoginAs(role) {
    // Сбрасываем текущую сессию и показываем экран Auth, чтобы возможная
    // ошибка /dummyLogin была видна пользователю в auth-alert.
    logout();
    await handleDummyLogin(role);
}

function devLogoutToAuth() {
    logout();
}

// Функция для запуска живых UTC-часов в углу экрана
function initUtcClock() {
    const timeVal = document.getElementById('utc-time-val');
    if (!timeVal) return;

    function updateClock() {
        const now = new Date();

        // Извлекаем компоненты времени UTC
        const day = String(now.getUTCDate()).padStart(2, '0');
        const month = String(now.getUTCMonth() + 1).padStart(2, '0'); // Месяцы с 0
        const year = now.getUTCFullYear();

        const hours = String(now.getUTCHours()).padStart(2, '0');
        const minutes = String(now.getUTCMinutes()).padStart(2, '0');
        const seconds = String(now.getUTCSeconds()).padStart(2, '0');

        // Собираем строку формата DD.MM.YYYY HH:MM:SS
        timeVal.textContent = `${day}.${month}.${year} ${hours}:${minutes}:${seconds}`;
    }

    // Запускаем сразу и вешаем интервал на каждую секунду
    updateClock();
    setInterval(updateClock, 1000);
}