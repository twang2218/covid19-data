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

	cItem              *colly.Collector
	cIndex             *colly.Collector
	listenersDaily     []func(model.Daily)
	listenersResidents []func(model.Residents)
}

const (
	LINK_DAILY_0 string = "https://wsjkw.sh.gov.cn/xwfb/index.html"
	LINK_DAILY_1 string = "https://wsjkw.sh.gov.cn/xwfb/index{page}.html"
	LINK_DAILY_2 string = "https://ss.shanghai.gov.cn/search?q=%E6%96%B0%E5%A2%9E%E6%9C%AC%E5%9C%9F%20%E5%B1%85%E4%BD%8F%E5%9C%B0%E4%BF%A1%E6%81%AF&page={page}&view=xwzx&contentScope=1&dateOrder=2&tr=4&dr=&format=1&re=2&all=1&siteId=wsjkw.sh.gov.cn&siteArea=all"
	MAX_PAGES    int    = 30
)

var DATE_RESIDENT_MERGED time.Time = time.Date(2022, 3, 18, 0, 0, 0, 0, time.Local)

func NewDailyCrawler(cache_dir string) *DailyCrawler {
	hc := DailyCrawler{}
	//	创建基础爬虫
	hc.cItem = colly.NewCollector(
		colly.AllowedDomains(
			"wsjkw.sh.gov.cn",
			"ss.shanghai.gov.cn",
			"mp.weixin.qq.com",
		),
		colly.DetectCharset(),
		colly.CacheDir(cache_dir),
		colly.Async(true),
		// colly.Debugger(&debug.LogDebugger{}),
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
		// log.Tracef("NewDailyCrawler(): cItem.Request(): %s\n", r.URL)
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
	///	疫情通报、居住地信息; 微信居住地信息
	hc.cItem.OnHTML(".Article, #page-content", hc.parseItem)

	//	创建索引爬虫
	hc.cIndex = hc.cItem.Clone()
	extensions.RandomUserAgent(hc.cIndex)
	hc.cIndex.OnRequest(func(r *colly.Request) {
		// r.ResponseCharacterEncoding = "gb2312"
		// log.Debugf("DailyCrawler => %s", r.URL)
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
	hc.cIndex.OnHTML(".list-date a, .result a", hc.parseIndex)

	hc.listenersDaily = []func(model.Daily){}

	return &hc
}

func (c *DailyCrawler) Collect() {
	//	开始抓取
	c.cItem.Visit("https://mp.weixin.qq.com/s/agdZHOqVZh9atNHOQEFTog")
	///（循环以抓取指定页数）
	// c.cIndex.Visit(LINK_DAILY_1)
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
		c.cIndex.Visit(link)
	}
	//	等待结束
	c.cIndex.Wait()
	c.cItem.Wait()
}

func (c *DailyCrawler) parseItem(e *colly.HTMLElement) {
	var d model.Daily
	d.Source = e.Request.URL.String()
	// d.Source = e.Response.Request.URL.String()
	// fmt.Println(d.Source)

	// 标题
	title := strings.TrimSpace(e.ChildText("#ivs_title, .rich_media_title"))
	// log.Tracef("DailyCrawler.parseItem(%s): %s => %q\n", e.Attr("id"), e.Request.URL, title)

	if err := parseDailyTitle(&d, title); err != nil {
		log.Errorf("解析文章标题失败：%s => %q", err, title)
	}

	// log.Tracef("DailyCrawler.parseItem(): [%s] 本土 (新增:%d, 无症状: %d), 境外输入 (新增:%d, 无症状: %d), 出院: %d, 解除医学观察: %d",
	// 	d.Date.Format("2006-01-02"),
	// 	d.LocalConfirmed,
	// 	d.LocalAsymptomatic,
	// 	d.ImportedConfirmed,
	// 	d.ImportedAsymptomatic,
	// 	d.DischargedFromHospital,
	// 	d.DischargedFromMedicalObservation,
	// )

	// content := strings.TrimSpace(e.ChildText("#ivs_content"))
	//	上面的代码会去掉所有换行，导致匹配失败，因此用下面的方式行于行之间用 '\n' 链接
	content_lines := []string{}
	e.ForEach("#ivs_content p, section > strong, section > span > strong, section > p", func(i int, h *colly.HTMLElement) {
		t := strings.TrimSpace(h.Text)
		if len(t) > 0 {
			content_lines = append(content_lines, t)
		}
	})
	content := strings.Join(content_lines, "\n")

	// log.Tracef("[%s] <%s>: %s", d.Date.Format("2006-01-02"), title, d.Source)

	if !strings.Contains(title, "居住地信息") {
		// log.Tracef("[%s] <%s>: 每日疫情信息统计", d.Date.Format("2006-01-02"), title)
		//	每日疫情信息统计
		if err := parseDailyContent(&d, content); err != nil {
			log.Errorf("[%s] 解析文章内容失败：%s => %q", d.Date.Format("2006-01-02"), err, title)
		}
		// log.Tracef("DailyCrawler.parseItem(): [%s] 本土（无症状=>确诊:%d, 出院:%d），境外输入 (出院:%d); 解除医学观察（本土:%d，境外输入:%d）",
		// 	d.Date.Format("2006-01-02"),
		// 	d.LocalConfirmedFromAsymptomatic,
		// 	d.LocalDischargedFromHospital,
		// 	d.ImportedDischargedFromHospital,
		// 	d.LocalDischargedFromMedicalObservation,
		// 	d.ImportedDischargedFromMedicalObservation,
		// )

		fixDaily(&d)

		//	通知 OnDailyListeners
		c.notifyAllOnDailyListeners(d)
	}
	if strings.Contains(title, "居住地信息") || d.Date.Before(DATE_RESIDENT_MERGED) {
		// log.Tracef("[%s] <%s>: 居住地信息统计", d.Date.Format("2006-01-02"), title)
		//	居住地信息统计
		rs := make(model.Residents, 0)
		if err := parseResidents(&rs, d.Date, content); err != nil {
			log.Errorf("[%s] 解析感染者居住地信息失败：%s => %q", d.Date.Format("2006-01-02"), err, title)
		} else {
			// log.Infof("[%s] 解析到 %d 个感染病例居住地信息。", d.Date.Format("2006-01-02"), len(rs))
			c.notifyAllOnResidentsListeners(rs)
		}
		// log.Tracef("[%s] <%s>: 居住地信息统计: rs => [%d]", d.Date.Format("2006-01-02"), title, len(rs))
	}
}

func (c *DailyCrawler) parseIndex(e *colly.HTMLElement) {
	link := e.Request.AbsoluteURL(strings.TrimSpace(e.Attr("href")))
	title := e.Text
	// if !strings.Contains(title, "上海2022") {
	if !strings.Contains(title, "本土新冠肺炎") && !strings.Contains(title, "居住地信息") {
		return
	}
	// log.Tracef("DailyCrawler.parseIndex() => %q --- %q", link, title)
	//	告知 cItem 抓取该链接
	c.cItem.Visit(link)
}

//	解析 Daily

var (
	reDailyDate                             = regexp.MustCompile(`(?:^|[：】海京]+)(?P<date>(?:\d+年)?\d+月\d+日)(?:[，（]+|0—24时|[^，]+新增)`)
	reDailyLocalConfirmed                   = regexp.MustCompile(`(?:[^累计]+本土[新冠肺炎]*确诊病例|新增)(?P<number>\d+)(?:例|例本土新冠肺炎确诊(?:病例)?)(?:[、，。（ ]|$)`)
	reDailyLocalAsymptomatic                = regexp.MustCompile(`新增(?:本土无症状感染者)?(?P<number>\d+)(?:例|例本土无症状感染者)(?:[、，。（ ]|$)`)
	reDailyImportedConfirmed                = regexp.MustCompile(`境外输入(?:性新冠肺炎确诊)?(?:病例)?(?P<number>\d+)例`)
	reDailyImportedAsymptomatic             = regexp.MustCompile(`境外输入性无症状感染者(?P<number>\d+)例`)
	reDailyDischargedFromHospital           = regexp.MustCompile(`治愈出院(?P<number>\d+)例`)
	reDailyDischargedFromMedicalObservation = regexp.MustCompile(`解除医学观察(?:无症状感染者)?(?P<number>\d+)例`)
)

//	解析 Daily 标题
func parseDailyTitle(d *model.Daily, title string) error {
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
		if !strings.Contains(m[1], "年") {
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
		d.LocalConfirmed, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中本土新增：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 本土无症状
	m = reDailyLocalAsymptomatic.FindStringSubmatch(title)
	if m == nil {
		// return fmt.Errorf("[%s] 无法解析文章标题中本土无症状感染者：%q", d.Date.Format("2006-01-02"), title)
	} else {
		d.LocalAsymptomatic, err = strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("[%s] 无法解析文章标题中本土无症状感染者：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 境外输入新增
	m = reDailyImportedConfirmed.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中境外输入新增：%q", d.Date.Format("2006-01-02"), title)
	} else {
		d.ImportedConfirmed, err = strconv.Atoi(m[1])
		if err != nil {
			log.Warnf("[%s] 无法解析文章标题中境外输入新增：%q", d.Date.Format("2006-01-02"), title)
		}
	}

	// 境外输入无症状
	m = reDailyImportedAsymptomatic.FindStringSubmatch(title)
	if m == nil {
		// log.Warnf("[%s] 无法解析文章标题中境外输入无症状感染者：%q", d.Date.Format("2006-01-02"), title)
	} else {
		d.ImportedAsymptomatic, err = strconv.Atoi(m[1])
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

var (
	reDailyUnderMedicalObservation                  = regexp.MustCompile(`24时[^。]+尚在医学观察中的[无症状]+感染者(?P<number>\d+)例`)
	reDailyDischargedFromMedicalObservation2        = regexp.MustCompile(`—24时.*解除医学观察无症状感染者(?P<number>\d+)例`)
	reDailyLocalConfirmedFromBubble                 = regexp.MustCompile(`—24时.*，(?:其中)?(?P<number>\d+)例确诊病例和.*在隔离管控中发现`)
	reDailyLocalConfirmedFromAsymptomatic           = regexp.MustCompile(`—24时.*本土.*(?:含|其中)(?P<number>\d+)例(?:确诊病例)?(?:由|为既往)无症状感染者(?:转为确诊病例|转归)`)
	reDailyLocalAsymptomaticFromBubble              = regexp.MustCompile(`—24时.*和(?P<number>\d+)例无症状感染者在隔离管控中发现`)
	reDailyLocalDischargedFromHospital              = regexp.MustCompile(`—24时.*本土.*治愈出院(?P<number>\d+)例`)
	reDailyLocalDischargedFromMedicalObservation    = regexp.MustCompile(`—24时.*(?:新增本土.*解除医学观察|解除医学观察.*本土)无症状感染者(?P<number>\d+)例`)
	reDailyLocalInHospital                          = regexp.MustCompile(`24时[^。]+累计[^。]*本土[^。]*在院治疗(?P<number>\d+)例`)
	reDailyLocalUnderMedicalObservation             = regexp.MustCompile(`24时[^。]+尚在医学观察中[^。]+本土无症状感染者(?P<number>\d+)[例，]`)
	reDailyLocalDeath                               = regexp.MustCompile(`—24时.*本土.*死亡(?:病例)?(?P<number>\d+)例`)
	reDailyImportedAsymptomatic2                    = regexp.MustCompile(`—24时.*境外输入性无症状感染者(?P<number>\d+)例`)
	reDailyImportedDischargedFromHospital           = regexp.MustCompile(`—24\s*时.*境外输入.*治愈出院(?P<number>\d+)例`)
	reDailyImportedDischargedFromMedicalObservation = regexp.MustCompile(`—24时.*解除医学观察.*境外输入性无症状感染者(?P<number>\d+)例`)
	reDailyImportedInHospital                       = regexp.MustCompile(`24时[^。]+累计[^。]*境外输入[^。]*在院治疗(?P<number>\d+)例`)
	reDailyImportedUnderMedicalObservation          = regexp.MustCompile(`24时[^。]+尚在医学观察中[^。]*境外输入性?无症状[感染者]+(?P<number>\d+)[例，。]`)
	reDailyImportedDeath                            = regexp.MustCompile(`—24时.*境外输入.*死亡(?:病例)?(?P<number>\d+)例`)
	reDailySevere                                   = regexp.MustCompile(`24时[^。]+累计[^。]*本土[^。危]*重[型症](?P<number>\d+)例`)
	reDailyCritical                                 = regexp.MustCompile(`24时[^。]+累计[^。]*本土[^。]*危重型(?P<number>\d+)例`)
	reDailyTotalLocalConfirmed                      = regexp.MustCompile(`24时[^。]+累计本土确诊(?:病例)?(?P<number>\d+)例`)
	reDailyTotalLocalDischargedFromHospital         = regexp.MustCompile(`24时[^。]+累计[^。]*本土[^。]*治愈出院(?P<number>\d+)例`)
	reDailyTotalLocalDeath                          = regexp.MustCompile(`24时[^。]+累计[^。]*(?:本土)?[^。]*死亡(?P<number>\d+)例`)
	reDailyTotalImportedConfirmed                   = regexp.MustCompile(`24时[^。]+累计[^。]*境外输入[^。]*确诊病例(?P<number>\d+)例`)
	reDailyTotalImportedDischargedFromHospital      = regexp.MustCompile(`24时[^。]+累计[^。]*境外输入[^。]*出院(?P<number>\d+)例`)
)

// 解析 Daily 内容
func parseDailyContent(d *model.Daily, content string) error {
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
			d.LocalConfirmed, err = strconv.Atoi(m[1])
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
			d.LocalAsymptomatic, err = strconv.Atoi(m[1])
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
			d.ImportedConfirmed, err = strconv.Atoi(m[1])
			if err != nil {
				return fmt.Errorf("[%s] 无法解析文章内容中境外输入确诊：%q", d.Date.Format("2006-01-02"), m[1])
			}
		}
	}

	// 境外输入确诊 (补充标题缺失)
	if d.ImportedConfirmed == 0 {
		m = reDailyImportedConfirmed.FindStringSubmatch(content)
		if m == nil {
			// log.Warnf("[%s] 无法解析文章内容中境外输入确诊：%q", d.Date.Format("2006-01-02"), content)
		} else {
			d.ImportedConfirmed, err = strconv.Atoi(m[1])
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
		d.LocalInHospital, err = strconv.Atoi(m[1])
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
		d.ImportedInHospital, err = strconv.Atoi(m[1])
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
	return parseDailyContentRegion(d, content)
}

var (
	patternCase                            = `(?:病例|无症状感染者)(?P<from>\d+)(?:[—、,]?(?:病例|无症状感染者)?(?P<to>\d+))?(?:，[^区\n]+)*(?:，(?:居住(?:于|地为))?(?P<district>[^区\n]+区)[^，]*)，`
	patternCaseList                        = `(?:` + patternCase + `\n?)+`
	reDailyRegionConfirmedFromBubble       = regexp.MustCompile(patternCaseList + `.*(?:隔离管控人员|无症状感染者的密切接触者).*确诊病例`)
	reDailyRegionConfirmedFromRisk         = regexp.MustCompile(patternCaseList + `.*风险人群筛查.*确诊病例`)
	reDailyRegionConfirmedFromAsymptomatic = regexp.MustCompile(patternCaseList + `.*本土无症状感染者[^密\n]+确诊病例`)
	reDailyRegionAsymptomaticFromBubble    = regexp.MustCompile(patternCaseList + `.*(?:隔离管控人员|无症状感染者的密切接触者).*诊断为无症状感染者`)
	reDailyRegionAsymptomaticFromRisk      = regexp.MustCompile(patternCaseList + `.*风险人群筛查.*诊断为无症状感染者`)
	reDailyRegionItem                      = regexp.MustCompile(patternCase)
)

func parseDailyContentRegion(d *model.Daily, content string) error {
	var mm [][]string
	var err error

	//	隔离管控 => 确诊病例
	mm = reDailyRegionConfirmedFromBubble.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalConfirmedFromBubble > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区确诊病例(来自隔离管控)：%q", d.Date.Format("2006-01-02"), content)
		}
	} else {
		d.DistrictConfirmedFromBubble = parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区确诊病例(来自隔离管控): %#v", cs.Date.Format("2006-01-02"), cs.DistrictConfirmedFromBubble)
	}
	//	风险人群 => 确诊病例
	mm = reDailyRegionConfirmedFromRisk.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalConfirmedFromRisk > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区确诊病例(来自风险人群)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictConfirmedFromRisk = parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区确诊病例(来自风险人群): %#v", cs.Date.Format("2006-01-02"), cs.DistrictConfirmedFromRisk)
	}
	//	无症状 => 确诊病例
	mm = reDailyRegionConfirmedFromAsymptomatic.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalConfirmedFromAsymptomatic > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区确诊病例(来自无症状感染者)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictConfirmedFromAsymptomatic = parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区确诊病例(来自无症状感染者): %#v", cs.Date.Format("2006-01-02"), cs.DistrictConfirmedFromAsymptomatic)
	}
	//	隔离管控 => 无症状
	mm = reDailyRegionAsymptomaticFromBubble.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalAsymptomaticFromBubble > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区无症状感染者(来自隔离管控)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictAsymptomaticFromBubble = parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区无症状感染者(来自隔离管控): %#v", cs.Date.Format("2006-01-02"), cs.DistrictAsymptomaticFromBubble)
	}
	//	风险人群 => 无症状
	mm = reDailyRegionAsymptomaticFromRisk.FindAllStringSubmatch(content, -1)
	if mm == nil {
		if d.LocalAsymptomaticFromRisk > 0 {
			log.Warnf("[%s] 无法解析文章内容中城区无症状感染者(来自风险人群)", d.Date.Format("2006-01-02"))
		}
	} else {
		d.DistrictAsymptomaticFromRisk = parseDailyContentRegionItems(d, mm)
		// log.Warnf("[%s] 城区无症状感染者(来自风险人群): %#v", cs.Date.Format("2006-01-02"), cs.DistrictAsymptomaticFromRisk)
	}

	return err
}

func parseDailyContentRegionItems(d *model.Daily, mm [][]string) map[string]int {
	dict := make(map[string]int)
	for _, m := range mm {
		items := parseDailyContentRegionItem(d, m[0])
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

func parseDailyContentRegionItem(d *model.Daily, text string) map[string]int {
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

func fixDaily(d *model.Daily) error {
	//	无症状
	if d.Asymptomatic == 0 {
		if d.LocalAsymptomatic != 0 || d.ImportedAsymptomatic != 0 {
			d.Asymptomatic = d.LocalAsymptomatic + d.ImportedAsymptomatic
		}
	} else {
		if d.Asymptomatic != (d.LocalAsymptomatic + d.ImportedAsymptomatic) {
			log.Warnf("[%s] 无症状数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"), d.Asymptomatic, d.LocalAsymptomatic, d.ImportedAsymptomatic)
		}
	}
	if d.LocalAsymptomaticFromBubble != 0 {
		//	检查分区数量
		c := 0
		for _, v := range d.DistrictAsymptomaticFromBubble {
			c += v
		}
		if d.LocalAsymptomaticFromBubble != c {
			log.Warnf("[%s] 无症状(来自隔离管控)数据不匹配：总共:%d (%d) (分区: %v)", d.Date.Format("2006-01-02"), d.LocalAsymptomaticFromBubble, c, d.DistrictAsymptomaticFromBubble)
		}
	}
	if d.LocalAsymptomaticFromRisk == 0 {
		d.LocalAsymptomaticFromRisk = d.LocalAsymptomatic - d.LocalAsymptomaticFromBubble
		if d.LocalAsymptomaticFromRisk < 0 {
			log.Warnf("[%s] 无症状(来自风险人群)数据不合理：本土无症状: %d => 来自隔离管控: %d + 来自风险人群: %d", d.Date.Format("2006-01-02"), d.LocalAsymptomatic, d.LocalAsymptomaticFromBubble, d.LocalAsymptomaticFromRisk)
		}
	} else {
		//	检查分区数量
		r := 0
		for _, v := range d.DistrictAsymptomaticFromRisk {
			r += v
		}
		if d.LocalAsymptomaticFromRisk != r {
			log.Warnf("[%s] 无症状(来自风险人群)数据不匹配：总共:%d (分区: %v)", d.Date.Format("2006-01-02"), d.LocalAsymptomaticFromRisk, d.DistrictAsymptomaticFromRisk)
		}
	}
	if d.DistrictAsymptomatic == nil {
		d.DistrictAsymptomatic = make(map[string]int)
		for r, v := range d.DistrictAsymptomaticFromBubble {
			d.DistrictAsymptomatic[r] = v
		}
		for r, v := range d.DistrictAsymptomaticFromRisk {
			if val, ok := d.DistrictAsymptomatic[r]; ok {
				d.DistrictAsymptomatic[r] = val + v
			} else {
				d.DistrictAsymptomatic[r] = v
			}
		}
	}

	//	确诊
	if d.Confirmed == 0 {
		if d.LocalConfirmed != 0 || d.ImportedConfirmed != 0 {
			d.Confirmed = d.LocalConfirmed + d.ImportedConfirmed
		}
	} else {
		if d.Confirmed != (d.LocalConfirmed + d.ImportedConfirmed) {
			log.Warnf("[%s] 确诊数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"), d.Confirmed, d.LocalConfirmed, d.ImportedConfirmed)
		}
	}
	if d.LocalConfirmedFromAsymptomatic != 0 {
		//	检查分区数量
		c := 0
		for _, v := range d.DistrictConfirmedFromAsymptomatic {
			c += v
		}
		if d.LocalConfirmedFromAsymptomatic != c {
			log.Warnf("[%s] 确证病例(来自无症状)数据不匹配：总共:%d (分区: %v)", d.Date.Format("2006-01-02"), d.LocalConfirmedFromAsymptomatic, d.DistrictConfirmedFromAsymptomatic)
		}
	}
	if d.LocalConfirmedFromBubble != 0 {
		//	检查分区数量
		c := 0
		for _, v := range d.DistrictConfirmedFromBubble {
			c += v
		}
		if d.LocalConfirmedFromBubble != c {
			log.Warnf("[%s] 确证病例(来自隔离管控)数据不匹配：总共:%d (分区: %v)", d.Date.Format("2006-01-02"), d.LocalConfirmedFromBubble, d.DistrictConfirmedFromBubble)
		}
	}
	if d.LocalConfirmedFromRisk == 0 {
		d.LocalConfirmedFromRisk = d.LocalConfirmed - (d.LocalConfirmedFromBubble + d.LocalConfirmedFromAsymptomatic)
		if d.LocalConfirmedFromRisk < 0 {
			log.Warnf("[%s] 确证病例(来自风险人群)数据不合理：本土确证病例: %d => 来自隔离管控: %d + 来自风险人群: %d", d.Date.Format("2006-01-02"), d.LocalConfirmed, d.LocalConfirmedFromBubble, d.LocalConfirmedFromRisk)
		}
	} else {
		//	检查分区数量
		r := 0
		for _, v := range d.DistrictConfirmedFromRisk {
			r += v
		}
		if d.LocalConfirmedFromRisk != r {
			log.Warnf("[%s] 确证病例(来自风险人群)数据不匹配：总共:%d (分区: %v)", d.Date.Format("2006-01-02"), d.LocalConfirmedFromRisk, d.DistrictConfirmedFromRisk)
		}
	}
	if d.DistrictConfirmed == nil {
		d.DistrictConfirmed = make(map[string]int)
		for r, v := range d.DistrictConfirmedFromBubble {
			d.DistrictConfirmed[r] = v
		}
		for r, v := range d.DistrictConfirmedFromRisk {
			if val, ok := d.DistrictConfirmed[r]; ok {
				d.DistrictConfirmed[r] = val + v
			} else {
				d.DistrictConfirmed[r] = v
			}
		}
		for r, v := range d.DistrictConfirmedFromAsymptomatic {
			if val, ok := d.DistrictConfirmed[r]; ok {
				d.DistrictConfirmed[r] = val + v
			} else {
				d.DistrictConfirmed[r] = v
			}
		}
	}

	// 治愈出院
	if d.DischargedFromHospital == 0 {
		if d.LocalDischargedFromHospital != 0 || d.ImportedDischargedFromHospital != 0 {
			d.DischargedFromHospital = d.LocalDischargedFromHospital + d.ImportedDischargedFromHospital
		}
	} else {
		if d.DischargedFromHospital != (d.LocalDischargedFromHospital + d.ImportedDischargedFromHospital) {
			if d.LocalDischargedFromHospital == 0 && d.ImportedDischargedFromHospital > 0 {
				//	应该是没能解析出本土治愈出院，可以计算获得
				d.LocalDischargedFromHospital = d.DischargedFromHospital - d.ImportedDischargedFromHospital
			} else if d.LocalDischargedFromHospital > 0 && d.ImportedDischargedFromHospital == 0 {
				//  应该是没能解析出境外输入治愈出院，可以计算获得
				d.ImportedDischargedFromHospital = d.DischargedFromHospital - d.LocalDischargedFromHospital
			} else {
				//	三者均不为0，因此必然是出错了。
				log.Warnf("[%s] 治愈出院数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"), d.DischargedFromHospital, d.LocalDischargedFromHospital, d.ImportedDischargedFromHospital)
			}
		}
	}

	// 解除医学观察
	if d.DischargedFromMedicalObservation == 0 {
		if d.LocalDischargedFromMedicalObservation != 0 || d.ImportedDischargedFromMedicalObservation != 0 {
			d.DischargedFromMedicalObservation = d.LocalDischargedFromMedicalObservation + d.ImportedDischargedFromMedicalObservation
		}
	} else {
		if d.DischargedFromMedicalObservation != (d.LocalDischargedFromMedicalObservation + d.ImportedDischargedFromMedicalObservation) {
			log.Warnf("[%s] 解除医学观察数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"), d.DischargedFromMedicalObservation, d.LocalDischargedFromMedicalObservation, d.ImportedDischargedFromMedicalObservation)
		}
	}

	// 死亡
	if d.Death == 0 {
		if d.LocalDeath != 0 || d.ImportedDeath != 0 {
			d.Death = d.LocalDeath + d.ImportedDeath
		}
	} else {
		if d.Death != (d.LocalDeath + d.ImportedDeath) {
			log.Warnf("[%s] 死亡数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"), d.Death, d.LocalDeath, d.ImportedDeath)
		}
	}

	// 在院治疗
	if d.InHospital == 0 {
		if d.LocalInHospital != 0 || d.ImportedInHospital != 0 {
			d.InHospital = d.LocalInHospital + d.ImportedInHospital
		}
	} else {
		if d.InHospital != (d.LocalInHospital + d.ImportedInHospital) {
			log.Warnf("[%s] 在院治疗数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"), d.InHospital, d.LocalInHospital, d.ImportedInHospital)
		}
	}

	// 尚在医疗观察
	if d.UnderMedicalObservation == 0 {
		if d.LocalUnderMedicalObservation != 0 || d.ImportedUnderMedicalObservation != 0 {
			d.UnderMedicalObservation = d.LocalUnderMedicalObservation + d.ImportedUnderMedicalObservation
		}
	} else {
		if d.UnderMedicalObservation != (d.LocalUnderMedicalObservation + d.ImportedUnderMedicalObservation) {
			log.Warnf("[%s] 尚在医疗观察数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"),
				d.UnderMedicalObservation, d.LocalUnderMedicalObservation, d.ImportedUnderMedicalObservation)
		}
	}

	//	本土确诊、出院、死亡、住院
	if d.LocalInHospital+d.TotalLocalDischargedFromHospital+d.TotalLocalDeath != d.TotalLocalConfirmed {
		log.Warnf("[%s] 本土确诊、出院、死亡、住院数据不匹配：累计本土确诊(%d) => 本土在院治疗(%d) + 累计本土治愈出院(%d) + 累计本土死亡(%d)", d.Date.Format("2006-01-02"),
			d.TotalLocalConfirmed, d.LocalInHospital, d.TotalLocalDischargedFromHospital, d.TotalLocalDeath)
	}

	//	境外输入确诊、出院、死亡、住院
	if d.ImportedInHospital+d.TotalImportedDischargedFromHospital != d.TotalImportedConfirmed {
		log.Warnf("[%s] 境外输入确诊、出院、死亡、住院数据不匹配：累计境外输入确诊(%d) => 境外输入在院治疗(%d) + 累计境外输入治愈出院(%d)", d.Date.Format("2006-01-02"),
			d.TotalImportedConfirmed, d.ImportedInHospital, d.TotalImportedDischargedFromHospital)
	}

	return nil
}

