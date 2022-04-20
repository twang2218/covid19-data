package crawler

import (
	"crawler/model"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

const (
	CRAWLER_CACHE_DIR       = "../data/.cache/"
	CRAWLER_REQUEST_TIMEOUT = 4 * time.Minute
	CRAWLER_RETRY_TIMEOUT   = 4 * time.Minute
	CRAWLER_VISIT_DELAY     = 100 * time.Millisecond
	CRAWLER_PARALLELISM     = 5
)

type DailyCrawler struct {
	PageVisited int32
	PageTotal   int32

	cItem     *colly.Collector
	cIndex    *colly.Collector
	listeners []func(model.Daily)
}

const (
	LINK_DAILY_1 string = "https://wsjkw.sh.gov.cn/xwfb/index.html"
	LINK_DAILY_2 string = "https://ss.shanghai.gov.cn/search?page={page}&view=xwzx&contentScope=1&dateOrder=1&tr=5&dr=2022-01-01+%E8%87%B3+2022-06-01&format=1&re=2&all=1&debug=&siteId=wsjkw.sh.gov.cn&siteArea=all&q=%E6%96%B0%E5%A2%9E%E6%9C%AC%E5%9C%9F"
)

func NewDailyCrawler(cache bool) *DailyCrawler {
	hc := DailyCrawler{}
	//	创建基础爬虫
	// u, err := url.Parse(LINK_DAILY)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	var cache_dir string
	if cache {
		cache_dir = CRAWLER_CACHE_DIR
	}

	hc.cItem = colly.NewCollector(
		colly.AllowedDomains(
			"wsjkw.sh.gov.cn",
			"ss.shanghai.gov.cn",
			"mp.weixin.qq.com",
		),
		colly.DetectCharset(),
		colly.CacheDir(cache_dir),
		colly.Async(true),
	)
	hc.cItem.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: CRAWLER_PARALLELISM,
	})
	hc.cItem.SetRequestTimeout(CRAWLER_REQUEST_TIMEOUT)
	extensions.RandomUserAgent(hc.cItem)
	hc.cItem.OnRequest(func(r *colly.Request) {
		r.ResponseCharacterEncoding = "utf-8"
		atomic.AddInt32(&hc.PageTotal, 1)
	})
	hc.cItem.OnScraped(func(r *colly.Response) {
		atomic.AddInt32(&hc.PageVisited, 1)
	})
	hc.cItem.OnError(func(resp *colly.Response, err error) {
		log.Warnf("DailyCrawler.OnError(): [Item] (%s) => '%s'", resp.Request.URL, err)
		//	重试。在另外线程等待一段时间，以不阻碍当前线程（爬虫）运行。
		go func(link string) {
			time.Sleep(CRAWLER_RETRY_TIMEOUT)
			hc.cItem.Visit(link)
		}(resp.Request.URL.String())
	})
	//	添加解析具体页面函数
	hc.cItem.OnHTML(".Article", hc.parseItem)

	//	创建索引爬虫
	hc.cIndex = hc.cItem.Clone()
	extensions.RandomUserAgent(hc.cIndex)
	hc.cIndex.OnRequest(func(r *colly.Request) {
		// r.ResponseCharacterEncoding = "gb2312"
		log.Debugf("DailyCrawler => %s", r.URL)
	})
	hc.cIndex.OnError(func(resp *colly.Response, err error) {
		log.Warnf("DailyIndexCrawler.OnError(): [Index] (%s) => '%s'", resp.Request.URL, err)
		//	重试。在另外线程等待一段时间，以不阻碍当前线程（爬虫）运行。
		go func(link string) {
			time.Sleep(CRAWLER_RETRY_TIMEOUT)
			hc.cIndex.Visit(link)
		}(resp.Request.URL.String())
	})
	//	添加索引解析函数
	// hc.cIndex.OnHTML(".list-date a", hc.parseIndex)
	hc.cIndex.OnHTML(".result a", hc.parseIndex)

	hc.listeners = []func(model.Daily){}

	return &hc
}

func (c *DailyCrawler) Collect() {
	//	开始抓取 （循环以抓取指定页数）
	// c.cIndex.Visit(LINK_DAILY_1)
	for i := 1; i < 20; i++ {
		// if i > 1 {
		// 	page = fmt.Sprintf("_%d", i)
		// }
		page := strconv.Itoa(i)
		link := strings.ReplaceAll(LINK_DAILY_2, "{page}", page)
		c.cIndex.Visit(link)
	}
	//	等待结束
	c.cIndex.Wait()
	c.cItem.Wait()
}

