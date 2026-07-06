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

const I18N = {
    ru: {
        auth_subtitle: 'Вход в корпоративную систему бронирования',
        tab_signin: 'Войти',
        tab_signup: 'Регистрация',
        auth_alert_default: 'Ошибка аутентификации',
        label_email: 'Email / Логин',
        label_password: 'Пароль',
        btn_signin_submit: 'Войти в систему',
        placeholder_signup_password: 'минимум 6 символов',
        label_role: 'Выбор роли',
        option_role_user: 'Сотрудник (User)',
        option_role_admin: 'Администратор (Admin)',
        btn_signup_submit: 'Зарегистрироваться',
        divider_quick_login: 'БЫСТРЫЙ ВХОД ДЛЯ ТЕСТОВ',
        btn_dummy_user_title: 'Вход как User',
        btn_dummy_user_sub: 'Тестовый сотрудник',
        btn_dummy_admin_title: 'Вход как Admin',
        btn_dummy_admin_sub: 'Тестовый админ',
        dummy_login_note: 'Использует детерминированные UUID бэкенда для мгновенной проверки логики без ввода паролей.',
        badge_employee: 'Сотрудник',
        badge_admin: 'Администратор',
        btn_logout: 'Выйти',
        rooms_heading: 'Переговорные комнаты',
        rooms_subheading: 'Выберите помещение из доступных',
        slots_heading: 'Временная сетка',
        slots_subheading: 'Слоты генерируются автоматически',
        slots_empty_message: 'Нет доступных слотов на выбранную дату. Возможно, у комнаты нет расписания или все слоты заняты.',
        slots_select_prompt: 'Выберите временной слот',
        my_bookings_heading: 'Мои бронирования',
        my_bookings_subheading: 'Ваши текущие запланированные встречи',
        modal_title: 'Подтверждение бронирования',
        modal_intro: 'Вы собираетесь забронировать временной интервал в системе.',
        modal_room_label: 'Комната:',
        modal_date_label: 'Дата:',
        modal_time_label: 'Время:',
        modal_create_conf: 'Создать онлайн-встречу (Conference Link)',
        modal_conf_note: 'Если внешний сервис ссылок недоступен, бронирование будет автоматически отменено базой.',
        modal_cancel: 'Отмена',
        modal_confirm: 'Подтвердить',
        tab_manage: 'Управление ресурсами',
        tab_log: 'Журнал всех броней',
        create_room_title: 'Создать новую переговорную',
        label_room_name: 'Название комнаты',
        placeholder_room_name: 'Например: Переговорка 147',
        label_room_capacity: 'Вместимость (чел.)',
        label_room_desc: 'Описание / Оборудование',
        placeholder_room_desc: 'Проектор, Смарт-ТВ, маркерная доска...',
        btn_add_room: 'Добавить комнату',
        schedule_title: 'Настройка доступности (Расписание)',
        schedule_warning: 'Внимание! После создания расписание нельзя изменить. Длительность слота жестко равна 30 минутам.',
        label_select_room: 'Выберите комнату',
        option_loading_rooms: 'Загрузка списка переговорок...',
        label_days_of_week: 'Дни недели доступности',
        day_mon: 'Пн', day_tue: 'Вт', day_wed: 'Ср', day_thu: 'Чт', day_fri: 'Пт', day_sat: 'Сб', day_sun: 'Вс',
        label_work_start: 'Начало работы',
        label_work_end: 'Конец работы',
        btn_activate_schedule: 'Активировать временную сетку',
        log_heading: 'Логи бронирования всей компании',
        log_subheading: 'Полный список операций с поддержкой серверной пагинации',
        th_booking_id: 'ID брони',
        th_slot: 'Слот',
        th_user: 'Пользователь',
        th_created: 'Создано (UTC)',
        th_status: 'Статус',
        btn_page_prev: '« Назад',
        btn_page_next: 'Вперед »',
        rights_full_access: 'Права: Полный доступ',

        // Динамика / JS-строки
        toast_close: 'Закрыть',
        toast_register_success: 'Регистрация успешна! Теперь вы можете войти.',
        toast_rooms_load_failed: 'Не удалось загрузить список переговорок: {error}',
        toast_slots_load_failed: 'Не удалось загрузить слоты: {error}',
        toast_booking_created: 'Бронирование успешно создано!',
        toast_bookings_load_failed: 'Не удалось загрузить ваши бронирования: {error}',
        toast_booking_cancelled: 'Бронирование отменено',
        toast_booking_cancel_failed: 'Не удалось отменить бронирование: {error}',
        toast_room_created: 'Комната успешно создана!',
        toast_room_create_failed: 'Ошибка добавления переговорки: {error}',
        toast_select_room_first: 'Сначала создайте и выберите переговорку',
        toast_select_day: 'Выберите хотя бы один день недели!',
        toast_schedule_activated: 'Временная сетка расписания успешно активирована!',
        toast_schedule_activate_failed: 'Ошибка активации расписания: {error}',
        toast_logs_load_failed: 'Не удалось загрузить журнал броней: {error}',

        rooms_empty: 'Переговорки пока не добавлены',
        room_capacity_suffix: '{n} чел',
        room_no_description: 'Описание отсутствует',
        slots_free: 'Свободно',
        slots_book_button: 'Забронировать выбранный слот ({time})',
        modal_selected_room: 'Выбранная комната',
        bookings_empty: 'У вас нет активных бронирований',
        bookings_load_failed_short: 'Не удалось загрузить бронирования',
        conference_label: 'Видеоконференция',
        conference_join: 'Войти',
        booking_room_unavailable: 'Переговорная комната (данные недоступны)',
        booking_time_unavailable: 'Время недоступно',
        booking_id_label: 'ID Брони:',
        booking_cancel_button: 'Отменить бронирование',
        booking_cancel_confirm: 'Вы уверены, что хотите отменить эту бронь?',
        admin_select_room_first_option: 'Сначала создайте переговорку',
        admin_room_capacity_suffix: 'до {n} чел.',
        admin_log_page_label: 'Страница {n}',
        admin_log_load_failed_row: 'Не удалось загрузить данные',
        admin_log_empty_row: 'Записи логов отсутствуют',
        id_prefix: 'ID: {id}...',

        // Ошибки бэкенда
        err_invalid_json: 'Некорректный формат запроса. Попробуйте ещё раз.',
        err_invalid_input: 'Проверьте правильность заполнения полей формы.',
        err_invalid_id_format: 'Некорректный идентификатор ресурса.',
        err_missing_auth: 'Вы не авторизованы. Пожалуйста, войдите заново.',
        err_invalid_token: 'Сессия истекла. Пожалуйста, войдите заново.',
        err_insufficient_permissions: 'У вас недостаточно прав для этого действия.',
        err_invalid_credentials: 'Неверный email или пароль.',
        err_user_not_found: 'Тестовый пользователь для этой роли не настроен на сервере.',
        err_room_not_found: 'Переговорка не найдена. Возможно, она была удалена.',
        err_schedule_not_found: 'Для этой переговорки ещё не задано расписание.',
        err_slot_not_found: 'Выбранный временной слот не найден.',
        err_booking_not_found: 'Бронирование не найдено.',
        err_email_exists: 'Пользователь с таким email уже зарегистрирован.',
        err_schedule_exists: 'Расписание для этой переговорки уже создано и не может быть изменено.',
        err_booking_exists: 'Этот слот уже занят другим бронированием. Обновите список слотов.',
        err_slot_past: 'Нельзя забронировать слот в прошлом.',
        err_internal: 'Внутренняя ошибка сервера. Попробуйте позже.',
        err_status_0: 'Не удалось подключиться к серверу. Убедитесь, что бэкенд запущен и доступен по адресу {url}.',
        err_status_400: 'Некорректный запрос.',
        err_status_401: 'Требуется авторизация.',
        err_status_403: 'Недостаточно прав для этого действия.',
        err_status_404: 'Запрашиваемый ресурс не найден.',
        err_status_409: 'Конфликт данных: операция невозможна в текущем состоянии.',
        err_status_500: 'Внутренняя ошибка сервера. Попробуйте позже.',
        err_unexpected: 'Произошла непредвиденная ошибка. Попробуйте ещё раз.'
    },
    en: {
        auth_subtitle: 'Sign in to the corporate booking system',
        tab_signin: 'Sign In',
        tab_signup: 'Sign Up',
        auth_alert_default: 'Authentication error',
        label_email: 'Email / Login',
        label_password: 'Password',
        btn_signin_submit: 'Sign in',
        placeholder_signup_password: 'minimum 6 characters',
        label_role: 'Select role',
        option_role_user: 'Employee (User)',
        option_role_admin: 'Administrator (Admin)',
        btn_signup_submit: 'Sign up',
        divider_quick_login: 'QUICK LOGIN FOR TESTING',
        btn_dummy_user_title: 'Sign in as User',
        btn_dummy_user_sub: 'Test employee',
        btn_dummy_admin_title: 'Sign in as Admin',
        btn_dummy_admin_sub: 'Test admin',
        dummy_login_note: 'Uses deterministic backend UUIDs for instant logic testing without entering passwords.',
        badge_employee: 'Employee',
        badge_admin: 'Administrator',
        btn_logout: 'Log out',
        rooms_heading: 'Meeting Rooms',
        rooms_subheading: 'Select a room from the available ones',
        slots_heading: 'Time Grid',
        slots_subheading: 'Slots are generated automatically',
        slots_empty_message: 'No slots available for the selected date. The room may have no schedule, or all slots are taken.',
        slots_select_prompt: 'Select a time slot',
        my_bookings_heading: 'My Bookings',
        my_bookings_subheading: 'Your current scheduled meetings',
        modal_title: 'Confirm Booking',
        modal_intro: 'You are about to book a time slot in the system.',
        modal_room_label: 'Room:',
        modal_date_label: 'Date:',
        modal_time_label: 'Time:',
        modal_create_conf: 'Create an online meeting (Conference Link)',
        modal_conf_note: 'If the external link service is unavailable, the booking will be automatically cancelled by the backend.',
        modal_cancel: 'Cancel',
        modal_confirm: 'Confirm',
        tab_manage: 'Manage Resources',
        tab_log: 'All Bookings Log',
        create_room_title: 'Create a new meeting room',
        label_room_name: 'Room name',
        placeholder_room_name: 'E.g.: Meeting Room 147',
        label_room_capacity: 'Capacity (people)',
        label_room_desc: 'Description / Equipment',
        placeholder_room_desc: 'Projector, Smart TV, whiteboard...',
        btn_add_room: 'Add room',
        schedule_title: 'Availability Settings (Schedule)',
        schedule_warning: 'Warning! The schedule cannot be changed once created. Slot duration is fixed at 30 minutes.',
        label_select_room: 'Select a room',
        option_loading_rooms: 'Loading room list...',
        label_days_of_week: 'Available days of the week',
        day_mon: 'Mon', day_tue: 'Tue', day_wed: 'Wed', day_thu: 'Thu', day_fri: 'Fri', day_sat: 'Sat', day_sun: 'Sun',
        label_work_start: 'Work start',
        label_work_end: 'Work end',
        btn_activate_schedule: 'Activate time grid',
        log_heading: 'Company-wide booking logs',
        log_subheading: 'Full list of operations with server-side pagination',
        th_booking_id: 'Booking ID',
        th_slot: 'Slot',
        th_user: 'User',
        th_created: 'Created (UTC)',
        th_status: 'Status',
        btn_page_prev: '« Prev',
        btn_page_next: 'Next »',
        rights_full_access: 'Rights: Full access',

        toast_close: 'Close',
        toast_register_success: 'Registration successful! You can now sign in.',
        toast_rooms_load_failed: 'Failed to load room list: {error}',
        toast_slots_load_failed: 'Failed to load slots: {error}',
        toast_booking_created: 'Booking created successfully!',
        toast_bookings_load_failed: 'Failed to load your bookings: {error}',
        toast_booking_cancelled: 'Booking cancelled',
        toast_booking_cancel_failed: 'Failed to cancel booking: {error}',
        toast_room_created: 'Room created successfully!',
        toast_room_create_failed: 'Error adding room: {error}',
        toast_select_room_first: 'Please create and select a room first',
        toast_select_day: 'Select at least one day of the week!',
        toast_schedule_activated: 'Schedule time grid activated successfully!',
        toast_schedule_activate_failed: 'Error activating schedule: {error}',
        toast_logs_load_failed: 'Failed to load booking log: {error}',

        rooms_empty: 'No meeting rooms added yet',
        room_capacity_suffix: '{n} ppl',
        room_no_description: 'No description',
        slots_free: 'Free',
        slots_book_button: 'Book selected slot ({time})',
        modal_selected_room: 'Selected room',
        bookings_empty: 'You have no active bookings',
        bookings_load_failed_short: 'Failed to load bookings',
        conference_label: 'Video conference',
        conference_join: 'Join',
        booking_room_unavailable: 'Meeting room (data unavailable)',
        booking_time_unavailable: 'Time unavailable',
        booking_id_label: 'Booking ID:',
        booking_cancel_button: 'Cancel booking',
        booking_cancel_confirm: 'Are you sure you want to cancel this booking?',
        admin_select_room_first_option: 'Create a meeting room first',
        admin_room_capacity_suffix: 'up to {n} ppl.',
        admin_log_page_label: 'Page {n}',
        admin_log_load_failed_row: 'Failed to load data',
        admin_log_empty_row: 'No log entries',
        id_prefix: 'ID: {id}...',

        err_invalid_json: 'Invalid request format. Please try again.',
        err_invalid_input: 'Please check the form fields.',
        err_invalid_id_format: 'Invalid resource identifier.',
        err_missing_auth: 'You are not authorized. Please sign in again.',
        err_invalid_token: 'Session expired. Please sign in again.',
        err_insufficient_permissions: 'You do not have permission to perform this action.',
        err_invalid_credentials: 'Invalid email or password.',
        err_user_not_found: 'A test user for this role is not configured on the server.',
        err_room_not_found: 'Meeting room not found. It may have been deleted.',
        err_schedule_not_found: 'No schedule has been set for this room yet.',
        err_slot_not_found: 'The selected time slot was not found.',
        err_booking_not_found: 'Booking not found.',
        err_email_exists: 'A user with this email is already registered.',
        err_schedule_exists: 'A schedule for this room already exists and cannot be changed.',
        err_booking_exists: 'This slot is already taken by another booking. Refresh the slot list.',
        err_slot_past: 'You cannot book a slot in the past.',
        err_internal: 'Internal server error. Please try again later.',
        err_status_0: 'Could not connect to the server. Make sure the backend is running and reachable at {url}.',
        err_status_400: 'Invalid request.',
        err_status_401: 'Authorization required.',
        err_status_403: 'Insufficient permissions for this action.',
        err_status_404: 'Requested resource not found.',
        err_status_409: 'Data conflict: the operation is not possible in the current state.',
        err_status_500: 'Internal server error. Please try again later.',
        err_unexpected: 'An unexpected error occurred. Please try again.'
    }
};

