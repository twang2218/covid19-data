package main

import (
	"crawler/crawler"
	"crawler/model"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func actionCrawlDaily(c *cli.Context) error {
	var ds model.Dailys
	var rs model.Residents
	rss := make(map[time.Time]int, 0)
	var lockR sync.Mutex

	crawler := crawler.NewDailyCrawler(!c.Bool("no-cache"))
	crawler.AddOnDailyListener(func(cs model.Daily) {
		d := ds.Find(cs.Date)
		if d == nil {
			ds.Add(cs)
			if len(ds)%100 == 0 {
				if err := ds.SaveToCSV(c.String("daily")); err != nil {
					log.Fatal(fmt.Errorf("无法写入文件 %q: %s", c.String("daily"), err))
				}
			}
		} else {
			log.Warnf("发现重复日期：%s", cs.Date.Format("2006-01-02"))
		}
	})
	crawler.AddOnResidentsListener(func(r model.Residents) {
		// log.Infof("%d + %d", len(rs), len(r))
		lockR.Lock()
		rs = append(rs, r...)
		for _, t := range r {
			if val, ok := rss[t.Date]; ok {
				rss[t.Date] = val + 1
			} else {
				rss[t.Date] = 1
			}
		}
		lockR.Unlock()
	})
	crawler.Collect()
	// bar.Finish()
	log.Infof("总共得到 %d 天疫情数据。", len(ds))

	//	排序
	ds.Sort()

	for _, d := range ds {
		log.Tracef("actionCrawlDaily(): [%s] 本土 (确诊:%d {无症状=>确诊:%d, 隔离管控:%d, 其它:%d},\t 无症状: %d (隔离管控:%d, 其它:%d));\t 境外输入 (确诊:%d, 无症状: %d);\t 出院: %d（本土:%d / 境外输入:%d）;\t 解除医学观察: %d （本土:%d / 境外输入:%d）;\t 死亡: %d (本土:%d, 境外输入:%d); \t %d 居住地信息",
			d.Date.Format("2006-01-02"),
			d.LocalConfirmed,
			d.LocalConfirmedFromAsymptomatic,
			d.LocalConfirmedFromBubble,
			d.LocalConfirmedFromRisk,
			d.LocalAsymptomatic,
			d.LocalAsymptomaticFromBubble,
			d.LocalAsymptomaticFromRisk,
			d.ImportedConfirmed,
			d.ImportedAsymptomatic,
			d.DischargedFromHospital,
			d.LocalDischargedFromHospital,
			d.ImportedDischargedFromHospital,
			d.DischargedFromMedicalObservation,
			d.LocalDischargedFromMedicalObservation,
			d.ImportedDischargedFromMedicalObservation,
			d.Death,
			d.LocalDeath,
			d.ImportedDeath,
			rss[d.Date],
		)
		// log.Tracef("actionCrawlDaily(): [%s] \t => 无症状：%d, \t 区域无症状： {%v}",
		// 	d.Date.Format("2006-01-02"),
		// 	d.LocalAsymptomatic,
		// 	d.DistrictAsymptomatic,
		// )
		// log.Tracef("actionCrawlDaily(): [%s] \t => 确诊病例：%d, \t 区域确诊病例： {%v}",
		// 	d.Date.Format("2006-01-02"),
		// 	d.LocalConfirmed,
		// 	d.DistrictConfirmed,
		// )
	}

	//	将最终结果写入 JSON
	if err := ds.SaveToCSV(c.String("daily")); err != nil {
		return fmt.Errorf("无法写入文件 %q: %s", c.String("daily"), err)
	}

	// rs.Sort()
	if err := rs.SaveToCSV(c.String("residents")); err != nil {
		return fmt.Errorf("无法写入文件 %q: %s", c.String("residents"), err)
	}

	return nil
}
