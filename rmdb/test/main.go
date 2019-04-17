package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"git.topvdn.com/NuclearPower/aid190630/components/treelog"

	"git.topvdn.com/NuclearPower/aid190630/components/pg"
	"github.com/lib/pq"
)

//ImgElem ImgElem
type ImgElem struct {
	fid     int
	feature []string
	urlbase string
	data    []byte
	dlmsg   string
	ctime   int64
}

var (
	elems []ImgElem
)

func read() {
	var l sync.Mutex

	SelOneData, _ := pg.InitSel(
		4, "postgres", "postgres://postgres:lingdasa@192.168.100.44:5432/chenlei1230?sslmode=disable",
		fmt.Sprintf("select id,features from  %s order by id limit 5000 ;", "features_570w"), 4,
		func(rowselems *sql.Rows) (interface{}, error) {
			var elem ImgElem
			err := rowselems.Scan(&elem.fid, (*pq.StringArray)(&elem.feature))
			if err == nil {
				l.Lock()
				elems = append(elems, elem)
				l.Unlock()
			} else {
				fmt.Println(err)
			}
			return nil, fmt.Errorf("rowselems is nil")
		})

	t1 := time.Now()
	SelOneData.DoSel()

	fmt.Println("ok", len(elems), time.Now().Sub(t1))
}

func main() {
	ntlog := treelog.NewNTlog()

	go func() {
		for {
			ntlog.Add(10)
			time.Sleep(time.Millisecond * 50)
		}
	}()

	for {
		t1 := time.Now()
		time.Sleep(time.Second)

		fmt.Println("month:", int(t1.Month()), "day:", t1.Day(), "hour:", t1.Hour(), "min:", t1.Minute(), "sec:", t1.Second())

		b1, _ := json.Marshal(ntlog)
		fmt.Println(string(b1))
	}

	// read()

	TestWrite, _ := pg.InitNoSel(
		4, "postgres", "postgres://postgres:lingdasa@192.168.100.44:5432/chenlei1230?sslmode=disable",
		fmt.Sprintf("insert into %s (id,features) VALUES ($1,$2);", "features_570w_temp"), 4)

	for _, v := range elems {
		TestWrite.DoExeAysn(v.fid, (pq.StringArray)(v.feature))
	}

	time.Sleep(time.Second)

	TestWrite.Info()

	stop := make(chan bool)
	<-stop

	return
}
