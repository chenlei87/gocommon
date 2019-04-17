package rmdb

import (
	"database/sql"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

//UdbSel 通用数据库配置
type UdbSel struct {
	db      *sql.DB
	StrSQL  string //table 换为%s的形式
	readAll func(stmt *sql.Rows) (interface{}, error)
	ElemNum int

	paralNum int
}

//CheckErr CheckErr
func CheckErr(err error) {
	if err != nil {
		pc, _, _, _ := runtime.Caller(1)
		f := runtime.FuncForPC(pc)
		fmt.Println(err)
		fmt.Println(f.Name(), string(debug.Stack()))
		panic(err)
	}
}

//PrintflnErr PrintflnErr
func PrintflnErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

//InitSel InitSel
func InitSel(ConnMax int, engine string, dsn string, sqlsel string, ParalNum int,
	ReadAll func(rows *sql.Rows) (interface{}, error)) (a *UdbSel, err error) {

	var v UdbSel
	v.db, err = sql.Open(engine, dsn)
	CheckErr(err)

	v.db.SetMaxOpenConns(ConnMax)
	v.db.SetMaxIdleConns(0)
	err = v.db.Ping()
	CheckErr(err)

	v.readAll = ReadAll

	sqlsel = strings.Replace(sqlsel, ";", "", -1)
	v.StrSQL = sqlsel

	num := 0
	//get count
	sqlselN := fmt.Sprintf("select count(0) from (%s) t1", sqlsel)
	err = v.db.QueryRow(sqlselN).Scan(&num)
	CheckErr(err)

	if num > 0 {
		v.ElemNum = num
	} else {
		fmt.Println("read empty data")
	}

	fmt.Println("InitSel: ", sqlselN, " num:", num)

	if ParalNum > 64 || ParalNum < 1 {
		CheckErr(fmt.Errorf("ParalNum should in [1,64] "))
	}
	v.paralNum = ParalNum

	return &v, nil
}

func doxx(strsql string, v *UdbSel, pnum *int64) {
	rowselems, err := v.db.Query(strsql)
	defer rowselems.Close()
	CheckErr(err)

	fmt.Println("read ", 0, "/ ", v.ElemNum)
	for rowselems.Next() {
		n := atomic.AddInt64(pnum, 1)
		if n%10000 == 0 {
			fmt.Println("read ", n, "/ ", v.ElemNum)
		}
		v.readAll(rowselems)
	}
}

//DoSel DoSel
//rowselem 已经关闭
func (v *UdbSel) DoSel() (interface{}, error) {
	readnum := int64(0)
	if v.readAll != nil {
		if v.paralNum == 1 {
			doxx(v.StrSQL, v, &readnum)
		} else {
			var wg sync.WaitGroup
			wg.Add(v.paralNum)
			for i := 0; i < v.paralNum; i++ {
				strsql := fmt.Sprintf("select * from (%s) t1 where id %% %d = %d", v.StrSQL, v.paralNum, i)
				go func(strsql string) {
					doxx(strsql, v, &readnum)
					wg.Add(-1)
				}(strsql)
			}
			wg.Wait()
		}
		fmt.Println("read over ", readnum, "/ ", v.ElemNum)
	}
	return nil, nil
}

//Dispose Dispose
func (v *UdbSel) Dispose() {
	if v.db != nil {
		v.db.Close()
	}
}

// func main() {
// 	var l sync.Mutex
// 	var elems []ImgElem

// 	SelOneData, _ := pg.InitSel(
// 		4, "postgres", "postgres://postgres:lingdasa@192.168.100.44:5432/chenlei1230?sslmode=disable",
// 		fmt.Sprintf("select id,features from  %s order by id limit 200001 ;", "features_570w"), 16,
// 		func(rowselems *sql.Rows) (interface{}, error) {
// 			var elem ImgElem
// 			err := rowselems.Scan(&elem.fid, (*pq.StringArray)(&elem.feature))
// 			if err == nil {
// 				l.Lock()
// 				elems = append(elems, elem)
// 				l.Unlock()
// 			} else {
// 				fmt.Println(err)
// 			}
// 			return nil, fmt.Errorf("rowselems is nil")
// 		})

// 	t1 := time.Now()
// 	SelOneData.DoSel()

// 	fmt.Println("ok", len(elems), time.Now().Sub(t1))
// }
