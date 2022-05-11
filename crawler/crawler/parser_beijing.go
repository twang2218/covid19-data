package crawler

import (
	"crawler/model"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type DailyParserBeijing struct{}

func (p DailyParserBeijing) GetSelector(t string) string {
	selectors := map[string]string{
		"index":   ".listLk a",
		"item":    ".article0",
		"title":   ".articleTitle",
		"content": ".article p",
	}

	if val, ok := selectors[t]; ok {
		return val
	} else {
		return ""
	}
}

func (p DailyParserBeijing) GetItemLinks() []string {
	return []string{
		// "https://mp.weixin.qq.com/s/agdZHOqVZh9atNHOQEFTog",
	}
}

func (p DailyParserBeijing) GetIndexLinks() []string {
	const (
		LINK_DAILY_0 string = "http://wjw.beijing.gov.cn/wjwh/ztzl/xxgzbd/gzbdyqtb/index.html"
		LINK_DAILY_1 string = "http://wjw.beijing.gov.cn/wjwh/ztzl/xxgzbd/gzbdyqtb/index{page}.html"
		MAX_PAGES    int    = 5
	)

	links := []string{}

	for i := 0; i < MAX_PAGES; i++ {
		//	首页>>卫生健康文化>>专题专栏>>新型冠状病毒感染的肺炎疫情防控>>疫情通报
		page := ""
		if i > 0 {
			page = fmt.Sprintf("_%d", i)
		}
		link := strings.ReplaceAll(LINK_DAILY_1, "{page}", page)

		//	访问
		links = append(links, link)
	}

	return links
}

func (p DailyParserBeijing) GetDistricts() []string {
	return []string{
		"朝阳区", "东城区", "西城区", "海淀区", "房山区", "丰台区", "石景山区", "门头沟区",
		"大兴区", "通州区", "顺义区", "昌平区", "怀柔区", "平谷区", "密云区", "延庆区",
	}
}

func (p DailyParserBeijing) IsDaily(date time.Time, title string) bool {
	return p.IsValidTitle(title)
}

func (p DailyParserBeijing) IsResidents(date time.Time, title string) bool {
	//	本轮疫情开始于 2022-04-15 以后
	return p.IsValidTitle(title) && date.After(time.Date(2022, 4, 15, 0, 0, 0, 0, time.Local))
}

func (p DailyParserBeijing) IsValidTitle(title string) bool {
	return strings.Contains(title, "日新增") || strings.Contains(title, "日无新增")
}

//	解析 Daily

//	解析 Daily 标题
func (p DailyParserBeijing) ParseDailyTitle(d *model.Daily, title string) error {
	if d == nil {
		return fmt.Errorf("输入对象为空")
	}

	var m []string
	var err error

	// 标题日期
	m = reDailyDate.FindStringSubmatch(title)
	if m == nil {
		// return fmt.Errorf("[%s] 无法解析文章标题中日期：%q", d.Date.Format("2006-01-02"), title)
	} else {
		s := m[1]
		if !strings.Contains(s, "年") {
			s = fmt.Sprintf("2022年%s", m[1])
		}
		d.Date, err = time.Parse("2006年1月2日", s)
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中日期：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// if p.IsResidents(d.Date, title) {
	// 	// 这里面只包含居住地信息，后面不需要解析
	// 	// log.Infof("[%s] 居住地信息：%q", cs.Date.Format("2006-01-02"), title)
	// 	return nil
	// }

	// 本土新增
	m = reDailyLocalConfirmed.FindStringSubmatch(title)
	if m == nil {
		// return fmt.Errorf("[%s] 无法解析文章标题中本土新增：%q", d.Date.Format("2006-01-02"), title)
	} else {
		///	2种情况
		n := m[1] + m[2]
		d.LocalConfirmed, err = strconv.Atoi(n)
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中本土新增：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 本土无症状
	m = reDailyLocalAsymptomatic.FindStringSubmatch(title)
	if m == nil || !strings.Contains(m[0], "无症状") {
		// return fmt.Errorf("[%s] 无法解析文章标题中本土无症状感染者：%q", d.Date.Format("2006-01-02"), title)
	} else {
		// 有3种情况
		n := m[1] + m[2] + m[3]
		d.LocalAsymptomatic, err = strconv.Atoi(n)
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中本土无症状感染者：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 境外输入新增
	m = reDailyImportedConfirmed.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中境外输入新增：%q", d.Date.Format("2006-01-02"), title)
	} else {
		//	数值有两个case
		n := m[1] + m[2]
		d.ImportedConfirmed, err = strconv.Atoi(n)
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中境外输入新增：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 境外输入无症状
	m = reDailyImportedAsymptomatic.FindStringSubmatch(title)
	if m == nil || !strings.Contains(m[0], "无症状") {
		// log.Warnf("[%s] 无法解析文章标题中境外输入无症状感染者：%q", d.Date.Format("2006-01-02"), title)
	} else {
		/// 正则包含4个可能性
		n := m[1] + m[2] + m[3] + m[4]
		d.ImportedAsymptomatic, err = strconv.Atoi(n)
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中境外输入无症状感染者：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 治愈出院
	m = reDailyDischargedFromHospital.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中治愈出院：%q", d.Date.Format("2006-01-02"), title)
	} else {
		d.DischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中治愈出院：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 解除医学观察
	m = reDailyDischargedFromMedicalObservation.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中解除医学观察：%q", d.Date.Format("2006-01-02"), title)
	} else {
		d.DischargedFromMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中解除医学观察：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	return nil
}

// 解析 Daily 内容
func (p DailyParserBeijing) ParseDailyContent(d *model.Daily, content string) error {
	if d == nil {
		return fmt.Errorf("输入对象为空")
	}

	var m []string
	var err error
	// 修正一些格式上的问题
	content = strings.ReplaceAll(content, "区，\n", "区，")
	content = strings.ReplaceAll(content, "居住于\n", "居住于")
	content = strings.ReplaceAll(content, "无症状\n", "无症状")
	content = strings.ReplaceAll(content, "无症状感染\n", "无症状感染")
	content = strings.ReplaceAll(content, "无症状感染者\n", "无症状感染者")

	// 日期 (补充标题缺失)
	if time.Time.IsZero(d.Date) {
		m = reDailyDate.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中日期：%q", d.Date.Format("2006-01-02"), content)
		} else {
			s := m[1]
			if !strings.Contains(s, "年") {
				s = "2022年" + s
			}
			d.Date, err = time.Parse("2006年1月2日", s)
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中日期：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// log.Tracef("[%s] 解析文章内容：(%d)", d.Date.Format("2006-01-02"), len(content))

	// 本土阳性感染者 (补充标题缺失)
	if d.LocalPositive == 0 {
		m = reDailyLocalPositive.FindStringSubmatch(content)
		if m == nil {
			//	可能没有阳性数据
			// fmt.Printf("[%s] 无法解析文章内容中本土阳性感染者：%q", d.Date.Format("2006-01-02"), content)
		} else {
			d.LocalPositive, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中本土阳性感染者：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	if d.LocalPositive > 0 && d.Mild == 0 {
		m = reDailyMild.FindStringSubmatch(content)
		if m == nil {
			fmt.Printf("[%s] 无法解析文章内容中本土轻型：%q", d.Date.Format("2006-01-02"), content)
		} else {
			d.Mild, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中本土轻型：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	if d.LocalPositive > 0 && d.Common == 0 {
		m = reDailyCommon.FindStringSubmatch(content)
		if m == nil {
			fmt.Printf("[%s] 无法解析文章内容中本土普通型：%q", d.Date.Format("2006-01-02"), content)
		} else {
			d.Common, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中本土普通型：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 本土新增 (补充标题缺失)
	if d.LocalConfirmed == 0 {
		m = reDailyLocalConfirmed.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中本土新增：%q", d.Date.Format("2006-01-02"), content)
		} else {
			///	2种情况
			n := m[1] + m[2]
			d.LocalConfirmed, err = strconv.Atoi(n)
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中本土新增：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 本土无症状 (补充标题缺失)
	if d.LocalAsymptomatic == 0 {
		m = reDailyLocalAsymptomatic.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中本土无症状：%q", d.Date.Format("2006-01-02"), content)
		} else {
			// 有3种情况
			n := m[1] + m[2] + m[3]
			d.LocalAsymptomatic, err = strconv.Atoi(n)
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中本土无症状：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 境外输入确诊 (补充标题缺失)
	if d.ImportedConfirmed == 0 {
		m = reDailyImportedConfirmed.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中境外输入确诊：%q", d.Date.Format("2006-01-02"), content)
		} else {
			//	数值有两个case
			n := m[1] + m[2]
			d.ImportedConfirmed, err = strconv.Atoi(n)
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中境外输入确诊：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 境外输入无症状 (补充标题缺失)
	if d.ImportedAsymptomatic == 0 {
		m = reDailyImportedAsymptomatic2.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中境外输入无症状：%q", d.Date.Format("2006-01-02"), content)
		} else {
			d.ImportedAsymptomatic, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中境外输入无症状：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 本土隔离管控中发现的阳性感染者
	m = reDailyLocalPositiveFromBubble.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土隔离管控中发现的阳性感染者：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalPositiveFromBubble, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土隔离管控中发现的阳性感染者：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土风险人群中发现的阳性感染者
	m = reDailyLocalPositiveFromRisk.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土风险人群中发现的阳性感染者：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalPositiveFromRisk, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土风险人群中发现的阳性感染者：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土无症状转为确诊病例
	m = reDailyLocalConfirmedFromAsymptomatic.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土无症状转为确诊病例：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalConfirmedFromAsymptomatic, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土无症状转为确诊病例：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土隔离管控中发现的确诊病例
	m = reDailyLocalConfirmedFromBubble.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土隔离管控中发现的确诊病例：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalConfirmedFromBubble, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土隔离管控中发现的确诊病例：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土隔离管控中发现的无症状病例
	m = reDailyLocalAsymptomaticFromBubble.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土隔离管控中发现的无症状病例：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalAsymptomaticFromBubble, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土隔离管控中发现的无症状病例：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 治愈出院 （补充标题缺失）
	m = reDailyDischargedFromHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中治愈出院：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.DischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章内容中治愈出院：%q", d.Date.Format("2006-01-02"), content)
		}
	}
	// 本土治愈出院
	m = reDailyLocalDischargedFromHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土治愈出院：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalDischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土治愈出院：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 境外输入治愈出院
	m = reDailyImportedDischargedFromHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入治愈出院：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.ImportedDischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入治愈出院：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 解除医学观察 (补充标题缺失)
	if d.DischargedFromMedicalObservation == 0 {
		m = reDailyDischargedFromMedicalObservation2.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中解除医学观察：%q", d.Date.Format("2006-01-02"), content)
		} else {
			d.DischargedFromMedicalObservation, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中解除医学观察：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 本土解除医学观察
	m = reDailyLocalDischargedFromMedicalObservation.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土解除医学观察：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalDischargedFromMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土解除医学观察：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 境外输入解除医学观察
	m = reDailyImportedDischargedFromMedicalObservation.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入解除医学观察：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.ImportedDischargedFromMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入解除医学观察：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土死亡病例
	m = reDailyLocalDeath.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土死亡病例：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalDeath, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土死亡病例：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 境外输入死亡病例
	m = reDailyImportedDeath.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入死亡病例：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.ImportedDeath, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入死亡病例：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 累计本土确诊
	m = reDailyTotalLocalConfirmed.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中累计本土确诊：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.TotalLocalConfirmed, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中累计本土确诊：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 累计本土治愈出院
	m = reDailyTotalLocalDischargedFromHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中累计本土治愈出院：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.TotalLocalDischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中累计本土治愈出院：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土在院治疗
	m = reDailyLocalInHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土在院治疗：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.CurrentLocalInHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土在院治疗：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 累计境外输入确诊
	m = reDailyTotalImportedConfirmed.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中累计境外输入确诊：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.TotalImportedConfirmed, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中累计境外输入确诊：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 累计境外输入治愈出院
	m = reDailyTotalImportedDischargedFromHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中累计境外输入治愈出院：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.TotalImportedDischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中累计境外输入治愈出院：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 境外输入在院治疗
	m = reDailyImportedInHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入在院治疗：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.CurrentImportedInHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入在院治疗：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 累计本土死亡
	m = reDailyTotalLocalDeath.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中累计本土死亡：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.TotalLocalDeath, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中累计本土死亡：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 重型
	m = reDailySevere.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中重型：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.Severe, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中重型：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 危重型
	m = reDailyCritical.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中危重型：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.Critical, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中危重型：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 尚在医学观察
	m = reDailyUnderMedicalObservation.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中尚在医学观察：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.UnderMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中尚在医学观察：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土尚在医学观察
	m = reDailyLocalUnderMedicalObservation.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土尚在医学观察：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.LocalUnderMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土尚在医学观察：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 境外输入尚在医学观察
	m = reDailyImportedUnderMedicalObservation.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入尚在医学观察：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.ImportedUnderMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入尚在医学观察：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	//	进一步解析城区信息
	return p.parseDailyContentRegion(d, content)
}

var (
	patternCaseBeijing       = `(?P<district>[^，；。、]{2,3}区(?:[和、].{2,3}区)*)各?(?P<number>\d+)例(?:[，；。、])`
	reDailyRegionListBeijing = regexp.MustCompile(`^[^累计]+(?:` + patternCaseBeijing + `)+`)
	reDailyRegionItemBeijing = regexp.MustCompile(patternCaseBeijing)
)

func (p DailyParserBeijing) parseDailyContentRegion(d *model.Daily, content string) error {
	var mm [][]string
	var err error

	//	隔离管控 => 确诊病例
	mm = reDailyRegionConfirmedFromBubble.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalConfirmedFromBubble > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区确诊病例(来自隔离管控)：%q", d.Date.Format("2006-01-02"), content)
		}
	} else {
		d.DistrictConfirmedFromBubble = p.parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区确诊病例(来自隔离管控): %#v", cs.Date.Format("2006-01-02"), cs.DistrictConfirmedFromBubble)
	}
	//	风险人群 => 确诊病例
	mm = reDailyRegionConfirmedFromRisk.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalConfirmedFromRisk > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区确诊病例(来自风险人群)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictConfirmedFromRisk = p.parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区确诊病例(来自风险人群): %#v", cs.Date.Format("2006-01-02"), cs.DistrictConfirmedFromRisk)
	}
	//	无症状 => 确诊病例
	mm = reDailyRegionConfirmedFromAsymptomatic.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalConfirmedFromAsymptomatic > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区确诊病例(来自无症状感染者)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictConfirmedFromAsymptomatic = p.parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区确诊病例(来自无症状感染者): %#v", cs.Date.Format("2006-01-02"), cs.DistrictConfirmedFromAsymptomatic)
	}
	//	隔离管控 => 无症状
	mm = reDailyRegionAsymptomaticFromBubble.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalAsymptomaticFromBubble > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区无症状感染者(来自隔离管控)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictAsymptomaticFromBubble = p.parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区无症状感染者(来自隔离管控): %#v", cs.Date.Format("2006-01-02"), cs.DistrictAsymptomaticFromBubble)
	}
	//	风险人群 => 无症状
	mm = reDailyRegionAsymptomaticFromRisk.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalAsymptomaticFromRisk > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区无症状感染者(来自风险人群)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictAsymptomaticFromRisk = p.parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区无症状感染者(来自风险人群): %#v", cs.Date.Format("2006-01-02"), cs.DistrictAsymptomaticFromRisk)
	}
	//	区域阳性感染者
	mm = reDailyRegionListBeijing.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalPositive > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区阳性感染者(来自风险人群)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictPositive = p.parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区阳性感染者(来自风险人群): %#v", cs.Date.Format("2006-01-02"), cs.DistrictPositive)
	}

	return err
}

func (p DailyParserBeijing) parseDailyContentRegionItems(d *model.Daily, mm [][]string) map[string]int {
	dict := make(map[string]int)
	for _, m := range mm {
		items := p.parseDailyContentRegionItem(d, m[0])
		for k, v := range items {
			if val, ok := dict[k]; ok {
				dict[k] = val + v
			} else {
				dict[k] = v
			}
		}
	}
	return dict
}

func (p DailyParserBeijing) parseDailyContentRegionItem(d *model.Daily, text string) map[string]int {
	item := make(map[string]int)
	m := reDailyRegionItemBeijing.FindAllStringSubmatch(text, -1)
	if m == nil {
		log.Warnf("[%s] 无法解析区域病例列表：%q", d.Date.Format("2006-01-02"), text)
	} else {
		for _, it := range m {
			var v int
			var err error

			if len(it[2]) > 0 {
				v, err = strconv.Atoi(it[2])
				if err != nil {
					log.Warnf("[%s] 无法解析区域病例列表：%q", d.Date.Format("2006-01-02"), it)
					continue
				}
			}
			regions := strings.Split(it[1], "和")
			for _, r := range regions {
				r = strings.TrimSpace(r)
				if len(r) > 0 {
					item[r] = v
				}
			}
		}
	}
	return item
}

var (
	reResidentDistrictBeijingSection = regexp.MustCompile(`(?:确诊病例|[无症状]*感染者)(?P<num_list>[\d\s、至]*)：[^\n]+`)
	reResidentDistrictBeijing1       = regexp.MustCompile(`(?:确诊病例|[无症状]*感染者)(?P<num_list>[\d\s、至]*)[：]?[^\n]*(?:现住址?|进入[地点]*|均为)均?[位于为]*(?P<district>[^区\n]{2,4}区)(?P<address>[^，。 \n]*)[^\n]*(?:(?P<date>\d+\s*月\d+\s*日|当日)均?诊断均?为(?P<type>确诊病例|无症状感染者))(?:[，]临床分型[均分别]*为(?P<level>(?:[^型，\n]+型)+))?`)
	reResidentDistrictBeijing1Level  = regexp.MustCompile(`(?P<date>\d+月\d+日)?(?:确诊病例|[无症状]*感染者)(?P<num_list>[\d\s、至]+)(?:诊断为确诊病例，)?临床分型均?为(?P<level>[^型]+型)`)
)

func (p DailyParserBeijing) ParseResidents(rs *model.Residents, date time.Time, content string) error {
	if rs == nil {
		return fmt.Errorf("输入对象为空")
	}

	var mm [][]string
	var err error

	// if date.Format("2006-01-02") == "2022-04-28" {
	// 	fmt.Printf("[%s] <<<<content>>>>\n%s\n<<<<<<<>>>>>>>>\n", date.Format("2006-01-02"), content)
	// 	// fmt.Printf("[%s] > %q\n", date.Format("2006-01-02"), m[0])
	// }

	// 提取分区信息
	mm = reResidentDistrictBeijingSection.FindAllStringSubmatch(content, -1)
	// log.Tracef("ParseResidents(): [%s] %#v", date.Format("2006-01-02"), len(mm))
	for _, m1 := range mm {
		m2 := reResidentDistrictBeijing1.FindStringSubmatch(m1[0])
		if len(m2) == 7 {
			// 解析日期
			var d time.Time
			if len(m2[4]) > 0 && m2[4] != "当日" {
				s := m2[4]
				if !strings.Contains(s, "年") {
					s = fmt.Sprintf("2022年%s", s)
				}
				var err error
				d, err = time.Parse("2006年1月2日", strings.ReplaceAll(s, " ", ""))
				if err != nil {
					log.Warnf("无法解析日期: %s => %s", m2[0], err)
				}
			} else {
				d = date
			}
			// 解析病例编号
			num_list := p.parseNumberList(m2[1])
			if len(num_list) == 0 {
				num_list = append(num_list, 1)
			}

			// fmt.Printf("num_list: %v\n", num_list)

			// 解析分型
			var t string
			if m2[5] == "无症状感染者" {
				t = m2[5]
			} else {
				t = m2[6]
			}
			var ts []string
			if len(t) == 0 {
				// 确诊病例，但是未能解析出临床分型。可能是由不同分型所致，因此需要独立解析。
				mm3 := reResidentDistrictBeijing1Level.FindAllStringSubmatch(m1[0], -1)
				for _, m3 := range mm3 {
					//	解析每一群分型病例号
					sub_list := p.parseNumberList(m3[2])
					for range sub_list {
						ts = append(ts, m3[3])
					}
				}
			} else {
				if strings.Contains(t, "、") {
					//	对于分型列表，直接拆
					ts = append(ts, strings.Split(t, "、")...)
				} else {
					//	对于单一类型，重复多次
					for range num_list {
						ts = append(ts, t)
					}
					// fmt.Printf("num_list: %#v; ts: %#v\n", num_list, ts)
				}
			}
			//	TODO: should be removed if parse correctly
			for len(ts) < len(num_list) {
				ts = append(ts, "")
			}
			// fmt.Printf("m2: %#v\n", m2)
			for i, n := range num_list {
				r := model.Resident{
					Date:     d,
					Name:     fmt.Sprintf("%s%d", m2[5], n),
					Type:     ts[i],
					City:     "北京市",
					District: strings.TrimSpace(m2[2]),
					Address:  strings.TrimSpace(m2[3]),
				}
				// log.Tracef("[%s] %v", date.Format("2006-01-02"), r)
				*rs = append(*rs, r)
			}
		} else {
			//	如果是“到达北京首都机场”这类，不在社区中，所以不必报错。
			if !strings.Contains(m1[0], "机场") {
				log.Warnf("[%s] 无法解析居住地信息：%q => %#v", date.Format("2006-01-02"), m1[0], m2)
			}
		}
	}

	return err
}

func (p DailyParserBeijing) parseNumberList(text string) []int {
	num_items := strings.Split(text, "、")
	var num_list []int
	for _, item := range num_items {
		n, err := strconv.Atoi(strings.TrimSpace(item))
		if err == nil {
			// 纯数字
			num_list = append(num_list, n)
		} else {
			range_item := strings.Split(item, "至")
			if len(range_item) == 2 {
				// "123至456"
				begin, err := strconv.Atoi(strings.TrimSpace(range_item[0]))
				if err != nil {
					log.Warnf("无法解析病例序号列表：%s => %s", text, err)
					continue
				}
				end, err := strconv.Atoi(strings.TrimSpace(range_item[1]))
				if err != nil {
					log.Warnf("无法解析病例序号列表：%s => %s", text, err)
					continue
				}
				for i := begin; i <= end; i = i + 1 {
					num_list = append(num_list, i)
				}
			}
		}
	}
	return num_list
}
