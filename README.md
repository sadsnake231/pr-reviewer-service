# Запуск

`docker compose up -d`

Сервис доступен на порте `8080`

Также есть makefile (выполните make help)

# Стек
`Go`, БД `PostgreSQL`

Библиотека для миграций `goose`

Web-фреймворк `gin`

Драйвер для БД `pgx`

# Выполненные доп. задания

Эндпоинт статистики (доступен на `/stats`)

Нагрузочное тестирование (результаты в файле `vegeta-report.txt` и чуть ниже, сам тест в `/cmd/loadtest`)

Метод массовой деактивации (доступен на `/users/deactivateTeam`, метод `POST`. Пример: 

`curl --location 'localhost:8080/users/deactivateTeam' \
--header 'Content-Type: application/json' \
--header 'Accept: application/json' \
--data '{
  "team_name": "backend"
}'`). У меня сработало за 17 мс на `JSON` с 200 users, лежащем в корне)

Безопасная переназначаемость PR через транзакции

Unit-тестирование (в папке `service`) (покрытие 73%)

Конфигурация линтера

## Возникшие вопросы

Я не понял, что нужно делать, если в создаваемой команде есть неуникальные `user_id`. Решил, что будет логично это запретить, добавил для удобства новый Error Code - `NONUNIQUE_USER`.

## UML-схема приложения

<img width="2155" height="1702" alt="изображение" src="https://github.com/user-attachments/assets/2a979f5e-ed27-406f-8f68-889b12cbc1e5" />

## Схема БД

<img width="942" height="407" alt="изображение" src="https://github.com/user-attachments/assets/01a8d4ba-10af-4e23-b782-96224d00dcb3" />

## Информация по БД

Индексы созданы для полей `team_name` и `is_active` в таблице `users`, для `status` и `author_id` в таблице `pull_requests`, а также для полей `pull_request_id` и `user_id` в таблице pr_reviewers.

БД соответствует третьей нормальной форме

## Нагрузочный тест

`RPS` = 5, 20 команд по 200 пользователей в каждой. Тесты распределены неравномерно - 40% - создание pullrequest, 30% - получение информации о команде, 20% - получение review, назначенных пользователю, 10% - получение общей статистики.

Результаты на ноутбучном i7-1165G7:

```
Requests      [total, rate, throughput]         150, 5.03, 5.03
Duration      [total, attack, wait]             29.801s, 29.8s, 1.362ms
Latencies     [min, mean, 50, 90, 95, 99, max]  425.712µs, 3.992ms, 1.296ms, 7.028ms, 15.354ms, 15.643ms, 15.977ms
Bytes In      [total, mean]                     1006716, 6711.44
Bytes Out     [total, mean]                     5642, 37.61
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:88  201:62  
Error Set:
```