let currentLang = localStorage.getItem('mr_lang') || 'ru';

function t(key, vars) {
    const dict = I18N[currentLang] || I18N.ru;
    let str = (dict && dict[key] !== undefined) ? dict[key] : (I18N.ru[key] !== undefined ? I18N.ru[key] : key);
    if (vars) {
        Object.keys(vars).forEach(k => {
            str = str.replace(new RegExp('\\{' + k + '\\}', 'g'), vars[k]);
        });
    }
    return str;
}

function applyStaticTranslations() {
    document.querySelectorAll('[data-i18n]').forEach(el => {
        el.textContent = t(el.getAttribute('data-i18n'));
    });
    document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
        el.setAttribute('placeholder', t(el.getAttribute('data-i18n-placeholder')));
    });
    document.documentElement.lang = currentLang;

    const btnRu = document.getElementById('lang-btn-ru');
    const btnEn = document.getElementById('lang-btn-en');
    if (btnRu && btnEn) {
        btnRu.classList.toggle('btn-primary', currentLang === 'ru');
        btnRu.classList.toggle('btn-ghost', currentLang !== 'ru');
        btnEn.classList.toggle('btn-primary', currentLang === 'en');
        btnEn.classList.toggle('btn-ghost', currentLang !== 'en');
    }
}

