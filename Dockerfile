FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ENV TODO_PORT=7540 \
    TODO_DBFILE=./scheduler.db \
    # положил сюда токен для пароля 12345 (eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.396KCDWMomWrMEImsF84AmFRjBEvSvnyLh3ZA_mB_Wg), тесты не прошли, через браузер зайти тоже не смог
    TODO_PASSWORD=

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /my_web_server

CMD [ "/my_web_server"]