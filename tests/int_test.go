package tests

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func Test_AllIntegrationTests(t *testing.T) {
	suite.Run(t, newTestSuite())
}