function setLang(lang) {
    if (lang !== 'ru' && lang !== 'en') return;
    currentLang = lang;
    localStorage.setItem('mr_lang', lang);
    applyStaticTranslations();

    if (state.role === 'user') {
        renderUserRooms();
        renderSlotsGrid();
        loadMyBookings();
    } else if (state.role === 'admin') {
        renderAdminRoomSelect();
        loadAdminLogs();
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const datePicker = document.getElementById('slots-date-picker');
    if (datePicker && !datePicker.value) {
        datePicker.value = new Date().toISOString().slice(0, 10);
    }

    applyStaticTranslations();
    setupEventListeners();
    initApp();
    initUtcClock();
});

function initApp() {
    if (state.token) {
        showScreen(state.role === 'admin' ? 'admin' : 'user');
        loadInitialData();
    } else {
        showScreen('auth');
    }
}

function setupEventListeners() {
    document.getElementById('form-signin').addEventListener('submit', handleSignIn);
    document.getElementById('form-signup').addEventListener('submit', handleSignUp);

    document.getElementById('btn-dummy-user').addEventListener('click', () => handleDummyLogin('user'));
    document.getElementById('btn-dummy-admin').addEventListener('click', () => handleDummyLogin('admin'));

    document.querySelectorAll('.btn-logout').forEach(btn => {
        btn.addEventListener('click', logout);
    });

    document.getElementById('slots-date-picker').addEventListener('change', () => {
        state.slots = [];
        if (state.selectedRoomId) loadSlots(state.selectedRoomId);
    });

    document.getElementById('btn-confirm-booking').addEventListener('click', createBooking);

    document.getElementById('form-create-room').addEventListener('submit', createRoom);

    document.getElementById('form-create-schedule').addEventListener('submit', createSchedule);

    document.getElementById('btn-log-prev').addEventListener('click', () => navigateAdminLogs(-1));
    document.getElementById('btn-log-next').addEventListener('click', () => navigateAdminLogs(1));

    // Set minimum date to today (only allow future dates)
    const datePicker = document.getElementById('slots-date-picker');
    const today = new Date().toISOString().slice(0, 10);
    const twoWeeksLater = new Date(new Date().getTime() + 14 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);

    datePicker.min = today;
    datePicker.max = twoWeeksLater;
}

