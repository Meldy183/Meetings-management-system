---
name: "meetings-console"
description: "Управление системой протоколов совещаний: создание совещаний, управление участниками, повесткой, председателем и экспорт документов .docx"
version: 2.0.0
metadata:
  openclaw:
    emoji: "📋"
    always: false
    requires:
      bins: ["meetings-console"]
      env: []
---

# Описание

Ты управляешь системой протоколов совещаний. Взаимодействуй с бэкендом **ТОЛЬКО** через консольную утилиту `meetings-console`. Не используй curl, wget, HTTP-клиенты или любые другие инструменты — только `meetings-console`.

## Сборка утилиты (если ещё не собрана)

```bash
cd /путь/к/проекту/console && go build -o meetings-console .
```

После сборки убедись, что бинарник доступен в PATH или используй полный путь.

## Переменные окружения (необязательны — есть умолчания)

| Переменная | Умолчание | Описание |
|---|---|---|
| `MEETING_API_BASE_URL` | `http://localhost:8081/api` | Адрес бэкенда |
| `MEETING_API_TOKEN` | `admin` | API-ключ авторизации |

## Формат вызова

```
meetings-console <КОМАНДА> '<JSON_АРГУМЕНТЫ>'
```

**Правила:**
- JSON всегда в одинарных кавычках
- Если в значении есть одинарные кавычки — экранируй: `'\''`
- Ответ всегда начинается с `HTTP <код>`, затем тело в JSON
- HTTP 200/201 — успех. HTTP 4xx/5xx — ошибка, тело содержит `{"message":"..."}`.
- Если получил ошибку валидации или 4xx — разбери сообщение, исправь JSON и вызови снова.
- **Никогда не угадывай UUID или ID** — всегда читай их из ответов предыдущих команд.

---

## Команды

### Система

#### `health`
Проверка доступности бэкенда.

```bash
meetings-console health '{}'
```

---

### Участники (People)

#### `list_people`
Список всех участников или поиск по имени. Возвращает до 100 результатов.

```bash
meetings-console list_people '{}'
meetings-console list_people '{"q":"Иванов"}'
meetings-console list_people '{"q":"Иван Петров"}'
```

Возвращает: массив объектов Person.

```json
[
  {
    "id": 5,
    "last_name": "Иванов",
    "first_name": "Иван",
    "middle_name": "Петрович",
    "info": "Главный аналитик"
  }
]
```

`middle_name` и `info` отсутствуют в JSON если не заданы.

---

#### `get_person`
Получить одного участника по ID.

```bash
meetings-console get_person '{"id":5}'
```

Ошибка 404 — участник не найден.

---

#### `create_person`
Создать нового участника. `last_name` и `first_name` обязательны.

```bash
meetings-console create_person '{"last_name":"Иванов","first_name":"Иван"}'
meetings-console create_person '{"last_name":"Иванов","first_name":"Иван","middle_name":"Петрович","info":"Директор"}'
```

Возвращает: созданный объект Person с присвоенным `id`.
Ошибка 409 — участник с таким именем уже существует.

---

#### `update_person`
Обновить данные участника. `last_name` и `first_name` обязательны — всегда передавай оба.

```bash
meetings-console update_person '{"id":5,"last_name":"Иванов","first_name":"Иван","middle_name":"Петрович","info":"Заместитель директора"}'
```

Возвращает: обновлённый объект Person.
Ошибка 404 — участник не найден.
Ошибка 409 — новое имя уже занято другим участником.

---

#### `sort_people`
Вернуть переданные ID участников, отсортированные по фамилии, имени, отчеству.

```bash
meetings-console sort_people '{"ids":[17,5,42]}'
```

Возвращает: `{"ids":[42,5,17]}` — те же ID в алфавитном порядке.

---

### Совещания (Meetings)

#### `list_meetings`
Список совещаний с пагинацией. `limit` по умолчанию 20, `offset` по умолчанию 0.
Фильтр по статусу: `"complete"` — только завершённые, `"incomplete"` — только черновики, `""` — все.

```bash
meetings-console list_meetings '{}'
meetings-console list_meetings '{"limit":20,"offset":0}'
meetings-console list_meetings '{"limit":20,"offset":0,"status":"complete"}'
meetings-console list_meetings '{"status":"incomplete"}'
```

Возвращает:
```json
{
  "total": 42,
  "limit": 20,
  "offset": 0,
  "items": [ /* массив MeetingSummary */ ]
}
```

---

#### `create_meeting`
Создать новое совещание. `title` и `date` обязательны. Дата в формате ISO 8601.
Создаётся со статусом `incomplete` (черновик).

