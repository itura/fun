package build

import (
	"testing"
)

func TestPrepareBuild(t *testing.T) {
	app := getHelmApplication()
	app.PrepareBuild()
}
