package model

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Daily struct {
	Date time.Time // 日期
	//	总共
	Confirmed                        int // 确诊病例
	Asymptomatic                     int // 无症状感染者
	Mild                             int // 轻型
	Common                           int // 普通型
	Severe                           int // 重型
	Critical                         int // 危重型
	Death                            int // 死亡
	InHospital                       int // 在院治疗
	DischargedFromHospital           int // 治愈出院
	DischargedFromMedicalObservation int // 解除医学观察
	UnderMedicalObservation          int // 尚在医学观察

	//	本土
	LocalConfirmed                        int // 本土确诊病例
	LocalAsymptomatic                     int // 本土无症状感染者
	LocalConfirmedFromAsymptomatic        int // 从无症状感染者转归确诊的病例
	LocalConfirmedFromBubble              int // 从闭环隔离中发现的本土病例
	LocalConfirmedFromRisk                int // 从风险人群中发现的本土病例
	LocalAsymptomaticFromBubble           int // 从闭环隔离中发现的无症状感染者
	LocalAsymptomaticFromRisk             int // 从风险人群中发现的无症状感染者
	LocalInHospital                       int // 本土在院治疗
	LocalDischargedFromHospital           int // 本土病例出院
	LocalDischargedFromMedicalObservation int // 本土解除医学观察
	LocalDeath                            int // 本土死亡病例
	LocalUnderMedicalObservation          int // 本土尚在医学观察

	//	境外输入
	ImportedConfirmed                        int // 境外输入病例
	ImportedAsymptomatic                     int // 境外输入无症状感染者
	ImportedInHospital                       int // 境外输入在院治疗
	ImportedDischargedFromHospital           int // 境外输入病例出院
	ImportedDischargedFromMedicalObservation int // 境外输入解除医学观察
	ImportedDeath                            int // 境外输入死亡
	ImportedUnderMedicalObservation          int // 境外输入尚在医学观察

	//	累计
	TotalLocalConfirmed                 int // 累计本土确诊
	TotalLocalDischargedFromHospital    int // 累计本土治愈出院
	TotalLocalDeath                     int // 累计本土死亡
	TotalImportedConfirmed              int // 累计境外输入确诊病例
	TotalImportedDischargedFromHospital int // 累计境外输入治愈出院

	//	分区
	DistrictConfirmed                 map[string]int // 城区确诊病例
	DistrictConfirmedFromBubble       map[string]int // 城区从闭环隔离中发现确诊病例
	DistrictConfirmedFromRisk         map[string]int // 城区从风险人群中发现确诊病例
	DistrictConfirmedFromAsymptomatic map[string]int // 城区从无症状感染者中发现确诊病例
	DistrictAsymptomatic              map[string]int // 城区无症状感染者
	DistrictAsymptomaticFromBubble    map[string]int // 城区从闭环隔离中发现无症状感染者
	DistrictAsymptomaticFromRisk      map[string]int // 城区从风险人群中发现无症状感染者

	// meta
	Source string // 来源
}

type Dailys []Daily

var lockDailys sync.Mutex