```bash
meetings-console create_meeting '{"title":"Совещание по вопросам бюджета","date":"2026-04-01T10:00:00Z"}'
meetings-console create_meeting '{"title":"Заседание комиссии","date":"2026-04-15T14:00:00Z","place":"г. Москва, ул. Тверская, д. 13"}'
```

Возвращает: полный объект Meeting. Поле `id` — UUID, используй его во всех последующих командах.

---

#### `get_meeting`
Получить полные данные совещания: участники, председатель, повестка с докладчиками.

```bash
meetings-console get_meeting '{"id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}'
```

Возвращает:
```json
{
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "title": "Совещание по бюджету",
  "date": "2026-04-01T10:00:00Z",
  "place": "г. Москва",
  "status": "complete",
  "chairperson": {"id": 5, "last_name": "Иванов", "first_name": "Иван"},
  "people": [ /* массив Person */ ],
  "agenda_items": [ /* массив AgendaItem */ ],
  "created_at": "2026-03-20T08:00:00Z"
}
```

---

#### `get_meeting_meta`
Получить только скалярные поля совещания (без участников и повестки). Используй для проверки статуса.

```bash
meetings-console get_meeting_meta '{"id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}'
```

---

#### `update_meeting`
Изменить тему, дату и место совещания. `title` и `date` обязательны — передавай оба всегда.

```bash
meetings-console update_meeting '{"id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","title":"Новая тема","date":"2026-04-02T10:00:00Z"}'
meetings-console update_meeting '{"id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","title":"Новая тема","date":"2026-04-02T10:00:00Z","place":"Зал заседаний №2"}'
```

Ошибка 404 — совещание не найдено.

---

### Участники совещания

#### `list_meeting_people`
Получить список участников конкретного совещания в текущем порядке.

```bash
meetings-console list_meeting_people '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}'
```

---

#### `add_meeting_person`
Добавить участника в совещание. Участник должен существовать в базе данных.

```bash
meetings-console add_meeting_person '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","person_id":5}'
```

Ошибка 409 — участник уже добавлен в совещание.
Ошибка 422 — участник с таким ID не существует в базе.

---

#### `remove_meeting_person`
Удалить участника из совещания.

```bash
meetings-console remove_meeting_person '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","person_id":5}'
```

Ошибка 409 — участник является председателем или докладчиком. Сначала убери эти роли.

---

#### `order_meeting_people`
Установить порядок отображения участников. Передай **все** текущие ID участников — только в нужном порядке.

```bash
meetings-console order_meeting_people '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","person_ids":[5,7,3]}'
```

Ошибка 422 — переданные ID не совпадают точно с текущим составом участников.

---

### Председатель

#### `set_chairperson`
Назначить или заменить председателя. Участник **уже должен быть в списке участников совещания**.

```bash
meetings-console set_chairperson '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","person_id":5}'
```

Ошибка 422 — участник не в списке участников совещания. Сначала вызови `add_meeting_person`.

---

### Повестка

#### `list_agenda_items`
Получить список пунктов повестки с докладчиками.

```bash
meetings-console list_agenda_items '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}'
```

---

#### `add_agenda_item`
Добавить пункт повестки. Текст и хотя бы один докладчик обязательны.
Все докладчики должны быть в списке участников совещания.

```bash
meetings-console add_agenda_item '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","text":"Утверждение бюджета","speaker_ids":[5]}'
meetings-console add_agenda_item '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","text":"Кадровые вопросы","speaker_ids":[5,7,3]}'
```

Ошибка 422 — кто-то из докладчиков не в списке участников. Сначала вызови `add_meeting_person`.

---

#### `update_agenda_item`
Полностью заменить текст и список докладчиков пункта повестки. Передай всех нужных докладчиков.

```bash
meetings-console update_agenda_item '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","item_id":3,"text":"Уточнённый бюджет","speaker_ids":[5,7]}'
```

Ошибка 422 — докладчик не в участниках совещания.

---

#### `delete_agenda_item`
Удалить пункт повестки.

```bash
meetings-console delete_agenda_item '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","item_id":3}'
```

---

#### `order_agenda_items`
Установить порядок пунктов повестки. Передай **все** текущие ID пунктов.

```bash
meetings-console order_agenda_items '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","agenda_item_ids":[3,1,2]}'
```

Ошибка 422 — ID не совпадают точно с текущим набором пунктов.

---

### Докладчики пунктов повестки

#### `add_speaker`
Добавить докладчика к пункту повестки. Участник должен быть в списке участников совещания.

```bash
meetings-console add_speaker '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","item_id":3,"person_id":7}'
```

