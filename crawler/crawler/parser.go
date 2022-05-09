package crawler

import (
	"crawler/model"
	"regexp"
	"time"
)

type DailyParser interface {
	GetSelector(t string) string
	GetItemLinks() []string
	GetIndexLinks() []string
	GetDistricts() []string

	IsValidTitle(title string) bool
	IsDaily(date time.Time, title string) bool
	IsResidents(date time.Time, title string) bool

	ParseDailyTitle(d *model.Daily, title string) error
	ParseDailyContent(d *model.Daily, content string) error

	ParseResidents(rs *model.Residents, date time.Time, content string) error
}

var (
	reDailyDate                 = regexp.MustCompile(`(?:^|[：】海京]+)(?P<date>(?:\d+年)?\d+月\d+日)(?:[，（]+|0—24时|0时至24时|新增|[^，\n]+新增)`)
	reDailyLocalConfirmed       = regexp.MustCompile(`(?:[^累计\n]+本土[新冠肺炎]*确诊病例|新增)(?P<number>\d+)(?:例|例本土[新冠肺炎]*确诊[病例]*)(?:[和、，。（ ]|$)`)
	reDailyLocalAsymptomatic    = regexp.MustCompile(`(?:新增|[\s，、和])(?:[本土]*无症状感染者)?(?P<number>\d+)(?:例|例[本土]*无症状感染者)(?:[、，。（ 和]|$)`)
	reDailyImportedConfirmed    = regexp.MustCompile(`境外输入(?:性新冠肺炎确诊)?(?:病例)?(?P<number>\d+)例`)
	reDailyImportedAsymptomatic = regexp.MustCompile(`(?:[新增]*境外输入性?无症状感染者|和)(?P<number>\d+)例(?:境外输入无症状感染者)?`)
)

var (
	reDailyDischargedFromHospital           = regexp.MustCompile(`治愈出院(?P<number>\d+)例`)
	reDailyDischargedFromMedicalObservation = regexp.MustCompile(`解除医学观察(?:无症状感染者)?(?P<number>\d+)例`)
)

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
