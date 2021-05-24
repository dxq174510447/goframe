package proxy

import "testing"

type RunFirst struct {
}

func TestApplicationRun(t *testing.T) {
	args := []string{}
	ApplicationRun(&RunFirst{}, args...)

}
