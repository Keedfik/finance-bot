# Telegram Financial Assistant Bot

Этот проект представляет собой Telegram-бота, помогающего пользователям отслеживать и управлять личными расходами. Бот позволяет добавлять расходы, распределять их по категориям, устанавливать лимиты на категории и отслеживать общие траты.

## Функциональность

- Добавление расходов с указанием суммы, даты, заметки и категории
- Просмотр списка всех расходов
- Создание и управление категориями расходов
- Установка лимита по каждой категории расходов
- Контроль за превышением лимита по категориям при добавлении новых расходов
- Удаление последнего добавленного расхода
- Получение справки с описанием доступных команд

## Установка

### Предварительные требования

- Установленный [Go](https://golang.org/doc/install) (версия 1.20 или новее)
- Установленный [MongoDB](https://docs.mongodb.com/manual/installation/) (локально или удаленно)
- Установленный [Telegram](https://telegram.org/) и зарегистрированный бот (получите токен у [BotFather](https://core.telegram.org/bots#6-botfather))

### Шаги по установке

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/Keedfik/finance-bot
   cd finance-bot
   ```

2. Установите зависимости:

   ```bash
   go mod tidy
   ```

3. Создайте файл `.env` в корне проекта и добавьте следующие переменные:

   ```plaintext
   BOT_TOKEN=your_telegram_bot_token
   MONGO_URI=mongodb://localhost:27017
   DB_NAME=finance_bot
   ```

4. Запустите MongoDB (если она не запущена):

   ```bash
   net start MongoDB
   ```

5. Запустите бота:

   ```bash
   go run main.go
   ```

## Использование

### Команды

- `/start` - Запустить бота и получить приветственное сообщение.
- `/addexpense` - Добавить новый расход. Бот запросит у вас сумму, дату, заметку и категорию.
- `/getexpenses` - Получить список всех расходов.
- `/addcategory` - Добавить новую категорию расходов.
- `/setlimit` - Установить лимит для существующей категории расходов.
- `/getcategories` - Получить список всех существующих категорий расходов.
- `/deletelastexpense` - Удалить последний добавленный расход.
- `/help` - Получить справку с описанием доступных команд.

### Примеры использования

1. **Добавление расхода**

   Введите команду `/addexpense` и следуйте инструкциям бота:
   ```
   /addexpense
   ```
   Бот запросит сумму, дату, заметку и категорию. Например:
   ```
   Сумма: 100.50
   Дата: 2024-07-03
   Заметка: Обед
   Категория: Питание
   ```

2. **Просмотр всех расходов**

   Введите команду `/getexpenses`:
   ```
   /getexpenses
   ```

3. **Добавление новой категории**

   Введите команду `/addcategory` и следуйте инструкциям бота:
   ```
   /addcategory
   ```
   Бот запросит имя и лимит для новой категории. Например:
   ```
   Имя категории: Транспорт
   Лимит: 500
   ```

4. **Установка лимита для категории**

   Введите команду `/setlimit` и следуйте инструкциям бота:
   ```
   /setlimit
   ```
   Бот запросит имя существующей категории и новый лимит. Например:
   ```
   Категория: Транспорт
   Новый лимит: 600
   ```

## Структура проекта

Проект состоит из следующих пакетов:

- `bot`: Содержит основную логику взаимодействия с Telegram Bot API и обработку команд пользователя.
- `config`: Загружает конфигурацию приложения из файла `.env`.
- `db`: Обеспечивает подключение к MongoDB и определяет модели данных.

В корне проекта также находятся файлы `main.go` (точка входа приложения) и `go.mod` (файл зависимостей Go).

## Проверка функционала бота

1. Запустите бота командой `/start` если еще не сделали этого ранее
2. Добавьте расход: `/addexpense`, Категория: `Транспорт` (если категория не найдена - он добавит в Общую), Сумма: `100.50`, Заметка: `Обед` 
3. Просмотрите расходы: `/getexpenses`
4. Добавьте категорию: `/addcategory`, Название: `Транспорт`, Лимит: `500`
5. Просмотрите категории: `/getcategories`
6. Установите лимит категории: `/setlimit`, Категория: `Транспорт`, Новый лимит: `600`
7. Добавьте расход превышающий лимит: `/addexpense`, Категория: `Транспорт`, Сумма: `650`, Заметка: `Бензин` (Бот сообщит о превышении лимита)
8. Добавьте расход в рамках лимита: `/addexpense`, Категория: `Транспорт`, Сумма: `30`, Заметка: `Парковка`
9. Просмотрите расходы: `/getexpenses` 
10. Просмотрите категории: `/getcategories` (Убедитесь, что новый лимит "Транспорт" 600)
11. Удалите последний расход: `/deletelastexpense`, Подтвердите: `Да`
12. Просмотрите расходы: `/getexpenses` (Последний расход удален)
13. Получите справку: `/help`

Выполнив эту инструкцию, вы проверите все основные команды бота: добавление/просмотр расходов и категорий, установку лимитов, удаление расходов, а также работу механизма лимитов. Бот должен корректно отрабатывать каждый шаг.