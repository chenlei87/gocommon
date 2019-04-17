package rmdb

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"git.topvdn.com/NuclearPower/aid190630/components/treelog"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

//UdbNoSel 通用数据库配置
type UdbNoSel struct {
	db       *sql.DB
	StrExe   string
	stmt     (*sql.Stmt)
	paralnum int
	data     chan ([]interface{})

	ctx context.Context
}

//Info Info
type Info struct {
	Errtime int
	Oktime  int
	Errmsg  map[string]int

	Ms []int
}

//InitNoSel InitNoSel
func InitNoSel(ConnMax int, engine string, dsn string, StrExe string, ParalNum int) (a *UdbNoSel, err error) {

	var v UdbNoSel
	v.db, err = sql.Open(engine, dsn)
	CheckErr(err)

	v.db.SetMaxOpenConns(ConnMax)
	v.db.SetMaxIdleConns(0)
	err = v.db.Ping()
	CheckErr(err)

	v.StrExe = StrExe
	fmt.Println(StrExe)

	v.stmt, err = v.db.Prepare(v.StrExe)
	CheckErr(err)

	if ParalNum > 64 || ParalNum < 1 {
		CheckErr(fmt.Errorf("ParalNum should in [1,64] "))
	}

	v.data = make(chan []interface{}, 100)
	v.paralnum = ParalNum

	ctxkay := "NoSel"
	ctxval := new(sync.Map)
	v.ctx = context.WithValue(context.Background(), ctxkay, ctxval)

	for i := 0; i < ParalNum; i++ {
		go func(ctx context.Context, i int) {
			stmt, err := v.db.Prepare(v.StrExe)
			CheckErr(err)

			staticinfo := Info{Errmsg: make(map[string]int), Ms: make([]int, 1)}
			ntk := time.NewTicker(time.Second).C
			for {
				select {
				case <-ntk:
					//fmt.Println("aaa", len(staticinfo.Ms), " : ", staticinfo.Ms)
					if len(staticinfo.Ms) > 10 {
						staticinfo.Ms = append(staticinfo.Ms[1:len(staticinfo.Ms)], 0)
					} else {
						staticinfo.Ms = append(staticinfo.Ms, 0)
					}

				case v1, opened := <-v.data:
					if !opened {
						break
					}
					ts := time.Now()
					//fmt.Println(v1)
					_, err := stmt.Exec(v1...)
					if err != nil {
						staticinfo.Errtime++
						errno, _ := staticinfo.Errmsg[err.Error()]
						staticinfo.Errmsg[err.Error()] = errno + 1
					} else {
						staticinfo.Oktime++
					}

					vif := ctx.Value("NoSel")
					ctxv := vif.(*sync.Map)
					ctxv.Store(i, staticinfo)

					staticinfo.Ms[len(staticinfo.Ms)-1] += int((time.Now().Sub(ts).Nanoseconds() + 500000) / int64(1000000))

				}
			}
		}(v.ctx, i)
	}

	return &v, nil
}

//DoExe DoExe
func (v *UdbNoSel) DoExe(elems ...interface{}) error {

	if v.stmt != nil {
		_, err := v.stmt.Exec(elems...)
		PrintflnErr(err)
		return err
	}

	return fmt.Errorf(fmt.Sprintln(v, "Exe stmt is not init"))
}

//DoExeAysn DoExeAysn
func (v *UdbNoSel) DoExeAysn(elems ...interface{}) {
	v.data <- elems
}

//Info Info
func (v *UdbNoSel) Info() {
	ctxvif := v.ctx.Value("NoSel")
	ctxv := ctxvif.(*sync.Map)

	msg := treelog.NewLogMsg("UdbNoSel")

	sum := Info{Errmsg: make(map[string]int)}

	ctxv.Range(func(k, v interface{}) bool {
		vinfo := v.(Info)
		sum.Oktime += vinfo.Oktime
		sum.Errtime += vinfo.Errtime

		for k, v := range vinfo.Errmsg {
			errno, _ := sum.Errmsg[k]
			sum.Errmsg[k] = errno + v
		}

		submsg := msg.NewSub(strconv.Itoa(k.(int)))
		submsg.SetMsg(vinfo)
		return true
	})

	msg.SetMsg("ok", sum)
	fmt.Println(msg.String())
}

//Dispose Dispose
func (v *UdbNoSel) Dispose() error {
	close(v.data)
	return nil
}
