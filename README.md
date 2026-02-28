<file name=0 path=README.md># go-loglint

**go-loglint** — кастомный линтер для `golangci-lint`, который проверяет сообщения логов и ловит утечки чувствительных данных.  
Фокус проекта: **максимально полезные проверки при минимуме false-positive** — анализируем **только те логгер‑вызовы, которые явно описаны в конфиге**, и не трогаем остальной код.

---

## Установка

### Требования

- Go **1.26+**
- `golangci-lint` **2.x**
- macOS/Linux (используется `-buildmode=plugin`, нужен `CGO_ENABLED=1`)

> Важно: Go, которым вы **собираете** `loglint.so`, должен совпадать с Go, которым `golangci-lint` **загружает** плагин. Иначе будет ошибка вида:  
> `plugin was built with a different version of package internal/goarch/encoding`.

### Шаги

1) Клонируйте репозиторий и перейдите в него:
```bash
git clone https://github.com/alnoi/go-loglint
cd go-loglint
```

2) Включите CGO (нужно для plugin mode):
```bash
go env -w CGO_ENABLED=1
```

3) Очистите кэши и соберите плагин:
```bash
go clean -cache -testcache
go build -buildmode=plugin -o loglint.so ./plugin
```

4) Установите `golangci-lint` (если нужно):
```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

5) Убедитесь, что используется правильный `golangci-lint`:
```bash
which golangci-lint
golangci-lint version
```

---

## Использование

Все настройки линтера передаются через `.golangci.yml`.

Минимальный рабочий пример:

```yaml
version: "2"

linters:
  default: none
  enable:
    - loglint
  settings:
    custom:
      loglint:
        path: ./loglint.so
        description: Selectel log message linter
```

После этого запустите:

```bash
golangci-lint run ./...
```

---

## Конфигурация

Линтер поддерживает **3 способа конфигурации**.

### 1) Рекомендуемый способ — конфигурация прямо в `.golangci.yml`

Основной и самый простой способ — описать настройки линтера прямо в `.golangci.yml`.

Пример полной конфигурации:

```yaml
version: "2"

linters:
  default: none
  enable:
    - loglint
  settings:
    custom:
      loglint:
        path: ./loglint.so
        description: Selectel log message linter
        settings:
          configPath: ./loglint.yaml
          rules:
            lowercase: true
            englishOnly: true
            noEmoji: true
            sensitive: true
          sensitive:
            patterns: [password, passwd, token, secret, api_Key, apikey, authorization, bearer]
            allowlist: []
          exclude:
            paths:
              - vendor/
            files:
              - "*.pb.go"
              - "*_gen.go"
          logAPIs:
            - packagePath: "log/slog"
              receiverPkgPath: "log/slog"
              receiverType: "Logger"
              methods: ["Debug", "Info", "Warn", "Error"]
            - packagePath: "log"
              receiverPkgPath: "log"
              receiverType: "Logger"
              methods: ["Print", "Printf", "Println", "Fatal", "Fatalf", "Fatalln", "Panic", "Panicf", "Panicln"]
            - receiverPkgPath: "go.uber.org/zap"
              receiverType: "Logger"
              methods: ["Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal"]
            - receiverPkgPath: "go.uber.org/zap"
              receiverType: "SugaredLogger"
              methods: ["Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal", "Debugf", "Infof", "Warnf", "Errorf", "DPanicf", "Panicf", "Fatalf", "Debugw", "Infow", "Warnw", "Errorw", "DPanicw", "Panicw", "Fatalw"]
```

В этом режиме линтер полностью настраивается через `.golangci.yml` и не требует отдельного YAML-файла.

---

### 2) Внешний конфиг через `settings.configPath`

Если конфигурация большая или используется корпоративный общий файл, можно передать путь до внешнего YAML:

```yaml
linters-settings:
  custom:
    loglint:
      path: ./loglint.so
      settings:
        configPath: ./loglint.yaml
