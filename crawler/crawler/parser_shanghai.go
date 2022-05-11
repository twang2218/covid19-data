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

type DailyParserShanghai struct{}

func (p DailyParserShanghai) GetSelector(t string) string {
	selectors := map[string]string{
		"index":   ".list-date a, .result a",
		"item":    ".Article, #page-content",
		"title":   "#ivs_title, .rich_media_title",
		"content": "#ivs_content p, section > strong, section > span > strong, section > p",
	}

	if val, ok := selectors[t]; ok {
		return val
	} else {
		return ""
	}
}

func (p DailyParserShanghai) GetItemLinks() []string {
	return []string{
		"https://mp.weixin.qq.com/s/bqZp2AqqE-FPzJpx6FlhPA", // 5月7日 居住地信息
		// "https://mp.weixin.qq.com/s/Sq1YN8oMu0RCSddtFnN3tg", // 5月5日 疫情通报
		// "https://mp.weixin.qq.com/s/xps19UKtpgZUEPhfj1GC9Q", // 5月4日 疫情通报
		// "https://mp.weixin.qq.com/s/KyTRqsRBWbM5cEa2sk2wbg", // 5月3日 居住地信息
		// "https://mp.weixin.qq.com/s/BNp0FTEIV33VRghIpWaXwg", // 5月3日 疫情通报
		// "https://mp.weixin.qq.com/s/s_spcc0OApRItbuq5DG2LA", // 5月2日 居住地信息
		// "https://mp.weixin.qq.com/s/6Zk1yLrGojy_5bU4oS9ZTA", // 5月2日 疫情通报
		// "https://mp.weixin.qq.com/s/agdZHOqVZh9atNHOQEFTog", // 5月1日 居住地信息
		// "https://mp.weixin.qq.com/s/C8CaP7iR8Bi1HizU9NnjDw", // 4月13日 疫情通报
		"https://mp.weixin.qq.com/s/LguiUZj-zxy4xy19WO0_UA", // 4月17日 居住地信息
		"https://mp.weixin.qq.com/s/wuZXG2rdCKi-A5sZQJdKfA", // 4月17日 疫情通报
		"https://mp.weixin.qq.com/s/dRa-PExJr1qkRis88eGCnQ", // 4月16日 居住地信息
		"https://mp.weixin.qq.com/s/9YaDe0nseAmv58IwTQfakQ", // 4月16日 疫情通报
		"https://mp.weixin.qq.com/s/ZkhimhWpa92I2EWn3hmd8w", //	4月15日 居住地信息
		"https://mp.weixin.qq.com/s/SE0_F-Bwc2JFM_qKLwXpyQ", // 4月15日 疫情通报
		"https://mp.weixin.qq.com/s/5T76lht3s6g_KTiIx3XAYw", // 4月14日 居住地信息
		"https://mp.weixin.qq.com/s/CuoDLOZXhBl5HREQZe_9IQ", // 4月14日 疫情通报
		"https://mp.weixin.qq.com/s/L9AffT-SoEBV4puBa_mRqg", // 4月13日 居住地信息
		"https://mp.weixin.qq.com/s/C8CaP7iR8Bi1HizU9NnjDw", // 4月13日 疫情通报
		"https://mp.weixin.qq.com/s/OZGM-pNkefZqWr0IFRJj1g", // 4月12日 居住地信息
		"https://mp.weixin.qq.com/s/SQoQiurUqYMz6xOvuBdVWw", // 4月12日 疫情通报
		"https://mp.weixin.qq.com/s/vxFiV2HeSvByINUlTmFKZA", // 4月11日 居住地信息
		"https://mp.weixin.qq.com/s/eun72mybh5Uy0k2m88ae_Q", // 4月11日 疫情通报
		"https://mp.weixin.qq.com/s/u0XfHF8dgfEp8vGjRtcwXA", // 4月10日 居住地信息
		"https://mp.weixin.qq.com/s/FVqVXKK8EBnUe9sG1Gxq8g", // 4月10日 疫情通报
		"https://mp.weixin.qq.com/s/_Je5_5_HqBcs5chvH5SFfA", // 4月9日 居住地信息
		"https://mp.weixin.qq.com/s/s_Ylm-oTP-frivKUR6Wo_A", // 4月9日 疫情通报
	}
}

