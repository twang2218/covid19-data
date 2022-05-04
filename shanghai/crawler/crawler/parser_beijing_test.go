package crawler

import (
	"crawler/model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func s2date(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestDailyParserBeijing_ParseResidents(t *testing.T) {
	tests := []struct {
		name    string
		content string
		rs      model.Residents
	}{
		{
			name:    "单一病例",
			content: "确诊病例2：现住朝阳区松榆东里。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，临床分型为轻型。",
			rs: model.Residents{
				{Date: s2date("2022-04-30"), Name: "轻型2", Type: "轻型", City: "北京市", District: "朝阳区", Address: "松榆东里"},
			},
		},
		{
			name:    "单一病例 - 日期多了空格",
			content: "确诊病例18：现住通州区永顺镇馨通家园。综合流行病史、临床表现、实验室检测和影像学检查等结果，5 月1日诊断为确诊病例，临床分型为轻型。",
			rs: model.Residents{
				{Date: s2date("2022-05-01"), Name: "轻型18", Type: "轻型", City: "北京市", District: "通州区", Address: "永顺镇馨通家园"},
			},
		},
		{
			name:    "2个独立病例 - 同为无症状感染者",
			content: "无症状感染者1、2：现住房山区窦店镇于庄村。4月23日作为密切接触者进行核酸检测，当日报告结果为阳性，已转至定点医院，综合流行病史、临床表现、实验室检测和影像学检查等结果，4月24日均诊断为无症状感染者。",
			rs: model.Residents{
				{Date: s2date("2022-04-24"), Name: "无症状感染者1", Type: "无症状感染者", City: "北京市", District: "房山区", Address: "窦店镇于庄村"},
				{Date: s2date("2022-04-24"), Name: "无症状感染者2", Type: "无症状感染者", City: "北京市", District: "房山区", Address: "窦店镇于庄村"},
			},
		},
		{
			name:    "2个独立病例 - 相同分型",
			content: "确诊病例1、9：为同一家庭成员，现住通州区北苑街道新仓路小区。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，临床分型均为轻型。",
			rs: []model.Resident{
				{Date: s2date("2022-04-30"), Name: "轻型1", Type: "轻型", City: "北京市", District: "通州区", Address: "北苑街道新仓路小区"},
				{Date: s2date("2022-04-30"), Name: "轻型9", Type: "轻型", City: "北京市", District: "通州区", Address: "北苑街道新仓路小区"},
			},
		},
		{
			name:    "2个独立病例 - 不同分型",
			content: "确诊病例49 、50：现住朝阳区双井街道广和南里二条。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，确诊病例49临床分型为轻型，确诊病例50临床分型为普通型。",
			rs: []model.Resident{
				{Date: s2date("2022-04-30"), Name: "轻型49", Type: "轻型", City: "北京市", District: "朝阳区", Address: "双井街道广和南里二条"},
				{Date: s2date("2022-04-30"), Name: "普通型50", Type: "普通型", City: "北京市", District: "朝阳区", Address: "双井街道广和南里二条"},
			},
		},
		{
			name:    "2个独立病例 - 货车司机",
			content: "确诊病例1、2：均为3月27日通报的确诊病例货车司机的同车人员。由外省途经首都环线高速，进入通州区西集服务区，3月26日作为密切接触者进行集中隔离，3月29日报告核酸检测结果均为阳性，已转至定点医院，综合流行病史、临床表现、实验室检测和影像学检查等结果，当日诊断为确诊病例，临床分型均为轻型。",
			rs: []model.Resident{
				{Name: "轻型1", Type: "轻型", City: "北京市", District: "通州区", Address: "西集服务区"},
				{Name: "轻型2", Type: "轻型", City: "北京市", District: "通州区", Address: "西集服务区"},
			},
		},
		{
			name:    "3个独立病例 - 相同分型",
			content: "确诊病例10、44、45：现住址均位于朝阳区，在校学生。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，临床分型均为轻型。",
			rs: []model.Resident{
				{Date: s2date("2022-04-30"), Name: "轻型10", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Date: s2date("2022-04-30"), Name: "轻型44", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Date: s2date("2022-04-30"), Name: "轻型45", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
			},
		},
		{
			name:    "3个独立病例 - 不同分型1",
			content: "确诊病例6、8、9：现住址均位于石景山区。综合流行病史、临床表现、实验室检测和影像学检查等结果，5月1日诊断为确诊病例，临床分型分别为普通型、轻型、轻型。",
			rs: []model.Resident{
				{Date: s2date("2022-05-01"), Name: "普通型6", Type: "普通型", City: "北京市", District: "石景山区", Address: ""},
				{Date: s2date("2022-05-01"), Name: "轻型8", Type: "轻型", City: "北京市", District: "石景山区", Address: ""},
				{Date: s2date("2022-05-01"), Name: "轻型9", Type: "轻型", City: "北京市", District: "石景山区", Address: ""},
			},
		},
		{
			name:    "3个独立病例 - 不同分型2",
			content: "确诊病例14、15、16：均为朝阳区护国寺小吃（光明桥店）员工。4月25日报告核酸检测结果均为阳性，已转至定点医院，综合流行病史、临床表现、实验室检测和影像学检查等结果，当日诊断为确诊病例，确诊病例14临床分型为普通型，确诊病例15、16临床分型均为轻型。",
			rs: []model.Resident{
				{Name: "普通型14", Type: "普通型", City: "北京市", District: "朝阳区", Address: "护国寺小吃（光明桥店）员工"},
				{Name: "轻型15", Type: "轻型", City: "北京市", District: "朝阳区", Address: "护国寺小吃（光明桥店）员工"},
				{Name: "轻型16", Type: "轻型", City: "北京市", District: "朝阳区", Address: "护国寺小吃（光明桥店）员工"},
			},
		},
		{
			name:    "病例范围",
			content: "确诊病例44至48：现住朝阳区建外街道光辉里小区。5月2日诊断为确诊病例，临床分型均为轻型。",
			rs: model.Residents{
				{Date: s2date("2022-05-02"), Name: "轻型44", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "轻型45", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "轻型46", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "轻型47", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "轻型48", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
			},
		},
		{
			name:    "多种情况综合",
			content: "感染者231、234、244至246、248、252、256、261：现住址均位于朝阳区，在校学生。4月23日作为感染者146的密切接触者进行集中隔离，4月25日、26日报告核酸检测结果均为阳性，已转至定点医院，4月26日感染者246诊断为无症状感染者，感染者231、234、244、245、248当日诊断为确诊病例，感染者231、234、244、245临床分型均为轻型，感染者248临床分型为普通型，4月27日感染者252、256、261诊断为确诊病例，临床分型均为轻型。",
			rs: model.Residents{
				{Name: "轻型231", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "轻型234", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "轻型244", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "轻型245", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "普通型248", Type: "普通型", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "无症状感染者246", Type: "无症状感染者", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "轻型252", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "轻型256", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Name: "轻型261", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
			},
		},
	}

	// tests := []string{
	// 	"确诊病例1、9：为同一家庭成员，现住通州区北苑街道新仓路小区。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，临床分型均为轻型。",
	// }

	p := DailyParserBeijing{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rs model.Residents
			var date time.Time
			err := p.ParseResidents(&rs, date, tt.content)
			assert.NoError(t, err)
			assert.EqualValues(t, tt.rs, rs)
		})
	}
}