```

В этом случае линтер:

1. Загружает внешний YAML
2. Парсит его в структуру конфигурации
3. Мержит его с настройками, указанными напрямую в `.golangci.yml`

Правила merge:

- Значения из внешнего файла **перекрывают дефолтные**.
- Если одновременно заданы поля в `.golangci.yml` и во внешнем YAML — приоритет имеет `.golangci.yml`.
- Если какие-то поля отсутствуют — используются дефолтные значения.

---

### 3) Без указания конфига (дефолтные настройки)

Если `settings.configPath` не указан, линтер запускается с дефолтной конфигурацией.

Дефолтные значения:

```yaml
rules:
  lowercase: true
  englishOnly: true
  noEmoji: true
  sensitive: true

sensitive:
  patterns: [password, passwd, token, secret, api_Key, apikey, authorization, bearer]
  allowlist: []

logAPIs:
  - packagePath: log/slog
    receiverPkgPath: log/slog
    receiverType: Logger
    methods: ["Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal", "Debugf", "Infof", "Warnf", "Errorf", "DPanicf", "Panicf", "Fatalf", "Debugw", "Infow", "Warnw", "Errorw", "DPanicw", "Panicw", "Fatalw"]
  
  - receiverPkgPath: go.uber.org/zap
    receiverType: Logger
    methods: ["Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal", "Debugf", "Infof", "Warnf", "Errorf", "DPanicf", "Panicf", "Fatalf", "Debugw", "Infow", "Warnw", "Errorw", "DPanicw", "Panicw", "Fatalw"]

  - receiverPkgPath: go.uber.org/zap
    receiverType: SugaredLogger
    methods: ["Debug", "Info", "Warn", "Error", "DPanic", "Panic", "Fatal", "Debugf", "Infof", "Warnf", "Errorf", "DPanicf", "Panicf", "Fatalf", "Debugw", "Infow", "Warnw", "Errorw", "DPanicw", "Panicw", "Fatalw"]

exclude:
  paths:
    - vendor/
  files:
    - "*.pb.go"
    - "*_gen.go"
