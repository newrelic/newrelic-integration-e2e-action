package runtime

import (
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"testing"
)

func TestMetricsTester_Test(t *testing.T) {
	log := logrus.New()
	log.SetOutput(ioutil.Discard)
	NewEntitiesTester(clientMock{}, log)
}