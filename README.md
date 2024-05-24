Web-server/application to-do-list.

Позволяет добавлять, просматривать, править и удалять задачи, а также устанавливать правила их периодической цикличности.

Задания со звездочкой выполнены все, кроме финального - с аутентификацией.

Запускать сервер НЕ через докер.

Допфлагов нет, переменные окружения:
TODO_PORT=:7540
TODO_DBFILE=./scheduler.db

Все тесты прошли по общей команде go test ./tests


settings.go:
package tests

var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = true
var Search = true
var Token = ``

// Докер, из-за которого потенциально всплыли остальные проблемы собирался и запускался следующими командами:

docker build --tag my-web-server:v4 .;
docker run -it -p 7540:7540 --name web-server-container my-web-server:v4

Прошу проверить корректность кода без докера, позже вернусь к нему.