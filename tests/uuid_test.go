package tests

import (
	"fmt"
	"github.com/golang-module/carbon/v2"
	uuid "github.com/satori/go.uuid"
	"megin/app/api/service/es"
	"testing"
	"time"
)

func v2() {
	u1 := uuid.NewV2(uuid.DomainGroup).String()
	fmt.Println("go", u1)
}

func TestTime(t *testing.T) {
	betAtStr := "2022-12-08 23:50:00"
	dt := carbon.Parse(betAtStr, carbon.Bangkok).SetTimezone(carbon.PRC)
	nBetDate := dt.Carbon2Time().Format(carbon.ShortDateLayout)
	fmt.Println(nBetDate)

	betAt := es.ParseDateTime(carbon.Bangkok, "2022-12-09 12:00:00")
	fmt.Println(betAt)
}

//都是一样的。。
func TestUuidV2(t *testing.T) {

	go v2()

	u1 := uuid.NewV2(uuid.DomainGroup).String()
	fmt.Println("testv2", u1)

	time.Sleep(time.Second * 3)
	go v2()
	u2 := uuid.NewV2(uuid.DomainGroup).String()
	fmt.Println("testv2", u2)
}

func TestUuidV5(t *testing.T) {

	u1 := uuid.NewV4().String()
	fmt.Println(u1)

	u2 := uuid.NewV4().String()
	fmt.Println(u2)
}
