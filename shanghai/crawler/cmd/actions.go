package main

import (
	"crawler/crawler"
	"crawler/model"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func actionCrawlDaily(c *cli.Context) error {
	var case_stats model.Dailys

	crawler := crawler.NewDailyCrawler(!c.Bool("no-cache"))
	crawler.AddOnDailyListener(func(cs model.Daily) {
		d := case_stats.Find(cs.Date)
		if d == nil {
			case_stats.Add(cs)
			if len(case_stats)%100 == 0 {
				if err := case_stats.SaveToCSV(c.String("output")); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			log.Warnf("发现重复日期：%s", cs.Date.Format("2006-01-02"))
		}
	})
	crawler.Collect()
	// bar.Finish()
	log.Infof("总共得到 %d 天疫情数据。", len(case_stats))

	//	排序
	case_stats.Sort()

	for _, cs := range case_stats {
		log.Tracef("actionCrawlDaily(): [%s] 本土 (确诊:%d {无症状=>确诊:%d, 隔离管控:%d, 其它:%d},\t 无症状: %d (隔离管控:%d, 其它:%d));\t 境外输入 (确诊:%d, 无症状: %d);\t 出院: %d（本土:%d / 境外输入:%d）;\t 解除医学观察: %d （本土:%d / 境外输入:%d）;\t 死亡: %d (本土:%d, 境外输入:%d)",
			cs.Date.Format("2006-01-02"),
			cs.LocalConfirmed,
			cs.LocalConfirmedFromAsymptomatic,
			cs.LocalConfirmedFromBubble,
			cs.LocalConfirmedFromRisk,
			cs.LocalAsymptomatic,
			cs.LocalAsymptomaticFromBubble,
			cs.LocalAsymptomaticFromRisk,
			cs.ImportedConfirmed,
			cs.ImportedAsymptomatic,
			cs.DischargedFromHospital,
			cs.LocalDischargedFromHospital,
			cs.ImportedDischargedFromHospital,
			cs.DischargedFromMedicalObservation,
			cs.LocalDischargedFromMedicalObservation,
			cs.ImportedDischargedFromMedicalObservation,
			cs.Death,
			cs.LocalDeath,
			cs.ImportedDeath,
		)
		// log.Tracef("actionCrawlDaily(): [%s] \t => 无症状：%d, \t 区域无症状： {%v}",
		// 	cs.Date.Format("2006-01-02"),
		// 	cs.LocalAsymptomatic,
		// 	cs.DistrictAsymptomatic,
		// )
		// log.Tracef("actionCrawlDaily(): [%s] \t => 确诊病例：%d, \t 区域确诊病例： {%v}",
		// 	cs.Date.Format("2006-01-02"),
		// 	cs.LocalConfirmed,
		// 	cs.DistrictConfirmed,
		// )
	}

	//	将最终结果写入 JSON
	if err := case_stats.SaveToCSV(c.String("output")); err != nil {
		return err
	}

	return nil
}