class ApiError extends Error {
    constructor(status, code, message) {
        super(message);
        this.name = 'ApiError';
        this.status = status;
        this.code = code;
    }
}

const ERROR_MESSAGES = {
    // Validation
    'invalid json': 'err_invalid_json',
    'invalid input': 'err_invalid_input',
    'invalid identifier format': 'err_invalid_id_format',

    // Authorization
    'missing auth header': 'err_missing_auth',
    'invalid or expired token': 'err_invalid_token',
    'insufficient permissions': 'err_insufficient_permissions',
    'invalid email or password': 'err_invalid_credentials',

    // Entities not found
    'user not found': 'err_user_not_found',
    'room not found': 'err_room_not_found',
    'schedule not found': 'err_schedule_not_found',
    'slot not found': 'err_slot_not_found',
    'booking not found': 'err_booking_not_found',

    // Conflicts
    'user with given email already exists': 'err_email_exists',
    'schedule for this room already exists': 'err_schedule_exists',
    'booking for this slot already exists': 'err_booking_exists',
    'specified slot is in the past': 'err_slot_past',

    // Server
    'internal error': 'err_internal'
};

const STATUS_FALLBACK_MESSAGES = {
    0: 'err_status_0',
    400: 'err_status_400',
    401: 'err_status_401',
    403: 'err_status_403',
    404: 'err_status_404',
    409: 'err_status_409',
    500: 'err_status_500'
};