func (p DailyParserShanghai) GetIndexLinks() []string {
	const (
		LINK_DAILY_0 string = "https://wsjkw.sh.gov.cn/xwfb/index.html"
		LINK_DAILY_1 string = "https://wsjkw.sh.gov.cn/xwfb/index{page}.html"
		LINK_DAILY_2 string = "https://ss.shanghai.gov.cn/search?q=%E6%96%B0%E5%A2%9E%E6%9C%AC%E5%9C%9F%20%E5%B1%85%E4%BD%8F%E5%9C%B0%E4%BF%A1%E6%81%AF&page={page}&view=xwzx&contentScope=1&dateOrder=2&tr=4&dr=&format=1&re=2&all=1&siteId=wsjkw.sh.gov.cn&siteArea=all"
		MAX_PAGES    int    = 25
	)

	links := []string{}

	for i := 1; i < MAX_PAGES; i++ {
		//	新闻中心 > 新闻发布
		page := ""
		if i > 1 {
			page = fmt.Sprintf("_%d", i)
		}
		link := strings.ReplaceAll(LINK_DAILY_1, "{page}", page)

		//	全文检索
		// page := strconv.Itoa(i)
		// link := strings.ReplaceAll(LINK_DAILY_2, "{page}", page)

		//	访问
		links = append(links, link)
	}

	return links
}

func (p DailyParserShanghai) GetDistricts() []string {
	return []string{
		"浦东新区", "徐汇区", "闵行区", "黄浦区", "嘉定区", "松江区", "虹口区", "长宁区",
		"青浦区", "静安区", "宝山区", "杨浦区", "普陀区", "崇明区", "金山区", "奉贤区",
	}
}

func (p DailyParserShanghai) IsDaily(date time.Time, title string) bool {
	return !strings.Contains(title, "居住地信息")
}

var SHANGHAI_DATE_RESIDENT_MERGED time.Time = time.Date(2022, 3, 18, 0, 0, 0, 0, time.Local)

func (p DailyParserShanghai) IsResidents(date time.Time, title string) bool {
	//	2022-03-18
	return strings.Contains(title, "居住地信息") || date.Before(SHANGHAI_DATE_RESIDENT_MERGED)
}

func (p DailyParserShanghai) IsValidTitle(title string) bool {
	return strings.Contains(title, "本土新冠肺炎") || strings.Contains(title, "居住地信息")
}

//	解析 Daily

