package consumer_bet

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	"megin/app"
	"megin/system"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	var ts int64
	ts = 1654378187000
	time := ParseDateTime(ts)

	fmt.Println(time)

}

func ParseDateTime(ts int64) string {
	return time.Unix(ts/1000, 0).Format(carbon.DateTimeLayout)
}

func ParseDateTime2(ts int64) time.Time {
	return time.Unix(ts/1000, 0)
}

func TestMain(m *testing.M) {
	system.ServerInit("../../resources/config-dev.yaml", app.OnAppInitialize)
	m.Run()
}
