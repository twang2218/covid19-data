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
				{Date: s2date("2022-04-30"), Name: "确诊病例2", Type: "轻型", City: "北京市", District: "朝阳区", Address: "松榆东里"},
			},
		},
		{
			name:    "单一病例 - 日期多了空格",
			content: "确诊病例18：现住通州区永顺镇馨通家园。综合流行病史、临床表现、实验室检测和影像学检查等结果，5 月1日诊断为确诊病例，临床分型为轻型。",
			rs: model.Residents{
				{Date: s2date("2022-05-01"), Name: "确诊病例18", Type: "轻型", City: "北京市", District: "通州区", Address: "永顺镇馨通家园"},
			},
		},
		{
			name:    "单一病例 - 地址以‘区’为结尾",
			content: "确诊病例42 ：现住通州区于家务乡于家务西里小区。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，临床分型为轻型。",
			rs: model.Residents{
				{Date: s2date("2022-04-30"), Name: "确诊病例42", Type: "轻型", City: "北京市", District: "通州区", Address: "于家务乡于家务西里小区"},
			},
		},
		{
			name:    "单一病例 - 地址以四字‘区’为结尾",
			content: "确诊病例19：现住通州区永顺镇榆东一街金地格林北区。4月26日报告核酸检测结果为阳性，综合流行病史、临床表现、实验室检测和影像学检查等结果，4月27日诊断为确诊病例，临床分型为轻型。",
			rs: model.Residents{
				{Date: s2date("2022-04-27"), Name: "确诊病例19", Type: "轻型", City: "北京市", District: "通州区", Address: "永顺镇榆东一街金地格林北区"},
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
				{Date: s2date("2022-04-30"), Name: "确诊病例1", Type: "轻型", City: "北京市", District: "通州区", Address: "北苑街道新仓路小区"},
				{Date: s2date("2022-04-30"), Name: "确诊病例9", Type: "轻型", City: "北京市", District: "通州区", Address: "北苑街道新仓路小区"},
			},
		},
		{
			name:    "2个独立病例 - 不同分型1",
			content: "确诊病例49 、50：现住朝阳区双井街道广和南里二条。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，确诊病例49临床分型为轻型，确诊病例50临床分型为普通型。",
			rs: []model.Resident{
				{Date: s2date("2022-04-30"), Name: "确诊病例49", Type: "轻型", City: "北京市", District: "朝阳区", Address: "双井街道广和南里二条"},
				{Date: s2date("2022-04-30"), Name: "确诊病例50", Type: "普通型", City: "北京市", District: "朝阳区", Address: "双井街道广和南里二条"},
			},
		},
		// {	//	暂时不解析 “感染者xxx” 格式的页面
		// 	name:    "2个独立病例 - 不同分型2",
		// 	content: "感染者386、389：现住昌平区沙河镇乐乎有朋公寓，公司职员。通过社区核酸筛查发现，4月29日感染者386诊断为确诊病例，临床分型为轻型，感染者389诊断为无症状感染者。",
		// 	rs: []model.Resident{
		// 		{Name: "确诊病例386", Type: "轻型", City: "北京市", District: "昌平区", Address: "沙河镇乐乎有朋公寓"},
		// 		{Name: "确诊病例389", Type: "无症状感染者", City: "北京市", District: "昌平区", Address: "沙河镇乐乎有朋公寓"},
		// 	},
		// },
		{
			name:    "2个独立病例 - 货车司机",
			content: "确诊病例1、2：均为3月27日通报的确诊病例货车司机的同车人员。由外省途经首都环线高速，进入通州区西集服务区，3月26日作为密切接触者进行集中隔离，3月29日报告核酸检测结果均为阳性，已转至定点医院，综合流行病史、临床表现、实验室检测和影像学检查等结果，当日诊断为确诊病例，临床分型均为轻型。",
			rs: []model.Resident{
				{Name: "确诊病例1", Type: "轻型", City: "北京市", District: "通州区", Address: "西集服务区"},
				{Name: "确诊病例2", Type: "轻型", City: "北京市", District: "通州区", Address: "西集服务区"},
			},
		},
		{
			name:    "3个独立病例 - 相同分型",
			content: "确诊病例10、44、45：现住址均位于朝阳区，在校学生。综合流行病史、临床表现、实验室检测和影像学检查等结果，4月30日诊断为确诊病例，临床分型均为轻型。",
			rs: []model.Resident{
				{Date: s2date("2022-04-30"), Name: "确诊病例10", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Date: s2date("2022-04-30"), Name: "确诊病例44", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
				{Date: s2date("2022-04-30"), Name: "确诊病例45", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
			},
		},
		{
			name:    "3个独立病例 - 不同分型1",
			content: "确诊病例6、8、9：现住址均位于石景山区。综合流行病史、临床表现、实验室检测和影像学检查等结果，5月1日诊断为确诊病例，临床分型分别为普通型、轻型、轻型。",
			rs: []model.Resident{
				{Date: s2date("2022-05-01"), Name: "确诊病例6", Type: "普通型", City: "北京市", District: "石景山区", Address: ""},
				{Date: s2date("2022-05-01"), Name: "确诊病例8", Type: "轻型", City: "北京市", District: "石景山区", Address: ""},
				{Date: s2date("2022-05-01"), Name: "确诊病例9", Type: "轻型", City: "北京市", District: "石景山区", Address: ""},
			},
		},
		{
			name:    "3个独立病例 - 不同分型2",
			content: "确诊病例14、15、16：均为朝阳区护国寺小吃（光明桥店）员工。4月25日报告核酸检测结果均为阳性，已转至定点医院，综合流行病史、临床表现、实验室检测和影像学检查等结果，当日诊断为确诊病例，确诊病例14临床分型为普通型，确诊病例15、16临床分型均为轻型。",
			rs: []model.Resident{
				{Name: "确诊病例14", Type: "普通型", City: "北京市", District: "朝阳区", Address: "护国寺小吃（光明桥店）员工"},
				{Name: "确诊病例15", Type: "轻型", City: "北京市", District: "朝阳区", Address: "护国寺小吃（光明桥店）员工"},
				{Name: "确诊病例16", Type: "轻型", City: "北京市", District: "朝阳区", Address: "护国寺小吃（光明桥店）员工"},
			},
		},
		{
			name:    "3个独立病例 - 不同分型3",
			content: "确诊病例40、41、42：现住丰台区王佐镇鑫湖家园。5月4日诊断均为确诊病例，临床分型均为轻型。",
			rs: []model.Resident{
				{Date: s2date("2022-05-04"), Name: "确诊病例40", Type: "轻型", City: "北京市", District: "丰台区", Address: "王佐镇鑫湖家园"},
				{Date: s2date("2022-05-04"), Name: "确诊病例41", Type: "轻型", City: "北京市", District: "丰台区", Address: "王佐镇鑫湖家园"},
				{Date: s2date("2022-05-04"), Name: "确诊病例42", Type: "轻型", City: "北京市", District: "丰台区", Address: "王佐镇鑫湖家园"},
			},
		},
		{
			name:    "病例范围",
			content: "确诊病例44至48：现住朝阳区建外街道光辉里小区。5月2日诊断为确诊病例，临床分型均为轻型。",
			rs: model.Residents{
				{Date: s2date("2022-05-02"), Name: "确诊病例44", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "确诊病例45", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "确诊病例46", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "确诊病例47", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
				{Date: s2date("2022-05-02"), Name: "确诊病例48", Type: "轻型", City: "北京市", District: "朝阳区", Address: "建外街道光辉里小区"},
			},
		},
		// {
		// 	name:    "多种情况综合",
		// 	content: "感染者231、234、244至246、248、252、256、261：现住址均位于朝阳区，在校学生。4月23日作为感染者146的密切接触者进行集中隔离，4月25日、26日报告核酸检测结果均为阳性，已转至定点医院，4月26日感染者246诊断为无症状感染者，感染者231、234、244、245、248当日诊断为确诊病例，感染者231、234、244、245临床分型均为轻型，感染者248临床分型为普通型，4月27日感染者252、256、261诊断为确诊病例，临床分型均为轻型。",
		// 	rs: model.Residents{
		// 		{Name: "确诊病例231", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "确诊病例234", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "确诊病例244", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "确诊病例245", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "确诊病例248", Type: "普通型", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "无症状感染者246", Type: "无症状感染者", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "确诊病例252", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "确诊病例256", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
		// 		{Name: "确诊病例261", Type: "轻型", City: "北京市", District: "朝阳区", Address: ""},
		// 	},
		// },
	}

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

func TestParseDailyContentBeijing(t *testing.T) {
	type Case struct {
		Content string
		Daily   model.Daily
	}
	testcases := []Case{
		{
			Content: "5月7日0时至15时，新增本土新冠肺炎病毒感染者45例。自5月6日发布会后（5月6日15时至7日15时），新增本土新冠肺炎病毒感染者78例（感染者749至826），其中，朝阳区37例，房山区24例，丰台区7例，通州区和顺义区各3例，海淀区2例，门头沟区和东城区各1例；普通型3例、轻型51例、无症状感染者24例；管控人员70例，社区筛查8例。均已转至定点医院隔离治疗，相关风险点位及人员均已管控落位。现将社区筛查感染者情况通报如下： \n感染者750、752：通过社区核酸筛查发现，为母女关系，感染者750为在校学生，现住朝阳区劲松街道珠江帝景B区。两人4月25日至5月3日期间5次报告核酸检测结果均为阴性。感染者750自述5月4日出现发热、咳嗽等症状，当日参加社区核酸筛查，5月5日报告结果为阳性。5月4日感染者752报告核酸检测结果为阴性，5月5日作为密切接触者进行核酸检测，报告结果为阳性，5月6日诊断为确诊病例，临床分型均为轻型。\n感染者761、762、806、807：通过社区核酸筛查发现，为邻里关系，现住房山区韩村河镇五侯村。感染者807为房山区窦店镇934路公交车安保人员，感染者806曾于5月1日到访感染者761、感染者762家中。四人4月26日至5月4日多次报告核酸检测结果均为阴性，5月6日报告核酸检测结果均为阳性。5月6日感染者761、762诊断为无症状感染者，5月7日感染者806、807诊断为无症状感染者。    \n感染者763、765：通过社区核酸筛查发现，为父女关系。现住顺义区南法信镇东杜兰村。感染者763、765自述分别于5月3日、4日出现发热等症状。两人4月23日至5月4日期间多次报告核酸检测结果均为阴性，5月6日报告核酸检测结果均为阳性，当日诊断为确诊病例，临床分型均为轻型。 \n4月22日至5月7日15时，本市累计报告688例新冠肺炎病毒感染者，涉及15个区，其中，朝阳区288例，房山区179例，通州区64例，丰台区50例，海淀区30例，顺义区25例，石景山区13例，昌平区12例，大兴区11例，东城区7例，西城区、密云区、门头沟区和延庆区各2例，经开区1例。\n经流行病学和基因测序结果提示，本市存在两条独立传播链：一是朝阳区涉及疫情，累计报告675例；二是丰台中西医结合医院涉及疫情，累计报告12例，另有1例为返京人员。目前传播链条基本清晰，已明确感染来源的新增感染者均与之前通报的感染者存在流行病学关联。社会面仍存在隐匿传染源，疫情传播尚未完全阻断。\n",
			Daily: model.Daily{
				LocalPositive:           78,
				Mild:                    51,
				Common:                  3,
				LocalAsymptomatic:       24,
				LocalPositiveFromBubble: 70,
				LocalPositiveFromRisk:   8,
				DistrictPositive: map[string]int{
					"朝阳区":  37,
					"房山区":  24,
					"丰台区":  7,
					"通州区":  3,
					"顺义区":  3,
					"海淀区":  2,
					"门头沟区": 1,
					"东城区":  1,
				},
			},
		},
		{
			Content: "5月9日0时至24时，新增61例本土确诊病例(含1例5月7日、2例5月8日诊断的无症状感染者转确诊病例)和13例无症状感染者，无新增疑似病例；新增1例境外输入确诊病例，无新增疑似病例和无症状感染者。治愈出院26例。",
			Daily: model.Daily{
				Date:                   s2date("2022-05-09"),
				LocalPositive:          0,
				Mild:                   0,
				Common:                 0,
				LocalConfirmed:         61,
				LocalAsymptomatic:      13,
				ImportedConfirmed:      1,
				DischargedFromHospital: 26,
			},
		},
	}

	for i, c := range testcases {
		p := DailyParserBeijing{}
		d := model.Daily{}
		p.ParseDailyContent(&d, c.Content)
		assert.EqualValuesf(t, c.Daily, d, "解析内容失败 (%d) '%s'", i, c.Content)
	}
}
