# Управление бинарными файлами

## Обзор

В проекте настроена система исключения бинарных файлов из git для поддержания чистоты репозитория и предотвращения случайного коммита исполняемых файлов.

## Исключенные файлы

### Go бинарные файлы
- `main` - основной исполняемый файл
- `app` - приложение
- `server` - сервер
- `driver_bot` - бот для водителей
- `admin_bot` - административный бот
- `*.exe` - Windows исполняемые файлы
- `*.dll` - Windows библиотеки
- `*.so` - Linux библиотеки
- `*.dylib` - macOS библиотеки

### Тестовые и отладочные файлы
- `*.test` - тестовые бинарные файлы
- `*.out` - файлы покрытия кода
- `*.cover` - отчеты покрытия

### Сборки
- `bin/` - папка с бинарными файлами
- `dist/` - папка с дистрибутивами
- `build/` - папка сборки
- `target/` - целевые файлы

## Команды управления

### Очистка бинарных файлов
```bash
make clean-bin
```

Эта команда удалит все бинарные файлы из проекта:
- Исполняемые файлы Go
- Тестовые бинарные файлы
- Файлы покрытия кода

### Полная очистка
```bash
make clean
```

Эта команда очистит:
- Docker контейнеры и образы
- Бинарные файлы
- Временные файлы

## .gitignore правила

### Основные правила
```gitignore
# Go binaries
main
app
server
driver_bot
admin_bot
*.exe
*.dll
*.so
*.dylib

# Test binary, built with `go test -c`
*.test

# Output of the go coverage tool
*.out
*.cover

# Build artifacts
bin/
dist/
build/
target/

# Compiled binaries (any name)
backend/*/main
backend/*/app
backend/*/server
backend/*/driver_bot
backend/*/admin_bot
```

## Рекомендации

### При разработке
1. **Не коммитьте бинарные файлы** - они генерируются автоматически
2. **Используйте `make clean-bin`** после сборки для очистки
3. **Проверяйте .gitignore** при добавлении новых типов файлов

### При сборке
```bash
# Сборка приложения
cd backend
go build ./cmd/app

# Очистка после сборки
make clean-bin
```

### При деплое
```bash
# Сборка в Docker (рекомендуется)
make build

# Или локальная сборка с очисткой
cd backend && go build ./cmd/app && cd .. && make clean-bin
```

## Проверка состояния

### Поиск бинарных файлов
```bash
# Найти все исполняемые файлы
find . -type f -executable -not -path "./.git/*"

# Найти конкретные бинарные файлы
find . -name "main" -o -name "app" -o -name "*.exe"
```

### Проверка git статуса
```bash
# Проверить, что бинарные файлы не отслеживаются
git status

# Проверить .gitignore
git check-ignore backend/main
```

## Автоматизация

### Pre-commit hooks
Рекомендуется настроить pre-commit hook для автоматической очистки:

```bash
#!/bin/bash
# .git/hooks/pre-commit
make clean-bin
```

### CI/CD
В CI/CD пайплайнах добавьте очистку бинарных файлов:

```yaml
# .github/workflows/build.yml
- name: Clean binaries
  run: make clean-bin
```

## Проблемы и решения

### Бинарный файл все еще отслеживается
```bash
# Удалить из индекса
git rm --cached backend/main

# Добавить в .gitignore
echo "backend/main" >> .gitignore

# Закоммитить изменения
git add .gitignore
git commit -m "Remove binary from tracking"
```

### Файл не исключается .gitignore
```bash
# Проверить правила
git check-ignore backend/main

# Добавить правило в .gitignore
echo "backend/main" >> .gitignore
```

### Очистка истории git
Если бинарные файлы попали в историю:

```bash
# Использовать BFG Repo-Cleaner или git filter-branch
# (требует осторожности при работе с историей)
```

## Лучшие практики

1. **Всегда используйте `make clean-bin`** после локальной сборки
2. **Проверяйте git status** перед коммитом
3. **Добавляйте новые типы бинарных файлов в .gitignore**
4. **Используйте Docker для сборки** в продакшене
5. **Документируйте процесс сборки** в README 