function resolveErrorMessage(status, code) {
    if (code && ERROR_MESSAGES[code]) return t(ERROR_MESSAGES[code], { url: API_BASE_URL });
    if (STATUS_FALLBACK_MESSAGES[status]) return t(STATUS_FALLBACK_MESSAGES[status], { url: API_BASE_URL });
    return t('err_unexpected');
}

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

        if (response.status === 401 && !AUTH_ENDPOINTS.includes(endpoint) && state.token) {
            forceLogoutWithMessage(apiError.message);
        }

        throw apiError;
    }

    return data;
}

function forceLogoutWithMessage(message) {
    logout();
    showAuthError(message);
}

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
        <button class="btn btn-ghost btn-xs" aria-label="${t('toast_close')}">✕</button>
    `;

    toast.querySelector('button').addEventListener('click', () => toast.remove());
    container.appendChild(toast);

    setTimeout(() => toast.remove(), 6000);
}

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

        const payload = decodeJwtPayload(data.token);
        const role = (payload && (payload.role === 'admin' || payload.role === 'user')) ? payload.role : 'user';
        applySession(data.token, role, email);
        initApp();
    } catch (err) {
        showAuthError(err.message);
    }
}

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

        showToast(t('toast_register_success'), 'success');
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

    localStorage.removeItem('mr_token');
    localStorage.removeItem('mr_role');
    localStorage.removeItem('mr_email');
    localStorage.removeItem('mr_user_id');

    showAuthError(null);
    showScreen('auth');
}


async function loadInitialData() {
    const userEmailEl = document.getElementById('user-display-email');
    if (userEmailEl) userEmailEl.textContent = state.email || '—';
    const userIdEl = document.getElementById('user-display-id');
    if (userIdEl) userIdEl.textContent = state.userId ? t('id_prefix', { id: state.userId.substring(0, 8) }) : '';

    const adminEmailEl = document.getElementById('admin-display-email');
    if (adminEmailEl) adminEmailEl.textContent = state.email || '—';
    const adminIdEl = document.getElementById('admin-display-id');
    if (adminIdEl) adminIdEl.textContent = state.userId ? t('id_prefix', { id: state.userId.substring(0, 8) }) : t('rights_full_access');

    if (state.role === 'user') {
        await loadRooms();
        await loadMyBookings();
    } else if (state.role === 'admin') {
        await loadRooms();
        await loadAdminLogs();
    }
}

async function loadRooms() {
    try {
        const data = await apiRequest('/rooms/list');
        state.rooms = data.rooms || [];
    } catch (err) {
        state.rooms = [];
        showToast(t('toast_rooms_load_failed', { error: err.message }), 'error');
    }

    if (state.role === 'user') {
        renderUserRooms();
        if (state.rooms.length > 0) {
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
        container.innerHTML = `<p class="text-xs text-base-content/40 italic text-center py-4">${t('rooms_empty')}</p>`;
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
                <span class="badge badge-sm badge-neutral font-semibold">${t('room_capacity_suffix', { n: room.capacity || 0 })}</span>
            </div>
            <p class="text-xs text-base-content/60 mt-1">${escapeHtml(room.description || t('room_no_description'))}</p>
        `;
        container.appendChild(div);
    });
}

