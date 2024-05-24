Web-server/application to-do-list.

Позволяет добавлять, просматривать, править и удалять задачи, а также устанавливать правила их периодической цикличности.

Со звездочкой пытался выполнить все, правда с переменным успехом - тесты пройдены все, кроме финального - с аутентификацией.
Видимо не до конца понял задание, в связи с чем и опростоволосился.

В .gitignore пихнул только .env, но он указан в Dockerfile (что-то запутался с тем, как иначе указать переменные окружения. А если это единственный путь, то почему тогда .env идет в .gitignore?)

Допфлагов нет, порт по умолчанию - 7540.

settings.go:
package tests

var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = true
var Search = true
var Token = ``

Собирал докер командой docker build --tag my-web-server:v4 .;
запускал docker run -it -p 7540:7540 --name web-server-container my-web-server:v4