func (c *DailyCrawler) parseItem(e *colly.HTMLElement) {
	var cs model.Daily
	cs.Source = e.Request.URL.String()
	// cs.Source = e.Response.Request.URL.String()
	// fmt.Println(cs.Source)

	// 标题
	title := strings.TrimSpace(e.ChildText("#ivs_title"))
	// log.Tracef("DailyCrawler.parseItem(): [%s] %q", cs.Date.Format("2006-01-02"), title)

	if err := parseTitle(&cs, title); err != nil {
		log.Errorf("解析文章标题失败：%s => %q", err, title)
	}

	// log.Tracef("DailyCrawler.parseItem(): [%s] 本土 (新增:%d, 无症状: %d), 境外输入 (新增:%d, 无症状: %d), 出院: %d, 解除医学观察: %d",
	// 	cs.Date.Format("2006-01-02"),
	// 	cs.LocalConfirmed,
	// 	cs.LocalAsymptomatic,
	// 	cs.ImportedConfirmed,
	// 	cs.ImportedAsymptomatic,
	// 	cs.DischargedFromHospital,
	// 	cs.DischargedFromMedicalObservation,
	// )

	// content := strings.TrimSpace(e.ChildText("#ivs_content"))
	//	上面的代码会去掉所有换行，导致匹配失败，因此用下面的方式行于行之间用 '\n' 链接
	content_lines := []string{}
	e.ForEach("#ivs_content p", func(i int, h *colly.HTMLElement) {
		t := strings.TrimSpace(h.Text)
		if len(t) > 0 {
			content_lines = append(content_lines, t)
		}
	})
	content := strings.Join(content_lines, "\n")

	if err := parseContent(&cs, content); err != nil {
		log.Errorf("解析文章内容失败：%s => %q", err, title)
	}
	// log.Tracef("DailyCrawler.parseItem(): [%s] 本土（无症状=>确诊:%d, 出院:%d），境外输入 (出院:%d); 解除医学观察（本土:%d，境外输入:%d）",
	// 	cs.Date.Format("2006-01-02"),
	// 	cs.LocalConfirmedFromAsymptomatic,
	// 	cs.LocalDischargedFromHospital,
	// 	cs.ImportedDischargedFromHospital,
	// 	cs.LocalDischargedFromMedicalObservation,
	// 	cs.ImportedDischargedFromMedicalObservation,
	// )

	fixDaily(&cs)

	//	通知 OnDailyListeners
	c.notifyAllOnDailyListeners(cs)
}

func (c *DailyCrawler) parseIndex(e *colly.HTMLElement) {
	link := e.Request.AbsoluteURL(strings.TrimSpace(e.Attr("href")))
	title := e.Text
	// if !strings.Contains(title, "上海2022") {
	if !strings.Contains(title, "本土新冠肺炎") {
		return
	}
	// log.Tracef("DailyCrawler.parseIndex() => %q --- %q", link, title)
	//	告知 cItem 抓取该链接
	c.cItem.Visit(link)
}

//	解析 Daily

var (
	reDailyDate1                            = regexp.MustCompile(`上海(?P<date>\d+年\d+月\d+日)`)
	reDailyDate2                            = regexp.MustCompile(`(?P<date>\d+月\d+日)`)
	reDailyLocalConfirmed                   = regexp.MustCompile(`本土新冠肺炎确诊病例(?P<number>\d+)例`)
	reDailyLocalAsymptomatic                = regexp.MustCompile(`本土无症状感染者(?P<number>\d+)例`)
	reDailyImportedConfirmed                = regexp.MustCompile(`境外输入(?:性新冠肺炎确诊)?(?:病例)?(?P<number>\d+)例`)
	reDailyImportedAsymptomatic             = regexp.MustCompile(`境外输入性无症状感染者(?P<number>\d+)例`)
	reDailyDischargedFromHospital           = regexp.MustCompile(`治愈出院(?P<number>\d+)例`)
	reDailyDischargedFromMedicalObservation = regexp.MustCompile(`解除医学观察(?:无症状感染者)?(?P<number>\d+)例`)
)

