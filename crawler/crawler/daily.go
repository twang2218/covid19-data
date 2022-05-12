package crawler

import (
	"crawler/model"
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

	parser             DailyParser
	cItem              *colly.Collector
	cIndex             *colly.Collector
	listenersDaily     []func(model.Daily)
	listenersResidents []func(model.Residents)
}

type OnDailyListener interface {
	OnDaily(model.Daily)
}

type OnResidentsListener interface {
	OnResidents(model.Residents)
}

func NewDailyCrawler(city, cache_dir string) *DailyCrawler {
	dc := DailyCrawler{}
	switch city {
	case "shanghai":
		dc.parser = DailyParserShanghai{}
	case "beijing":
		dc.parser = DailyParserBeijing{}
	}
	dc.init(cache_dir)
	return &dc
}

func (dc *DailyCrawler) init(cache_dir string) {
	dc.cItem = colly.NewCollector(
		colly.DetectCharset(),
		colly.CacheDir(cache_dir),
		colly.Async(true),
		// colly.Debugger(&debug.LogDebugger{}),
	)
	dc.cItem.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: CRAWLER_PARALLELISM,
	})
	dc.cItem.SetRequestTimeout(CRAWLER_REQUEST_TIMEOUT)
	extensions.RandomUserAgent(dc.cItem)
	dc.cItem.OnRequest(func(r *colly.Request) {
		r.ResponseCharacterEncoding = "utf-8"
		atomic.AddInt32(&dc.PageTotal, 1)
		// log.Tracef("DailyCrawler.init(): cItem.Request(): %s\n", r.URL)
	})
	dc.cItem.OnScraped(func(r *colly.Response) {
		atomic.AddInt32(&dc.PageVisited, 1)
	})

	dc.cItem.OnError(func(resp *colly.Response, err error) {
		log.Warnf("DailyCrawler.OnError(): [Item] (%s) => '%s'", resp.Request.URL, err)
		//	重试。在另外线程等待一段时间，以不阻碍当前线程（爬虫）运行。
		go func(link string) {
			time.Sleep(CRAWLER_RETRY_TIMEOUT)
			dc.cItem.Visit(link)
		}(resp.Request.URL.String())
	})
	//	添加解析具体页面函数
	///	疫情通报、居住地信息; 微信居住地信息
	dc.cItem.OnHTML(dc.parser.GetSelector("item"), dc.ParseItem)

	//	创建索引爬虫
	dc.cIndex = dc.cItem.Clone()
	extensions.RandomUserAgent(dc.cIndex)
	dc.cIndex.OnRequest(func(r *colly.Request) {
		// r.ResponseCharacterEncoding = "gb2312"
		// log.Tracef("DailyCrawler => cIndex.Request(): %s", r.URL)
	})
	dc.cIndex.OnError(func(resp *colly.Response, err error) {
		log.Warnf("DailyCrawler.OnError(): [Index] (%s) => '%s'", resp.Request.URL, err)
		//	重试。在另外线程等待一段时间，以不阻碍当前线程（爬虫）运行。
		go func(link string) {
			time.Sleep(CRAWLER_RETRY_TIMEOUT)
			dc.cIndex.Visit(link)
		}(resp.Request.URL.String())
	})
	//	添加索引解析函数
	dc.cIndex.OnHTML(dc.parser.GetSelector("index"), dc.ParseIndex)
}

func (c *DailyCrawler) Collect() {
	//	先抓取指定内容页面
	for _, l := range c.parser.GetItemLinks() {
		c.cItem.Visit(l)
	}
	//	再抓取索引页面
	for _, l := range c.parser.GetIndexLinks() {
		c.cIndex.Visit(l)
	}
	//	等待结束
	c.cIndex.Wait()
	c.cItem.Wait()
}