var (
	reResidentDistrict1 = regexp.MustCompile(`(?:\n)(?P<district>[^\d\n：]+区)(?:\n[^\n]+(?:(?:\n分别)?居住于[^\n]?|）))*(?P<addrs>(?:\n[^\n2已][^\n]+[，。、]?)+)?`)
	reResidentDistrict2 = regexp.MustCompile(`(?P<type>病例|无症状感染者)(?P<number>\d+)，(?P<gender>男|女)，(?P<age>\d+月?)[岁龄]，(?:[^，]+，)?居住(?:于|地为)(?P<district>[^，。]+区)?(?P<addr>[^，。]+)`)
)

func parseResidents(rs *model.Residents, date time.Time, content string) error {
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
	if date.Before(DATE_RESIDENT_MERGED) {
		//	2022年3月18日之前的居住地信息是包含于疫情通告中的

		mm = reResidentDistrict2.FindAllStringSubmatch(content, -1)
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
		mm = reResidentDistrict1.FindAllStringSubmatch(content, -1)
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

//	Listener functions

///	Daily
func (c *DailyCrawler) AddOnDailyListener(f func(model.Daily)) {
	if f == nil {
		log.Warn("DailyCrawler.AddOnDailyListener(): couldn't add 'nil' as listener.")
		return
	}

	c.listenersDaily = append(c.listenersDaily, f)
}

func (c *DailyCrawler) ClearOnDailyListener() {
	c.listenersDaily = []func(model.Daily){}
}

func (c *DailyCrawler) notifyAllOnDailyListeners(h model.Daily) {
	for _, listener := range c.listenersDaily {
		listener(h)
	}
}

/// Residents
func (c *DailyCrawler) AddOnResidentsListener(f func(model.Residents)) {
	if f == nil {
		log.Warn("DailyCrawler.AddOnResidentsListener(): couldn't add 'nil' as listener.")
		return
	}

	c.listenersResidents = append(c.listenersResidents, f)
}

func (c *DailyCrawler) ClearOnResidentsListener() {
	c.listenersResidents = []func(model.Residents){}
}

func (c *DailyCrawler) notifyAllOnResidentsListeners(h model.Residents) {
	for _, listener := range c.listenersResidents {
		listener(h)
	}
}
