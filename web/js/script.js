// Конфигурация
const API_BASE_URL = '/admin';
const editModal = new bootstrap.Modal(document.getElementById('video-modal'));
const playerModal = new bootstrap.Modal(document.getElementById('video-player-modal'));
const uploadModal = new bootstrap.Modal(document.getElementById('upload-modal'));
const fileSelectModal = new bootstrap.Modal(document.getElementById('file-select-modal'));
const uploadImageModal = new bootstrap.Modal(document.getElementById('upload-image-modal'));
const contentTypesModal = new bootstrap.Modal(document.getElementById('content-type-modal'));
const userModal = new bootstrap.Modal(document.getElementById('user-modal'));
const categoryModal = new bootstrap.Modal(document.getElementById('category-modal'));
const categoryTypesModal = new bootstrap.Modal(document.getElementById('category-types-modal'));
const imageViewModal = new bootstrap.Modal(document.getElementById('image-view-modal'));
const videoCategoriesModal = new bootstrap.Modal(document.getElementById('video-categories-modal'));
let selectedVideoCategories = [];
let selectedTypes = [];
let categoriesList = []
let currentFileInput = null;
let selectedFile = null;
let filesList = [];
let currentPage = 'videos';
let currentSort = { field: 'mod_time', order: 'desc' };
let contentTypesList = [];

const parseDate = (dateStr) => {
    if (!dateStr) return new Date(0); // Возвращаем минимальную дату для null/undefined

    const parts = dateStr.split('.');
    if (parts.length === 3) {
        const [day, month, year] = parts;
        // Создаем дату в формате YYYY-MM-DD (корректно парсится в Date)
        return new Date(`${year}-${month}-${day}`);
    }

    // Если формат не DD.MM.YYYY, пробуем стандартный парсинг
    const date = new Date(dateStr);
    return isNaN(date.getTime()) ? new Date(0) : date;
}

// Theme Management
function applyTheme(theme) {
    if (theme === 'system') {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        document.documentElement.setAttribute('data-bs-theme', prefersDark ? 'dark' : 'light');
    } else {
        document.documentElement.setAttribute('data-bs-theme', theme);
    }
    localStorage.setItem('theme', theme);
    updateThemeIcons(theme);
}

function updateThemeIcons(theme) {
    $('.theme-option').each(function() {
        const icon = $(this).find('i');
        if ($(this).data('theme') === theme) {
            icon.addClass('text-primary');
        } else {
            icon.removeClass('text-primary');
        }
    });
}

function initTheme() {
    const savedTheme = localStorage.getItem('theme') || 'system';
    applyTheme(savedTheme);

    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
        if (localStorage.getItem('theme') === 'system') {
            applyTheme('system');
        }
    });
}

// Хеширование пароля
function hashPassword(password) {
    return CryptoJS.SHA256(password).toString(CryptoJS.enc.Hex);
}

// Универсальный AJAX-запрос
async function makeRequest(config) {
    try {
        const response = await $.ajax({
            url: `${API_BASE_URL}${config.endpoint}`,
            type: config.method,
            contentType: config.contentType || 'application/json',
            data: config.method === 'GET' ? config.data : JSON.stringify(config.data),
            dataType: 'json',
            beforeSend: (xhr) => {
                const token = localStorage.getItem('auth_token');
                if (token) {
                    xhr.setRequestHeader('Authorization', `Bearer ${token}`);
                }
            }
        });
        if (config.success) config.success(response);
        return response;
    } catch (error) {
        console.error('Ошибка запроса:', error);
        if (config.error) config.error(error);
        throw error;
    }
}

// Показать ошибку
function showError(message, isFatal = false) {
    const $error = $('#error-message');
    $error.text(message).removeClass('d-none').addClass('show');
    if (!isFatal) setTimeout(() => $error.removeClass('show'), 3000);
}

// Функция для переключения между страницами
function switchPage(page) {
    currentPage = page;
    localStorage.setItem('currentAdminPage', page);

    if (page === 'videos') {
        $('#admin-panel').removeClass('d-none');
        $('#content-types-panel').addClass('d-none');
        $('#users-panel').addClass('d-none');
        $('#categories-panel').addClass('d-none');
        loadVideos();
    } else if (page === 'content-types') {
        $('#admin-panel').addClass('d-none');
        $('#content-types-panel').removeClass('d-none');
        $('#users-panel').addClass('d-none');
        $('#categories-panel').addClass('d-none');
        loadContentTypes();
    } else if (page === 'users') {
        $('#admin-panel').addClass('d-none');
        $('#content-types-panel').addClass('d-none');
        $('#users-panel').removeClass('d-none');
        $('#categories-panel').addClass('d-none');
        loadUsers();
    } else if (page === 'categories') {
        $('#admin-panel').addClass('d-none');
        $('#content-types-panel').addClass('d-none');
        $('#users-panel').addClass('d-none');
        $('#categories-panel').removeClass('d-none');
        loadCategories();
    }

    $('.page-nav-btn').removeClass('active');
    $(`.page-nav-btn[data-page="${page}"]`).addClass('active');
}

// Функция для загрузки категорий
function loadCategories() {
    makeRequest({
        endpoint: '/category',
        method: 'GET',
        success: (categories) => {
            categoriesList = categories;
            renderCategories(categories);
        },
        error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки категорий')
    });
}