func (c *DailyCrawler) ParseItem(e *colly.HTMLElement) {
	var d model.Daily
	d.Source = e.Request.URL.String()

	// 标题
	title := strings.TrimSpace(e.ChildText(c.parser.GetSelector("title")))
	// log.Tracef("DailyCrawler.parseItem(%s): %s => %q\n", e.Attr("id"), e.Request.URL, title)

	if err := c.parser.ParseDailyTitle(&d, title); err != nil {
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
	e.ForEach(c.parser.GetSelector("content"), func(i int, h *colly.HTMLElement) {
		t := strings.TrimSpace(h.Text)
		if len(t) > 0 {
			content_lines = append(content_lines, t)
		}
	})
	content := strings.Join(content_lines, "\n")

	// log.Tracef("[%s] <%s>: %s", d.Date.Format("2006-01-02"), title, d.Source)

	if c.parser.IsDaily(d.Date, title) {
		// log.Tracef("[%s] <%s>: 每日疫情信息统计", d.Date.Format("2006-01-02"), title)
		//	每日疫情信息统计
		if err := c.parser.ParseDailyContent(&d, content); err != nil {
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

		c.FixDaily(&d)

		//	通知 OnDailyListeners
		//	使用 defer 将提交推迟到整个函数运行结束，这样给 c.FixDailyByResidents() 一个机会来修复缺失数据
		defer func() { c.OnDaily(d) }()
	}
	if c.parser.IsResidents(d.Date, title) {
		// log.Tracef("[%s] <%s>: 居住地信息统计", d.Date.Format("2006-01-02"), title)
		//	居住地信息统计
		rs := make(model.Residents, 0)
		if err := c.parser.ParseResidents(&rs, d.Date, content); err != nil {
			log.Errorf("[%s] 解析感染者居住地信息失败：%s => %q", d.Date.Format("2006-01-02"), err, title)
		} else {
			// log.Infof("[%s] 解析到 %d 个感染病例居住地信息。", d.Date.Format("2006-01-02"), len(rs))
			c.OnResidents(rs)

			// 此时还有机会通过居住地信息补充 Daily 缺失信息
			c.FixDailyByResidents(&d, rs)
		}
		// log.Tracef("[%s] <%s>: 居住地信息统计: rs => [%d]", d.Date.Format("2006-01-02"), title, len(rs))
	}
}

func (c *DailyCrawler) ParseIndex(e *colly.HTMLElement) {
	link := e.Request.AbsoluteURL(strings.TrimSpace(e.Attr("href")))
	title := e.Text
	// log.Tracef("DailyCrawler.ParseIndex(): %s => %s", title, link)
	if c.parser.IsValidTitle(title) {
		//	告知 cItem 抓取该链接
		c.cItem.Visit(link)
	}
}

func (c *DailyCrawler) FixDaily(d *model.Daily) error {
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
			log.Warnf("[%s] 无症状(来自隔离管控)数据不匹配：总共:%d => %d: (分区: %v)",
				d.Date.Format("2006-01-02"),
				d.LocalAsymptomaticFromBubble,
				c,
				d.DistrictAsymptomaticFromBubble)
		}
	}
	if d.LocalAsymptomaticFromRisk == 0 {
		if d.LocalAsymptomaticFromBubble > 0 {
			d.LocalAsymptomaticFromRisk = d.LocalAsymptomatic - d.LocalAsymptomaticFromBubble
			if d.LocalAsymptomaticFromRisk < 0 {
				log.Warnf("[%s] 无症状(来自风险人群)数据不合理：本土无症状: %d => 来自隔离管控: %d + 来自风险人群: %d", d.Date.Format("2006-01-02"), d.LocalAsymptomatic, d.LocalAsymptomaticFromBubble, d.LocalAsymptomaticFromRisk)
			}
		}
	} else {
		//	检查分区数量
		r := 0
		for _, v := range d.DistrictAsymptomaticFromRisk {
			r += v
		}
		if d.LocalAsymptomaticFromRisk != r {
			log.Warnf("[%s] 无症状(来自风险人群)数据不匹配：总共:%d => %d: (分区: %v)",
				d.Date.Format("2006-01-02"),
				d.LocalAsymptomaticFromRisk,
				r,
				d.DistrictAsymptomaticFromRisk)
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
	mcsc := d.Mild + d.Common + d.Severe + d.Critical
	if d.LocalConfirmed == 0 {
		if d.Mild != 0 || d.Common != 0 || d.Severe != 0 || d.Critical != 0 {
			d.LocalConfirmed = mcsc
		}
	} else {
		//	当临床分型数据存在，但是本地确诊数据却不一致的时候报错
		if (d.Mild != 0 || d.Common != 0 || d.Severe != 0 || d.Critical != 0) &&
			d.LocalConfirmed != mcsc {
			log.Warnf("[%s] 本土确诊数据不匹配：本土确诊:%d => %d: (轻型:%d / 普通型:%d / 重型:%d / 危重型:%d)",
				d.Date.Format("2006-01-02"),
				d.LocalConfirmed,
				mcsc,
				d.Mild, d.Common, d.Severe, d.Critical)
		}
	}
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
			log.Warnf("[%s] 确证病例(来自无症状)数据不匹配：总共:%d => %d: (分区: %v)",
				d.Date.Format("2006-01-02"),
				d.LocalConfirmedFromAsymptomatic,
				c,
				d.DistrictConfirmedFromAsymptomatic)
		}
	}
	if d.LocalConfirmedFromBubble != 0 {
		//	检查分区数量
		c := 0
		for _, v := range d.DistrictConfirmedFromBubble {
			c += v
		}
		if d.LocalConfirmedFromBubble != c {
			log.Warnf("[%s] 确证病例(来自隔离管控)数据不匹配：总共:%d => %d: (分区: %v)",
				d.Date.Format("2006-01-02"),
				d.LocalConfirmedFromBubble,
				c,
				d.DistrictConfirmedFromBubble)
		}
	}
	if d.LocalConfirmedFromRisk == 0 {
		if d.LocalConfirmedFromBubble > 0 || d.LocalConfirmedFromAsymptomatic > 0 {
			d.LocalConfirmedFromRisk = d.LocalConfirmed - (d.LocalConfirmedFromBubble + d.LocalConfirmedFromAsymptomatic)
			if d.LocalConfirmedFromRisk < 0 {
				log.Warnf("[%s] 确证病例(来自风险人群)数据不合理：本土确证病例: %d => 来自隔离管控: %d + 来自风险人群: %d", d.Date.Format("2006-01-02"), d.LocalConfirmed, d.LocalConfirmedFromBubble, d.LocalConfirmedFromRisk)
			}
		}
	} else {
		//	检查分区数量
		r := 0
		for _, v := range d.DistrictConfirmedFromRisk {
			r += v
		}
		if d.LocalConfirmedFromRisk != r {
			log.Warnf("[%s] 确证病例(来自风险人群)数据不匹配：总共:%d => %d: (分区: %v)",
				d.Date.Format("2006-01-02"),
				d.LocalConfirmedFromRisk,
				r,
				d.DistrictConfirmedFromRisk)
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

	//	阳性感染者
	if d.LocalPositive == 0 {
		if d.LocalConfirmed != 0 || d.LocalAsymptomatic != 0 {
			d.LocalPositive = d.LocalConfirmed + d.LocalAsymptomatic
		}
	} else {
		if d.LocalPositive != (d.LocalConfirmed + d.LocalAsymptomatic) {
			log.Warnf("[%s] 本土阳性感染者数据不匹配：本土阳性:%d (确诊:%d / 无症状:%d)", d.Date.Format("2006-01-02"), d.LocalPositive, d.LocalConfirmed, d.LocalAsymptomatic)
		}
	}
	if d.ImportedPositive == 0 {
		if d.ImportedConfirmed != 0 || d.ImportedAsymptomatic != 0 {
			d.ImportedPositive = d.ImportedConfirmed + d.ImportedAsymptomatic
		}
	} else {
		if d.ImportedPositive != (d.ImportedConfirmed + d.ImportedAsymptomatic) {
			log.Warnf("[%s] 境外输入阳性感染者数据不匹配：境外输入阳性:%d (确诊:%d / 无症状:%d)", d.Date.Format("2006-01-02"), d.ImportedPositive, d.ImportedConfirmed, d.ImportedAsymptomatic)
		}
	}

	if d.Positive == 0 {
		if d.Confirmed != 0 || d.Asymptomatic != 0 {
			d.Positive = d.Confirmed + d.Asymptomatic
		}
	} else {
		if d.Positive != (d.Confirmed + d.Asymptomatic) {
			log.Warnf("[%s] 阳性感染者数据不匹配：总共:%d (确诊:%d / 无症状:%d)", d.Date.Format("2006-01-02"), d.Positive, d.Confirmed, d.Asymptomatic)
		}
	}
	if d.LocalPositiveFromBubble != 0 {
		//	检查分区数量
		c := 0
		for _, v := range d.DistrictPositiveFromBubble {
			c += v
		}
		if d.LocalPositiveFromBubble != c {
			log.Warnf("[%s] 阳性感染者(来自隔离管控)数据不匹配：总共:%d => (%d): (分区: %v)",
				d.Date.Format("2006-01-02"),
				d.LocalPositiveFromBubble,
				c,
				d.DistrictPositiveFromBubble)
		}
	}
	if d.LocalPositiveFromRisk == 0 {
		if d.LocalPositiveFromBubble > 0 {
			d.LocalPositiveFromRisk = d.LocalPositive - d.LocalPositiveFromBubble
			if d.LocalPositiveFromRisk < 0 {
				log.Warnf("[%s] 阳性感染者(来自风险人群)数据不合理：本土阳性感染者: %d => 来自隔离管控: %d + 来自风险人群: %d", d.Date.Format("2006-01-02"), d.LocalPositive, d.LocalPositiveFromBubble, d.LocalPositiveFromRisk)
			}
		}
	} else {
		//	检查分区数量
		r := 0
		for _, v := range d.DistrictPositiveFromRisk {
			r += v
		}
		if d.LocalPositiveFromRisk != r {
			log.Warnf("[%s] 阳性感染者(来自风险人群)数据不匹配：总共:%d => (%d): (分区: %v)",
				d.Date.Format("2006-01-02"),
				d.LocalPositiveFromRisk,
				r,
				d.DistrictPositiveFromRisk)
		}
	}
	if len(d.DistrictPositive) == 0 {
		d.DistrictPositive = make(map[string]int)
		if len(d.DistrictPositiveFromBubble) > 0 || len(d.DistrictPositiveFromRisk) > 0 {
			//	如果存在分项的阳性分区数据，就从这里计算整体阳性数据
			for r, v := range d.DistrictPositiveFromBubble {
				d.DistrictPositive[r] = v
			}
			for r, v := range d.DistrictPositiveFromRisk {
				if val, ok := d.DistrictPositive[r]; ok {
					d.DistrictPositive[r] = val + v
				} else {
					d.DistrictPositive[r] = v
				}
			}
		} else if len(d.DistrictConfirmed) > 0 || len(d.DistrictAsymptomatic) > 0 {
			//	如果存在确诊和无症状分区数据，从这里计算阳性数据
			for r, v := range d.DistrictConfirmed {
				d.DistrictPositive[r] = v
			}
			for r, v := range d.DistrictAsymptomatic {
				if val, ok := d.DistrictPositive[r]; ok {
					d.DistrictPositive[r] = val + v
				} else {
					d.DistrictPositive[r] = v
				}
			}
		}
	}

	// 治愈出院
	if d.DischargedFromHospital == 0 {
		if d.LocalDischargedFromHospital != 0 || d.ImportedDischargedFromHospital != 0 {
			d.DischargedFromHospital = d.LocalDischargedFromHospital + d.ImportedDischargedFromHospital
		}
	} else {
		if (d.DischargedFromHospital != (d.LocalDischargedFromHospital + d.ImportedDischargedFromHospital)) && (d.LocalDischargedFromHospital > 0 || d.ImportedDischargedFromHospital > 0) {
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
	if d.CurrentInHospital == 0 {
		if d.CurrentLocalInHospital != 0 || d.CurrentImportedInHospital != 0 {
			d.CurrentInHospital = d.CurrentLocalInHospital + d.CurrentImportedInHospital
		}
	} else {
		if d.CurrentInHospital != (d.CurrentLocalInHospital + d.CurrentImportedInHospital) {
			log.Warnf("[%s] 在院治疗数据不匹配：总共:%d (本土:%d / 境外输入:%d)", d.Date.Format("2006-01-02"), d.CurrentInHospital, d.CurrentLocalInHospital, d.CurrentImportedInHospital)
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
	if d.CurrentLocalInHospital+d.TotalLocalDischargedFromHospital+d.TotalLocalDeath != d.TotalLocalConfirmed {
		log.Warnf("[%s] 本土确诊、出院、死亡、住院数据不匹配：累计本土确诊(%d) => (%d): 本土在院治疗(%d) + 累计本土治愈出院(%d) + 累计本土死亡(%d)",
			d.Date.Format("2006-01-02"),
			d.TotalLocalConfirmed,
			d.CurrentLocalInHospital+d.TotalLocalDischargedFromHospital+d.TotalLocalDeath,
			d.CurrentLocalInHospital, d.TotalLocalDischargedFromHospital, d.TotalLocalDeath)
	}

	//	境外输入确诊、出院、死亡、住院
	if d.CurrentImportedInHospital+d.TotalImportedDischargedFromHospital != d.TotalImportedConfirmed {
		log.Warnf("[%s] 境外输入确诊、出院、死亡、住院数据不匹配：累计境外输入确诊(%d) => (%d): 境外输入在院治疗(%d) + 累计境外输入治愈出院(%d)",
			d.Date.Format("2006-01-02"),
			d.TotalImportedConfirmed,
			d.CurrentImportedInHospital+d.TotalImportedDischargedFromHospital,
			d.CurrentImportedInHospital, d.TotalImportedDischargedFromHospital)
	}

	return nil
}

func (c *DailyCrawler) FixDailyByResidents(d *model.Daily, rs model.Residents) error {
	// 可以从居住地信息统计分区数据
	if len(d.DistrictConfirmed) == 0 || len(d.DistrictAsymptomatic) == 0 {
		// log.Tracef("FixDailyByResidents(): %#v", *d)
		if d.DistrictAsymptomatic == nil {
			d.DistrictAsymptomatic = make(map[string]int)
			// log.Tracef("FixDailyByResidents() - DistrictAsymptomatic: %#v", *d)
		}
		if d.DistrictConfirmed == nil {
			d.DistrictConfirmed = make(map[string]int)
			// log.Tracef("FixDailyByResidents() - DistrictConfirmed: %#v", *d)
		}
		if d.DistrictPositive == nil {
			d.DistrictPositive = make(map[string]int)
			// log.Tracef("FixDailyByResidents() - DistrictPositive: %#v", *d)
		}
		for _, r := range rs {
			if r.Date.Equal(d.Date) && len(r.District) > 0 && len(r.Type) > 0 {
				if r.Type == "无症状感染者" {
					//	无症状感染者
					if val, ok := d.DistrictAsymptomatic[r.District]; ok {
						d.DistrictAsymptomatic[r.District] = val + 1
						d.DistrictPositive[r.District] = val + 1
					} else {
						d.DistrictAsymptomatic[r.District] = 1
						d.DistrictPositive[r.District] = 1
					}
				} else {
					//	轻型、普通型、重型、危重型
					if val, ok := d.DistrictConfirmed[r.District]; ok {
						d.DistrictConfirmed[r.District] = val + 1
						d.DistrictPositive[r.District] = val + 1
					} else {
						d.DistrictConfirmed[r.District] = 1
						d.DistrictPositive[r.District] = 1
					}
				}
			}
		}
	}
	// 从居住地信息统计分型数据
	if d.Confirmed > 0 && d.Mild == 0 && d.Common == 0 && d.Severe == 0 && d.Critical == 0 {
		for _, r := range rs {
			if r.Date.Equal(d.Date) {
				switch r.Type {
				case "轻型":
					d.Mild = d.Mild + 1
				case "普通型":
					d.Common = d.Common + 1
				case "重型":
					d.Severe = d.Severe + 1
				case "危重型":
					d.Critical = d.Critical + 1
				}
			}
		}
	}
	return nil
}

//	Listener functions

///	OnDailyListener
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

func (c *DailyCrawler) OnDaily(h model.Daily) {
	for _, listener := range c.listenersDaily {
		listener(h)
	}
}

/// OnResidentsListener
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

func (c *DailyCrawler) OnResidents(h model.Residents) {
	for _, listener := range c.listenersResidents {
		listener(h)
	}
}