//	解析 Daily 标题
func (p DailyParserShanghai) ParseDailyTitle(d *model.Daily, title string) error {
	if d == nil {
		return fmt.Errorf("输入对象为空")
	}

	var m []string
	var err error

	// 标题日期
	m = reDailyDate.FindStringSubmatch(title)
	if m == nil {
		return fmt.Errorf("[%s] 无法解析文章标题中日期：%q", d.Date.Format("2006-01-02"), title)
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

	if strings.Contains(title, "居住地信息") {
		// 这里面只包含居住地信息，后面不需要解析
		// log.Infof("[%s] 居住地信息：%q", cs.Date.Format("2006-01-02"), title)
		return nil
	}

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
		/// 正则包含3个可能性，因此有3个数值的匹配，但是只可能有一个有值，因此字符串合并后就是那个有值的值
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
func (p DailyParserShanghai) ParseDailyContent(d *model.Daily, content string) error {
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
			d.Date, err = time.Parse("2006年1月2日", m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中日期：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// log.Tracef("[%s] 解析文章内容：(%d)", d.Date.Format("2006-01-02"), len(content))

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

	// 当前重型
	m = reDailySevere.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中当前重型：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.CurrentSevere, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中当前重型：%q", d.Date.Format("2006-01-02"), m[1])
		}
	}

	// 当前危重型
	m = reDailyCritical.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中当前危重型：%q", d.Date.Format("2006-01-02"), content)
	} else {
		d.CurrentCritical, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中当前危重型：%q", d.Date.Format("2006-01-02"), m[1])
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

func (p DailyParserShanghai) parseDailyContentRegion(d *model.Daily, content string) error {
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

	return err
}

func (p DailyParserShanghai) parseDailyContentRegionItems(d *model.Daily, mm [][]string) map[string]int {
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

func (p DailyParserShanghai) parseDailyContentRegionItem(d *model.Daily, text string) map[string]int {
	item := make(map[string]int)
	m := reDailyRegionItem.FindAllStringSubmatch(text, -1)
	if m == nil {
		log.Warnf("[%s] 无法解析区域病例列表：%q", d.Date.Format("2006-01-02"), text)
	} else {
		for _, it := range m {
			var from int
			var to int
			var err error

			if len(it) < 3 {
				log.Warnf("[%s] 无法解析区域病例列表：%q", d.Date.Format("2006-01-02"), it)
				continue
			} else if len(it) == 3 {
				from_str := strings.TrimSpace(it[1])
				_, err := strconv.Atoi(from_str)
				if err != nil {
					log.Warnf("[%s] 无法解析区域病例列表：%q", d.Date.Format("2006-01-02"), it)
				}
				region := strings.TrimSpace(it[2])
				item[region] = 1
			} else if len(it) == 4 {
				from_str := strings.TrimSpace(it[1])
				if len(from_str) > 0 {
					from, err = strconv.Atoi(from_str)
					if err != nil {
						log.Warnf("[%s] 无法解析区域病例列表：%s => %q", d.Date.Format("2006-01-02"), err, it)
					}
				}

				to_str := strings.TrimSpace(it[2])
				if len(to_str) > 0 {
					to, err = strconv.Atoi(to_str)
					if err != nil {
						log.Warnf("[%s] 无法解析区域病例列表：%s => %q", d.Date.Format("2006-01-02"), err, it)
					}
				}

				region := strings.TrimSpace(it[3])
				if to == 0 {
					if val, ok := item[region]; ok {
						item[region] = val + 1
					} else {
						item[region] = 1
					}
				} else {
					if val, ok := item[region]; ok {
						item[region] = val + (to - from + 1)
					} else {
						item[region] = to - from + 1
					}
				}
			} else {
				log.Warnf("[%s] 无法解析区域病例列表：%q", d.Date.Format("2006-01-02"), it)
				continue
			}
		}
	}
	return item
}

var (
	reResidentDistrictShanghai1 = regexp.MustCompile(`(?:\n)(?P<district>[^\d\n：]+区)(?:\n[^\n]+(?:(?:\n分别)?居住于[^\n]?|）))*(?P<addrs>(?:\n[^\n2已][^\n]+[，。、]?)+)?`)
	reResidentDistrictShanghai2 = regexp.MustCompile(`(?P<type>病例|无症状感染者)(?P<number>\d+)，(?P<gender>男|女)，(?P<age>\d+月?)[岁龄]，(?:[^，]+，)?居住(?:于|地为)(?P<district>[^，。]+区)?(?P<addr>[^，。]+)`)
)

func (p DailyParserShanghai) ParseResidents(rs *model.Residents, date time.Time, content string) error {
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
	if date.Before(SHANGHAI_DATE_RESIDENT_MERGED) {
		//	2022年3月18日之前的居住地信息是包含于疫情通告中的

		mm = reResidentDistrictShanghai2.FindAllStringSubmatch(content, -1)
		for _, m := range mm {

			// d := m[1]
			// log.Tracef("[%s] %#v", date.Format("2006-01-02"), m)
			if len(m) == 7 {
				age_str := strings.TrimSpace(m[4])
				is_infant := false
				if strings.HasSuffix(age_str, "月") {
					is_infant = true
					age_str = strings.TrimRight(age_str, "月")
				}
				var age float64
				if is_infant {
					n, err := strconv.Atoi(age_str)
					if err != nil {
						log.Warnf("[%s] 无法解析居住地信息中的婴儿年龄：%q", date.Format("2006-01-02"), m[0])
					}
					age = float64(n) / 12
				} else {
					n, err := strconv.Atoi(strings.TrimSpace(m[4]))
					if err != nil {
						log.Warnf("[%s] 无法解析居住地信息中的年龄：%q", date.Format("2006-01-02"), m[0])
					}
					age = float64(n)
				}
				t := "确诊病例"
				if strings.TrimSpace(m[1]) == "无症状感染者" {
					t = "无症状感染者"
				}

				r := model.Resident{
					Date:     date,
					Name:     fmt.Sprintf("%s%s", m[1], m[2]),
					Type:     t,
					Gender:   strings.TrimSpace(m[3]),
					Age:      age,
					City:     "上海市",
					District: strings.TrimSpace(m[5]),
					Address:  strings.TrimSpace(m[6]),
				}
				// log.Tracef("[%s] %v", date.Format("2006-01-02"), r)
				*rs = append(*rs, r)
			} else {
				log.Warnf("[%s] 无法解析居住地信息：%q => %#v", date.Format("2006-01-02"), m[0], m)
			}
		}
	} else {
		//	自 2022年3月18日 开始，使用独立的“居住地信息”公告
		mm = reResidentDistrictShanghai1.FindAllStringSubmatch(content, -1)
		for _, m := range mm {

			d := m[1]

			s := m[2]
			for _, c := range "\n，。、 " {
				s = strings.ReplaceAll(s, string(c), ",")
			}
			s = strings.ReplaceAll(s, ",,", ",")
			s = strings.ReplaceAll(s, ",,", ",")
			s = strings.Trim(s, ", ")

			addrs := strings.Split(s, ",")
			// log.Tracef("[%s] > %q => (%d) %#v\n", date.Format("2006-01-02"), d, len(addrs), addrs)
			for _, addr := range addrs {
				if len(addr) > 0 {
					r := model.Resident{
						Date:     date,
						Name:     fmt.Sprintf("%s%s", d, addr),
						City:     "上海市",
						District: d,
						Address:  addr,
					}
					*rs = append(*rs, r)
				}
			}
		}
	}

	return err
}