func (cs Dailys) SaveToCSV(filename string) error {
	districts := []string{
		"浦东新区",
		"徐汇区",
		"闵行区",
		"黄浦区",
		"嘉定区",
		"松江区",
		"虹口区",
		"长宁区",
		"青浦区",
		"静安区",
		"宝山区",
		"杨浦区",
		"普陀区",
		"崇明区",
		"金山区",
		"奉贤区",
	}
	records := [][]string{}
	//	Header
	header := []string{
		"日期",
		//	总体
		"确诊病例",
		"无症状感染者",
		"轻型",
		"普通型",
		"重型",
		"危重型",
		"死亡",
		"在院治疗",
		"治愈出院",
		"解除医学观察",
		"尚在医学观察",
		//	本土
		"本土确诊病例",
		"本土无症状感染者",
		"从无症状感染者转归确诊的病例",
		"从闭环隔离中发现的本土病例",
		"从风险人群中发现的本土病例",
		"从闭环隔离中发现的无症状感染者",
		"从风险人群中发现的无症状感染者",
		"本土在院治疗",
		"本土病例出院",
		"本土解除医学观察",
		"本土死亡病例",
		"本土尚在医学观察",
		//	境外输入
		"境外输入病例",
		"境外输入无症状感染者",
		"境外输入在院治疗",
		"境外输入病例出院",
		"境外输入解除医学观察",
		"境外输入死亡",
		"境外输入尚在医学观察",
		//	累计
		"累计本土确诊",
		"累计治愈出院",
		"累计本土死亡",
		"累计境外输入确诊病例",
		"累计境外输入治愈出院",
	}
	//	分区
	for _, d := range districts {
		header = append(header, fmt.Sprintf("%s_确诊", d))
	}
	for _, d := range districts {
		header = append(header, fmt.Sprintf("%s_确诊_来自闭环隔离", d))
	}
	for _, d := range districts {
		header = append(header, fmt.Sprintf("%s_确诊_来自无症状感染者", d))
	}
	for _, d := range districts {
		header = append(header, fmt.Sprintf("%s_确诊_来自风险人群", d))
	}
	for _, d := range districts {
		header = append(header, fmt.Sprintf("%s_无症状", d))
	}
	for _, d := range districts {
		header = append(header, fmt.Sprintf("%s_无症状_来自闭环隔离", d))
	}
	for _, d := range districts {
		header = append(header, fmt.Sprintf("%s_无症状_来自风险人群", d))
	}
	// meta
	header = append(header, "来源")

	records = append(records, header)

	appendByDistricts := func(r []string, districts []string, dict map[string]int) []string {
		for _, d := range districts {
			if val, ok := dict[d]; ok {
				r = append(r, strconv.Itoa(val))
			} else {
				r = append(r, "0")
			}
		}
		return r
	}

	for _, c := range cs {
		r := []string{
			c.Date.Format("2006-01-02"),
			//	总共
			strconv.Itoa(c.Confirmed),
			strconv.Itoa(c.Asymptomatic),
			strconv.Itoa(c.Mild),
			strconv.Itoa(c.Common),
			strconv.Itoa(c.Severe),
			strconv.Itoa(c.Critical),
			strconv.Itoa(c.Death),
			strconv.Itoa(c.InHospital),
			strconv.Itoa(c.DischargedFromHospital),
			strconv.Itoa(c.DischargedFromMedicalObservation),
			strconv.Itoa(c.UnderMedicalObservation),
			//	本土
			strconv.Itoa(c.LocalConfirmed),
			strconv.Itoa(c.LocalAsymptomatic),
			strconv.Itoa(c.LocalConfirmedFromAsymptomatic),
			strconv.Itoa(c.LocalConfirmedFromBubble),
			strconv.Itoa(c.LocalConfirmedFromRisk),
			strconv.Itoa(c.LocalAsymptomaticFromBubble),
			strconv.Itoa(c.LocalAsymptomaticFromRisk),
			strconv.Itoa(c.LocalInHospital),
			strconv.Itoa(c.LocalDischargedFromHospital),
			strconv.Itoa(c.LocalDischargedFromMedicalObservation),
			strconv.Itoa(c.LocalDeath),
			strconv.Itoa(c.LocalUnderMedicalObservation),
			//	境外输入
			strconv.Itoa(c.ImportedConfirmed),
			strconv.Itoa(c.ImportedAsymptomatic),
			strconv.Itoa(c.ImportedInHospital),
			strconv.Itoa(c.ImportedDischargedFromHospital),
			strconv.Itoa(c.ImportedDischargedFromMedicalObservation),
			strconv.Itoa(c.ImportedDeath),
			strconv.Itoa(c.ImportedUnderMedicalObservation),
			//	累计
			strconv.Itoa(c.TotalLocalConfirmed),
			strconv.Itoa(c.TotalLocalDischargedFromHospital),
			strconv.Itoa(c.TotalLocalDeath),
			strconv.Itoa(c.TotalImportedConfirmed),
			strconv.Itoa(c.TotalImportedDischargedFromHospital),
		}
		//	分区
		r = appendByDistricts(r, districts, c.DistrictConfirmed)
		r = appendByDistricts(r, districts, c.DistrictConfirmedFromBubble)
		r = appendByDistricts(r, districts, c.DistrictConfirmedFromAsymptomatic)
		r = appendByDistricts(r, districts, c.DistrictConfirmedFromRisk)
		r = appendByDistricts(r, districts, c.DistrictAsymptomatic)
		r = appendByDistricts(r, districts, c.DistrictAsymptomaticFromBubble)
		r = appendByDistricts(r, districts, c.DistrictAsymptomaticFromRisk)
		// meta
		r = append(r, c.Source)

		records = append(records, r)
	}

	return SaveToCSV(filename, records)
}

