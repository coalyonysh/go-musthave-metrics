# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Бенчмарки и профилирование

Проект включает бенчмарки для ключевых компонентов системы:

### Хранилище (internal/storage)

- `BenchmarkMemStorage_SetGauge` - Установка gauge метрик
- `BenchmarkMemStorage_SetCounter` - Установка counter метрик
- `BenchmarkMemStorage_GetGauge` - Чтение gauge метрик
- `BenchmarkMemStorage_GetCounter` - Чтение counter метрик
- `BenchmarkMemStorage_GetAllGauges` - Получение всех gauge метрик
- `BenchmarkMemStorage_GetAllCounters` - Получение всех counter метрик
- `BenchmarkMemStorage_SetMetricsBatch` - Пакетная запись метрик
- `BenchmarkMemStorage_ConcurrentSetGauge` - Параллельная запись gauge
- `BenchmarkMemStorage_ConcurrentGetGauge` - Параллельное чтение gauge

### Сигнатура (pkg/signature)

- `BenchmarkCalculateHash` - Вычисление HMAC-SHA256 хеша
- `BenchmarkVerifyHash` - Проверка хеша
- `BenchmarkCalculateHash_LargeData` - Хеширование больших данных

### Сервис аудита (internal/service)

- `BenchmarkAuditService_Log` - Логирование событий аудита
- `BenchmarkAuditService_Log_ManyMetrics` - ЛогированиеMany метрик

### Результаты оптимизации памяти

После профилирования и оптимизации кода получены следующие результаты:

```
File: storage.test
Type: alloc_space
Showing nodes accounting for -5183.73kB, 17.87% of 29015.36kB total

      flat  flat%   sum%        cum   cum%
-4149.23kB 14.30% 14.30% -4149.23kB 14.30%  github.com/coalyonysh/go-musthave-metrics/internal/storage.(*MemStorage).GetAllGauges
-1037.31kB  3.58% 17.88% -1037.31kB  3.58%  github.com/coalyonysh/go-musthave-metrics/internal/storage.(*MemStorage).GetAllCounters
```

**Результаты бенчмарка (до/после оптимизации):**

| Операция | До (B/op) | После (B/op) | Улучшение |
|----------|-----------|--------------|----------|
| GetAllGauges | 109016 | 54656 | **49.9%** |
| GetAllCounters | 109016 | 54656 | **49.9%** |

| Операция | До (allocs) | После (allocs) | Улучшение |
|----------|-------------|----------------|----------|
| GetAllGauges | 22 | 6 | **72.7%** |
| GetAllCounters | 22 | 6 | **72.7%** |

### Профилирование

Для запуска профилирования используйте:

```bash
# Снятие профиля памяти
go test -v -run TestProfileMemory ./internal/storage/... -o profiles/base.pprof

# Оптимизация кода...

# Снятие профиля после оптимизации
go test -v -run TestProfileMemoryResult ./internal/storage/... -o profiles/result.pprof

# Сравнение профилей
go tool pprof -top -sample_index=alloc_space -diff_base=profiles/base.pprof profiles/result.pprof
```

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**
