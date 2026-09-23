# isposch

Сервис скачивает XML расписания с `polytech-shedule.ru`, оставляет занятия конкретной группы и отдает календарь iCalendar.

## Запуск

```bash
go run .
```

Переменные окружения:

- `SCHEDULE_GROUP` — группа (по умолчанию `25290901/3091`);
- `SCHEDULE_URL` — URL XML;
- `HTTP_ADDR` — адрес HTTP-сервера (по умолчанию `:8080`).

Календарь доступен по `http://localhost:8080/calendar.ics`.

