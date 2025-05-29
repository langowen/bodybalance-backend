// Конфигурация
const API_BASE_URL = '/admin';
const editModal = new bootstrap.Modal(document.getElementById('video-modal'));
const playerModal = new bootstrap.Modal(document.getElementById('video-player-modal'));
const uploadModal = new bootstrap.Modal(document.getElementById('upload-modal'));
const fileSelectModal = new bootstrap.Modal(document.getElementById('file-select-modal'));
const uploadImageModal = new bootstrap.Modal(document.getElementById('upload-image-modal'));
const contentTypesModal = new bootstrap.Modal(document.getElementById('content-type-modal'));
let currentFileInput = null;
let selectedFile = null;
let filesList = [];
let currentPage = 'videos';
let currentSort = { field: 'mod_time', order: 'desc' };

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
                if (config.loading) $('#loading-spinner').removeClass('d-none');
            },
            complete: () => {
                if (config.loading) $('#loading-spinner').addClass('d-none');
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
    if (page === 'videos') {
        $('#admin-panel').removeClass('d-none');
        $('#content-types-panel').addClass('d-none');
        loadVideos();
    } else {
        $('#admin-panel').addClass('d-none');
        $('#content-types-panel').removeClass('d-none');
        loadContentTypes();
    }
}

// Функция для загрузки типов контента
function loadContentTypes() {
    makeRequest({
        endpoint: '/type',
        method: 'GET',
        loading: true,
        success: (types) => {
           // console.log('Server response:', types);
            renderContentTypes(types);
        },
        error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки типов контента')
    });
}

function formatDate(dateString) {
   // console.log('Raw date string:', dateString);
    if (!dateString) return 'Не указана';

    // Пробуем разобрать дату в формате DD.MM.YYYY
    const parts = dateString.split('.');
    if (parts.length === 3) {
        const [day, month, year] = parts;
        // Создаем дату в формате YYYY-MM-DD (месяцы в JS 0-11)
        const date = new Date(`${year}-${month}-${day}`);
        if (!isNaN(date.getTime())) {
            return date.toLocaleDateString('ru-RU', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                //hour: '2-digit',
                //minute: '2-digit'
            });
        }
    }

    // Если формат не распознан, покажем как есть
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

    // Добавляем сортировку
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
            return currentSort.order === 'asc'
                ? new Date(a.created_at) - new Date(b.created_at)
                : new Date(b.created_at) - new Date(a.created_at);
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
        contentTypesModal.show();
    }
}


