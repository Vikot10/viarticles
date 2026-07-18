# ViArticles

ViArticles — лёгкий персональный список материалов «прочитать позже». Приложение
состоит из одного Go-бинарника, хранит данные в локальном SQLite-файле и отдаёт
серверный HTML-интерфейс, улучшенный HTMX.

## Возможности

- ручное добавление ссылок и заметок;
- автоматическое сохранение ссылок из сообщений Telegram-бота;
- отметки «прочитано» и «нравится»;
- мягкое удаление материалов;
- поиск и фильтры по статусу и источнику;
- защита Telegram-бота разрешённым `chat_id`;
- отсутствие отдельного frontend build, PostgreSQL и внешней инфраструктуры.

Telegram Bot API не читает личный раздел «Избранное». Чтобы сохранить материал,
отправьте или перешлите боту сообщение со ссылкой. Без настроенного бота ручное
добавление продолжает работать.

## Запуск

Требуется запущенный Docker Desktop с Docker Compose.

Создайте локальный файл настроек:

```powershell
Copy-Item env.example .env
```

Для Telegram укажите в `.env` токен BotFather и свой числовой `chat_id`. Если бот пока
не нужен, оставьте эти значения пустыми. Загрузите `.env` в текущую PowerShell-сессию:

```powershell
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]*)=(.*)$') {
    Set-Item -Path "Env:$($matches[1].Trim())" -Value $matches[2].Trim()
  }
}
$env:COMPOSE_DISABLE_ENV_FILE = 'true'
```

Запустите приложение:

```powershell
docker compose up --build
```

Откройте `http://localhost:8080`. Порт можно изменить через `VIARTICLES_PORT`.
SQLite-база хранится в Docker volume `viarticles-sqlite-data` и сохраняется между
перезапусками контейнера.

Остановка:

```powershell
docker compose down
```

Команда `docker compose down -v` дополнительно удалит SQLite-базу и все материалы.
