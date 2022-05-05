package main

import (
	"crawler/crawler"
	"crawler/geocoder"
	"crawler/model"
	"fmt"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pmezard/go-difflib/difflib"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

var gc_count int

func consume(gc *geocoder.Geocoder, rs *model.Residents, stats *map[time.Time]int, in chan model.Resident) {
	for r := range in {
		//	地理编码
		if gc != nil {
			s := fmt.Sprintf("%s%s%s", r.City, r.District, r.Address)
			addr, err := gc.Geocode(s)
			if err != nil {
				log.Warnf("解析地址 %q 失败：%s", s, err)
			} else {
				r.Longitude = addr.Longitude
				r.Latitude = addr.Latitude
				gc_count += 1
			}
		}
		if gc_count%100000 == 0 {
			fmt.Println()
		} else if gc_count%10000 == 0 {
			fmt.Print(":")
		} else if gc_count%1000 == 0 {
			fmt.Print(".")
		}
		//	追加
		*rs = append(*rs, r)
		//	统计
		if val, ok := (*stats)[r.Date]; ok {
			(*stats)[r.Date] = val + 1
		} else {
			(*stats)[r.Date] = 1
		}
	}

	fmt.Println()
}

var spewConfig = spew.ConfigState{
	Indent:                  "  ",
	DisablePointerAddresses: true,
	DisableCapacities:       true,
	SortKeys:                true,
	DisableMethods:          true,
	MaxDepth:                10,
}

func diff[T any](old, fresh T) string {
	o := spewConfig.Sdump(old)
	f := spewConfig.Sdump(fresh)
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(o),
		B:        difflib.SplitLines(f),
		FromFile: "旧数据",
		FromDate: "",
		ToFile:   "新抓取",
		ToDate:   "",
		Context:  0,
	})
	if err != nil {
		log.Warnf("diff(): 出错：%s", err)
	}
	return diff
}

func updateDailys(old, fresh model.Dailys) model.Dailys {
	//	将新数据添加到旧的表中
	for _, fd := range fresh {
		found := false
		for _, od := range old {
			if od.Date.Equal(fd.Date) {
				// 存在一样的数据，则比对一致性
				if !assert.ObjectsAreEqualValues(od, fd) {
					//	同样的天，数据却不同
					log.Warnf("[%s] 数据不一致：\n%s", od.Date.Format("2006-01-02"), diff(od, fd))
				}
				found = true
				break
			}
		}
		//	这是新的数据，添加到旧表中
		if !found {
			old = append(old, fd)
		}
	}
	//	整理数据
	old.Sort()

	return old
}

func updateResidents(old, fresh model.Residents) model.Residents {
	//	将新数据添加到旧的表中
	for _, fd := range fresh {
		found := false
		for _, od := range old {
			if od.Date.Equal(fd.Date) && od.Name == fd.Name && od.District == fd.District && od.Address == fd.Address {
				// 存在一样的数据，则比对一致性
				if !assert.ObjectsAreEqualValues(od, fd) {
					//	同样的天，数据却不同
					log.Warnf("[%s - %s] 数据不一致：\n%s", od.Date.Format("2006-01-02"), od.Name, diff(od, fd))
				}
				found = true
				break
			}
		}
		//	这是新的数据，添加到旧表中
		if !found {
			old = append(old, fd)
		}
	}
	//	整理数据
	old.Sort()

	return old
}

func actionCrawlDaily(c *cli.Context) error {
	var ds model.Dailys
	var rs model.Residents
	var ds_old model.Dailys
	var rs_old model.Residents

	city := c.String("city")
	file_daily := strings.ReplaceAll(c.String("daily"), "{city}", city)
	file_daily_csv := file_daily + ".csv"
	file_daily_json := file_daily + ".json"
	file_residents := strings.ReplaceAll(c.String("residents"), "{city}", city)
	file_residents_csv := file_residents + ".csv"
	file_residents_json := file_residents + ".json"

	ds_old.LoadFromJSON(file_daily_json)
	rs_old.LoadFromJSON(file_residents_json)

	var districts []string
	switch city {
	case "beijing":
		districts = crawler.DailyParserBeijing{}.GetDistricts()
	case "shanghai":
		districts = crawler.DailyParserShanghai{}.GetDistricts()
	}

	stats := make(map[time.Time]int, 0)
	ch := make(chan model.Resident)

	// log.Tracef("geo_cache: %q, web_cache: %q", c.String("geo_cache"), c.String("web_cache"))

	gc := geocoder.NewGeocoderBaidu(c.String("key_baidu_map"), c.String("geo_cache"))
	defer gc.Close()

	go consume(&gc, &rs, &stats, ch)

	var web_cache string
	if !c.Bool("no-cache") {
		web_cache = c.String("web_cache")
	}
	crawler := crawler.NewDailyCrawler(c.String("city"), web_cache)
	crawler.AddOnDailyListener(func(cs model.Daily) {
		d := ds.Find(cs.Date)
		if d == nil {
			ds.Add(cs)
			if len(ds_old) == 0 {
				//	只在第一次下载数据文件的时候才进行数据暂存。
				//	因为只有第一次下载出错几率最高，而且不完整下载不应该覆盖历史数据。
				if len(ds)%100 == 0 {
					if err := ds.SaveToJSON(file_daily_json); err != nil {
						log.Fatal(fmt.Errorf("无法写入文件(daily) %q: %s", file_daily, err))
					}
				}
			}
		} else {
			log.Warnf("发现重复日期：%s", cs.Date.Format("2006-01-02"))
		}
	})
	crawler.AddOnResidentsListener(func(rs2 model.Residents) {
		// log.Infof("%d + %d", len(rs), len(r))
		for _, r := range rs2 {
			// 送给数据处理通道
			ch <- r
			// 中间保存
			if len(rs_old) == 0 {
				//	只在第一次下载数据文件的时候才进行数据暂存。
				//	因为只有第一次下载出错几率最高，而且不完整下载不应该覆盖历史数据。
				if len(rs)%100 == 0 {
					if err := rs.SaveToJSON(file_residents_json); err != nil {
						log.Fatal(fmt.Errorf("无法写入文件(residents) %q: %s", file_residents_json, err))
					}
				}
			}
		}
	})
	crawler.Collect()

	//	爬虫结束
	close(ch)

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
			stats[d.Date],
			// d.LocalConfirmed+d.LocalAsymptomatic-stats[d.Date],
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

	//	用新的数据更新旧的，以增加新的数据，但是要检查旧数据是否有所改动
	ds = updateDailys(ds_old, ds)
	rs = updateResidents(rs_old, rs)

	//	将最终结果写入文件
	if err := ds.SaveToCSV(file_daily_csv, districts); err != nil {
		return fmt.Errorf("无法写入文件(daily) %q: %s", file_daily_csv, err)
	}

	if err := ds.SaveToJSON(file_daily_json); err != nil {
		return fmt.Errorf("无法写入文件(daily) %q: %s", file_daily_json, err)
	}

	rs.Sort()
	if err := rs.SaveToCSV(file_residents_csv); err != nil {
		return fmt.Errorf("无法写入文件(resident) %q: %s", file_residents_csv, err)
	}
	if err := rs.SaveToJSON(file_residents_json); err != nil {
		return fmt.Errorf("无法写入文件(resident) %q: %s", file_residents_json, err)
	}

	return nil
}