//	解析 Daily 标题
func parseTitle(cs *model.Daily, title string) error {
	if cs == nil {
		return fmt.Errorf("输入对象为空")
	}

	var m []string
	var err error

	// 标题日期
	m = reDailyDate1.FindStringSubmatch(title)
	if m == nil {
		m = reDailyDate2.FindStringSubmatch(title)
		if m == nil {
			return fmt.Errorf("[%s] 无法解析文章标题中日期：%q", cs.Date.Format("2006-01-02"), title)
		}
		cs.Date, err = time.Parse("2006年1月2日", fmt.Sprintf("2022年%s", m[1]))
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中日期：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	} else {
		cs.Date, err = time.Parse("2006年1月2日", m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中日期：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土新增
	m = reDailyLocalConfirmed.FindStringSubmatch(title)
	if m == nil {
		// return fmt.Errorf("[%s] 无法解析文章标题中本土新增：%q", cs.Date.Format("2006-01-02"), title)
	} else {
		cs.LocalConfirmed, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中本土新增：%q", cs.Date.Format("2006-01-02"), title)
		}
	}

	// 本土无症状
	m = reDailyLocalAsymptomatic.FindStringSubmatch(title)
	if m == nil {
		// return fmt.Errorf("[%s] 无法解析文章标题中本土无症状感染者：%q", cs.Date.Format("2006-01-02"), title)
	} else {
		cs.LocalAsymptomatic, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中本土无症状感染者：%q", cs.Date.Format("2006-01-02"), title)
		}
	}

	// 境外输入新增
	m = reDailyImportedConfirmed.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中境外输入新增：%q", cs.Date.Format("2006-01-02"), title)
	} else {
		cs.ImportedConfirmed, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中境外输入新增：%q", cs.Date.Format("2006-01-02"), title)
		}
	}

	// 境外输入无症状
	m = reDailyImportedAsymptomatic.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中境外输入无症状感染者：%q", cs.Date.Format("2006-01-02"), title)
	} else {
		cs.ImportedAsymptomatic, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中境外输入无症状感染者：%q", cs.Date.Format("2006-01-02"), title)
		}
	}

	// 治愈出院
	m = reDailyDischargedFromHospital.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中治愈出院：%q", cs.Date.Format("2006-01-02"), title)
	} else {
		cs.DischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中治愈出院：%q", cs.Date.Format("2006-01-02"), title)
		}
	}

	// 解除医学观察
	m = reDailyDischargedFromMedicalObservation.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中解除医学观察：%q", cs.Date.Format("2006-01-02"), title)
	} else {
		cs.DischargedFromMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中解除医学观察：%q", cs.Date.Format("2006-01-02"), title)
		}
	}

	return nil
}

var (
	reDailyLocalConfirmedFromAsymptomatic           = regexp.MustCompile(`—24时.*本土.*含(?P<number>\d+)例由无症状感染者转为确诊病例`)
	reDailyLocalDischargedFromHospital              = regexp.MustCompile(`—24时.*本土.*治愈出院(?P<number>\d+)例`)
	reDailyImportedDischargedFromHospital           = regexp.MustCompile(`—24\s*时.*境外输入.*治愈出院(?P<number>\d+)例`)
	reDailyDischargedFromMedicalObservation2        = regexp.MustCompile(`—24时.*解除医学观察无症状感染者(?P<number>\d+)例`)
	reDailyLocalDischargedFromMedicalObservation    = regexp.MustCompile(`—24时.*(?:新增本土.*解除医学观察|解除医学观察.*本土)无症状感染者(?P<number>\d+)例`)
	reDailyImportedDischargedFromMedicalObservation = regexp.MustCompile(`—24时.*解除医学观察.*境外输入性无症状感染者(?P<number>\d+)例`)
	reDailyImportedAsymptomatic2                    = regexp.MustCompile(`—24时.*境外输入性无症状感染者(?P<number>\d+)例`)
	reDailyLocalConfirmedFromBubble                 = regexp.MustCompile(`—24时.*，(?:其中)?(?P<number>\d+)例确诊病例和.*在隔离管控中发现`)
	reDailyLocalAsymptomaticFromBubble              = regexp.MustCompile(`—24时.*和(?P<number>\d+)例无症状感染者在隔离管控中发现`)
	reDailyLocalDeath                               = regexp.MustCompile(`—24时.*本土.*死亡病例(?P<number>\d+)例`)
	reDailyImportedDeath                            = regexp.MustCompile(`—24时.*境外输入.*死亡病例(?P<number>\d+)例`)
)

