# Web-server/Application to-do-list 	

## Описание
Данный веб-сервер позволяет хранить задачи с указанием даты дедлайна и заголовка с комментарием. Задачи могут повторяться по заданному правилу, например, ежегодно, через определённое количество дней, в заданные дни месяца или недели. Если отметить задачу как выполненную, она будет перенесена на следующую дату согласно установленному правилу. Обычные задачи удаляются после выполнения.

## Возможности API
API проекта предоставляет следующие операции:

- [x] **Добавить задачу**
- [x] **Получить список задач**
- [x] **Удалить задачу**
- [x] **Получить параметры задачи**
- [x] **Изменить параметры задачи**
- [x] **Отметить задачу как выполненную**

## Дополнительные функции
- [x] _Возможность задавать порт и путь к файлу базы данных извне при запуске сервера_
- [x] _Поиск задач по ключевому слову и дате_
- [x] _Сборка и запуск проекта через Docker_