// Функция для отрисовки категорий
function renderCategories(categories) {
    //console.log('Received categories:', categories);
    const $container = $('#categories-list').empty();

    if (!categories || !Array.isArray(categories) || categories.length === 0) {
        $container.html('<tr><td colspan="7" class="text-center">Нет доступных категорий</td></tr>');
        return;
    }

    categories.forEach(category => {
        try {
            const createdAt = formatDate(category.date_created);
            // Безопасная обработка types (если null или undefined)
            const typeIds = category.types ? category.types.map(t => t.id).join(', ') : '-';
            const typeNames = category.types ? category.types.map(t => t.name).join(', ') : '-';
            const imageUrl = category.img_url ? `/img/${category.img_url}` : '/img/placeholder.jpg';

            $container.append(`
                <tr>
                    <td>${category.id}</td>
                    <td>
                        <img src="${imageUrl}" alt="Превью" class="category-thumbnail" 
                             style="width: 50px; height: 50px; object-fit: cover; cursor: pointer;"
                             data-src="${imageUrl}">
                    </td>
                    <td>${category.name}</td>
                    <td>${typeIds}</td>
                    <td>${typeNames}</td>
                    <td>${createdAt}</td>
                    <td>
                        <div class="action-buttons">
                            <button class="btn btn-sm btn-outline-primary edit-category" data-id="${category.id}">
                                <i class="bi bi-pencil"></i>
                            </button>
                        </div>
                    </td>
                </tr>
            `);
        } catch (error) {
            console.error('Error rendering category:', category, error);
        }
    });

    // Остальной код обработчиков событий...
    $('.edit-category').click(function() {
        //console.log('Edit button clicked, ID:', $(this).data('id'));
        openCategoryModal($(this).data('id'));
    });

    $('.category-thumbnail').click(function() {
        $('#image-view').attr('src', $(this).data('src'));
        imageViewModal.show();
    });

    $('.sortable').off('click').click(function() {
        const sortField = $(this).data('sort');
        $('.sortable').removeClass('sorted-asc sorted-desc');

        if (currentSort.field === sortField) {
            currentSort.order = currentSort.order === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.field = sortField;
            currentSort.order = 'asc';
        }

        $(this).addClass(`sorted-${currentSort.order}`);
        renderCategories(sortCategories(categories));
    });
}

// Функция для сортировки категорий
function sortCategories(categories) {
    return [...categories].sort((a, b) => {
        if (currentSort.field === 'id') {
            return currentSort.order === 'asc' ? a.id - b.id : b.id - a.id;
        } else if (currentSort.field === 'name') {
            return currentSort.order === 'asc'
                ? a.name.localeCompare(b.name)
                : b.name.localeCompare(a.name);
        } else if (currentSort.field === 'date_created') {
            const dateA = parseDate(a.date_created);
            const dateB = parseDate(b.date_created);
            return currentSort.order === 'asc'
                ? dateA - dateB
                : dateB - dateA;
        }
        return 0;
    });
}

// Функция для открытия модального окна категории
function openCategoryModal(categoryId = null) {
   // console.log('Opening category modal for ID:', categoryId); // Добавьте эту строку
    if (categoryId) {
        // Проверка, что категория с таким ID существует
        const category = categoriesList.find(c => c.id === categoryId);
        if (!category) {
            console.error('Category not found:', categoryId);
            showError('Категория не найдена');
            return;
        }
        $('#category-modal-title').text('Редактировать категорию');
        $('#delete-category-btn').removeClass('d-none');

        makeRequest({
            endpoint: `/category/${categoryId}`,
            method: 'GET',
            success: (category) => {
                $('#category-id').val(category.id);
                $('#category-name').val(category.name);
                $('#category-img').val(category.img_url || '');

                // Заполняем выбранные типы контента
                selectedTypes = category.types.map(t => ({ id: t.id, name: t.name }));
                updateSelectedTypesDisplay();

                categoryModal.show();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки категории')
        });
    } else {
        $('#category-modal-title').text('Добавить категорию');
        $('#delete-category-btn').addClass('d-none');
        $('#category-form')[0].reset();
        $('#category-id').val('');
        selectedTypes = [];
        updateSelectedTypesDisplay();
        categoryModal.show();
    }
}

// Функция для обновления отображения выбранных типов контента
function updateSelectedTypesDisplay() {
    const $container = $('#selected-types-list').empty();

    if (selectedTypes.length === 0) {
        $('#category-types').val('');
        return;
    }

    $('#category-types').val(selectedTypes.map(t => t.name).join(', '));

    selectedTypes.forEach(type => {
        $container.append(`
            <span class="badge bg-primary me-1 mb-1">
                ${type.name}
                <button type="button" class="btn-close btn-close-white btn-sm ms-1" 
                        data-id="${type.id}" aria-label="Удалить"></button>
            </span>
        `);
    });

    $('.btn-close[data-id]').click(function() {
        const typeId = parseInt($(this).data('id'));
        selectedTypes = selectedTypes.filter(t => t.id !== typeId);
        updateSelectedTypesDisplay();
    });
}

// Функция для открытия модального окна выбора типов контента
function openCategoryTypesModal() {
    if (contentTypesList.length === 0) {
        makeRequest({
            endpoint: '/type',
            method: 'GET',
            success: (types) => {
                contentTypesList = types;
                renderCategoryTypesList(types);
                categoryTypesModal.show();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки типов контента')
        });
    } else {
        renderCategoryTypesList(contentTypesList);
        categoryTypesModal.show();
    }
}

// Функция для отрисовки списка типов контента
function renderCategoryTypesList(types) {
    const $container = $('#category-types-list').empty();

    types.forEach(type => {
        const isSelected = selectedTypes.some(t => t.id === type.id);

        $container.append(`
            <tr>
                <td>${type.id}</td>
                <td>${type.name}</td>
                <td>
                    <input type="checkbox" class="form-check-input type-checkbox" 
                           data-id="${type.id}" data-name="${type.name}"
                           ${isSelected ? 'checked' : ''}>
                </td>
            </tr>
        `);
    });

    $('.sortable').off('click').click(function() {
        const sortField = $(this).data('sort');
        $('.sortable').removeClass('sorted-asc sorted-desc');

        if (currentSort.field === sortField) {
            currentSort.order = currentSort.order === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.field = sortField;
            currentSort.order = 'asc';
        }

        $(this).addClass(`sorted-${currentSort.order}`);
        renderCategoryTypesList(sortContentTypes(types));
    });
}

// Функция для загрузки типов контента
function loadContentTypes() {
    makeRequest({
        endpoint: '/type',
        method: 'GET',
        success: (types) => {
            contentTypesList = types;
            renderContentTypes(types);
        },
        error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки типов контента')
    });
}

function formatDate(dateString) {
    if (!dateString) return 'Не указана';
    const parts = dateString.split('.');
    if (parts.length === 3) {
        const [day, month, year] = parts;
        const date = new Date(`${year}-${month}-${day}`);
        if (!isNaN(date.getTime())) {
            return date.toLocaleDateString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric'
            });
        }
    }
    return dateString;
}

