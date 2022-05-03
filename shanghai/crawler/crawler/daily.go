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

type DailyParser interface {
	GetSelector(t string) string
	GetItemLinks() []string
	GetIndexLinks() []string

	// ParseItem(ld OnDailyListener, lr OnResidentsListener) colly.HTMLCallback
	// ParseIndex(f func(string) error) colly.HTMLCallback

	IsValidTitle(title string) bool
	IsDaily(date time.Time, title string) bool
	IsResidents(date time.Time, title string) bool

	ParseDailyTitle(d *model.Daily, title string) error
	ParseDailyContent(d *model.Daily, content string) error
	FixDaily(d *model.Daily) error

	ParseResidents(rs *model.Residents, date time.Time, content string) error
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
		// log.Debugf("DailyCrawler => %s", r.URL)
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

		c.parser.FixDaily(&d)

		//	通知 OnDailyListeners
		c.OnDaily(d)
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
		}
		// log.Tracef("[%s] <%s>: 居住地信息统计: rs => [%d]", d.Date.Format("2006-01-02"), title, len(rs))
	}
}

func (c *DailyCrawler) ParseIndex(e *colly.HTMLElement) {
	link := e.Request.AbsoluteURL(strings.TrimSpace(e.Attr("href")))
	title := e.Text
	if c.parser.IsValidTitle(title) {
		//	告知 cItem 抓取该链接
		c.cItem.Visit(link)
	}
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
