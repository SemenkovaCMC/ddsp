# ТРХиОД/ВПиРОД (ВМК МГУ, Осень 2018)
Репозиторий с заданием по курсам:
* Теория распределенного хранения и обработки данных
* Высокопроизводительная и распределенная обработка данных

Вопросы задавать в [Slack](https://trhiod.slack.com/).

## Задание ##
Целью практической части нашего курса является создание распределённого key-value хранилища (далее просто *хранилище*).

Replication Factor: 3, quorum: 2, язык разработки: [go](https://golang.org), для взаимодействия между компонентами используется [gRPC](https://grpc.io/).

*Хранилище* состоит из трех основных компонент:
1. Node -- узлы, на которых хранятся данные
2. Frontend -- точка входа в хранилище
3. Router -- узел, в котором хранится схема *хранилища*

Предполагается, что *хранилище* будет запускаться в конфигурации,
содержащей не менее трёх Node, один Router, и один или несколько Frontend.

В качестве *ключа* используется uint32, в качестве *значения* -- произовольный набор байт.

**Основной документацией** к *хранилищу* являются тесты,
содержащиеся в репозитории и godoc ([node](https://godoc.org/github.com/n-canter/ddsp/src/node/node), [router](https://godoc.org/github.com/n-canter/ddsp/src/router/router), [frontend](https://godoc.org/github.com/n-canter/ddsp/src/frontend/frontend)).

*Хранилище* позволяет выполнить следующие операции:
1. Put(key, value) -- добавление элемента
2. Delete(key) -- удаление элемента
3. Get(key) -- получение элемента

Рассмотрим подробнее эти операции.
### Put (key, value) ###
Frontend выполняет следующие действия: 
1. Отправляет запрос NodesFind(key, nodes) в Router (nodes -- список доступных Nodes).
2. Параллельно кладет значение value по ключу key в каждый из node, которые вернул запрос NodesFor(key, nodes).

Если NodesFor(key) возвращает ошибку, то ошибка немедленно возвращается клиенту.
Если NodesFor(key) прошел успешно, то результат возвращается клиенту после того, как завершатся все запросы из шага 2.

### Delete(key) ###
Delete работает по аналогии с Put.

Frontend выполняет следующие действия: 
1. Отправляет запрос NodesFind(key, nodes) в Router (nodes -- список доступных Nodes).
2. Параллельно отправляет Delete(key) в каждый из node, которые вернул запрос NodesFor(key, nodes).

Если NodesFor(key) возвращает ошибку, то ошибка немедленно возвращается клиенту.
Если NodesFor(key) прошел успешно, то результат возвращается клиенту после того, как завершатся все запросы из шага 2.

### Get(key) ###
При первом Get запросе Frontend отправляет запрос List() в Router для получения списка обслуживаемых Nodes.
Если List() запрос возвращает ошибку, то запрос должен быть повторен через специальный таймаут.
До момента успешного завершения List() запроса в Router ни один Get() не может быть выполнен.

Далее предполагается, что List() запрос выполнен, и во Frontend уже сохранен список обслуживаемых Nodes.

Frontend выполняет следующие действия для выполнения запроса Get(key):
1. Вызывает метод NodesFind(key, nodes) -- локальный запрос без похода в Router (nodes -- список обслуживаемых Nodes).
2. Параллельно отправляет запрос Get(key) в каждый из Node, которые вернул вызов метода NodesFor(key).

Результат возвращается клиенту, как только два из запросов в Node завершатся с одинаковым результатом,
либо после того, как все три запроса завершатся с разным результатом.

### NodesFind(key, nodes) ###
Для вычисления узлов, на которых должны хранится данные по данному ключу key используется разновидность метода
[HRW](https://en.wikipedia.org/wiki/Rendezvous_hashing).

Для каждой пары key, node (node -- элемент nodes) вычисляется хэш.

После этого nodes сортируется в соответствии со значением хэш (в убывающем порядке).
Если для различных пар key,node значение хэшей совпало,
то данные пары сортируются в соответствии с лексикографическим порядком node (в убывающем порядке).

### Heartbeats ###
Каждый Node отправляет специальные запросы heartbeats к Router через ранвые интервалы времени.
Если какой-либо Node не отправлял запросы в течении значения, задаваемого в конфигурации Router,
то днный Node считается недоступным.


## Сдача заданий и система оценивания ##
Необходимо реализовать методы помеченные комментарием `TODO: implement`.

**Запрещается** удалять/изменять существующий код, кроме методов, помеченных комментарием `TODO: implement`.

**Разрешается** добавление кода в существующие файлы и добавление новых файлов в репозиторий.

Задание состоит из трех частей.

Часть считается сданной тогда и только тогда, когда она проходит **все** относящиеся к ней тесты.

Список частей (в скобочках команда для запуска тестов):
1. Node (`make test-node`)
2. Frontend (`make test-fe`)
3. Router + Integration Tests (`make test-router && make test-integration`)

Количество частей, сданных студентом влияет на финальную оценку за курс.

Если сдана первая часть, то максимальная оценка, на которую претендует студент -- **уд**.

Если сданы первые две части, то максимальная оценка, на которую претендует студент -- **хор**.

Если сданы все части, то студент претендует на оценку **отл**.

Сдача первой части является необходимым и достаточным условием **допуска к экзамену**.

Для того чтобы сдать задание, нужно отправить Pull Request.

Первичное тестирование производится **автоматически** с использованием системы [Travic CI](https://travis-ci.org/).

Описание Pull Request **должно содержать фамилию, имя, номер группы** студента, сдающего задание.

После проверки преподавателем Pull Request будет присвоена одна из меток (label):
* ok - все хорошо
* refactor - необходим рефакторинг
* error - есть ошибки
