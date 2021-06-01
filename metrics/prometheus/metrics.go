package prometheus

import (
	"fmt"
)

var namespace = "opera"

// SetNamespace for metrics.
func SetNamespace(s string) {
	namespace = s
}

func prometheusDelims(name string) string {
	return fmt.Sprintf("%s_%s", namespace, name)
}
