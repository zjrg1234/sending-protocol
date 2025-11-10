package pg_test

import (
	"megin/app"
	"megin/system"
	"testing"
)

func TestConnect(t *testing.T) {

}

// ====================
func TestMain(m *testing.M) {
	system.ServerInit("../../resources/config-dev.yaml", app.OnAppInitialize)
	m.Run()
}
