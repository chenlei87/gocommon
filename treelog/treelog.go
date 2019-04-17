package treelog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

//MsgT MsgT
type MsgT struct {
	My   string      `json:"my,omitempty"`
	Time string      `json:"time,omitempty"`
	Msg  interface{} `json:"msg,omitempty"`
	Sub  [](*MsgT)   `json:"sub,omitempty"`
	t    time.Time

	sync.Mutex
}

//NewLogMsg NewLogMsg
func NewLogMsg(a ...interface{}) *MsgT {
	var msg MsgT
	msg.My = fmt.Sprint(a...) //msg.My + " " + strmy
	msg.t = time.Now()
	msg.Msg = ""
	return &msg
}

//StringIdent StringIdent
func (msg *MsgT) StringIdent() string {

	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln(err)
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")

	if err != nil {
		log.Fatalln(err)
	}

	return out.String()
}

//String String
func (msg *MsgT) String() string {

	b, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln(err)
	}

	return string(b)
}

//SetMy SetMy
func (msg *MsgT) SetMy(a ...interface{}) {
	msg.My = fmt.Sprint(a...) //msg.My + " " + strmy
}

//AddMy AddMy
func (msg *MsgT) addMy(a ...interface{}) {
	msg.My = fmt.Sprint(msg.My, a) //msg.My + " " + strmy
}

//SetMsg SetMsg
func (msg *MsgT) SetMsg(a ...interface{}) {
	//msg.Msg = fmt.Sprint(a...)
	//b, _ := json.Marshal(a)
	//msg.Msg = fmt.Sprint(string(b))
	msg.Msg = a

	if time.Now().After(msg.t.Add(time.Millisecond)) {
		msg.Time = time.Now().Sub(msg.t).String()
	}
}

//addSub addSub
func (msg *MsgT) addSub(sub *MsgT) {
	msg.Lock()
	msg.Sub = append(msg.Sub, sub)
	msg.Unlock()
}

//NewSub NewSub
func (msg *MsgT) NewSub(a ...interface{}) *MsgT {
	sub := NewLogMsg(a...)
	msg.addSub(sub)
	//msg.SetMsg(a...)
	return sub
}

//NTlog NTlog
type NTlog struct {
	AllN      int64
	MonthN    []int64
	DayN      []int64
	HourN     []int64
	MinuteN   []int64
	SecondN   []int64
	StartTime time.Time
	uptime    time.Time
}

//NewNTlog NewNTlog
func NewNTlog() *NTlog {
	v := NTlog{
		StartTime: time.Now(),
		AllN:      0,
		MonthN:    make([]int64, 12),
		DayN:      make([]int64, 31),
		HourN:     make([]int64, 24),
		MinuteN:   make([]int64, 60),
		SecondN:   make([]int64, 60)}

	return &v
}

//Add Add
func (ntlog *NTlog) Add(n int64) {
	tn := time.Now()

	if tn.Second() != ntlog.uptime.Second() {
		ntlog.SecondN[tn.Second()] = 0

		if tn.Minute() != ntlog.uptime.Minute() {
			ntlog.MinuteN[tn.Minute()] = 0

			if tn.Hour() != ntlog.uptime.Hour() {
				ntlog.HourN[tn.Hour()] = 0

				if tn.Day() != ntlog.uptime.Day() {
					ntlog.DayN[tn.Day()-1] = 0

					if tn.Month() != ntlog.uptime.Month() {
						ntlog.MonthN[tn.Month()-1] = 0
					}
				}
			}
		}

	}

	ntlog.uptime = tn

	ntlog.AllN += n
	ntlog.MonthN[tn.Month()-1] += n
	ntlog.DayN[tn.Day()-1] += n
	ntlog.HourN[tn.Hour()] += n
	ntlog.MinuteN[tn.Minute()] += n
	ntlog.SecondN[tn.Second()] += n
}

//addSlice addSlice
func addSlice(dst *[]int64, src []int64) {
	if len(*dst) != len(src) || len(src) == 0 {
		return
	}

	for i, v := range src {
		(*dst)[i] += v
	}
	return
}

//AddNT AddNT
func (ntlog *NTlog) AddNT(ntlog1 *NTlog) {
	ntlog.AllN += ntlog1.AllN
	addSlice(&ntlog.MonthN, ntlog1.MonthN)
	addSlice(&ntlog.DayN, ntlog1.DayN)
	addSlice(&ntlog.HourN, ntlog1.HourN)
	addSlice(&ntlog.MinuteN, ntlog1.MinuteN)
	addSlice(&ntlog.SecondN, ntlog1.SecondN)
}

//AVGInfo AVGInfo
type AVGInfo struct {
	SecAvg   int64
	MinAvg   int64
	HourAvg  int64
	DayAvg   int64
	MonthAvg int64
	Exepire  time.Duration
	AllN     int64
}

//Average Average
//上一秒的avg
func (ntlog *NTlog) Average() AVGInfo {

	tn := time.Now().Add((-1) * time.Second)

	v := AVGInfo{
		AllN:     ntlog.AllN,
		Exepire:  tn.Sub(ntlog.StartTime),
		SecAvg:   ntlog.SecondN[tn.Second()],
		MinAvg:   ntlog.MinuteN[tn.Minute()] / (int64)(tn.Second()+1),
		HourAvg:  ntlog.HourN[tn.Hour()] / (int64)(tn.Second()+1+(tn.Minute()*60)),
		DayAvg:   ntlog.DayN[tn.Day()-1] / (int64)(tn.Second()+1+(tn.Minute()*60)+(tn.Hour()*60*60)),
		MonthAvg: ntlog.MonthN[tn.Month()-1] / (int64)(tn.Second()+1+(tn.Minute()*60)+(tn.Hour()*60*60)+(tn.Day()*60*60*24))}
	return v
}
