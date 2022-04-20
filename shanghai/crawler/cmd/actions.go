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
	// bar := pb.StartNew(10)
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
		//	update progress bar
		// bar.SetTotal(int64(crawler.PageTotal))
		// bar.SetCurrent(int64(crawler.PageVisited))
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
	}

	//	将最终结果写入 JSON
	if err := case_stats.SaveToCSV(c.String("output")); err != nil {
		return err
	}

	return nil
}

// func actionCrawlLocations(c *cli.Context) error {
// 	var locations model.Locations

// 	crawler := crawler.NewLocationsCrawler()
// 	// bar := pb.StartNew(10)
// 	crawler.AddOnHerbListener(func(l model.Location) {
// 		locations.Add(l)
// 		if len(locations)%100 == 0 {
// 			if err := model.SaveToJSON(c.String("output"), locations); err != nil {
// 				log.Fatal(err)
// 			}
// 		}
// 		//	update progress bar
// 		// bar.SetTotal(int64(crawler.PageTotal))
// 		// bar.SetCurrent(int64(crawler.PageVisited))
// 	})
// 	crawler.Collect()
// 	// bar.Finish()
// 	log.Infof("总共得到 %d 天疫情数据。", len(locations))

// 	//	排序
// 	locations.Sort()

// 	//	将最终结果写入 JSON
// 	if err := model.SaveToJSON(c.String("output"), locations); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func actionCrawlPrescription(c *cli.Context) error {
// 	var pres model.Prescriptions

// 	crawler := crawler.NewPrescriptionCrawler()
// 	bar := pb.StartNew(10)
// 	crawler.AddOnPrescriptionListener(func(p model.Prescription) {
// 		pres.Add(p)
// 		// log.Debug(".")
// 		// // log.Infof("【%s】 %s", p.Name, p.Link)
// 		if len(pres)%100 == 0 {
// 			log.Infof("-----========< 已下载 %d 中药方剂 (%d / %d) >=======------", len(pres), crawler.PageVisited, crawler.PageTotal)
// 			if err := model.SaveToJSON(c.String("output"), pres); err != nil {
// 				log.Fatal(err)
// 			}
// 			// log.Debugf("分项列表：%+v", crawler.keys)
// 		}
// 		//	update progress bar
// 		bar.SetTotal(int64(crawler.PageTotal))
// 		bar.SetCurrent(int64(crawler.PageVisited))
// 	})

// 	//	如果存在药材数据库，则添加药材数据，用以处理处方信息
// 	var herbs model.Herbs
// 	if err := herbs.LoadFromJSON(c.String("herb")); err != nil {
// 		log.Warnf("无法加载药材数据文件，处方信息将不被处理。 '%s'", err.Error())
// 	} else {
// 		crawler.SetHerbs(herbs)
// 	}

// 	crawler.Collect()
// 	bar.Finish()
// 	log.Infof("总共得到 %d 个中药方剂。", len(pres))

// 	//	排序
// 	pres.AssignID()

// 	//	将最终结果写入 JSON
// 	if err := model.SaveToJSON(c.String("output"), pres); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func actionStats(c *cli.Context) error {
// 	//	加载药材数据
// 	var herbs model.Herbs
// 	if err := herbs.LoadFromJSON(c.String("herb")); err != nil {
// 		return fmt.Errorf("无法加载数据文件。'%s'", err.Error())
// 	}
// 	log.Infof("读取了 %d 味中药药材。", len(herbs))

// 	//	加载方剂数据
// 	var pres model.Prescriptions
// 	if err := pres.LoadFromJSON(c.String("prescription")); err != nil {
// 		return fmt.Errorf("无法加载数据文件。'%s'", err.Error())
// 	}
// 	log.Infof("读取了 %d 个中药方剂。", len(pres))

// 	processor.Statistics(herbs, pres)

// 	return nil
// }

// func actionProcess(c *cli.Context) error {
// 	//	加载药材数据
// 	var herbs model.Herbs
// 	if err := herbs.LoadFromJSON(c.String("herb")); err != nil {
// 		return fmt.Errorf("无法加载数据文件。'%s'", err.Error())
// 	}
// 	log.Infof("读取了 %d 味中药药材。", len(herbs))

// 	//	加载方剂数据
// 	var pres model.Prescriptions
// 	if err := pres.LoadFromJSON(c.String("prescription")); err != nil {
// 		return fmt.Errorf("无法加载数据文件。'%s'", err.Error())
// 	}
// 	log.Infof("读取了 %d 个中药方剂。", len(pres))

// 	processor.Process(herbs, pres)

// 	return nil
// }