// 564例确诊病例和24230例无症状感染者在隔离管控中发现

// 解析 Daily 内容
func parseContent(cs *model.Daily, content string) error {
	if cs == nil {
		return fmt.Errorf("输入对象为空")
	}

	var m []string
	var err error

	// 境外输入无症状 (补充标题缺失)
	if cs.ImportedAsymptomatic == 0 {
		m = reDailyImportedAsymptomatic2.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中境外输入无症状：%q", cs.Date.Format("2006-01-02"), content)
		} else {
			cs.ImportedAsymptomatic, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中境外输入无症状：%q", cs.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 本土无症状转为确诊病例
	m = reDailyLocalConfirmedFromAsymptomatic.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土无症状转为确诊病例：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.LocalConfirmedFromAsymptomatic, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土无症状转为确诊病例：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土隔离管控中发现的确诊病例
	m = reDailyLocalConfirmedFromBubble.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土隔离管控中发现的确诊病例：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.LocalConfirmedFromBubble, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土隔离管控中发现的确诊病例：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土隔离管控中发现的无症状病例
	m = reDailyLocalAsymptomaticFromBubble.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土隔离管控中发现的无症状病例：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.LocalAsymptomaticFromBubble, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土隔离管控中发现的无症状病例：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土治愈出院
	m = reDailyLocalDischargedFromHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土治愈出院：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.LocalDischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土治愈出院：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// // 境外输入无症状转为确诊病例
	// m = reDailyImportedConfirmedFromAsymptomatic.FindStringSubmatch(content)
	// if m == nil {
	// 	// log.Warnf("[%s] 无法解析文章内容中境外输入无症状转为确诊病例：%q", cs.Date.Format("2006-01-02"), content)
	// } else {
	// 	cs.ImportedConfirmedFromAsymptomatic, err = strconv.Atoi(m[1])
	// 	if err != nil {
	// 		return fmt.Errorf("[%s] 无法解析文章内容中境外输入无症状转为确诊病例：%q", cs.Date.Format("2006-01-02"), m[1])
	// 	}
	// }

	// 境外输入治愈出院
	m = reDailyImportedDischargedFromHospital.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入治愈出院：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.ImportedDischargedFromHospital, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入治愈出院：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 解除医学观察 (补充标题缺失)
	if cs.DischargedFromMedicalObservation == 0 {
		m = reDailyDischargedFromMedicalObservation2.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中解除医学观察：%q", cs.Date.Format("2006-01-02"), content)
		} else {
			cs.DischargedFromMedicalObservation, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中解除医学观察：%q", cs.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 本土解除医学观察
	m = reDailyLocalDischargedFromMedicalObservation.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土解除医学观察：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.LocalDischargedFromMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土解除医学观察：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 境外输入解除医学观察
	m = reDailyImportedDischargedFromMedicalObservation.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入解除医学观察：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.ImportedDischargedFromMedicalObservation, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入解除医学观察：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 本土死亡病例
	m = reDailyLocalDeath.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中本土死亡病例：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.LocalDeath, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中本土死亡病例：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	// 境外输入死亡病例
	m = reDailyImportedDeath.FindStringSubmatch(content)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章内容中境外输入死亡病例：%q", cs.Date.Format("2006-01-02"), content)
	} else {
		cs.ImportedDeath, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章内容中境外输入死亡病例：%q", cs.Date.Format("2006-01-02"), m[1])
		}
	}

	return nil
}

