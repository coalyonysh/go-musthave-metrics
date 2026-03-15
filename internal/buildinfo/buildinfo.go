// Package buildinfo предоставляет функции для вывода информации о сборке.
//
// Пакет используется для вывода версии, даты и коммита сборки при старте приложения.
// Значения могут быть установлены во время сборки через переменные:
//   - buildVersion - версия сборки
//   - buildDate - дата сборки
//   - buildCommit - коммит сборки
//
// Пример использования:
//
//	buildinfo.PrintBuildInfo(buildVersion, buildDate, buildCommit)
package buildinfo

import "fmt"

// PrintBuildInfo выводит в stdout информацию о сборке приложения.
//
// Если переданные значения пустые, вместо них выводится "N/A".
// Формат вывода:
//
//	Build version: <buildVersion>
//	Build date: <buildDate>
//	Build commit: <buildCommit>
func PrintBuildInfo(buildVersion, buildDate, buildCommit string) {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n",
		normalizeValue(buildVersion),
		normalizeValue(buildDate),
		normalizeValue(buildCommit))
}

// normalizeValue возвращает значение или "N/A", если значение пустое
func normalizeValue(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