// В функции $(document).ready() добавим обработчики:
$(document).ready(function() {
    // ... существующий код ...

    // Обработчики для работы с типами контента
    $('#add-content-type-btn').click(() => openContentTypeModal());

    $('#content-type-form').submit(function(e) {
        e.preventDefault();
        const typeData = {
            name: $('#content-type-name').val()
        };

        const typeId = $('#content-type-id').val();
        // Явно проверяем, есть ли ID
        const isEdit = typeId && typeId !== 'undefined' && typeId !== '';
        const method = isEdit ? 'PUT' : 'POST';
        const endpoint = isEdit ? `/type/${typeId}` : '/type';

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

    // Добавим кнопки переключения между страницами
    $('.admin-header h1').after(`
        <div class="page-nav">
            <a href="#" class="page-nav-btn ${currentPage === 'videos' ? 'active' : ''}" data-page="videos">Видео</a>
            <a href="#" class="page-nav-btn ${currentPage === 'content-types' ? 'active' : ''}" data-page="content-types">Типы контента</a>
        </div>
    `);

    $('.page-nav-btn').click(function(e) {
        e.preventDefault();
        switchPage($(this).data('page'));
    });

    // Инициализация текущей страницы
    switchPage(currentPage);
});

// Загрузка видео
function loadVideos() {
    makeRequest({
        endpoint: '/video',
        method: 'GET',
        loading: true,
        success: (videos) => renderVideos(videos),
        error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки видео')
    });
}

// Отрисовка видео
function renderVideos(videos) {
    const $container = $('#videos-container').empty();
   // console.log("Получены видео:", videos); // <- Добавьте эту строку
    if (!videos?.length) {
        $container.html('<div class="col-12"><div class="alert alert-info">Нет доступных видео</div></div>');
        return;
    }

    videos.forEach(video => {
        const imageUrl = video.img_url ? `/img/${video.img_url}` : '/img/placeholder.jpg';
        const videoFilename = video.url.split('/').pop();

        $container.append(`
            <div class="video-card">
                <div class="video-thumbnail">
                    <img src="${imageUrl}" alt="${video.name}" 
                         onerror="this.src='/img/placeholder.jpg'">
                    
                    <div class="video-overlay">
                        <button class="play-button" data-video="${videoFilename}">
                            <svg class="play-icon" viewBox="0 0 24 24">
                                <path d="M8 5v14l11-7z"/>
                            </svg>
                        </button>
                    </div>
                    
                    <div class="edit-icon-wrapper">
                        <div class="edit-icon" data-id="${video.id}">
                            <svg width="20" height="20" viewBox="0 0 24 24">
                                <path fill="currentColor" d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/>
                            </svg>
                        </div>
                    </div>
                </div>
                
                <div class="video-info">
                    <h3>${video.name}</h3>
                    <p>${video.description || 'Нет описания'}</p>
                </div>
            </div>
        `);
    });

    $('.play-button').off('click').on('click', function(e) {
        e.stopPropagation();
        playVideo($(this).data('video'));
    });

    $('.edit-icon').off('click').on('click', function(e) {
        e.stopPropagation();
        openVideoModal($(this).data('id'));
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
                editModal.show();
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка загрузки видео')
        });
    } else {
        $('#modal-title').text('Добавить видео');
        $('#delete-btn').addClass('d-none');
        $('#video-form')[0].reset();
        editModal.show();
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

    // Устанавливаем заголовок в зависимости от типа файлов
    const title = fileType === 'video' ? 'Выберите видеофайл' : 'Выберите изображение';
    $('#file-select-title').text(title);

    // Загружаем список файлов соответствующего типа
    loadFilesList(fileType);
    fileSelectModal.show();
}

// Обновленная функция загрузки списка файлов
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

// Обновляем обработчики для кнопок выбора файлов
$('#select-video-btn').click(() => openFileSelectModal($('#video-url')[0], 'video'));
$('#select-image-btn').click(() => openFileSelectModal($('#video-img')[0], 'img'));

// Добавляем обработчики для загрузки изображений
$('#select-images-btn').click(() => $('#image-input').click());
$('#image-input').change(function() {
    handleImages(this.files);
    $(this).val('');
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

function renderFilesList(fileType = 'video') {
    const $fileList = $('#file-list').empty();
    const searchTerm = $('#file-search').val().toLowerCase();

    let filteredFiles = filesList.filter(file =>
        file.name.toLowerCase().includes(searchTerm)
    );

    // Фильтрация по расширениям файлов на клиенте (дополнительная проверка)
    const allowedExtensions = fileType === 'video'
        ? ['.mp4', '.webm', '.ogg']
        : ['.jpg', '.jpeg', '.png', '.gif', '.webp'];

    filteredFiles = filteredFiles.filter(file => {
        const ext = file.name.substring(file.name.lastIndexOf('.')).toLowerCase();
        return allowedExtensions.includes(ext);
    });

    filteredFiles.sort((a, b) => {
        if (currentSort.field === 'size') {
            return currentSort.order === 'asc' ? a.size - b.size : b.size - a.size;
        } else if (currentSort.field === 'mod_time') {
            return currentSort.order === 'asc'
                ? new Date(a.mod_time) - new Date(b.mod_time)
                : new Date(b.mod_time) - new Date(a.mod_time);
        } else {
            return currentSort.order === 'asc'
                ? a.name.localeCompare(b.name)
                : b.name.localeCompare(a.name);
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

// Инициализация
$(document).ready(function() {
    initTheme();

    $('.theme-option').click(function(e) {
        e.preventDefault();
        const theme = $(this).data('theme');
        applyTheme(theme);
    });

    // Проверка авторизации при загрузке
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

            // Условие изменено: проверяем success или message
            if (response.message === "Authentication successful" || response.token) {
                // Если токен есть — сохраняем, если нет — запрашиваем видео
                if (response.token) {
                    localStorage.setItem('auth_token', response.token);
                }
                $('#login-container').addClass('d-none');
                $('#admin-panel').removeClass('d-none');
                loadVideos(); // Принудительно загружаем видео
            } else {
                showError('Неверный логин или пароль');
            }
        } catch (error) {
            console.error("Ошибка входа:", error);
            showError(error.responseJSON?.error || 'Ошибка авторизации');
        }
    });

    // Проверка авторизации
    function checkAuth() {
        makeRequest({
            endpoint: '/video',
            method: 'GET',
            success: () => {
                $('#login-container').addClass('d-none');
                $('#admin-panel').removeClass('d-none');
                loadVideos();
            },
            error: (err) => {
                if (err.status === 401) {
                    $('#login-container').removeClass('d-none');
                    $('#admin-panel').addClass('d-none');
                    localStorage.removeItem('auth_token');
                }
            }
        });
    }

    // Выход
    $('#logout-btn').click(() => {
        makeRequest({
            endpoint: '/logout',
            method: 'POST',
            success: () => {
                localStorage.removeItem('auth_token');
                $('#admin-panel').removeClass('show').addClass('d-none');
                $('#login-container').removeClass('d-none').addClass('show');
            },
            error: (err) => showError(err.responseJSON?.error || 'Ошибка при выходе')
        });
    });

    // Управление видео
    $('#add-video-btn').click(() => openVideoModal());
    $('#upload-image-btn').click(() => uploadImageModal.show());
    $('#upload-video-btn').click(() => uploadModal.show());

    $('#video-form').submit(function(e) {
        e.preventDefault();
        const videoData = {
            name: $('#video-name').val(),
            url: $('#video-url').val(),
            img_url: $('#video-img').val(),
            description: $('#video-desc').val()
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

    // Загрузка файлов
    $('#select-files-btn').click(() => $('#file-input').click());
    $('#file-input').change(function() {
        handleFiles(this.files);
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

    // Выбор файлов
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