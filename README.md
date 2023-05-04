## Порядок выполнения

* Одна задача - один Pull Request. Название ветки и Pull Request должно быть в формате `taskName-userName`, например `files-unspectd`.
* Перед тем как выполнять задачу - внимательно прочитайте README.me в директории задачи.
* Перед тем как отправлять Pull Request
  * пройдите секцию `Тест` в README.md и убедитесь что все работает правильно
  * отформатируйте ваш код с помощью `gofumpt`

## Рекомендуемый порядок задач
* [files](files) - работа с файлами
* [stdio](stdio) - работа со стандартными потоками ввода/вывода
* [filehashes](filehashes) - работа с файловой системой, использование интерфейсов для тестирования
* [wordscount](wordscount) - работа с многопоточной обработкой данных
* [watchcmd](watchcmd) - работа с сигналами, использование time.Ticker и context.Context
* [tcp-guessgame](tcp-guessgame) - работа с TCP
* [http-booklibrary](http-booklibrary) - работа с пакетом `net/http`
* [http-userroles](http-userroles) - работа с Unix Domain Socket, `boltdb/bolt`, `labstack/echo`, `spf13/cobra`, `go-resty/resty`
* [sql-querier](sql-querier) - работа с `database/sql`, написание простых sql запросов
* [http-userroles: follow-up 1](http-userroles) - работа со `swaggo`. Генерация OpenAPI спецификации