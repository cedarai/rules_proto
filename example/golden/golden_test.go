package golden

import (
	"testing"

	"github.com/stackb/rules_proto/pkg/goldentest"
)

func TestGoldens(t *testing.T) {
	goldentest.
		FromDir("example/golden", goldentest.WithOnlyTests("proto_compiled_sources")).
		Run(t, "gazelle")
}
