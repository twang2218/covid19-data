package model

import (
	"encoding/csv"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Daily struct {
	Date time.Time // 日期
	//	总共
	Asymptomatic                     int // 无症状感染者
	Confirmed                        int // 确诊病例
	Mild                             int // 轻型
	Common                           int // 普通型
	Severe                           int // 重型
	Critical                         int // 危重型
	Death                            int // 死亡
	DischargedFromHospital           int // 治愈出院
	DischargedFromMedicalObservation int // 解除医学观察
	ReleasedFromIsolation            int // 解除集中隔离

	//	本土
	LocalConfirmed                        int // 本土确诊病例
	LocalAsymptomatic                     int // 本土无症状感染者
	LocalConfirmedFromAsymptomatic        int // 从无症状感染者转归确诊的病例
	LocalConfirmedFromBubble              int // 从闭环隔离中发现的本土病例
	LocalConfirmedFromRisk                int // 从风险人群中发现的本土病例
	LocalAsymptomaticFromBubble           int // 从闭环隔离中发现的无症状感染者
	LocalAsymptomaticFromRisk             int // 从风险人群中发现的无症状感染者
	LocalDischargedFromHospital           int // 本土病例出院
	LocalDischargedFromMedicalObservation int // 本土解除医学观察
	LocalDeath                            int // 本土死亡病例

	//	境外输入
	ImportedConfirmed                        int // 境外输入病例
	ImportedAsymptomatic                     int // 境外输入无症状感染者
	ImportedDischargedFromHospital           int // 境外输入病例出院
	ImportedDischargedFromMedicalObservation int // 境外输入解除医学观察
	ImportedDeath                            int // 境外输入死亡

	//	分区
	DistrictPudong    int // 浦东新区
	DistrictXuhui     int // 徐汇区
	DistrictMinhang   int // 闵行区
	DistrictHuangpu   int // 黄浦区
	DistrictJiading   int // 嘉定区
	DistrictSongjiang int // 松江区
	DistrictHongkou   int // 虹口区
	DistrictChangning int // 长宁区
	DistrictQingpu    int // 青浦区
	DistrictJingan    int // 静安区
	DistrictBaoshan   int // 宝山区
	DistrictYangpu    int // 杨浦区
	DistrictPutuo     int // 普陀区
	DistrictChongming int // 崇明区
	DistrictJinshan   int // 金山区
	DistrictFengxian  int // 奉贤区

	// meta
	Source string // 来源
}

type Dailys []Daily

var lockDailys sync.Mutex

func (cs Dailys) SaveToCSV(filename string) error {
	records := [][]string{
		//	Header
		{
			"日期",
			//	总体
			"无症状感染者",
			"确诊病例",
			"轻型",
			"普通型",
			"重型",
			"危重型",
			"死亡",
			"治愈出院",
			"解除医学观察",
			"解除集中隔离",
			//	本土
			"本土确诊病例",
			"本土无症状感染者",
			"从无症状感染者转归确诊的病例",
			"从闭环隔离中发现的本土病例",
			"从风险人群中发现的本土病例",
			"从闭环隔离中发现的无症状感染者",
			"从风险人群中发现的无症状感染者",
			"本土病例出院",
			"本土解除医学观察",
			"本土死亡病例",
			//	境外输入
			"境外输入病例",
			"境外输入无症状感染者",
			"境外输入病例出院",
			"境外输入解除医学观察",
			"境外输入死亡",
			//	分区
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
			// meta
			"来源",
		},
	}

	for _, c := range cs {
		records = append(records, []string{
			c.Date.Format("2006-01-02"),
			//	总共
			strconv.Itoa(c.Asymptomatic),
			strconv.Itoa(c.Confirmed),
			strconv.Itoa(c.Mild),
			strconv.Itoa(c.Common),
			strconv.Itoa(c.Severe),
			strconv.Itoa(c.Critical),
			strconv.Itoa(c.Death),
			strconv.Itoa(c.DischargedFromHospital),
			strconv.Itoa(c.DischargedFromMedicalObservation),
			strconv.Itoa(c.ReleasedFromIsolation),
			//	本土
			strconv.Itoa(c.LocalConfirmed),
			strconv.Itoa(c.LocalAsymptomatic),
			strconv.Itoa(c.LocalConfirmedFromAsymptomatic),
			strconv.Itoa(c.LocalConfirmedFromBubble),
			strconv.Itoa(c.LocalConfirmedFromRisk),
			strconv.Itoa(c.LocalAsymptomaticFromBubble),
			strconv.Itoa(c.LocalAsymptomaticFromRisk),
			strconv.Itoa(c.LocalDischargedFromHospital),
			strconv.Itoa(c.LocalDischargedFromMedicalObservation),
			strconv.Itoa(c.LocalDeath),
			//	境外输入
			strconv.Itoa(c.ImportedConfirmed),
			strconv.Itoa(c.ImportedAsymptomatic),
			strconv.Itoa(c.ImportedDischargedFromHospital),
			strconv.Itoa(c.ImportedDischargedFromMedicalObservation),
			strconv.Itoa(c.ImportedDeath),
			//	各区
			strconv.Itoa(c.DistrictPudong),
			strconv.Itoa(c.DistrictXuhui),
			strconv.Itoa(c.DistrictMinhang),
			strconv.Itoa(c.DistrictHuangpu),
			strconv.Itoa(c.DistrictJiading),
			strconv.Itoa(c.DistrictSongjiang),
			strconv.Itoa(c.DistrictHongkou),
			strconv.Itoa(c.DistrictChangning),
			strconv.Itoa(c.DistrictQingpu),
			strconv.Itoa(c.DistrictJingan),
			strconv.Itoa(c.DistrictBaoshan),
			strconv.Itoa(c.DistrictYangpu),
			strconv.Itoa(c.DistrictPutuo),
			strconv.Itoa(c.DistrictChongming),
			strconv.Itoa(c.DistrictJinshan),
			strconv.Itoa(c.DistrictFengxian),
			//	meta
			c.Source,
		})
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

func SaveToCSV(filename string, records [][]string) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	w := csv.NewWriter(f)
	return w.WriteAll(records)
}