// Функция для отрисовки типов контента
function renderContentTypes(types) {
    const $container = $('#content-types-list').empty();

    if (!types?.length) {
        $container.html('<tr><td colspan="4" class="text-center">Нет доступных типов контента</td></tr>');
        return;
    }

    types.forEach(type => {
        const createdAt = formatDate(type.created_at);
        $container.append(`
            <tr>
                <td>${type.id}</td>
                <td>${type.name}</td>
                <td>${createdAt}</td>
                <td>
                    <div class="action-buttons">
                        <button class="btn btn-sm btn-outline-primary edit-content-type" data-id="${type.id}">
                            <i class="bi bi-pencil"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `);
    });

    $('.edit-content-type').click(function() {
        openContentTypeModal($(this).data('id'));
    });

    $('.sortable').off('click').click(function() {
        const sortField = $(this).data('sort');
        $('.sortable').removeClass('sorted-asc sorted-desc');

        if (currentSort.field === sortField) {
            currentSort.order = currentSort.order === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.field = sortField;
            currentSort.order = 'asc';
        }

        $(this).addClass(`sorted-${currentSort.order}`);
        renderContentTypes(sortContentTypes(types));
    });
}

// Функция для сортировки типов контента
function sortContentTypes(types) {
    return [...types].sort((a, b) => {
        if (currentSort.field === 'id') {
            return currentSort.order === 'asc' ? a.id - b.id : b.id - a.id;
        } else if (currentSort.field === 'name') {
            return currentSort.order === 'asc'
                ? a.name.localeCompare(b.name)
                : b.name.localeCompare(a.name);
        } else if (currentSort.field === 'created_at') {
            const dateA = parseDate(a.created_at);
            const dateB = parseDate(b.created_at);
            return currentSort.order === 'asc'
                ? dateA - dateB
                : dateB - dateA;
        }
        return 0;
    });
}

