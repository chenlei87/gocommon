package rps

import (
	"encoding/json"
	"time"
)

//rps rps
type rps struct {
	AllN      int64
	MonthN    []int64
	DayN      []int64
	HourN     []int64
	MinuteN   []int64
	SecondN   []int64
	StartTime time.Time
	uptime    time.Time
}

//Newrps Newrps
func Newrps() *rps {
	v := rps{
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
func (rps *rps) Add(n int64) {
	tn := time.Now()

	if tn.Second() != rps.uptime.Second() {
		rps.SecondN[tn.Second()] = 0

		if tn.Minute() != rps.uptime.Minute() {
			rps.MinuteN[tn.Minute()] = 0

			if tn.Hour() != rps.uptime.Hour() {
				rps.HourN[tn.Hour()] = 0

				if tn.Day() != rps.uptime.Day() {
					rps.DayN[tn.Day()-1] = 0

					if tn.Month() != rps.uptime.Month() {
						rps.MonthN[tn.Month()-1] = 0
					}
				}
			}
		}

	}

	rps.uptime = tn

	rps.AllN += n
	rps.MonthN[tn.Month()-1] += n
	rps.DayN[tn.Day()-1] += n
	rps.HourN[tn.Hour()] += n
	rps.MinuteN[tn.Minute()] += n
	rps.SecondN[tn.Second()] += n
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
func (rps *rps) AddNT(rps1 *rps) {
	rps.AllN += rps1.AllN
	addSlice(&rps.MonthN, rps1.MonthN)
	addSlice(&rps.DayN, rps1.DayN)
	addSlice(&rps.HourN, rps1.HourN)
	addSlice(&rps.MinuteN, rps1.MinuteN)
	addSlice(&rps.SecondN, rps1.SecondN)
}

func (rps *rps) String() string {
	b, _ := json.Marshal(rps)

	return string(b)
}

//AVGInfo AVGInfo
type AVGInfo struct {
	SecRPS         float64
	MinRPS         float64
	HourRPS        float64
	DayRPS         float64
	MonthRPS       float64
	ExepireSeconds float64
	AllN           int64
}

func funa(left, allsec float64) float64 {
	if left > allsec {
		return float64(allsec)
	}
	return float64(left)
}

//Average Average
//上一秒的avg
func (rps *rps) Average() AVGInfo {

	tnlast := time.Now().Add((-1) * time.Second)
	tn := time.Now()

	allsec := time.Now().Sub(rps.StartTime).Seconds()
	//本分钟经过的秒数
	ms := float64(tn.UnixNano()%1000000000/1000000+1) / float64(1000)
	//fmt.Println(tn.Unix(), tn.UnixNano(), ms)
	minSec := funa(float64(tn.Second())+ms, allsec)
	hourSec := funa(float64((tn.Second()+tn.Minute()*60))+ms, allsec)
	daySec := funa(float64((tn.Second()+(tn.Minute()*60)+(tn.Hour()*60*60)))+ms, allsec)
	monthSec := funa(float64((tn.Second()+(tn.Minute()*60)+(tn.Hour()*60*60)+(tn.Day()*60*60*24)))+ms, allsec)

	//fmt.Println("rps: N", allsec, rps.MinuteN[tn.Minute()], rps.HourN[tn.Hour()], rps.DayN[tn.Day()-1], rps.MonthN[tn.Month()-1])
	//fmt.Println("rps: T", allsec, minSec, hourSec, daySec, monthSec)

	v := AVGInfo{
		AllN:           rps.AllN,
		ExepireSeconds: tn.Sub(rps.StartTime).Seconds(),
		SecRPS:         float64(rps.SecondN[tnlast.Second()]),      //上一秒的平均速度
		MinRPS:         float64(rps.MinuteN[tn.Minute()]) / minSec, // 这一分组到目前的平均速度
		HourRPS:        float64(rps.HourN[tn.Hour()]) / hourSec,
		DayRPS:         float64(rps.DayN[tn.Day()-1]) / daySec,
		MonthRPS:       float64(rps.MonthN[tn.Month()-1]) / monthSec}
	return v
}

func (avg AVGInfo) String() string {
	b, _ := json.MarshalIndent(avg, "", "\t")

	return string(b)
}
