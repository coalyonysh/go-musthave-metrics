// multichecker объединяет статические анализаторы для проверки качества кода.
package main

import (
	"github.com/coalyonysh/go-musthave-metrics/cmd/staticlint/noexit"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(noexit.NoExitAnalyzer)
}