// Функция для открытия модального окна типа контента
function openContentTypeModal(typeId = null) {
    if (typeId) {
        $('#content-type-modal-title').text('Редактировать тип контента');
        $('#delete-content-type-btn').removeClass('d-none');

        makeRequest({
            endpoint: `/type/${typeId}`,
            method: 'GET',
            success: (type) => {
                $('#content-type-id').val(type.id);
                $('#content-type-name').val(type.name);
                contentTypesModal.show();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки типа контента')
        });
    } else {
        $('#content-type-modal-title').text('Добавить тип контента');
        $('#delete-content-type-btn').addClass('d-none');
        $('#content-type-form')[0].reset();
        $('#content-type-id').val(''); // Явно сбрасываем ID
        contentTypesModal.show();
    }
}

// Загрузка видео
function loadVideos() {
    makeRequest({
        endpoint: '/video',
        method: 'GET',
        success: (videos) => renderVideos(videos),
        error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки видео')
    });
}

// Отрисовка видео
function renderVideos(videos) {
    const $container = $('#videos-container').empty();

    if (!videos?.length) {
        $container.html('<div class="alert alert-info">Нет доступных видео</div>');
        return;
    }

    $container.html(`
        <div class="table-responsive">
            <table class="table table-hover">
                <thead>
                    <tr>
                        <th>Превью</th>
                        <th class="sortable" data-sort="id">ID</th>
                        <th class="sortable" data-sort="name">Название</th>
                        <th>Описание</th>
                        <th>ID категорий</th>
                        <th>Категории</th>
                        <th class="sortable" data-sort="created_at">Дата создания</th>
                        <th>Действия</th>
                    </tr>
                </thead>
                <tbody id="videos-table-body"></tbody>
            </table>
        </div>
    `);

    const $tbody = $('#videos-table-body');

    videos.forEach(video => {
       // console.log(video)
        const imageUrl = video.img_url ? `/img/${video.img_url}` : '/img/placeholder.jpg';
        const categories = video.categories || [];

        // Формируем списки ID и названий категорий
        const categoryIds = categories.length > 0
            ? categories.map(c => c.id).join(', ')
            : '-';

        const categoryNames = categories.length > 0
            ? categories.map(c => c.name).join(', ')
            : '-';

        const createdAt = formatDate(video.created_at);
        const videoFilename = video.url.split('/').pop();

        $tbody.append(`
            <tr>
                <td>
                    <img src="${imageUrl}" alt="Превью" class="video-thumbnail" 
                         style="width: 60px; height: 40px; object-fit: cover; cursor: pointer;"
                         data-src="${imageUrl}">
                </td>
                <td>${video.id}</td>
                <td>${video.name || '-'}</td>
                <td>${video.description || '-'}</td>
                <td>${categoryIds}</td>
                <td>${categoryNames}</td>
                <td>${createdAt}</td>
                <td>
                    <div class="d-flex gap-2">
                        <button class="btn btn-sm btn-outline-primary play-video-btn" 
                                data-video="${videoFilename}" title="Воспроизвести">
                            <i class="bi bi-play-fill"></i>
                        </button>
                        <button class="btn btn-sm btn-outline-secondary edit-video-btn" 
                                data-id="${video.id}" title="Редактировать">
                            <i class="bi bi-pencil"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `);
    });

    // Обработчики событий остаются без изменений
    $('.play-video-btn').click(function() {
        playVideo($(this).data('video'));
    });

    $('.edit-video-btn').click(function() {
        openVideoModal($(this).data('id'));
    });

    $('.video-thumbnail').click(function() {
        $('#image-view').attr('src', $(this).data('src'));
        imageViewModal.show();
    });

    $('.sortable').off('click').click(function() {
        const sortField = $(this).data('sort');
        $('.sortable').removeClass('sorted-asc sorted-desc');

        if (currentSort.field === sortField) {
            currentSort.order = currentSort.order === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.field = sortField;
            currentSort.order = 'asc';
        }

        $(this).addClass(`sorted-${currentSort.order}`);
        renderVideos(sortVideos(videos));
    });
}

// Функция сортировки видео
function sortVideos(videos) {
    return [...videos].sort((a, b) => {
        if (currentSort.field === 'id') {
            return currentSort.order === 'asc' ? a.id - b.id : b.id - a.id;
        } else if (currentSort.field === 'name') {
            return currentSort.order === 'asc'
                ? a.name.localeCompare(b.name)
                : b.name.localeCompare(a.name);
        } else if (currentSort.field === 'created_at') {
            const dateA = parseDate(a.created_at);
            const dateB = parseDate(b.created_at);
            return currentSort.order === 'asc'
                ? dateA - dateB
                : dateB - dateA;
        }
        return 0;
    });
}

// Воспроизведение видео
function playVideo(filename) {
    const player = document.getElementById('video-player');
    const title = $(`.play-button[data-video="${filename}"]`).closest('.video-card').find('h3').text();

    $('#video-player-title').text(title);
    player.src = `/video/${filename}`;
    player.load();
    playerModal.show();

    playerModal._element.addEventListener('hidden.bs.modal', () => {
        player.pause();
        player.currentTime = 0;
    }, { once: true });
}

// Открытие модального окна редактирования
function openVideoModal(videoId = null) {
    if (videoId) {
        $('#modal-title').text('Редактировать видео');
        $('#delete-btn').removeClass('d-none');

        makeRequest({
            endpoint: `/video/${videoId}`,
            method: 'GET',
            success: (video) => {
                $('#video-id').val(video.id);
                $('#video-name').val(video.name);
                $('#video-url').val(video.url);
                $('#video-img').val(video.img_url || '');
                $('#video-desc').val(video.description || '');

                // Заполняем выбранные категории
                selectedVideoCategories = video.categories || [];
                updateSelectedVideoCategoriesDisplay();

                editModal.show();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки видео')
        });
    } else {
        $('#modal-title').text('Добавить видео');
        $('#delete-btn').addClass('d-none');
        $('#video-form')[0].reset();
        $('#video-id').val('');
        selectedVideoCategories = [];
        updateSelectedVideoCategoriesDisplay();
        editModal.show();
    }
}

function openVideoCategoriesModal() {
    if (categoriesList.length === 0) {
        makeRequest({
            endpoint: '/category',
            method: 'GET',
            success: (categories) => {
                categoriesList = categories;
                renderVideoCategoriesList(categories);
                videoCategoriesModal.show();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки категорий')
        });
    } else {
        renderVideoCategoriesList(categoriesList);
        videoCategoriesModal.show();
    }
}

function renderVideoCategoriesList(categories) {
    const $container = $('#video-categories-list').empty();

    categories.forEach(category => {
        const isSelected = selectedVideoCategories.some(c => c.id === category.id);
        const imageUrl = category.img_url ? `/img/${category.img_url}` : '/img/placeholder.jpg';
        const createdAt = formatDate(category.date_created);

        $container.append(`
            <tr>
                <td>${category.id}</td>
                <td>
                    <img src="${imageUrl}" alt="Превью" style="width: 30px; height: 30px; object-fit: cover;">
                </td>
                <td>${category.name}</td>
                <td>${createdAt}</td>
                <td>
                    <input type="checkbox" class="form-check-input category-checkbox" 
                           data-id="${category.id}" data-name="${category.name}"
                           ${isSelected ? 'checked' : ''}>
                </td>
            </tr>
        `);
    });

    $('.sortable').off('click').click(function() {
        const sortField = $(this).data('sort');
        $('.sortable').removeClass('sorted-asc sorted-desc');

        if (currentSort.field === sortField) {
            currentSort.order = currentSort.order === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.field = sortField;
            currentSort.order = 'asc';
        }

        $(this).addClass(`sorted-${currentSort.order}`);
        renderVideoCategoriesList(sortCategories(categories));
    });
}

// Функция для обновления отображения выбранных категорий
function updateSelectedVideoCategoriesDisplay() {
    const $container = $('#selected-video-categories-list').empty();

    if (selectedVideoCategories.length === 0) {
        $('#video-categories').val('');
        return;
    }

    $('#video-categories').val(selectedVideoCategories.map(c => c.name).join(', '));

    selectedVideoCategories.forEach(category => {
        $container.append(`
            <span class="badge bg-primary me-1 mb-1">
                ${category.name}
                <button type="button" class="btn-close btn-close-white btn-sm ms-1" 
                        data-id="${category.id}" aria-label="Удалить"></button>
            </span>
        `);
    });

    $('.btn-close[data-id]').click(function() {
        const categoryId = parseInt($(this).data('id'));
        selectedVideoCategories = selectedVideoCategories.filter(c => c.id !== categoryId);
        updateSelectedVideoCategoriesDisplay();
    });
}

// Функция для загрузки пользователей
function loadUsers() {
    makeRequest({
        endpoint: '/users',
        method: 'GET',
        success: (users) => renderUsers(users),
        error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки пользователей')
    });
}

// Функция для отрисовки пользователей
function renderUsers(users) {
    const $container = $('#users-list').empty();

    if (!users?.length) {
        $container.html('<tr><td colspan="7" class="text-center">Нет доступных пользователей</td></tr>');
        return;
    }

    users.forEach(user => {
        const createdAt = formatDate(user.date_created);
        const isAdmin = user.admin ? 'Да' : 'Нет';

        $container.append(`
            <tr>
                <td>${user.id}</td>
                <td>${user.username}</td>
                <td>${user.content_type_id || '-'}</td>
                <td>${user.content_type_name || '-'}</td>
                <td>${isAdmin}</td>
                <td>${createdAt}</td>
                <td>
                    <div class="action-buttons">
                        <button class="btn btn-sm btn-outline-primary edit-user" data-id="${user.id}">
                            <i class="bi bi-pencil"></i>
                        </button>
                    </div>
                </td>
            </tr>
        `);
    }); // Закрывающая скобка для forEach

    // Обработчики событий должны быть после forEach
    $('.edit-user').click(function() {
        openUserModal($(this).data('id'));
    });

    $('.sortable').off('click').click(function() {
        const sortField = $(this).data('sort');
        $('.sortable').removeClass('sorted-asc sorted-desc');

        if (currentSort.field === sortField) {
            currentSort.order = currentSort.order === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.field = sortField;
            currentSort.order = 'asc';
        }

        $(this).addClass(`sorted-${currentSort.order}`);
        renderUsers(sortUsers(users));
    });
} // Закрывающая скобка для функции renderUsers

// Функция для сортировки пользователей
function sortUsers(users) {
    return [...users].sort((a, b) => {
        if (currentSort.field === 'id') {
            return currentSort.order === 'asc' ? a.id - b.id : b.id - a.id;
        } else if (currentSort.field === 'username') {
            return currentSort.order === 'asc'
                ? a.username.localeCompare(b.username)
                : b.username.localeCompare(a.username);
        } else if (currentSort.field === 'content_type_id') {
            return currentSort.order === 'asc'
                ? (a.content_type_id || 0) - (b.content_type_id || 0)
                : (b.content_type_id || 0) - (a.content_type_id || 0);
        } else if (currentSort.field === 'content_type_name') {
            const nameA = String(a.content_type_name || '').trim().toLowerCase();
            const nameB = String(b.content_type_name || '').trim().toLowerCase();
            const result = nameA.localeCompare(nameB);
            return currentSort.order === 'asc' ? result : -result;
        } else if (currentSort.field === 'admin') {
            return currentSort.order === 'asc'
                ? (a.admin === b.admin ? 0 : a.admin ? 1 : -1)
                : (a.admin === b.admin ? 0 : a.admin ? -1 : 1);
        } else if (currentSort.field === 'date_created') {
            const dateA = parseDate(a.date_created);
            const dateB = parseDate(b.date_created);
            return currentSort.order === 'asc'
                ? dateA - dateB
                : dateB - dateA;
        }
        return 0;
    });
}

// Функция для открытия модального окна пользователя
function openUserModal(userId = null) {
    if (contentTypesList.length === 0) {
        makeRequest({
            endpoint: '/type',
            method: 'GET',
            success: (types) => {
                contentTypesList = types;
                showUserModal(userId);
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки типов контента')
        });
    } else {
        showUserModal(userId);
    }
}

function showUserModal(userId = null) {
    if (userId) {
        $('#user-modal-title').text('Редактировать пользователя');
        $('#delete-user-btn').removeClass('d-none');

        makeRequest({
            endpoint: `/users/${userId}`,
            method: 'GET',
            success: (user) => {
                $('#user-id').val(user.id);
                $('#user-username').val(user.username);
                $('#user-admin').prop('checked', user.admin);
                $('#user-content-type-id').val(user.content_type_id || '');
                $('#user-content-type-name').val(user.content_type_name || '');
                $('#user-password').val('');
                userModal.show();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки пользователя')
        });
    } else {
        $('#user-modal-title').text('Добавить пользователя');
        $('#delete-user-btn').addClass('d-none');
        $('#user-form')[0].reset();
        $('#user-id').val(''); // Явно сбрасываем ID
        userModal.show();
    }
}

// Функция для загрузки изображений
function handleImages(files) {
    const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    const validExtensions = ['.jpg', '.jpeg', '.png', '.gif', '.webp'];
    const errors = [];
    const filesToUpload = [];

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const extension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();

        if (!validTypes.includes(file.type) && !validExtensions.includes(extension)) {
            errors.push(`Файл "${file.name}" имеет неподдерживаемый формат`);
            continue;
        }

        if (file.size > 10 * 1024 * 1024) {
            errors.push(`Файл "${file.name}" слишком большой (макс. 10MB)`);
            continue;
        }

        filesToUpload.push(file);
    }

    if (errors.length > 0) {
        $('#upload-image-errors').html(errors.join('<br>')).removeClass('d-none');
    }

    if (filesToUpload.length > 0) {
        uploadImages(filesToUpload);
    }
}

function uploadImages(files) {
    $('#upload-image-progress').removeClass('d-none');
    $('#upload-image-progress-bar').css('width', '0%');
    $('#upload-image-status').html(`Загрузка ${files.length} файла(ов)...`);

    const formData = new FormData();
    files.forEach(file => formData.append('image', file));

    $.ajax({
        url: `${API_BASE_URL}/files/img`,
        type: 'POST',
        data: formData,
        processData: false,
        contentType: false,
        beforeSend: (xhr) => {
            const token = localStorage.getItem('auth_token');
            if (token) {
                xhr.setRequestHeader('Authorization', `Bearer ${token}`);
            }
        },
        xhr: function() {
            const xhr = new XMLHttpRequest();
            xhr.upload.addEventListener('progress', function(e) {
                if (e.lengthComputable) {
                    const percent = Math.round((e.loaded / e.total) * 100);
                    $('#upload-image-progress-bar').css('width', percent + '%');
                }
            }, false);
            return xhr;
        },
        success: function() {
            $('#upload-image-status').html('<div class="alert alert-success">Изображения успешно загружены!</div>');
        },
        error: function(xhr) {
            $('#upload-image-status').html('<div class="alert alert-danger">Ошибка загрузки изображений</div>');
            $('#upload-image-errors').html(xhr.responseJSON?.error || 'Неизвестная ошибка').removeClass('d-none');
        }
    });
}

// Загрузка файлов
function handleFiles(files) {
    const validTypes = ['video/mp4', 'video/webm', 'video/ogg'];
    const validExtensions = ['.mp4', '.webm', '.ogg'];
    const errors = [];
    const filesToUpload = [];

    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const extension = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();

        if (!validTypes.includes(file.type) && !validExtensions.includes(extension)) {
            errors.push(`Файл "${file.name}" имеет неподдерживаемый формат`);
            continue;
        }

        if (file.size > 500 * 1024 * 1024) {
            errors.push(`Файл "${file.name}" слишком большой (макс. 500MB)`);
            continue;
        }

        filesToUpload.push(file);
    }

    if (errors.length > 0) {
        $('#upload-errors').html(errors.join('<br>')).removeClass('d-none');
    }

    if (filesToUpload.length > 0) {
        uploadFiles(filesToUpload);
    }
}

function uploadFiles(files) {
    $('#upload-progress').removeClass('d-none');
    $('#upload-progress-bar').css('width', '0%');
    $('#upload-status').html(`Загрузка ${files.length} файла(ов)...`);

    const formData = new FormData();
    files.forEach(file => formData.append('video', file));

    $.ajax({
        url: `${API_BASE_URL}/files/video`,
        type: 'POST',
        data: formData,
        processData: false,
        contentType: false,
        beforeSend: (xhr) => {
            const token = localStorage.getItem('auth_token');
            if (token) {
                xhr.setRequestHeader('Authorization', `Bearer ${token}`);
            }
        },
        xhr: function() {
            const xhr = new XMLHttpRequest();
            xhr.upload.addEventListener('progress', function(e) {
                if (e.lengthComputable) {
                    const percent = Math.round((e.loaded / e.total) * 100);
                    $('#upload-progress-bar').css('width', percent + '%');
                }
            }, false);
            return xhr;
        },
        success: function() {
            $('#upload-status').html('<div class="alert alert-success">Файлы успешно загружены!</div>');
            loadFilesList();
        },
        error: function(xhr) {
            $('#upload-status').html('<div class="alert alert-danger">Ошибка загрузки файлов</div>');
            $('#upload-errors').html(xhr.responseJSON?.error || 'Неизвестная ошибка').removeClass('d-none');
        }
    });
}

// Функции для выбора файлов
function openFileSelectModal(inputField, fileType = 'video') {
    currentFileInput = inputField;
    selectedFile = null;
    $('#file-search').val('');
    $('#select-file-btn').prop('disabled', true);

    const title = fileType === 'video' ? 'Выберите видеофайл' : 'Выберите изображение';
    $('#file-select-title').text(title);

    loadFilesList(fileType);
    fileSelectModal.show();
}

function loadFilesList(fileType = 'video') {
    const endpoint = fileType === 'video' ? '/files/video' : '/files/img';

    makeRequest({
        endpoint: endpoint,
        method: 'GET',
        success: function(files) {
            filesList = files;
            renderFilesList(fileType);
        },
        error: function(err) {
            showError(err.responseJSON?.error || `Ошибка загрузки списка ${fileType === 'video' ? 'видео' : 'изображений'}`);
        }
    });
}

function renderFilesList(fileType = 'video') {
    const $fileList = $('#file-list').empty();
    const searchTerm = $('#file-search').val().toLowerCase();

    let filteredFiles = filesList.filter(file =>
        file.name.toLowerCase().includes(searchTerm)
    );

    const allowedExtensions = fileType === 'video'
        ? ['.mp4', '.webm', '.ogg']
        : ['.jpg', '.jpeg', '.png', '.gif', '.webp'];

    filteredFiles = filteredFiles.filter(file => {
        const ext = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
        return allowedExtensions.includes(ext);
    });

    filteredFiles.sort((a, b) => {
        if (currentSort.field === 'size') {
            return currentSort.order === 'asc'
                ? (a.size || 0) - (b.size || 0)
                : (b.size || 0) - (a.size || 0);
        }
        else if (currentSort.field === 'mod_time') {
            const dateA = parseDate(a.mod_time);
            const dateB = parseDate(b.mod_time);
            return currentSort.order === 'asc'
                ? dateA - dateB
                : dateB - dateA;
        }
        else {
            // Сортировка по имени с учетом регистра
            const nameA = a.name || '';
            const nameB = b.name || '';
            return currentSort.order === 'asc'
                ? nameA.localeCompare(nameB, undefined, { sensitivity: 'base' })
                : nameB.localeCompare(nameA, undefined, { sensitivity: 'base' });
        }
    });

    filteredFiles.forEach(file => {
        const isSelected = selectedFile && selectedFile.name === file.name;
        const sizeMB = (file.size / (1024 * 1024)).toFixed(2);
        const modDate = new Date(file.mod_time).toLocaleString();

        $fileList.append(`
            <tr class="${isSelected ? 'selected' : ''}" data-name="${file.name}">
                <td>${file.name}</td>
                <td class="file-size">${sizeMB} MB</td>
                <td>${modDate}</td>
            </tr>
        `);
    });

    $('#file-list tr').click(function() {
        const fileName = $(this).data('name');
        selectedFile = filesList.find(f => f.name === fileName);
        $('#file-list tr').removeClass('selected');
        $(this).addClass('selected');
        $('#select-file-btn').prop('disabled', false);
    });
}

// Проверка авторизации
function checkAuth() {
    makeRequest({
        endpoint: '/video',
        method: 'GET',
        success: () => {
            document.body.classList.add('logged-in');
            $('#login-container').addClass('d-none');

            // Восстанавливаем сохранённую страницу или используем 'videos' по умолчанию
            const savedPage = localStorage.getItem('currentAdminPage') || 'videos';
            switchPage(savedPage);
        },
        error: (err) => {
            if (err.status === 401) {
                document.body.classList.remove('logged-in');
                $('#login-container').removeClass('d-none');
                $('#admin-panel, #content-types-panel, #users-panel').addClass('d-none');
                localStorage.removeItem('auth_token');
            }
        }
    });
}

function initNavigation() {
    $('.admin-header h1').after(`
        <div class="page-nav">
            <a href="#" class="page-nav-btn ${currentPage === 'videos' ? 'active' : ''}" data-page="videos">Видео</a>
            <a href="#" class="page-nav-btn ${currentPage === 'content-types' ? 'active' : ''}" data-page="content-types">Типы контента</a>
            <a href="#" class="page-nav-btn ${currentPage === 'users' ? 'active' : ''}" data-page="users">Пользователи</a>
            <a href="#" class="page-nav-btn ${currentPage === 'categories' ? 'active' : ''}" data-page="categories">Категории</a>
        </div>
    `);
}

// Инициализация при загрузке страницы
$(document).ready(function() {
    initTheme();
    initNavigation();

    // Обработчики тем
    $('.theme-option').click(function(e) {
        e.preventDefault();
        applyTheme($(this).data('theme'));
    });

    // Проверка авторизации
    checkAuth();

    // Обработчик входа
    $('#login-form').submit(async function(e) {
        e.preventDefault();
        const login = $('#login').val().trim();
        const password = $('#password').val().trim();

        if (!login || !password) {
            showError('Заполните все поля');
            return;
        }

        try {
            const hashedPassword = hashPassword(password);
            const response = await $.ajax({
                url: `${API_BASE_URL}/signin`,
                type: 'POST',
                contentType: 'application/json',
                data: JSON.stringify({ login, password: hashedPassword }),
                dataType: 'json'
            });

            if (response.message === "Authentication successful" || response.token) {
                if (response.token) {
                    localStorage.setItem('auth_token', response.token);
                }
                document.body.classList.add('logged-in');
                $('#login-container').addClass('d-none');

                // Используем сохраненную страницу или 'videos' по умолчанию
                const savedPage = localStorage.getItem('currentAdminPage') || 'videos';
                switchPage(savedPage);
            } else {
                showError('Неверный логин или пароль');
            }
        } catch (error) {
            showError(error.responseJSON?.error || 'Ошибка авторизации');
        }
    });

    // Выход
    $('#logout-btn').click(() => {
        makeRequest({
            endpoint: '/logout',
            method: 'POST',
            success: () => {
                localStorage.removeItem('currentAdminPage'); // Очищаем сохраненную страницу
                document.body.classList.remove('logged-in');
                localStorage.removeItem('auth_token');
                $('#admin-panel, #content-types-panel, #users-panel').addClass('d-none');
                $('#login-container').removeClass('d-none');
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка при выходе')
        });
    });

    // Обработчик навигации
    $(document).on('click', '.page-nav-btn', function(e) {
        e.preventDefault();
        switchPage($(this).data('page'));
    });

    // Обработчики для видео
    $('#add-video-btn').click(() => openVideoModal());
    $('#upload-image-btn').click(() => uploadImageModal.show());
    $('#upload-video-btn').click(() => uploadModal.show());
    $('#select-video-categories-btn').click(openVideoCategoriesModal);

    $('#save-selected-video-categories-btn').click(function() {
        selectedVideoCategories = [];
        $('.category-checkbox:checked').each(function() {
            selectedVideoCategories.push({
                id: parseInt($(this).data('id')),
                name: $(this).data('name')
            });
        });
        updateSelectedVideoCategoriesDisplay();
        videoCategoriesModal.hide();
    });

    $('#video-form').submit(function(e) {
        e.preventDefault();
        const videoData = {
            name: $('#video-name').val(),
            url: $('#video-url').val(),
            img_url: $('#video-img').val(),
            description: $('#video-desc').val(),
            category_ids: selectedVideoCategories.map(c => c.id)
        };

        const videoId = $('#video-id').val();
        const method = videoId ? 'PUT' : 'POST';
        const endpoint = videoId ? `/video/${videoId}` : '/video';

        makeRequest({
            endpoint,
            method,
            data: videoData,
            success: () => {
                editModal.hide();
                loadVideos();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка сохранения видео')
        });
    });

    $('#delete-btn').click(function() {
        if (confirm('Удалить это видео?')) {
            makeRequest({
                endpoint: `/video/${$('#video-id').val()}`,
                method: 'DELETE',
                success: () => {
                    editModal.hide();
                    loadVideos();
                },
                error: (err) => showError(err.responseJSON?.error || 'Ошибка удаления видео')
            });
        }
    });

    // Обработчики для типов контента
    $('#add-content-type-btn').click(() => openContentTypeModal());

    $('#content-type-form').submit(function(e) {
        e.preventDefault();
        const typeData = {
            name: $('#content-type-name').val()
        };

        const typeId = $('#content-type-id').val();
        const method = typeId ? 'PUT' : 'POST';
        const endpoint = typeId ? `/type/${typeId}` : '/type';

        makeRequest({
            endpoint,
            method,
            data: typeData,
            success: () => {
                contentTypesModal.hide();
                loadContentTypes();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка сохранения типа контента')
        });
    });

    $('#delete-content-type-btn').click(function() {
        if (confirm('Удалить этот тип контента?')) {
            makeRequest({
                endpoint: `/type/${$('#content-type-id').val()}`,
                method: 'DELETE',
                success: () => {
                    contentTypesModal.hide();
                    loadContentTypes();
                },
                error: (err) => showError(err.responseJSON?.error || 'Ошибка удаления типа контента')
            });
        }
    });

    // Обработчики для пользователей
    $('#add-user-btn').click(() => openUserModal());

    $('#user-form').submit(function(e) {
        e.preventDefault();
        const userData = {
            username: $('#user-username').val(),
            content_type_id: $('#user-content-type-id').val() || null,
            content_type_name: $('#user-content-type-name').val() || null,
            admin: $('#user-admin').is(':checked'),
            password: $('#user-password').val()
        };

        if (userData.password) {
            userData.password = hashPassword(userData.password);
        } else {
            delete userData.password;
        }

        const userId = $('#user-id').val();
        const method = userId ? 'PUT' : 'POST';
        const endpoint = userId ? `/users/${userId}` : '/users';

        makeRequest({
            endpoint,
            method,
            data: userData,
            success: () => {
                userModal.hide();
                loadUsers();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка сохранения пользователя')
        });
    });

    $('#delete-user-btn').click(function() {
        if (confirm('Удалить этого пользователя?')) {
            makeRequest({
                endpoint: `/users/${$('#user-id').val()}`,
                method: 'DELETE',
                success: () => {
                    userModal.hide();
                    loadUsers();
                },
                error: (err) => showError(err.responseJSON?.error || 'Ошибка удаления пользователя')
            });
        }
    });

    // Обработчик выбора типа контента
    $('#select-content-type-btn').click(function() {
        const $modal = $(`
        <div class="modal fade" tabindex="-1">
            <div class="modal-dialog">
                <div class="modal-content">
                    <div class="modal-header">
                        <h5 class="modal-title">Выберите тип контента</h5>
                        <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                    </div>
                    <div class="modal-body">
                        <div class="table-responsive">
                            <table class="table table-hover">
                                <thead>
                                    <tr>
                                        <th>ID</th>
                                        <th>Название</th>
                                    </tr>
                                </thead>
                                <tbody id="content-type-select-list"></tbody>
                            </table>
                        </div>
                    </div>
                    <div class="modal-footer">
                        <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Отмена</button>
                    </div>
                </div>
            </div>
        </div>
    `);

        const $list = $modal.find('#content-type-select-list');
        contentTypesList.forEach(type => {
            $list.append(`
            <tr>
                <td>${type.id}</td>
                <td><button type="button" class="btn btn-link p-0 text-start" data-id="${type.id}">${type.name}</button></td>
            </tr>
        `);
        });

        $list.find('button').click(function() {
            const typeId = $(this).data('id');
            const typeName = $(this).text();
            $('#user-content-type-id').val(typeId);
            $('#user-content-type-name').val(typeName);
            $modal.modal('hide');
        });

        $modal.modal('show');
    });

    // Обработчики для загрузки файлов
    $('#select-files-btn').click(() => $('#file-input').click());
    $('#file-input').change(function() {
        handleFiles(this.files);
        $(this).val('');
    });

    $('#select-images-btn').click(() => $('#image-input').click());
    $('#image-input').change(function() {
        handleImages(this.files);
        $(this).val('');
    });

    $('#upload-dropzone')
        .on('dragover', function(e) {
            e.preventDefault();
            $(this).addClass('active');
        })
        .on('dragleave', function() {
            $(this).removeClass('active');
        })
        .on('drop', function(e) {
            e.preventDefault();
            $(this).removeClass('active');
            handleFiles(e.originalEvent.dataTransfer.files);
        });

    $('#upload-image-dropzone')
        .on('dragover', function(e) {
            e.preventDefault();
            $(this).addClass('active');
        })
        .on('dragleave', function() {
            $(this).removeClass('active');
        })
        .on('drop', function(e) {
            e.preventDefault();
            $(this).removeClass('active');
            handleImages(e.originalEvent.dataTransfer.files);
        });

    // Обработчики для категорий
    $('#add-category-btn').click(() => openCategoryModal());
    $('#select-category-image-btn').click(() => openFileSelectModal($('#category-img')[0], 'img'));
    $('#select-category-types-btn').click(openCategoryTypesModal);

    $('#save-selected-types-btn').click(function() {
        selectedTypes = [];
        $('.type-checkbox:checked').each(function() {
            selectedTypes.push({
                id: parseInt($(this).data('id')),
                name: $(this).data('name')
            });
        });
        updateSelectedTypesDisplay();
        categoryTypesModal.hide();
    });

    $('#category-form').submit(function(e) {
        e.preventDefault();
        const categoryData = {
            name: $('#category-name').val(),
            img_url: $('#category-img').val(),
            type_ids: selectedTypes.map(t => t.id)
        };

        const categoryId = $('#category-id').val();
        const method = categoryId ? 'PUT' : 'POST';
        const endpoint = categoryId ? `/category/${categoryId}` : '/category';

        makeRequest({
            endpoint,
            method,
            data: categoryData,
            success: () => {
                categoryModal.hide();
                loadCategories();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка сохранения категории')
        });
    });

    $('#delete-category-btn').click(function() {
        if (confirm('Удалить эту категорию?')) {
            makeRequest({
                endpoint: `/category/${$('#category-id').val()}`,
                method: 'DELETE',
                success: () => {
                    categoryModal.hide();
                    loadCategories();
                },
                error: (err) => showError(err.responseJSON?.error || 'Ошибка удаления категории')
            });
        }
    });

    // Обработчики для выбора файлов
    $('#select-video-btn').click(() => openFileSelectModal($('#video-url')[0], 'video'));
    $('#select-image-btn').click(() => openFileSelectModal($('#video-img')[0], 'img'));

    $('#file-search').on('input', renderFilesList);
    $('#clear-search').click(function() {
        $('#file-search').val('');
        renderFilesList();
    });

    $('.sort-btn').click(function() {
        const sortField = $(this).data('sort');
        $('.sort-btn').removeClass('active').find('.sort-icon').html('');

        if (currentSort.field === sortField) {
            currentSort.order = currentSort.order === 'asc' ? 'desc' : 'asc';
        } else {
            currentSort.field = sortField;
            currentSort.order = 'desc';
        }

        $(this).addClass('active')
            .find('.sort-icon')
            .html(`<i class="bi bi-arrow-${currentSort.order === 'asc' ? 'up' : 'down'}"></i>`);

        renderFilesList();
    });

    $('#select-file-btn').click(function() {
        if (selectedFile && currentFileInput) {
            $(currentFileInput).val(selectedFile.name);
            fileSelectModal.hide();
        }
    });
});