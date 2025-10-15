# Noted-auth

Сервис для авторизации в Noted (https://dnk33.com) с помощью keycloak

ЗАПУСК С ЛОГАМИ УРОВНЯ DEBUG В PRODUCTION СТРОГО ЗАПРЕЩЁН, ТАК КАК В ДЕБАГ ПИШУТСЯ ЧУВСТВИТЕЛЬНЫЕ ДАННЫЕ (ТОКЕНЫ ПОЛЬЗОВАТЕЛЕЙ)

Known issues:
- Нет тестов
- Не настроен CI/CD + codecov
- Не настроен viper + healthcheck для vault + yml path