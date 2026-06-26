# Boxmind iOS

Нативное iOS-приложение Boxmind. Работает с тем же backend API, что web и Chrome extension.

## Что уже есть (MVP v0.1)

- Вход по email + OTP код
- Список закладок
- Добавление ссылки вручную
- Удаление свайпом
- Открытие закладки по тапу
- Pull-to-refresh
- Автообновление, пока карточка дорабатывается
- Share Extension: сохранение ссылки через системное «Поделиться»
- Токен хранится в Keychain (безопаснее, чем UserDefaults)

## Структура проекта

```text
ios/
  project.yml              # описание Xcode-проекта (XcodeGen)
  Boxmind/
    App/                   # точка входа, RootView
    Core/                  # API, модели, Keychain
    Features/              # Auth, Bookmarks (MVVM)
    Shared/                # общие хелперы
    Resources/             # Info.plist, Assets
  BoxmindShare/            # Share Extension
```

Архитектура: **SwiftUI + MVVM + async/await**.

## Что тебе нужно на Mac

1. **Xcode** из App Store (бесплатно, ~12+ GB)
2. **XcodeGen** — генерирует `.xcodeproj` из `project.yml`:

```bash
brew install xcodegen
```

3. Apple ID (бесплатный) — чтобы запустить на своём iPhone/simulator

## Как открыть и запустить

```bash
cd ios
xcodegen generate
open Boxmind.xcodeproj
```

В Xcode:

1. Подожди, пока проект проиндексируется
2. Выбери симулятор iPhone (например iPhone 16)
3. Нажми ▶ Run (Cmd+R)

Приложение по умолчанию ходит в production API:

`https://api.boxmind.link/api/v1`

## Как проверить Share Extension

1. Запусти основное приложение Boxmind и войди в аккаунт
2. Открой Safari на симуляторе или iPhone
3. Открой любой сайт
4. Нажми **Share** → выбери **Boxmind**
5. Ссылка сохранится **сразу**, без окна подтверждения — share sheet просто закроется
6. Открой Boxmind — новая карточка уже в списке (или pull-to-refresh)

Extension читает JWT из App Group и сам делает `POST /bookmarks`. Основное приложение открывать не нужно.

### Локальный backend (опционально)

В Xcode: **Product → Scheme → Edit Scheme → Run → Arguments → Environment Variables**

```
BOXMIND_API_URL = http://127.0.0.1:8080/api/v1
```

## Как работает login

1. Вводишь email → `POST /auth/login`
2. Получаешь код на почту
3. Вводишь код → `POST /auth/verify`
4. JWT сохраняется в Keychain
5. Открывается список закладок

## Следующие шаги (не в этом MVP)

- Share Extension (сохранение из Safari/YouTube через «Поделиться»)
- Секции как на web (Смотреть / Слушать / Игры…)
- Открытие ссылки в Safari
- Push-уведомления
- App Store release ($99/год Apple Developer)

## Полезно знать новичку

- **SwiftUI** — декларативный UI (как React, но для iOS)
- **ViewModel** — логика экрана, View — только отображение
- **Keychain** — системное защищённое хранилище для токенов
- **Simulator** — виртуальный iPhone на Mac, для первых тестов достаточно
- **Share Extension** — отдельный mini-target, появится на следующем этапе