func fixDaily(cs *model.Daily) error {
	//	无症状
	if cs.Asymptomatic == 0 {
		if cs.LocalAsymptomatic != 0 || cs.ImportedAsymptomatic != 0 {
			cs.Asymptomatic = cs.LocalAsymptomatic + cs.ImportedAsymptomatic
		}
	} else {
		if cs.Asymptomatic != (cs.LocalAsymptomatic + cs.ImportedAsymptomatic) {
			log.Warnf("[%s] 无症状数据不匹配：总共:%d (本土:%d / 境外输入:%d)", cs.Date.Format("2006-01-02"), cs.Asymptomatic, cs.LocalAsymptomatic, cs.ImportedAsymptomatic)
		}
	}
	if cs.LocalAsymptomaticFromRisk == 0 {
		cs.LocalAsymptomaticFromRisk = cs.LocalAsymptomatic - cs.LocalAsymptomaticFromBubble
	}

	//	确诊
	if cs.Confirmed == 0 {
		if cs.LocalConfirmed != 0 || cs.ImportedConfirmed != 0 {
			cs.Confirmed = cs.LocalConfirmed + cs.ImportedConfirmed
		}
	} else {
		if cs.Confirmed != (cs.LocalConfirmed + cs.ImportedConfirmed) {
			log.Warnf("[%s] 确诊数据不匹配：总共:%d (本土:%d / 境外输入:%d)", cs.Date.Format("2006-01-02"), cs.Confirmed, cs.LocalConfirmed, cs.ImportedConfirmed)
		}
	}
	if cs.LocalConfirmedFromRisk == 0 {
		cs.LocalConfirmedFromRisk = cs.LocalConfirmed - (cs.LocalConfirmedFromBubble + cs.LocalConfirmedFromAsymptomatic)
	}

	// 治愈出院
	if cs.DischargedFromHospital == 0 {
		if cs.LocalDischargedFromHospital != 0 || cs.ImportedDischargedFromHospital != 0 {
			cs.DischargedFromHospital = cs.LocalDischargedFromHospital + cs.ImportedDischargedFromHospital
		}
	} else {
		if cs.DischargedFromHospital != (cs.LocalDischargedFromHospital + cs.ImportedDischargedFromHospital) {
			if cs.LocalDischargedFromHospital == 0 && cs.ImportedDischargedFromHospital > 0 {
				//	应该是没能解析出本土治愈出院，可以计算获得
				cs.LocalDischargedFromHospital = cs.DischargedFromHospital - cs.ImportedDischargedFromHospital
			} else if cs.LocalDischargedFromHospital > 0 && cs.ImportedDischargedFromHospital == 0 {
				//  应该是没能解析出境外输入治愈出院，可以计算获得
				cs.ImportedDischargedFromHospital = cs.DischargedFromHospital - cs.LocalDischargedFromHospital
			} else {
				//	三者均不为0，因此必然是出错了。
				log.Warnf("[%s] 治愈出院数据不匹配：总共:%d (本土:%d / 境外输入:%d)", cs.Date.Format("2006-01-02"), cs.DischargedFromHospital, cs.LocalDischargedFromHospital, cs.ImportedDischargedFromHospital)
			}
		}
	}

	// 解除医学观察
	if cs.DischargedFromMedicalObservation == 0 {
		if cs.LocalDischargedFromMedicalObservation != 0 || cs.ImportedDischargedFromMedicalObservation != 0 {
			cs.DischargedFromMedicalObservation = cs.LocalDischargedFromMedicalObservation + cs.ImportedDischargedFromMedicalObservation
		}
	} else {
		if cs.DischargedFromMedicalObservation != (cs.LocalDischargedFromMedicalObservation + cs.ImportedDischargedFromMedicalObservation) {
			log.Warnf("[%s] 解除医学观察数据不匹配：总共:%d (本土:%d / 境外输入:%d)", cs.Date.Format("2006-01-02"), cs.DischargedFromMedicalObservation, cs.LocalDischargedFromMedicalObservation, cs.ImportedDischargedFromMedicalObservation)
		}
	}

	// 死亡
	if cs.Death == 0 {
		if cs.LocalDeath != 0 || cs.ImportedDeath != 0 {
			cs.Death = cs.LocalDeath + cs.ImportedDeath
		}
	} else {
		if cs.Death != (cs.LocalDeath + cs.ImportedDeath) {
			log.Warnf("[%s] 死亡数据不匹配：总共:%d (本土:%d / 境外输入:%d)", cs.Date.Format("2006-01-02"), cs.Death, cs.LocalDeath, cs.ImportedDeath)
		}
	}

	return nil
}

//	Listener functions

func (c *DailyCrawler) AddOnDailyListener(f func(model.Daily)) {
	if f == nil {
		log.Warn("DailyCrawler.AddOnDailyListener(): couldn't add 'nil' as listener.")
		return
	}

	c.listeners = append(c.listeners, f)
}

func (c *DailyCrawler) ClearOnDailyListener() {
	c.listeners = []func(model.Daily){}
}

func (c *DailyCrawler) notifyAllOnDailyListeners(h model.Daily) {
	for _, listener := range c.listeners {
		listener(h)
	}
}