function selectRoom(roomId) {
    state.selectedRoomId = roomId;
    renderUserRooms();
    loadSlots(roomId);
}

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
        cacheSlotMeta(state.slots, roomId);
    } catch (err) {
        state.slots = [];
        showToast(t('toast_slots_load_failed', { error: err.message }), 'error');
    }
    renderSlotsGrid();
}

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

    const keys = Object.keys(cache);
    if (keys.length > SLOT_CACHE_LIMIT) {
        keys.slice(0, keys.length - SLOT_CACHE_LIMIT).forEach(k => delete cache[k]);
    }

    try {
        localStorage.setItem(SLOT_CACHE_KEY, JSON.stringify(cache));
    } catch (e) {
        // localStorage isn't available
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
    actionBtn.textContent = t('slots_select_prompt');

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
            <span class="text-[10px] font-normal tracking-wide opacity-60">${t('slots_free')}</span>
        `;

        btn.onclick = () => {
            Array.from(grid.children).forEach(child => child.classList.remove('btn-accent'));
            btn.classList.add('btn-accent');

            state.selectedSlotId = slot.id;
            actionBtn.disabled = false;
            actionBtn.textContent = t('slots_book_button', { time: startTime });

            const currentRoom = state.rooms.find(r => r.id === state.selectedRoomId);
            document.getElementById('modal-room-name').textContent = currentRoom ? currentRoom.name : t('modal_selected_room');
            document.getElementById('modal-booking-date').textContent = selectedDate;
            document.getElementById('modal-booking-time').textContent = `${startTime} - ${endTime}`;
            showBookingModalError(null);
        };

        grid.appendChild(btn);
    });
}

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
        showToast(t('toast_booking_created'), 'success');

        await loadSlots(state.selectedRoomId);
        await loadMyBookings();
    } catch (err) {
        showBookingModalError(err.message);
        if (err.status === 409) {
            await loadSlots(state.selectedRoomId);
        }
    } finally {
        confirmBtn.disabled = false;
    }
}

async function loadMyBookings() {
    const container = document.getElementById('user-bookings-list');
    container.innerHTML = '';

    let bookings = [];
    try {
        const data = await apiRequest('/bookings/my');
        bookings = data.bookings || [];
    } catch (err) {
        showToast(t('toast_bookings_load_failed', { error: err.message }), 'error');
        container.innerHTML = `<p class="text-xs text-error/70 italic text-center py-4">${t('bookings_load_failed_short')}</p>`;
        return;
    }

    if (bookings.length === 0) {
        container.innerHTML = `<p class="text-xs text-base-content/40 italic text-center py-4">${t('bookings_empty')}</p>`;
        return;
    }

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

    enriched.sort((a, b) => {
        const statusRank = s => (s.booking.status === 'active' ? 0 : 1);
        const statusDiff = statusRank(a) - statusRank(b);
        if (statusDiff !== 0) return statusDiff;

        if (a.roomName && b.roomName) {
            return a.roomName.localeCompare(b.roomName, currentLang === 'en' ? 'en' : 'ru');
        }
        if (a.roomName && !b.roomName) return -1;
        if (!a.roomName && b.roomName) return 1;
        return 0;
    });

    enriched.forEach(({ booking, roomName, start, end }) => {
        const div = document.createElement('div');
        div.className = 'p-4 bg-base-200 rounded-xl border border-base-300 relative flex flex-col justify-between gap-3';

        const displayRoomName = roomName || t('booking_room_unavailable');
        const timeRange = (start && end) ? `${formatIsoTime(start)} - ${formatIsoTime(end)}` : null;
        const dateStr = start ? formatDateOnly(start) : null;
        const statusBadge = booking.status === 'active' ? 'badge-success' : 'badge-neutral opacity-50';

        let confLinkHtml = '';
        if (booking.conference_link) {
            confLinkHtml = `
                <div class="p-2 bg-primary/10 rounded-lg flex items-center justify-between border border-primary/20">
                    <div class="flex flex-col">
                        <span class="text-[10px] uppercase font-bold text-primary tracking-wider">${t('conference_label')}</span>
                        <span class="text-xs opacity-70 font-mono truncate max-w-[180px]">${escapeHtml(booking.conference_link)}</span>
                    </div>
                    <a href="${booking.conference_link}" target="_blank" class="btn btn-xs btn-primary font-bold px-3">${t('conference_join')}</a>
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
                    ${timeRange ? `📅 ${dateStr} | 🕒 ${timeRange}` : t('booking_time_unavailable')}
                </div>
                <div class="text-[10px] font-mono text-base-content/40 mt-0.5">
                    ${t('booking_id_label')} ${booking.id.substring(0, 8)}...
                </div>
            </div>
            ${confLinkHtml}
            ${booking.status === 'active' ? `<button class="btn btn-xs btn-outline btn-error w-full mt-1" onclick="cancelBooking('${booking.id}')">${t('booking_cancel_button')}</button>` : ''}
        `;
        container.appendChild(div);
    });
}

async function cancelBooking(bookingId) {
    if (!confirm(t('booking_cancel_confirm'))) return;
    try {
        await apiRequest(`/bookings/${bookingId}/cancel`, { method: 'POST' });
        showToast(t('toast_booking_cancelled'), 'success');
        await loadMyBookings();
        if (state.selectedRoomId) loadSlots(state.selectedRoomId);
    } catch (err) {
        showToast(t('toast_booking_cancel_failed', { error: err.message }), 'error');
    }
}

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
        showToast(t('toast_room_created'), 'success');
        document.getElementById('form-create-room').reset();
        await loadRooms();
    } catch (err) {
        showToast(t('toast_room_create_failed', { error: err.message }), 'error');
    }
}

function renderAdminRoomSelect() {
    const select = document.getElementById('admin-sched-room-id');
    select.innerHTML = '';

    if (state.rooms.length === 0) {
        select.innerHTML = `<option value="" disabled selected>${t('admin_select_room_first_option')}</option>`;
        return;
    }

    state.rooms.forEach(room => {
        const opt = document.createElement('option');
        opt.value = room.id;
        opt.textContent = `${room.name} (${t('admin_room_capacity_suffix', { n: room.capacity || '?' })})`;
        select.appendChild(opt);
    });
}

async function createSchedule(e) {
    e.preventDefault();
    const roomId = document.getElementById('admin-sched-room-id').value;
    if (!roomId) {
        showToast(t('toast_select_room_first'), 'warning');
        return;
    }

    const startTime = document.getElementById('admin-sched-start').value; // HH:MM
    const endTime = document.getElementById('admin-sched-end').value;     // HH:MM

    const checkedDays = Array.from(document.querySelectorAll('input[name="days"]:checked'))
        .map(cb => parseInt(cb.value, 10));

    if (checkedDays.length === 0) {
        showToast(t('toast_select_day'), 'warning');
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
        showToast(t('toast_schedule_activated'), 'success');
    } catch (err) {
        showToast(t('toast_schedule_activate_failed', { error: err.message }), 'error');
    }
}

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
        showToast(t('toast_logs_load_failed', { error: err.message }), 'error');
        tableBody.innerHTML = `<tr><td colspan="5" class="text-center text-error/70 py-4">${t('admin_log_load_failed_row')}</td></tr>`;
        return;
    }

    document.getElementById('btn-log-page').textContent = t('admin_log_page_label', { n: pagination.page || state.adminLogPage });

    const prevBtn = document.getElementById('btn-log-prev');
    const nextBtn = document.getElementById('btn-log-next');
    prevBtn.disabled = (pagination.page || state.adminLogPage) <= 1;
    if (typeof pagination.total === 'number' && typeof pagination.page_size === 'number' && pagination.page_size > 0) {
        nextBtn.disabled = (pagination.page * pagination.page_size) >= pagination.total;
    } else {
        nextBtn.disabled = bookings.length < state.adminLogPageSize;
    }

    if (bookings.length === 0) {
        tableBody.innerHTML = `<tr><td colspan="5" class="text-center opacity-50 py-4">${t('admin_log_empty_row')}</td></tr>`;
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

function formatIsoTime(dateTimeString) {
    if (!dateTimeString) return '—';

    const match = String(dateTimeString).match(/(\d{2}):(\d{2})(?::\d{2})?/);
    if (match) {
        return `${match[1]}:${match[2]}`;
    }

    const d = new Date(dateTimeString);
    if (!isNaN(d.getTime())) {
        return d.toISOString().substring(11, 16);
    }

    return '—';
}

function formatDateTime(dateTimeString) {
    if (!dateTimeString) return '—';

    const normalized = String(dateTimeString)
        .replace(' +0000 UTC', 'Z')
        .replace(' ', 'T');

    const d = new Date(normalized);
    if (isNaN(d.getTime())) return dateTimeString;

    return d.toLocaleString(currentLang === 'en' ? 'en-GB' : 'ru-RU', {
        timeZone: 'UTC',
        day: '2-digit', month: '2-digit', year: 'numeric',
        hour: '2-digit', minute: '2-digit'
    });
}

function formatDateOnly(dateTimeString) {
    if (!dateTimeString) return '—';

    const normalized = String(dateTimeString)
        .replace(' +0000 UTC', 'Z')
        .replace(' ', 'T');

    const d = new Date(normalized);
    if (isNaN(d.getTime())) return '—';

    return d.toLocaleDateString(currentLang === 'en' ? 'en-GB' : 'ru-RU', {
        timeZone: 'UTC',
        day: '2-digit', month: '2-digit', year: 'numeric'
    });
}

function escapeHtml(str) {
    if (!str) return '';
    return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

async function devLoginAs(role) {
    logout();
    await handleDummyLogin(role);
}

function devLogoutToAuth() {
    logout();
}

function initUtcClock() {
    const timeVal = document.getElementById('utc-time-val');
    if (!timeVal) return;

    function updateClock() {
        const now = new Date();

        const day = String(now.getUTCDate()).padStart(2, '0');
        const month = String(now.getUTCMonth() + 1).padStart(2, '0'); // Месяцы с 0
        const year = now.getUTCFullYear();

        const hours = String(now.getUTCHours()).padStart(2, '0');
        const minutes = String(now.getUTCMinutes()).padStart(2, '0');
        const seconds = String(now.getUTCSeconds()).padStart(2, '0');

        timeVal.textContent = `${day}.${month}.${year} ${hours}:${minutes}:${seconds}`;
    }

    updateClock();
    setInterval(updateClock, 1000);
}