func (cs Dailys) Find(d time.Time) *Daily {
	for i, c := range cs {
		if c.Date.Truncate(24 * time.Hour).Equal(d.Truncate(24 * time.Hour)) {
			return &cs[i]
		}
	}
	return nil
}

func (cs *Dailys) Add(c Daily) {
	lockDailys.Lock()
	*cs = append(*cs, c)
	lockDailys.Unlock()
}

func (cs *Dailys) Sort() {
	sort.SliceStable(*cs, func(i, j int) bool {
		return (*cs)[i].Date.After((*cs)[j].Date)
	})
}

type Resident struct {
	Date      time.Time // 日期
	Type      string    // 分型 （确诊病例/无症状感染者）
	Gender    string    // 性别
	Age       float64   // 年龄
	City      string    // 城市
	District  string    // 区
	Address   string    // 居住地
	Longitude float64   // 经度
	Latitude  float64   // 纬度
}

type Residents []Resident

// var lockResidents sync.Mutex

func (rs Residents) SaveToCSV(filename string) error {
	records := [][]string{}
	//	Header
	header := []string{
		"日期",
		"分型",
		"性别",
		"年龄",
		"市",
		"区",
		"居住地",
		"经度",
		"纬度",
	}

	records = append(records, header)

	for _, r := range rs {
		rec := []string{
			r.Date.Format("2006-01-02"),
			r.Type,
			r.Gender,
			strconv.FormatFloat(r.Age, 'f', 0, 32),
			r.City,
			r.District,
			r.Address,
			strconv.FormatFloat(r.Longitude, 'f', -1, 64),
			strconv.FormatFloat(r.Latitude, 'f', -1, 64),
		}
		records = append(records, rec)
	}

	return SaveToCSV(filename, records)
}

// func (rs Residents) Find(d time.Time) *Resident {
// 	for i, c := range rs {
// 		if c.Date.Truncate(24 * time.Hour).Equal(d.Truncate(24 * time.Hour)) {
// 			return &rs[i]
// 		}
// 	}
// 	return nil
// }

// func (rs *Residents) Add(r Resident) {
// 	lockResidents.Lock()
// 	*rs = append(*rs, r)
// 	lockResidents.Unlock()
// }

func (rs *Residents) Sort() {
	sort.SliceStable(*rs, func(i, j int) bool {
		l := (*rs)[i]
		r := (*rs)[j]

		if !l.Date.Equal(r.Date) {
			return l.Date.After(r.Date)
		}

		if l.District != r.District {
			return strings.Compare(l.District, r.District) < 0
		}

		if l.Address != r.Address {
			return strings.Compare(l.Address, r.Address) < 0
		}

		return l.Age < r.Age
	})
}

func SaveToCSV(filename string, records [][]string) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	return w.WriteAll(records)
}
