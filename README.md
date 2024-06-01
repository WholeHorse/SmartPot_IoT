## Содержание

* [Содержание](#содержание)
* [О проекте](#о-проекте)
    * [Описание](#описание)
    * [Технологии](#технологии)
* [Использование](#использование)
    * [Запуск приложения](#запуск-приложения)
    * [Запросы Postman](#запросы-postman)
* [Авторы](#авторы)

## О проекте

### Описание

SmartPot — это система для управления умными горшками, включающая в себя работу с датчиками и устройствами, объединенными под одним горшком. Пользователь может добавлять, изменять и удалять горшки, датчики и устройства, а также получать актуальные данные с датчиков и управлять устройствами.

Проект выполнен в рамках зачетной работы по дисциплине "Технологии Интернета вещей".

### Технологии

- **Язык программирования**: Go
- **Веб-фреймворк**: Gin
- **База данных**: PostgreSQL
- **Средства разработки и тестирования API**: Postman

## Использование

### Запуск приложения

Для запуска приложения выполните следующие шаги:

1. Установите необходимые зависимости.
    ```shell
    go get -u github.com/gin-gonic/gin
    go get -u github.com/lib/pq
    ```

2. Убедитесь, что у вас установлен PostgreSQL и запущен сервер базы данных. Создайте базу данных и выполните необходимые SQL-запросы для создания таблиц.
    ```sql
    CREATE TABLE pots (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );

    CREATE TABLE sensors (
        id VARCHAR(50) PRIMARY KEY,
        type VARCHAR(50) NOT NULL,
        value FLOAT NOT NULL,
        status VARCHAR(50) NOT NULL,
        pot_id INT REFERENCES pots(id)
    );

    CREATE TABLE devices (
        id VARCHAR(50) PRIMARY KEY,
        type VARCHAR(50) NOT NULL,
        status VARCHAR(50) NOT NULL,
        pot_id INT REFERENCES pots(id)
    );

    CREATE TABLE logs (
        id SERIAL PRIMARY KEY,
        sensor_id VARCHAR(50),
        device_id VARCHAR(50),
        action VARCHAR(255) NOT NULL,
        timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    ```

3. Заполните переменные окружения в файле `.env`:
    ```
    DB_USER=iot_admin
    DB_PASSWORD=iotpass
    DB_NAME=iot-db
    DB_HOST=localhost
    DB_PORT=5432
    ```

4. Запустите приложение:
    ```shell
    go run main.go
    ```

Приложение будет доступно по адресу http://127.0.0.1:8080/.

### Запросы Postman

Примеры запросов для использования в Postman:

1. **Добавить горшок**
    ```
    POST /pots/add
    Content-Type: application/json
    Body: 
    {
        "name": "MyPot"
    }
    ```

2. **Удалить горшок**
    ```
    DELETE /pots/delete/:id
    ```

3. **Получить все горшки**
    ```
    GET /pots
    ```

4. **Добавить датчик**
    ```
    POST /sensors/add
    Content-Type: application/json
    Body: 
    {
        "id": "sensor1",
        "type": "humidity",
        "value": 0.0,
        "status": "active",
        "pot_id": 1
    }
    ```

5. **Удалить датчик**
    ```
    DELETE /sensors/delete/:id
    ```

6. **Получить все датчики**
    ```
    GET /sensors
    ```

7. **Добавить устройство**
    ```
    POST /devices/add
    Content-Type: application/json
    Body: 
    {
        "id": "device1",
        "type": "watering_system",
        "status": "inactive",
        "pot_id": 1
    }
    ```

8. **Удалить устройство**
    ```
    DELETE /devices/delete/:id
    ```

9. **Обновить статус устройства**
    ```
    PUT /devices/update/:id
    Content-Type: application/json
    Body: 
    {
        "status": "active"
    }
    ```

10. **Получить все устройства**
    ```
    GET /devices
    ```

11. **Очистить логи**
    ```
    DELETE /logs/clear
    ```

## Авторы

- Бурыкин Михаил
- Дурыничев Дмитрий
- Громов Алексей
