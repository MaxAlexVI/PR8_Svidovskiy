# Практическое занятие 8. Работа с MongoDB: подключение, создание коллекции, CRUD-операции
# Свидовский М.А. ЭФМО-01-25

## Структура проекта
<img width="398" height="403" alt="image" src="https://github.com/user-attachments/assets/85659f9b-2092-4712-89e0-dc50b0e0dcf3" />

## Запуск
Запуск контейнера - docker compose up -d

Копирование файла с примером переменных окружения - cp .env.example .env   # PowerShell: Copy-Item .env.example .env

Запуск приложения - go run ./cmd/api 

### POST /api/v1/notes — создать 
<img width="748" height="575" alt="image" src="https://github.com/user-attachments/assets/cfaa63ba-2946-491d-b088-fe0dce76d93c" />

<img width="1080" height="556" alt="image" src="https://github.com/user-attachments/assets/f03533bf-ac2f-4552-ad55-ee899bb16216" />


### GET /api/v1/notes?q=&limit=&skip= — список + поиск по заголовку (ищем записи где в "title" есть "note")
<img width="803" height="832" alt="image" src="https://github.com/user-attachments/assets/a1cddfc1-1af3-46b0-9dce-1cc7fced3839" />

<img width="646" height="669" alt="image" src="https://github.com/user-attachments/assets/22c78d45-c934-4920-8746-9e6a1c13a379" />

<img width="625" height="530" alt="image" src="https://github.com/user-attachments/assets/2544ac96-8b2d-4563-b201-81ebef4ade47" />

### GET /api/v1/notes/{id} — получить по id
<img width="658" height="484" alt="image" src="https://github.com/user-attachments/assets/844aec0d-2471-458d-ac7c-2859f81295c5" />

<img width="640" height="499" alt="image" src="https://github.com/user-attachments/assets/bef544a5-b3e9-4615-9f3a-3570da8dec78" />



### PATCH /api/v1/notes/{id} — частично обновить

<img width="559" height="538" alt="image" src="https://github.com/user-attachments/assets/79aeac05-5346-4c07-a44d-4c864eb9946e" />

### DELETE /api/v1/notes/{id} — удалить (таск с начальником)

<img width="757" height="407" alt="image" src="https://github.com/user-attachments/assets/26a060ad-0c49-4be9-b82b-825e95612c28" />
<img width="747" height="463" alt="image" src="https://github.com/user-attachments/assets/abdd3e0e-907f-419e-acd3-28e8122ac7dd" />

## Доп. плюшки
### 1.	Текстовый поиск. Добавьте текстовый индекс и эндпоинт q переключите на $text:{$search:q}.

<img width="1614" height="757" alt="image" src="https://github.com/user-attachments/assets/648141e2-fe0c-4728-b0be-09c9a97937c9" />
<img width="979" height="460" alt="image" src="https://github.com/user-attachments/assets/23104754-2681-4ea8-9be1-e3f347810125" />


### 2. TTL-индекс. Поле expiresAt + индекс вида { expiresAt: 1 } с expireAfterSeconds: 0 для авто-удаления старых заметок.
 
<img width="1612" height="940" alt="image" src="https://github.com/user-attachments/assets/7de4002a-8e62-42f6-9759-8521fd41c10c" />
<img width="1615" height="993" alt="image" src="https://github.com/user-attachments/assets/30c5becf-5677-45db-86d2-45b1839d272e" />
<img width="751" height="475" alt="image" src="https://github.com/user-attachments/assets/03d423a0-ae8f-4cc6-a21b-f4d9e00db65d" />

<img width="805" height="760" alt="image" src="https://github.com/user-attachments/assets/f692587a-f12d-43b5-8c63-1ea7d1e95477" />

### 3. Aggregation pipeline. Верните статистику: количество заметок, средняя длина content.

<img width="972" height="600" alt="image" src="https://github.com/user-attachments/assets/1fda09f3-f6a1-432a-b257-47ccc389b40e" />

<img width="1434" height="604" alt="image" src="https://github.com/user-attachments/assets/8250a441-d963-44bc-b16e-df019fd71646" />

<img width="842" height="578" alt="image" src="https://github.com/user-attachments/assets/4efed497-6146-4973-8280-8551a28727a0" />




