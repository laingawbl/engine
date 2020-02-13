package controller_test

import (
	"testing"

	"github.com/laingawbl/engine/controller"
	"github.com/laingawbl/engine/controller/testsuite"
)

func TestInMemSuite(t *testing.T) {
	s := controller.InMemStore()
	testsuite.Suite(t, s, func() { s.(interface{ Clear() }).Clear() })
}