Ошибка 409 — участник уже является докладчиком этого пункта.
Ошибка 422 — участник не в списке участников совещания.

---

#### `remove_speaker`
Убрать докладчика из пункта повестки.

```bash
meetings-console remove_speaker '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","item_id":3,"person_id":7}'
```

Ошибка 409 — нельзя убрать последнего докладчика. Сначала добавь другого.

---

#### `order_speakers`
Установить порядок докладчиков пункта повестки. Передай **все** текущие ID докладчиков этого пункта.

```bash
meetings-console order_speakers '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6","item_id":3,"person_ids":[7,5]}'
```

---

### Экспорт документов

Экспорт доступен только для совещаний со статусом `complete`. При `incomplete` — ошибка 409.

#### `export_agenda`
Скачать документ «Повестка» в формате .docx. Файл сохраняется на диск.

```bash
meetings-console export_agenda '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}'
```

Файл сохраняется как `agenda-<meeting_id>.docx` в текущей директории.

---

#### `export_participants`
Скачать документ «Список участников» в формате .docx.

```bash
meetings-console export_participants '{"meeting_id":"3fa85f64-5717-4562-b3fc-2c963f66afa6"}'
```

Файл сохраняется как `participants-<meeting_id>.docx` в текущей директории.

---

## Статус совещания

Статус вычисляется автоматически при каждом запросе — не хранится.

| Статус | Условие |
|---|---|
| `incomplete` | нет председателя, ИЛИ нет участников, ИЛИ нет пунктов повестки |
| `complete` | есть председатель И ≥1 участник И ≥1 пункт повестки (каждый с ≥1 докладчиком) |

---

## Стандартный порядок создания полного совещания

```bash
# 1. Найти или создать участников
meetings-console list_people '{"q":"Иванов"}'
meetings-console create_person '{"last_name":"Иванов","first_name":"Иван","middle_name":"Петрович","info":"Директор"}'
meetings-console create_person '{"last_name":"Петров","first_name":"Сергей","info":"Аналитик"}'
# → запомни id каждого участника (например 5 и 7)

# 2. Создать совещание
meetings-console create_meeting '{"title":"Совещание по итогам квартала","date":"2026-04-01T10:00:00Z"}'
# → запомни id совещания (UUID)

# 3. Добавить участников в совещание
meetings-console add_meeting_person '{"meeting_id":"<uuid>","person_id":5}'
meetings-console add_meeting_person '{"meeting_id":"<uuid>","person_id":7}'

# 4. Назначить председателя (должен уже быть в участниках)
meetings-console set_chairperson '{"meeting_id":"<uuid>","person_id":5}'

# 5. Добавить пункты повестки
meetings-console add_agenda_item '{"meeting_id":"<uuid>","text":"Вступительное слово","speaker_ids":[5]}'
meetings-console add_agenda_item '{"meeting_id":"<uuid>","text":"Итоги квартала","speaker_ids":[5,7]}'
meetings-console add_agenda_item '{"meeting_id":"<uuid>","text":"Планы на следующий квартал","speaker_ids":[7]}'

# 6. Проверить статус
meetings-console get_meeting_meta '{"id":"<uuid>"}'
# → "status": "complete"

# 7. Экспортировать документы
meetings-console export_agenda '{"meeting_id":"<uuid>"}'
meetings-console export_participants '{"meeting_id":"<uuid>"}'
```

---

## Справочник ошибок

| Код | Команда | Причина | Решение |
|---|---|---|---|
| 404 | любая | Ресурс не найден | Проверь ID через list/get команду |
| 409 | `create_person` | Имя уже занято | Найди через `list_people` |
| 409 | `remove_meeting_person` | Участник — председатель или докладчик | Убери роли перед удалением |
| 409 | `remove_speaker` | Последний докладчик пункта | Добавь другого докладчика сначала |
| 409 | `add_meeting_person` | Участник уже в совещании | Не нужно добавлять повторно |
| 409 | `add_speaker` | Уже является докладчиком | Не нужно добавлять повторно |
| 409 | `export_*` | Совещание в статусе incomplete | Проверь `get_meeting_meta`, заполни недостающее |
| 422 | `set_chairperson` | Участник не в списке совещания | Сначала `add_meeting_person` |
| 422 | `add_meeting_person` | ID не существует в базе | Проверь через `list_people` или `get_person` |
| 422 | `add_agenda_item` / `update_agenda_item` | Докладчик не в участниках | Сначала `add_meeting_person` для каждого |
| 422 | `add_speaker` | Участник не в списке совещания | Сначала `add_meeting_person` |
| 422 | `order_*` | Переданные ID не совпадают с текущим набором | Получи актуальный список, передай все ID |
