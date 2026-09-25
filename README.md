# isposch

Сервис скачивает XML расписания с `polytech-shedule.ru`, оставляет занятия конкретной группы и отдает календарь в формате iCalendar.

Расписание доступно по адресу:

```text
/calendar.ics
```


## Запуск

```bash
go run .
```

Переменные окружения:

- `SCHEDULE_GROUP`: группа 
- `HTTP_ADDR`: адрес сервера
- `PORT`: порт сервера. Если переменная задана, она используется вместо `HTTP_ADDR`.

После запуска локально календарь доступен по адресу:

`http://localhost:8080/calendar.ics`

## Хост на Render

Создайте **Web Service** из GitHub-репозитория и укажите:

- **Build Command:** `go build -o app .`
- **Start Command:** `./app`

В переменной `SCHEDULE_GROUP` укажите нужную группу.

Переменную `PORT` задавать не нужно — Render передаст ее автоматически. После деплоя календарь будет доступен по адресу:

`https://имясервиса.onrender.com/calendar.ics`.