```

---

### Как происходит загрузка и merge

Если указан `settings.configPath`, линтер:

1. Загружает YAML-файл
2. Валидирует структуру
3. Использует его как runtime-конфигурацию

Если файл не указан — используются дефолтные значения.

Линтер **не требует пересборки** при изменении YAML — достаточно изменить конфиг и перезапустить `golangci-lint`.

---

## Что именно настраивается

В конфиге можно управлять:

- правилами сообщений (`lowercase`, `noEmoji`, `englishOnly`)
- списком чувствительных паттернов (`patterns`)
- исключениями (`allowlist`)
- описанием поддерживаемых логгеров (`logAPIs`)

Ключевая идея: линтер анализирует **только те логгер‑вызовы, которые явно описаны в `logAPIs`**.

Это позволяет использовать его с любым логгером (zap, slog, logrus, собственные обёртки и т.д.), просто добавив описание в конфиг.

---

## Что именно проверяется

### 1) Правила сообщений (message rules)

Настраиваются в `rules`:

- `lowercase`: сообщение должно начинаться с маленькой буквы (после пробелов/табов)
- `noEmoji`: запрещены emoji и “спецсимволы” (например `...`, `…`, `?`, `!`, ZWJ/VS16 и т.п.)
- `englishOnly`: только ASCII (защищает от скрытых Unicode-символов)

Важно: линтер может сообщать **несколько нарушений** для одного сообщения (например, `noEmoji` + `englishOnly`).

### 2) Чувствительные данные (sensitive)

Настраиваются в `sensitive`:

- `patterns`: список подстрок/идентификаторов, которые считаются “опасными”
- `allowlist`: исключения (если совпало с allowlist — не считаем утечкой)

Что мы ловим:

- переменные: `password`, `token`
- поля структур: `user.Password`, `req.Token`
- вложенные выражения: конкатенации `x + token`, вызовы функций `foo(bar(password))`, аргументы внутри `zap.String("k", token)` и т.п.

⚠️ **результат произвольных функций не вычисляется** (например, `fmt.Sprintf`)

---

## Любой логгер — через конфиг

Ключевая фишка: линтер **не привязан** к конкретному логгеру.  
Он работает с любыми API, если вы описали их в `logAPIs`.

### package calls & receiver calls

Мы поддерживаем два основных паттерна, чтобы не путать разные вызовы:

1) **Package call**: `pkg.Func(...)`  
   Пример: `slog.Info("msg")`  
   Настраивается через:
   - `packagePath`
   - `methods`

2) **Receiver call**: `obj.Method(...)`  
   Пример: `logger.Info("msg")`  
   Настраивается через:
   - `receiverPkgPath`
   - `receiverType`
   - `methods`

Про поля `receiverPkgPath` и `receiverType`:

`receiverPkgPath` — путь пакета, в котором объявлен тип receiver’а.

`receiverType` — имя типа, для которого разрешены методы логирования.

При анализе вызова вида `obj.Method(...)` линтер:
1. Получает тип `obj` из type-checker.
2. Извлекает имя типа.
3. Извлекает путь пакета, в котором этот тип объявлен.
4. Сравнивает их с `receiverType` и `receiverPkgPath` из конфига.

Вызов считается логгером только при совпадении **обоих** значений.

---

## Поддержка compile-time сообщений

Линтер умеет доставать сообщение как **compile-time строку**:
- строковые литералы `"msg"`
- конкатенации констант `"a" + "b"`
- локальные `const` строки

Если значение не является compile-time строкой (например `slog.Info(f)` где `f` — переменная), правило формата сообщения не будет применяться, но **чувствительные аргументы** всё равно проверяются.

---

## Тесты

### Структура

- Юнит‑тесты для правил/детекторов:
  - `internal/analyzer/loglint/*_test.go`

- Интеграционные тесты анализатора через `analysistest`:
  - `internal/analyzer/loglint/testdata/src/...`
  - несколько категорий (basic, apis, sensitive, constmsg и т.д.)
  - mock для go.uber.org для работы тестов
  - конфиги для тестов:
    - `internal/analyzer/loglint/testdata/cfg/*.yaml`

### Запуск

Все тесты:
```bash
go test ./...
```

С покрытием:
```bash
go test ./... -cover
```

---

## Подход к минимизации false-positive

Линтер спроектирован таким образом, чтобы минимизировать ложные срабатывания.

- Анализируется только строго ограниченный набор логгеров, явно описанных в `logAPIs`.
- Различаются `pkg_call` и `receiver_call`; для receiver-вызовов дополнительно валидируются тип и пакет объявления.
- Не производится попытка вычислять произвольные выражения, которые не гарантируются type-checker (например, форматирующие функции), чтобы избежать шумных результатов.
- Для поиска чувствительных данных используется рекурсивный обход выражений, что позволяет находить утечки во вложенных аргументах.

---

## Troubleshooting (частые проблемы)

### `plugin was built with a different version of package internal/goarch/encoding`

Проверьте версии:

```bash
go version
golangci-lint version
```

Версия Go, которой собран `loglint.so`, и версия Go, которой собран `golangci-lint`, **должны совпадать**.

Если версии отличаются — пересоберите плагин тем же Go, который использует `golangci-lint`.

Пример пересборки:

```bash
rm -f loglint.so
go clean -cache -testcache
go build -buildmode=plugin -o loglint.so ./plugin
golangci-lint run ./...
```

---
### Примеры использования

в `cmd/main.go` можно найти простенький код с логами

вот результат запуска логера на нем

```bash
alan@MacBook-Pro-Sulejmanov go-loglint % golangci-lint run
cmd/main.go:19:12: loglint(lowercase): message must start with a lowercase letter (loglint)
        slog.Info("Login started")
                  ^
cmd/main.go:23:51: loglint(no-sensitive): log contains sensitive field (loglint)
        slog.Info("auth successes with user", u.Email, u.Password, u.Token)
                                                         ^
cmd/main.go:25:9: loglint(no-emoji): message must not contain emoji or special characters: found emoji '🥳' (loglint)
        z.Info("login ended 🥳")
               ^
3 issues:
* loglint: 3

```