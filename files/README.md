## Описание
Программа должна читать файл, вычитывать все слова, считать наиболее частые и писать первые N пар слово: использований в выходной файл.

## Конфигурация
Для того чтобы сконфигурировать команду нужно использовать флаги:

```
-in: путь к входному файлу
-out: путь к выходному файлу
-limit: сколько пар записать в файл (если не установлено: 10)
-min-length: минимальная длина для слова (если не установлено: 5)
```
Сортировать слова по популярности использования, если оно равно, то лексиграфически (стандартным оператором < для строк)

## Полезные материалы

Как использовать флаги в Go (https://gobyexample.com/command-line-flags, не забудьте использовать `flag.Parse()`)

Как открыть файл для чтения или записи (https://metanit.com/go/tutorial/8.2.php, не забудьте после обработки ошибки написать `defer file.Close()` - что это обсудим позже)

Как вычитать все из `io.Reader` ([io.ReadAll](https://pkg.go.dev/io#ReadAll))

Как обрабатывать ошибки [log.Fatal](https://pkg.go.dev/log#Fatal)

Как сортировать мапу: нужно переложить все пары в слайс любого типа который вам нужен, например `type result { word string, count int }` и сортировать его

Как удобно получить слова из текста: [strings.Fields()](https://www.geeksforgeeks.org/strings-fields-function-in-golang-with-examples/)
Разбить текст на слова - считать, что слова состоят только из букв, а все остальное - разделитель
```go
strings.FieldsFunc(string(content), func(r rune) bool {
	return !unicode.IsLetter(r)
})
```

## Тест

Как понять что все работает правильно: выполнить `make run` в директории files и получить следующий ответ

```
reasoning: 11
science: 11
scientific: 10
based: 6
inductive: 6
method: 6
problem: 6
thinking: 6
deductive: 5
logical: 5
observation: 5
specific: 5
brain: 4
descriptive: 4
general: 4
```