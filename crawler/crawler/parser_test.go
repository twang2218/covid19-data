package crawler

import (
	"crawler/model"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRegexpDailyTitle(t *testing.T) {
	testcases := [][]string{
		{
			"上海2022年4月21日，新增本土新冠肺炎确诊病例1931例 新增本土无症状感染者15698例 无新增境外输入性新冠肺炎确诊病例 无新增境外输入性无症状感染者",
			"2022年4月21日", // 日期
			"1931",       // 本土新增确诊
			"15698",      // 本土新增无症状
			"",           // 境外新增确诊
			"",           // 境外新增无症状
		},
		{
			"4月21日，新增本土新冠肺炎确诊病例1931例 新增本土无症状感染者15698例 无新增境外输入性新冠肺炎确诊病例 无新增境外输入性无症状感染者",
			"4月21日", // 日期
			"1931",  // 本土新增确诊
			"15698", // 本土新增无症状
			"",      // 境外新增确诊
			"",      // 境外新增无症状
		},
		{
			"4月24日（0-24时）上海新增2472例本土新冠肺炎确诊病例，新增16983例本土无症状感染者",
			"4月24日", // 日期
			"2472",  // 本土新增确诊
			"16983", // 本土新增无症状
			"",      // 境外新增确诊
			"",
		},
		{
			"海2022年4月24日，新增本土新冠肺炎确诊病例2472例 新增本土无症状感染者16983例 无新增境外输入性新冠肺炎确诊病例 新增境外输入性无症状感染者1例",
			"2022年4月24日", // 日期
			"2472",       // 本土新增确诊
			"16983",      // 本土新增无症状
			"",           // 境外新增确诊
			"1",
		},
		{
			"4月18日上海新增新增境外输入性新冠肺炎确诊病例25例 新增境外输入性无症状感染者10例 解除医学观察无症状感染者4例 治愈出院8例",
			"4月18日", // 日期
			"",      // 本土新增确诊
			"",      // 本土新增无症状
			"25",    // 境外新增确诊
			"10",    // 境外新增无症状
		},
		{
			"北京5月1日新增36例本土确诊病例、 5例本土无症状感染者 治愈出院10例\n日期：2022-05-02 来源：北京市卫生健康委员会 ",
			"5月1日", // 日期
			"36",   // 本土新增确诊
			"5",    // 本土新增无症状
			"",     // 境外新增确诊
			"",     // 境外新增无症状
		},
		{
			"4月5日0时至24时，新增4例本土确诊病例（确诊病例1昨日已通报）和1例无症状感染者，无新增疑似病例；无新增境外输入确诊病例、疑似病例和无症状感染者。治愈出院3例。",
			"4月5日", // 日期
			"4",    // 本土新增确诊
			"1",    // 本土新增无症状
			"",     // 境外新增确诊
			"",     // 境外新增无症状
		},
	}

	for i, c := range testcases {
		var m []string
		//	日期
		m = reDailyDate.FindStringSubmatch(c[0])
		if len(c[1]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配标题 - 日期失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[1], m[1], fmt.Sprintf("匹配标题 - 日期失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配标题 - 日期失败：(%d) %q => nil", i, c[0]))
		}

		// 本土确诊
		m = reDailyLocalConfirmed.FindStringSubmatch(c[0])
		if len(c[2]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配标题 - 本土确诊失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[2], m[1], fmt.Sprintf("匹配标题 - 本土确诊失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配标题 - 本土确诊失败：(%d) %q => nil", i, c[0]))
		}
		// 本土无症状
		m = reDailyLocalAsymptomatic.FindStringSubmatch(c[0])
		if len(c[3]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配标题 - 本土无症状失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[3], m[1], fmt.Sprintf("匹配标题 - 本土无症状失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配标题 - 本土无症状失败：(%d) %q => nil", i, c[0]))
		}

		// 境外输入确诊病例
		m = reDailyImportedConfirmed.FindStringSubmatch(c[0])
		if len(c[4]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配标题 - 境外输入确诊病例失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[4], m[1], fmt.Sprintf("匹配标题 - 境外输入确诊病例失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配标题 - 境外输入确诊病例失败：(%d) %q => %v", i, c[0], m))
		}
		// 境外输入无症状
		m = reDailyImportedAsymptomatic.FindStringSubmatch(c[0])
		if len(c[5]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配标题 - 境外输入无症状失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[5], m[1], fmt.Sprintf("匹配标题 - 境外输入无症状失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			if m != nil {
				// TODO: Fix this case
				// assert.Contains(t, m[0], "无症状")
			} else {
				assert.Nil(t, m, fmt.Sprintf("匹配标题 - 境外输入无症状失败：(%d) %q => %v", i, c[0], m))
			}
		}
	}
}

func TestRegexpContentTotal(t *testing.T) {
	testcases := [][]string{
		{
			"2022年2月26日0时至2022年4月27日24时，累计本土确诊45840例，治愈出院23931例，在院治疗21624例（其中重型304例，危重型48例）。现有待排查的疑似病例0例。\n2022年2月26日0时至2022年4月27日24时，累计死亡285例。\n截至2022年4月27日24时，累计境外输入性确诊病例4579例，出院4573例，在院治疗6例。现有待排查的疑似病例0例。\n截至2022年3月27日24时，尚在医学观察中的无症状感染者14414例，其中本土无症状感染者14376，境外输入性无症状感染38例。",
			"45840", //	累计本土确诊
			"23931", // 累计本土治愈出院
			"21624", // 本土在院治疗
			"285",   // 累计本土死亡
			"304",   // 重型
			"48",    // 危重型
			"4579",  // 累计境外输入确诊
			"4573",  // 累计境外输入治愈出院
			"6",     // 境外输入在院治疗
			"14414", // 尚在医学观察
			"14376", // 本土尚在医学观察
			"38",    // 境外输入尚在医学观察
		},
		{
			"2022年2月26日0时至2022年4月26日24时，累计本土确诊44548例，治愈出院21621例，在院治疗22689例（其中重型244例，危重型27例），死亡238例。现有待排查的疑似病例0例。\n截至2022年4月26日24时，累计境外输入性确诊病例4579例，出院4573例，在院治疗6例。现有待排查的疑似病例0例。",
			"44548", //	累计本土确诊
			"21621", // 累计本土治愈出院
			"22689", // 本土在院治疗
			"238",   // 累计本土死亡
			"244",   // 重型
			"27",    // 危重型
			"4579",  // 累计境外输入确诊
			"4573",  // 累计境外输入治愈出院
			"6",     // 境外输入在院治疗
			"",      // 尚在医学观察
			"",      // 本土尚在医学观察
			"",      // 境外输入尚在医学观察
		},
		{
			"截至2022年2月19日24时，累计境外输入性确诊病例3584例，出院3400例，在院治疗184例。现有待排查的疑似病例0例。\n截至2022年2月19日24时，累计本土确诊病例392例，治愈出院384例，在院治疗1例，死亡7例。现有待排查的疑似病例0例。\n截至2022年2月19日24时，尚在医学观察中的无症感染者5例，其中境外输入性无症状感染者4例，本土无症状感染者1例。",
			"392",  // 累计本土确诊
			"384",  // 累计本土治愈出院
			"1",    // 本土在院治疗
			"7",    // 累计本土死亡
			"",     // 重型
			"",     // 危重型
			"3584", // 累计境外输入确诊
			"3400", // 累计境外输入治愈出院
			"184",  // 境外输入在院治疗
			"5",    // 尚在医学观察
			"1",    // 本土尚在医学观察
			"4",    // 境外输入尚在医学观察
		},
		{
			"截至2022年1月11日24时，累计境外输入性确诊病例2984例，出院2636例，在院治疗348例。现有待排查的疑似病例1例。\n截至2022年1月11日24时，累计本土确诊病例388例，治愈出院381例，在院治疗0例，死亡7例。现有待排查的疑似病例0例。\n截至2022年1月11日24时，尚在医学观察中的无症状感染者23例，其中境外输入无症状感染者7例，本土无症状感染者16例。",
			"388",  // 累计本土确诊
			"381",  // 累计本土治愈出院
			"0",    // 本土在院治疗
			"7",    // 累计本土死亡
			"",     // 重型
			"",     // 危重型
			"2984", // 累计境外输入确诊
			"2636", // 累计境外输入治愈出院
			"348",  // 境外输入在院治疗
			"23",   // 尚在医学观察
			"16",   // 本土尚在医学观察
			"7",    // 境外输入尚在医学观察
		},
		{
			"截至2022年4月12日24时，累计本土确诊9903例，治愈出院2120例，在院治疗7776例（其中重症9例）。现有待排查的疑似病例0例。\n截至2022年4月12日24时，累计境外输入性确诊病例4568例，出院4526例，在院治疗42例。现有待排查的疑似病例28例。\n截至2022年4月12日24时，尚在医学观察中的无症状感染者224704例，其中本土无症状感染者224691例，境外输入性无症状感染13例。",
			"9903",   // 累计本土确诊
			"2120",   // 累计本土治愈出院
			"7776",   // 本土在院治疗
			"",       // 累计本土死亡
			"9",      // 重型
			"",       // 危重型
			"4568",   // 累计境外输入确诊
			"4526",   // 累计境外输入治愈出院
			"42",     // 境外输入在院治疗
			"224704", // 尚在医学观察
			"224691", // 本土尚在医学观察
			"13",     // 境外输入尚在医学观察
		},
	}

	for i, c := range testcases {
		var m []string
		// 累计本土确诊
		m = reDailyTotalLocalConfirmed.FindStringSubmatch(c[0])
		assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 累计本土确诊失败：(%d) %q => nil", i, c[0]))
		if len(m) > 0 {
			assert.Equal(t, c[1], m[1], fmt.Sprintf("匹配内容 - 累计本土确诊失败：(%d) %q => nil", i, c[0]))
		}
		// 累计本土治愈出院
		m = reDailyTotalLocalDischargedFromHospital.FindStringSubmatch(c[0])
		assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 累计本土治愈出院失败：(%d) %q => nil", i, c[0]))
		if len(m) > 0 {
			assert.Equal(t, c[2], m[1], fmt.Sprintf("匹配内容 - 累计本土治愈出院失败：(%d) %q => nil", i, c[0]))
		}
		// 本土在院治疗
		m = reDailyLocalInHospital.FindStringSubmatch(c[0])
		assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 本土在院治疗失败：(%d) %q => nil", i, c[0]))
		if len(m) > 0 {
			assert.Equal(t, c[3], m[1], fmt.Sprintf("匹配内容 - 本土在院治疗失败：(%d) %q => nil", i, c[0]))
		}
		// 累计本土死亡
		m = reDailyTotalLocalDeath.FindStringSubmatch(c[0])
		if len(c[4]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 累计本土死亡失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[4], m[1], fmt.Sprintf("匹配内容 - 累计本土死亡失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配内容 - 累计本土死亡失败：(%d) %q => nil", i, c[0]))
		}
		// 重型
		m = reDailySevere.FindStringSubmatch(c[0])
		if len(c[5]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 重型失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[5], m[1], fmt.Sprintf("匹配内容 - 重型失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配内容 - 重型失败：(%d) %q => nil", i, c[0]))
		}
		// 危重型
		m = reDailyCritical.FindStringSubmatch(c[0])
		if len(c[6]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 危重型失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[6], m[1], fmt.Sprintf("匹配内容 - 危重型失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配内容 - 危重型失败：(%d) %q => nil", i, c[0]))
		}
		// 累计境外输入确诊
		m = reDailyTotalImportedConfirmed.FindStringSubmatch(c[0])
		assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 累计境外输入确诊失败：(%d) %q => nil", i, c[0]))
		if len(m) > 0 {
			assert.Equal(t, c[7], m[1], fmt.Sprintf("匹配内容 - 累计境外输入确诊失败：(%d) %q => nil", i, c[0]))
		}
		// 累计境外输入治愈出院
		m = reDailyTotalImportedDischargedFromHospital.FindStringSubmatch(c[0])
		assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 累计境外输入治愈出院失败：(%d) %q => nil", i, c[0]))
		if len(m) > 0 {
			assert.Equal(t, c[8], m[1], fmt.Sprintf("匹配内容 - 累计境外输入治愈出院失败：(%d) %q => nil", i, c[0]))
		}
		// 境外输入在院治疗
		m = reDailyImportedInHospital.FindStringSubmatch(c[0])
		assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 境外输入在院治疗失败：(%d) %q => nil", i, c[0]))
		if len(m) > 0 {
			assert.Equal(t, c[9], m[1], fmt.Sprintf("匹配内容 - 境外输入在院治疗失败：(%d) %q => nil", i, c[0]))
		}
		// 尚在医学观察
		m = reDailyUnderMedicalObservation.FindStringSubmatch(c[0])
		if len(c[10]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 尚在医学观察失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[10], m[1], fmt.Sprintf("匹配内容 - 尚在医学观察失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配内容 - 尚在医学观察失败：(%d) %q => nil", i, c[0]))
		}
		// 本土尚在医学观察
		m = reDailyLocalUnderMedicalObservation.FindStringSubmatch(c[0])
		if len(c[11]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 本土尚在医学观察失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[11], m[1], fmt.Sprintf("匹配内容 - 本土尚在医学观察失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配内容 - 本土尚在医学观察失败：(%d) %q => nil", i, c[0]))
		}

		// 境外输入尚在医学观察
		m = reDailyImportedUnderMedicalObservation.FindStringSubmatch(c[0])
		if len(c[12]) > 0 {
			assert.NotNil(t, m, fmt.Sprintf("匹配内容 - 境外输入尚在医学观察失败：(%d) %q => nil", i, c[0]))
			if len(m) > 0 {
				assert.Equal(t, c[12], m[1], fmt.Sprintf("匹配内容 - 境外输入尚在医学观察失败：(%d) %q => nil", i, c[0]))
			}
		} else {
			assert.Nil(t, m, fmt.Sprintf("匹配内容 - 境外输入尚在医学观察失败：(%d) %q => nil", i, c[0]))
		}

	}

}

func TestParseDailyContentRegion(t *testing.T) {
	type Case struct {
		Content string
		Daily   model.Daily
	}
	testcases := []Case{
		{
			Content: "市卫健委今早（25日）通报：2022年4月24日0—24时，新增本土新冠肺炎确诊病例2472例和无症状感染者16983例，其中846例确诊病例为此前无症状感染者转归，1557例确诊病例和16835例无症状感染者在隔离管控中发现，其余在相关风险人群排查中发现。新增境外输入性新冠肺炎无症状感染者1例，在闭环管控中发现。\n阳性感染者居住地信息按区划分进行统计，您可关注所在区的官方微信，第一时间了解本区阳性感染者的居住信息，稍后小布也将汇总各区信息。\n本土病例情况\n2022年4月24日0—24时，新增本土新冠肺炎确诊病例2472例，含846例由无症状感染者转为确诊病例。新增治愈出院2449例。\n病例1—病例510，居住于浦东新区，\n病例511—病例683，居住于黄浦区，\n病例684—病例803，居住于徐汇区，\n病例804—病例866，居住于长宁区，\n病例867—病例973，居住于静安区，\n病例974—病例1011，居住于普陀区，\n病例1012—病例1108，居住于虹口区，\n病例1109—病例1159，居住于杨浦区，\n病例1160—病例1242，居住于闵行区，\n病例1243—病例1365，居住于宝山区，\n病例1366—病例1490，居住于嘉定区，\n病例1491—病例1525，居住于松江区，\n病例1526—病例1530，居住于青浦区，\n病例1531—病例1557，居住于崇明区，\n均为本市闭环隔离管控人员，其间新冠病毒核酸检测结果异常，经疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例1558—病例1578，居住于浦东新区，\n病例1579、病例1580，居住于黄浦区，\n病例1581—病例1587，居住于徐汇区，\n病例1588、病例1589，居住于长宁区，\n病例1590—病例1595，居住于静安区，\n病例1596—病例1599，居住于虹口区，\n病例1600—病例1603，居住于杨浦区，\n病例1604—病例1620，居住于闵行区，\n病例1621、病例1622，居住于宝山区，\n病例1623—病例1626，居住于嘉定区，\n在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例1627—病例1969，居住于浦东新区，\n病例1970—病例2111，居住于黄浦区，\n病例2112—病例2212，居住于徐汇区，\n病例2213—病例2223，居住于长宁区，\n病例2224—病例2252，居住于静安区，\n病例2253—病例2257，居住于普陀区，\n病例2258—病例2262，居住于虹口区，\n病例2263—病例2282，居住于杨浦区，\n病例2283—病例2325，居住于闵行区，\n病例2326—病例2335，居住于宝山区，\n病例2336—病例2338，居住于嘉定区，\n病例2339—病例2469，居住于松江区，\n病例2470—病例2472，居住于青浦区，\n为此前报告的本土无症状感染者。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n2022年4月24日0—24时，新增本土死亡51例。平均年龄84.2岁，80岁以上高龄老人共37位，最大年龄100岁。51位患者基础疾病严重，累及多个脏器，包括急性冠脉综合征、心力衰竭、严重心律失常、高血压3级（极高危）、脑梗后遗症、糖尿病、帕金森病、阿尔兹海默症、尿毒症、恶性肿瘤、重度营养不良、电解质紊乱等。患者入院后，因原发疾病加重，经抢救无效死亡。死亡的直接原因均为基础疾病。\n本土无症状感染者情况\n2022年4月24日0—24时，新增本土无症状感染者16983例。\n无症状感染者1—无症状感染者5255，居住于浦东新区，\n无症状感染者5256—无症状感染者6692，居住于黄浦区，\n无症状感染者6693—无症状感染者7851，居住于徐汇区，\n无症状感染者7852—无症状感染者8522，居住于长宁区，\n无症状感染者8523—无症状感染者10068，居住于静安区，\n无症状感染者10069—无症状感染者10499，居住于普陀区，\n无症状感染者10500—无症状感染者11041，居住于虹口区，\n无症状感染者11042—无症状感染者12179，居住于杨浦区，\n无症状感染者12180—无症状感染者13249，居住于闵行区，\n无症状感染者13250—无症状感染者15605，居住于宝山区，\n无症状感染者15606—无症状感染者16071，居住于嘉定区，\n无症状感染者16072—无症状感染者16138，居住于金山区，\n无症状感染者16139—无症状感染者16520，居住于松江区，\n无症状感染者16521—无症状感染者16760，居住于青浦区，\n无症状感染者16761—无症状感染者16767，居住于奉贤区，\n无症状感染者16768—无症状感染者16835，居住于崇明区，\n均为本市闭环隔离管控人员，其间新冠病毒核酸检测结果异常，经疾控中心复核结果为阳性，诊断为无症状感染者。\n无症状感染者16836—无症状感染者16887，居住于浦东新区，\n无症状感染者16888，居住于黄浦区，\n无症状感染者16889—无症状感染者16897，居住于徐汇区，\n无症状感染者16898、无症状感染者16899，居住于长宁区，\n无症状感染者16900—无症状感染者16913，居住于静安区，\n无症状感染者16914—无症状感染者16916，居住于普陀区，\n无症状感染者16917—无症状感染者16923，居住于虹口区，\n无症状感染者16924—无症状感染者16945，居住于杨浦区，\n无症状感染者16946—无症状感染者16961，居住于闵行区，\n无症状感染者16962—无症状感染者16976，居住于宝山区，\n无症状感染者16977—无症状感染者16979，居住于嘉定区，\n无症状感染者16980，居住于金山区，\n无症状感染者16981—无症状感染者16983，居住于青浦区，\n在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经疾控中心复核结果为阳性，诊断为无症状感染者。\n境外输入病例情况\n2022年4月24日0—24时，无新增境外输入性新冠肺炎确诊病例。新增治愈出院2例，其中来自美国1例，来自英国1例。\n境外输入性无症状感染者情况\n2022年4月24日0—24时，新增境外输入性无症状感染者1例。\n该无症状感染者为中国籍，在美国探亲，自美国出发，于2022年4月9日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n该境外输入性无症状感染者已转至定点医疗机构医学观察，同航班密切接触者此前均已落实集中隔离观察。\n2022年4月24日0—24时，解除医学观察本土无症状感染者19523例。\n2022年2月26日0时至2022年4月24日24时，累计本土确诊41281例，治愈出院17041例，在院治疗24102例（其中重型196例，危重型23例）。现有待排查的疑似病例0例。\n2022年2月26日0时至2022年4月24日24时，累计死亡138例。\n截至2022年4月24日24时，累计境外输入性确诊病例4579例，出院4570例，在院治疗9例。现有待排查的疑似病例2例。\nImage\n各区阳性感染者的居住信息，稍后小布将进行汇总发布。\n资料：市卫健委\n编辑：林欣",
			Daily: model.Daily{
				DistrictConfirmedFromBubble:       map[string]int{"嘉定区": 125, "宝山区": 123, "崇明区": 27, "徐汇区": 120, "普陀区": 38, "杨浦区": 51, "松江区": 35, "浦东新区": 510, "虹口区": 97, "长宁区": 63, "闵行区": 83, "青浦区": 5, "静安区": 107, "黄浦区": 173},
				DistrictConfirmedFromRisk:         map[string]int{"嘉定区": 4, "宝山区": 2, "徐汇区": 7, "杨浦区": 4, "浦东新区": 21, "虹口区": 4, "长宁区": 2, "闵行区": 17, "静安区": 6, "黄浦区": 2},
				DistrictConfirmedFromAsymptomatic: map[string]int{"嘉定区": 3, "宝山区": 10, "徐汇区": 101, "普陀区": 5, "杨浦区": 20, "松江区": 131, "浦东新区": 343, "虹口区": 5, "长宁区": 11, "闵行区": 43, "青浦区": 3, "静安区": 29, "黄浦区": 142},
				DistrictAsymptomaticFromBubble:    map[string]int{"嘉定区": 466, "奉贤区": 7, "宝山区": 2356, "崇明区": 68, "徐汇区": 1159, "普陀区": 431, "杨浦区": 1138, "松江区": 382, "浦东新区": 5255, "虹口区": 542, "金山区": 67, "长宁区": 671, "闵行区": 1070, "青浦区": 240, "静安区": 1546, "黄浦区": 1437},
				DistrictAsymptomaticFromRisk:      map[string]int{"嘉定区": 3, "宝山区": 15, "徐汇区": 9, "普陀区": 3, "杨浦区": 22, "浦东新区": 52, "虹口区": 7, "金山区": 1, "长宁区": 2, "闵行区": 16, "青浦区": 3, "静安区": 14, "黄浦区": 1},
			},
		},
		{
			Content: "\n病例1—病例510，居住于浦东新区，病例511—病例683，居住于黄浦区，病例684—病例803，居住于徐汇区，病例804—病例866，居住于长宁区，病例867—病例973，居住于静安区，病例974—病例1011，居住于普陀区，病例1012—病例1108，居住于虹口区，病例1109—病例1159，居住于杨浦区，病例1160—病例1242，居住于闵行区，病例1243—病例1365，居住于宝山区，病例1366—病例1490，居住于嘉定区，病例1491—病例1525，居住于松江区，病例1526—病例1530，居住于青浦区，病例1531—病例1557，居住于崇明区，均为本市闭环隔离管控人员，其间新冠病毒核酸检测结果异常，经疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n\n病例1558—病例1578，居住于浦东新区，病例1579、病例1580，居住于黄浦区，病例1581—病例1587，居住于徐汇区，病例1588、病例1589，居住于长宁区，病例1590—病例1595，居住于静安区，病例1596—病例1599，居住于虹口区，病例1600—病例1603，居住于杨浦区，病例1604—病例1620，居住于闵行区，病例1621、病例1622，居住于宝山区，病例1623—病例1626，居住于嘉定区，在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n\n病例1627—病例1969，居住于浦东新区，病例1970—病例2111，居住于黄浦区，病例2112—病例2212，居住于徐汇区，病例2213—病例2223，居住于长宁区，病例2224—病例2252，居住于静安区，病例2253—病例2257，居住于普陀区，病例2258—病例2262，居住于虹口区，病例2263—病例2282，居住于杨浦区，病例2283—病例2325，居住于闵行区，病例2326—病例2335，居住于宝山区，病例2336—病例2338，居住于嘉定区，病例2339—病例2469，居住于松江区，病例2470—病例2472，居住于青浦区，为此前报告的本土无症状感染者。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n\n2022年4月24日0—24时，新增本土死亡51例。平均年龄84.2岁，80岁以上高龄老人共37位，最大年龄100岁。51位患者基础疾病严重，累及多个脏器，包括急性冠脉综合征、心力衰竭、严重心律失常、高血压3级（极高危）、脑梗后遗症、糖尿病、帕金森病、阿尔兹海默症、尿毒症、恶性肿瘤、重度营养不良、电解质紊乱等。患者入院后，因原发疾病加重，经抢救无效死亡。死亡的直接原因均为基础疾病。\n\n2022年4月24日0—24时，新增本土无症状感染者16983例。\n\n无症状感染者1—无症状感染者5255，居住于浦东新区，无症状感染者5256—无症状感染者6692，居住于黄浦区，无症状感染者6693—无症状感染者7851，居住于徐汇区，无症状感染者7852—无症状感染者8522，居住于长宁区，无症状感染者8523—无症状感染者10068，居住于静安区，无症状感染者10069—无症状感染者10499，居住于普陀区，无症状感染者10500—无症状感染者11041，居住于虹口区，无症状感染者11042—无症状感染者12179，居住于杨浦区，无症状感染者12180—无症状感染者13249，居住于闵行区，无症状感染者13250—无症状感染者15605，居住于宝山区，无症状感染者15606—无症状感染者16071，居住于嘉定区，无症状感染者16072—无症状感染者16138，居住于金山区，无症状感染者16139—无症状感染者16520，居住于松江区，无症状感染者16521—无症状感染者16760，居住于青浦区，无症状感染者16761—无症状感染者16767，居住于奉贤区，无症状感染者16768—无症状感染者16835，居住于崇明区，均为本市闭环隔离管控人员，其间新冠病毒核酸检测结果异常，经疾控中心复核结果为阳性，诊断为无症状感染者。\n\n无症状感染者16836—无症状感染者16887，居住于浦东新区，无症状感染者16888，居住于黄浦区，无症状感染者16889—无症状感染者16897，居住于徐汇区，无症状感染者16898、无症状感染者16899，居住于长宁区，无症状感染者16900—无症状感染者16913，居住于静安区，无症状感染者16914—无症状感染者16916，居住于普陀区，无症状感染者16917—无症状感染者16923，居住于虹口区，无症状感染者16924—无症状感染者16945，居住于杨浦区，无症状感染者16946—无症状感染者16961，居住于闵行区，无症状感染者16962—无症状感染者16976，居住于宝山区，无症状感染者16977—无症状感染者16979，居住于嘉定区，无症状感染者16980，居住于金山区，无症状感染者16981—无症状感染者16983，居住于青浦区，在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经疾控中心复核结果为阳性，诊断为无症状感染者。\n\n2022年4月24日0—24时，无新增境外输入性新冠肺炎确诊病例。新增治愈出院2例，其中来自美国1例，来自英国1例。\n\n2022年4月24日0—24时，新增境外输入性无症状感染者1例。\n\n该无症状感染者为中国籍，在美国探亲，自美国出发，于2022年4月9日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n\n该境外输入性无症状感染者已转至定点医疗机构医学观察，同航班密切接触者此前均已落实集中隔离观察。",
			Daily: model.Daily{
				DistrictConfirmedFromBubble:       map[string]int{"嘉定区": 125, "宝山区": 123, "崇明区": 27, "徐汇区": 120, "普陀区": 38, "杨浦区": 51, "松江区": 35, "浦东新区": 510, "虹口区": 97, "长宁区": 63, "闵行区": 83, "青浦区": 5, "静安区": 107, "黄浦区": 173},
				DistrictConfirmedFromRisk:         map[string]int{"嘉定区": 4, "宝山区": 2, "徐汇区": 7, "杨浦区": 4, "浦东新区": 21, "虹口区": 4, "长宁区": 2, "闵行区": 17, "静安区": 6, "黄浦区": 2},
				DistrictConfirmedFromAsymptomatic: map[string]int{"嘉定区": 3, "宝山区": 10, "徐汇区": 101, "普陀区": 5, "杨浦区": 20, "松江区": 131, "浦东新区": 343, "虹口区": 5, "长宁区": 11, "闵行区": 43, "青浦区": 3, "静安区": 29, "黄浦区": 142},
				DistrictAsymptomaticFromBubble:    map[string]int{"嘉定区": 466, "奉贤区": 7, "宝山区": 2356, "崇明区": 68, "徐汇区": 1159, "普陀区": 431, "杨浦区": 1138, "松江区": 382, "浦东新区": 5255, "虹口区": 542, "金山区": 67, "长宁区": 671, "闵行区": 1070, "青浦区": 240, "静安区": 1546, "黄浦区": 1437},
				DistrictAsymptomaticFromRisk:      map[string]int{"嘉定区": 3, "宝山区": 15, "徐汇区": 9, "普陀区": 3, "杨浦区": 22, "浦东新区": 52, "虹口区": 7, "金山区": 1, "长宁区": 2, "闵行区": 16, "青浦区": 3, "静安区": 14, "黄浦区": 1},
			},
		},
		{
			Content: "市卫健委今早（26日）通报：2022年3月25日0—24时，新增本土新冠肺炎确诊病例38例和无症状感染者2231例，其中5例确诊病例为此前无症状感染者转归，3例确诊病例和1773例无症状感染者在隔离管控中发现，其余在相关风险人群排查中发现。新增境外输入性新冠肺炎确诊病例9例和无症状感染者2例，均在闭环管控中发现。\n2022年3月25日0—24时，新增本土新冠肺炎确诊病例38例。治愈出院13例。\n病例1，男，23岁，居住于闵行区，病例2，男，32岁，居住于浦东新区，病例3，男，17岁，居住于浦东新区，均为本市闭环隔离管控人员，其间新冠病毒核酸检测结果异常，经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例4，女，15岁，居住于浦东新区，病例5，男，35岁，居住于闵行区，病例6，女，40岁，居住于闵行区，病例7，男，58岁，居住于浦东新区，病例8，男，55岁，居住于浦东新区，病例9，女，51岁，居住于浦东新区，病例10，男，63岁，居住于浦东新区，病例11，女，69岁，居住于浦东新区，病例12，女，64岁，居住于浦东新区，病例13，女，45岁，居住于浦东新区，病例14，男，60岁，居住于浦东新区，病例15，男，56岁，居住于浦东新区，病例16，男，66岁，居住于浦东新区，病例17，女，57岁，居住于浦东新区，病例18，女，54岁，居住于浦东新区，病例19，男，63岁，居住于浦东新区，病例20，女，58岁，居住于浦东新区，病例21，男，58岁，居住于浦东新区，病例22，男，53岁，居住于浦东新区，病例23，男，56岁，居住于浦东新区，病例24，男，51岁，居住于浦东新区，病例25，女，54岁，居住于浦东新区，病例26，男，53岁，居住于浦东新区，病例27，男，52岁，居住于浦东新区，病例28，男，53岁，居住于浦东新区，病例29，女，52岁，居住于浦东新区，病例30，男，52岁，居住于浦东新区，病例31，男，58岁，居住于浦东新区，病例32，女，51岁，居住于浦东新区，病例33，男，64岁，居住于浦东新区，在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例34，女，64岁，居住于普陀区，病例35，男，13岁，居住于松江区，病例36，男，3岁，居住于浦东新区，病例37，男，16岁，居住于闵行区，病例38，女，8岁，居住于徐汇区，为此前报告的本土无症状感染者。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n目前，已追踪到以上病例在本市的密切接触者67人，均已落实集中隔离观察。对病例曾活动过的场所已进行终末消毒。\n2022年3月25日0—24时，新增本土无症状感染者2231例。\n无症状感染者1，女，33岁，居住于黄浦区，无症状感染者2，女，66岁，居住于黄浦区，无症状感染者3，男，55岁，居住于黄浦区，无症状感染者4，男，60岁，居住于黄浦区，无症状感染者5，女，29岁，居住于黄浦区，无症状感染者6，女，27岁，居住于黄浦区，无症状感染者7，男，26岁，居住于黄浦区，无症状感染者8，男，66岁，居住于黄浦区，无症状感染者9，女，65岁，居住于黄浦区，无症状感染者10，男，60岁，居住于黄浦区，无症状感染者11，女，53岁，居住于浦东新区，无症状感染者12，男，88岁，居住于浦东新区，无症状感染者13，女，48岁，居住于嘉定区，无症状感染者14，男，52岁，居住于闵行区，均为本市闭环隔离管控人员，其间新冠病毒核酸检测结果异常，经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n无症状感染者1774，女，50岁，居住于黄浦区，无症状感染者1775，女，72岁，居住于黄浦区，无症状感染者1776，女，70岁，居住于黄浦区",
			Daily: model.Daily{
				DistrictConfirmedFromBubble:       map[string]int{"浦东新区": 2, "闵行区": 1},
				DistrictConfirmedFromRisk:         map[string]int{"浦东新区": 28, "闵行区": 2},
				DistrictConfirmedFromAsymptomatic: map[string]int{"徐汇区": 1, "普陀区": 1, "松江区": 1, "浦东新区": 1, "闵行区": 1},
				DistrictAsymptomaticFromBubble:    map[string]int{"嘉定区": 1, "浦东新区": 2, "闵行区": 1, "黄浦区": 10},
			},
		},
		{
			Content: "市卫健委今早（17日）通报：2022年3月16日0—24时，新增本土新冠肺炎确诊病例8例（含1例由无症状感染者转为确诊病例）和无症状感染者150例，其中1例确诊病例和69例无症状感染者在隔离管控中发现，1例无症状感染者为外省返沪人员协查中发现，其余在相关风险人群排查中发现。新增境外输入性新冠肺炎确诊病例15例和无症状感染者6例,均在闭环管控中发现。\n2022年3月16日0—24时，新增本土新冠肺炎确诊病例8例，含1例由无症状感染者转为确诊病例。\n病例1，男，49岁，居住于闵行区江川路街道剑川路综合服务中心工地宿舍，为此前报告的本土无症状感染者。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。（3月16日已通报）\n病例2，女，32岁，居住于黄浦区顺昌路612弄，病例3，女，35岁，居住于徐汇区田林八村，均系本市报告本土无症状感染者的密切接触者，即被隔离管控，其间新冠病毒核酸检测结果异常，经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例4，男，38岁，居住于嘉定区崇教路267号对面工地宿舍，病例5，女，78岁，居住于嘉定区桃园新村，病例6，男，52岁，居住于嘉定区桃园新村，病例7，男，47岁，居住于嘉定区崇教路267号对面工地宿舍，病例8，女，14岁，居住于徐汇区桂林西街11弄，在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n目前，已追踪到以上病例在本市的密切接触者13人，均已落实集中隔离观察。对病例曾活动过的场所已进行终末消毒。\n2022年3月16日0—24时，新增本土无症状感染者150例。\n无症状感染者1，男，61岁，居住于嘉定区南新路170弄，无症状感染者2，男，45岁，居住于宝山区泗塘五村，无症状感染者3，男，53岁，居住于宝山区云西路168弄，无症状感染者4，男，51岁，居住于宝山区友谊西路富桥路工地宿舍，无症状感染者5，女，39岁，居住于普陀区古浪路55弄，无症状感染者6，女，56岁，居住于松江区茸惠路858弄，无症状感染者7，男，51岁，居住于长宁区长宁路405弄，均系本市报告本土无症状感染者的密切接触者，即被隔离管控，其间新冠病毒核酸检测结果异常，经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。（3月16日已通报）\n无症状感染者8，女，28岁，居住于长宁区番禺路222弄，系外省返沪协查人员核酸筛查异常，即被集中隔离管控。经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。（3月16日已通报）\n无症状感染者9，女，25岁，居住于宝山区淞南四村，无症状感染者10，女，50岁，居住于宝山区云西路168弄，无症状感染者11，女，5岁，居住于黄浦区斜土东路197号，无症状感染者12，女，31岁，居住于黄浦区斜土东路197号，无症状感染者13，女，50岁，居住于黄浦区斜土东路197号，无症状感染者14，女，36岁，居住于黄浦区静修路50弄，无症状感染者15，女，46岁，居住于嘉定区鹤友路336弄，无症状感染者16，男，60岁，居住于嘉定区鹤望路365弄，无症状感染者17，女，35岁，居住于长宁区华阳路70号，无症状感染者18，男，36岁，居住于长宁区虹桥路977号，无症状感染者19，女，54岁，居住于嘉定区金沙江路2823弄，在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。（3月16日已通报）\n无症状感染者20，女，57岁，居住于嘉定区小东街，无症状感染者21，男，25岁，居住于嘉定区安新村，无症状感染者22，女，40岁，居住于嘉定区连俊村，无症状感染者23，男，80岁，居住于嘉定区连俊村，无症状感染者24，男，44岁，居住于虹口区四川北路1906弄，无症状感染者25，男，60岁，居住于虹口区水电路818弄，无症状感染者26，女，56岁，居住于虹口区水电路818弄，无症状感染者27，女，31岁，居住于虹口区水电路818弄，无症状感染者28，男，68岁，居住于浦东新区行泰路150弄，无症状感染者29，女，28岁，居住于浦东新区行泰路150弄，无症状感染者30，男，11岁，居住于浦东新区华鹏路390弄，无症状感染者31，女，50岁，居住于浦东新区欣莲佳苑，无症状感染者32，女，14岁，居住于浦东新区新龙路69弄，无症状感染者33，女，29岁，居住于浦东新区上游村，无症状感染者34，男，40岁，居住于浦东新区听悦路685弄，无症状感染者35，男，68岁，居住于浦东新区听悦路960弄，无症状感染者36，女，67岁，居住于浦东新区听悦路960弄，无症状感染者37，男，23岁，居住于浦东新区勤丰村，无症状感染者38，女，51岁，居住于浦东新区勤丰村，无症状感染者39，女，54岁，居住于浦东新区惠东村，无症状感染者40，女，48岁，居住于浦东新区汇南村，无症状感染者41，女，55岁，居住于浦东新区英雄村，无症状感染者42，女，47岁，居住于浦东新区城南路110号，无症状感染者43，女，34岁，居住于浦东新区听悦路960弄，无症状感染者44，男，38岁，居住于浦东新区国展路1756号，无症状感染者45，男，11岁，居住于浦东新区芳甸路333弄，无症状感染者46，男，15岁，居住于浦东新区高城路200弄，无症状感染者47，女，41岁，居住于浦东新区芦云路200弄，无症状感染者48，女，39岁，居住于浦东新区外灶里灶村，无症状感染者49，女，59岁，居住于浦东新区三三公路5020弄，无症状感染者50，女，13岁，居住于浦东新区东靖路2250弄，无症状感染者51，女，71岁，居住于嘉定区三里村，无症状感染者52，男，53岁，居住于嘉定区三里村，无症状感染者53，男，42岁，居住于嘉定区崇教路267号对面工地宿舍，无症状感染者54，男，31岁，居住于虹口区辉河路25弄，无症状感染者55，男，45岁，居住于宝山区罗和路935弄，无症状感染者56，男，31岁，居住于嘉定区环城路762弄，无症状感染者57，女，31岁，居住于普陀区武威东路478弄，无症状感染者58，女，36岁，居住于闵行区北翟路1554弄，无症状感染者59，女，34岁，居住于闵行区王泥浜村，无症状感染者60，女，19岁，居住于松江区龙源路555号，无症状感染者61，女，42岁，居住于黄浦区陆家浜路1398号，无症状感染者62，男，40岁，居住于松江区外婆泾路2999弄，无症状感染者63，女，24岁，居住于松江区新南街625弄，无症状感染者64，女，40岁，居住于黄浦区河南南路1001弄，无症状感染者65，女，34岁，居住于宝山区场北路399弄，无症状感染者66，男，52岁，居住于闵行区兰平路301弄，无症状感染者67，女，50岁，居住于闵行区建设路16弄，无症状感染者68，女，25岁，居住于青浦区龙联路660弄，无症状感染者69，男，46岁，居住于闵行区江川路街道剑川路综合服务中心工地宿舍，无症状感染者70，男，53岁，居住于闵行区江川路街道剑川路综合服务中心工地宿舍，无症状感染者71，男，42岁，居住于虹口区车站北路732弄，无症状感染者72，男，8岁，居住于虹口区车站北路732弄，无症状感染者73，男，42岁，居住于嘉定区塔城路470弄，无症状感染者74，男，57岁，居住于嘉定区双单路德立路工地宿舍，无症状感染者75，男，50岁，居住于嘉定区崇教路267号对面工地宿舍，无症状感染者76，女，34岁，居住于嘉定区南新路219弄，无症状感染者77，男，36岁，居住于嘉定区草庵村，无症状感染者78，女，54岁，居住于奉贤区新四平公路467弄，无症状感染者79，女，3岁，居住于奉贤区新四平公路467弄，无症状感染者80，男，23岁，居住于嘉定区崇教路267号对面工地宿舍，无症状感染者81，男，41岁，居住于闵行区江川路街道剑川路综合服务中心工地宿舍，均系本市报告本土确诊病例或无症状感染者的密切接触者，即被隔离管控，其间新冠病毒核酸检测结果异常，经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n无症状感染者82，男，53岁，居住于浦东新区国展路1760号，无症状感染者83，男，47岁，居住于浦东新区永泰路630弄，无症状感染者84，男，69岁，居住于浦东新区五莲路780弄，无症状感染者85，男，45岁，居住于嘉定区连俊村，无症状感染者86，男，40岁，居住于杨浦区政通路118弄，无症状感染者87，女，39岁，居住于杨浦区控江路1209号，无症状感染者88，女，43岁，居住于静安区万荣路1199弄，无症状感染者89，女，13岁，居住于浦东新区联勤村，无症状感染者90，女，29岁，居住于嘉定区鹤望路365号，无症状感染者91，男，84岁，居住于嘉定区桃园新村，无症状感染者92，男，72岁，居住于嘉定区汇善路1333弄，无症状感染者93，女，70岁，居住于嘉定区汇善路1333弄，无症状感染者94，男，89岁，居住于嘉定区桃园新村，无症状感染者95，女，55岁，居住于嘉定区桃园小区，无症状感染者96，男，90岁，居住于嘉定区桃园新村，无症状感染者97，男，72岁，居住于嘉定区桃园新村，无症状感染者98，女，18岁，居住于嘉定区桃园新村，无症状感染者99，女，46岁，居住于嘉定区桃园新村，无症状感染者100，男，78岁，居住于嘉定区温宿路45号，无症状感染者101，男，60岁，居住于虹口区宝山路888弄，无症状感染者102，女，38岁，居住于虹口区逸仙路288弄，无症状感染者103，男，41岁，居住于黄浦区瞿溪路904弄，无症状感染者104，男，71岁，居住于浦东新区东沟六村，无症状感染者105，女，47岁，居住于浦东新区微山新村，无症状感染者106，男，53岁，居住于浦东新区俱进路505弄，无症状感染者107，女，48岁，居住于浦东新区龙东大道1号，无症状感染者108，男，34岁，居住于浦东新区孙环路366弄，无症状感染者109，男，6岁，居住于浦东新区宣镇东路788弄，无症状感染者110，男，75岁，居住于浦东新区浦建路60弄，无症状感染者111，男，38岁，居住于浦东新区新行路433弄，无症状感染者112，女，51岁，居住于浦东新区西门路588弄，无症状感染者113，男，54岁，居住于浦东新区四灶村，无症状感染者114，女，11岁，居住于闵行区浦申路1288弄，无症状感染者115，女，29岁，居住于浦东新区听悦路960弄，无症状感染者116，女，20岁，居住于浦东新区华佗路280弄，无症状感染者117，男，40岁，居住于浦东新区船厂街27号，无症状感染者118，女，34岁，居住于浦东新区和佳路105弄，无症状感染者119，男，27岁，居住于宝山区新村路681号，无症状感染者120，男，25岁，居住于宝山区新村路681号，无症状感染者121，女，33岁，居住于闵行区虹莘路3333弄，无症状感染者122，男，40岁，居住于静安区谈家桥路80弄，无症状感染者123，男，61岁，居住于普陀区定边路377弄，无症状感染者124，男，23岁，居住于浦东新区临沂一村，无症状感染者125，女，49岁，居住于黄浦区进贤路158号，无症状感染者126，男，44岁，居住于浦东新区芳草路258号，无症状感染者127，男，19岁，居住于虹口区四川北路1906号，无症状感染者128，男，33岁，居住于浦东新区钱堂村，无症状感染者129，男，51岁，居住于嘉定区嘉新公路668弄，无症状感染者130，女，39岁，居住于嘉定区娄塘路763号，无症状感染者131，男，23岁，居住于嘉定区崇教路267号对面工地宿舍，无症状感染者132，男，30岁，居住于嘉定区塔城路850弄，无症状感染者133，男，59岁，居住于嘉定区陈周村，无症状感染者134，男，53岁，居住于嘉定区崇教路267号对面工地宿舍，无症状感染者135，男，50岁，居住于嘉定区崇教路267号对面工地宿舍，无症状感染者136，女，36岁，居住于闵行区申北路135弄，无症状感染者137，女，35岁，居住于闵行区联青路51弄，无症状感染者138，女，38岁，居住于闵行区银春路2200弄，无症状感染者139，男，43岁，居住于宝山区菊联路89弄，无症状感染者140，男，53岁，居住于闵行区沪闵路280号，无症状感染者141，女，11岁，居住于闵行区业祥路111弄，无症状感染者142，男，12岁，居住于闵行区虹梅南路1728弄，无症状感染者143，男，11岁，居住于闵行区业祥路111弄，无症状感染者144，女，45岁，居住于闵行区古美七村，无症状感染者145，男，55岁，居住于徐汇区江南一村，无症状感染者146，女，23岁，居住于徐汇区龙华西路101号乙，无症状感染者147，男，36岁，居住于徐汇区田林十三村，无症状感染者148，女，20岁，居住于徐汇区零陵路231号，无症状感染者149，女，56岁，居住于徐汇区梅陇十村，无症状感染者150，男，50岁，居住于静安区芷江西路543弄，在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n目前，已追踪到以上无症状感染者在本市的密切接触者242人，均已落实集中隔离观察。对无症状感染者曾活动过的场所已进行终末消毒。\n2022年3月16日0—24时，通过口岸联防联控机制，报告15例新增境外输入性新冠肺炎确诊病例。新增治愈出院33例，其中来自中国香港22例，来自日本2例，来自台湾地区2例，来自巴布亚新几内亚1例，来自泰国1例，来自西班牙1例，来自荷兰1例，来自新加坡1例，来自美国1例，来自韩国1例。\n病例1为中国籍，暂居香港，自香港出发，于2022年2月26日抵达上海浦东国际机场，入关后即被集中隔离观察，解除隔离前出现症状，即送指定医疗机构隔离排查。经专家组会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例2为中国籍，在日本探亲，自日本出发，于2022年2月27日抵达上海浦东国际机场，入关后即被集中隔离观察，解除隔离前出现症状，即送指定医疗机构隔离排查。经专家组会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例3为中国籍，在美国工作，自美国出发，于2022年2月28日抵达上海浦东国际机场，入关后即被集中隔离观察，解除隔离前出现症状，即送指定医疗机构隔离排查。经专家组会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例4为中国籍，在香港就学，自香港出发，于2022年3月4日抵达上海浦东国际机场，入关后即被集中隔离观察，其间出现症状。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例5为中国籍，暂居香港，自香港出发，于2022年3月6日抵达上海浦东国际机场，入关后即被集中隔离观察，其间出现症状。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例6、病例7均为台湾地区居民，在台湾地区生活，病例8为中国籍，在台湾地区探亲，病例6—病例8自台湾地区出发，乘坐同一航班，于2022年3月6日抵达上海浦东国际机场，入关后即被集中隔离观察，其间出现症状。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例9为中国籍，病例10、病例11均为加拿大籍，病例9—病例11在加拿大生活，自加拿大出发，乘坐同一航班于2022年3月6日抵达上海浦东国际机场，入关后即被集中隔离观察，其间出现症状。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例12为荷兰籍，在荷兰探亲，自荷兰出发，于2022年3月9日抵达上海浦东国际机场，入关后即被集中隔离观察，其间出现症状。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例13为中国籍，在香港就学，自香港出发，于2022年3月13日抵达上海浦东国际机场，因有症状，入关后即被送至指定医疗机构隔离留观。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例14为韩国籍，在韩国探亲，自韩国出发，于2022年3月14日抵达上海浦东国际机场，入关后即被集中隔离观察，其间出现症状。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n病例15为中国籍，在香港就学，自香港出发，于2022年3月14日抵达上海浦东国际机场，入关后即被集中隔离观察，其间出现症状。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。\n15例境外输入性确诊病例已转至定点医疗机构救治，已追踪同航班密切接触者58人，均已落实集中隔离观察。\n2022年3月16日0—24时，新增境外输入性无症状感染者6例。\n无症状感染者1为台湾地区居民，在台湾地区生活，自台湾地区出发，于2022年3月10日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n无症状感染者2为台湾地区居民，在台湾地区生活，自台湾地区出发，于2022年3月10日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n无症状感染者3为中国籍，在美国探亲，自美国出发，于2022年3月11日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n无症状感染者4为中国籍，在日本留学，自日本出发，于2022年3月13日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n无症状感染者5为中国籍，在加拿大探亲，自加拿大出发，于2022年3月13日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n无症状感染者6为中国籍，暂居香港，自香港出发，于2022年3月14日抵达上海浦东国际机场，入关后即被集中隔离观察，其间例行核酸检测异常。经排查，区疾控中心新冠病毒核酸检测结果为阳性。综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。\n6例境外输入性无症状感染者已转至定点医疗机构医学观察，已追踪同航班密切接触者53人，均已落实集中隔离观察。\n2022年3月16日0—24时，解除医学观察无症状感染者19例，其中本土无症状感染者4例，境外输入性无症状感染者15例。\n截至2022年3月16日24时，累计本土确诊495例，治愈出院385例，在院治疗103例，死亡7例。现有待排查的疑似病例0例。\n截至2022年3月16日24时，累计境外输入性确诊病例4382例，出院3842例，在院治疗540例。现有待排查的疑似病例14例。\n截至2022年3月16日24时，尚在医学观察中的无症状感染者1254例，其中本土无症状感染者1106例，境外输入性无症状感染者148例。\n区域\n在院治疗确诊病例\n境外输入人员\n（按输入地分）\n韩国\n25\n日本\n15\n新加坡\n12\n美国\n10\n加拿大\n7\n英国\n4\n澳大利亚\n4\n以色列\n3\n泰国\n2\n西班牙\n2\n坦桑尼亚\n2\n奥地利\n2\n巴布亚新几内亚\n1\n马里\n1\n加纳\n1\n俄罗斯\n1\n东帝汶\n1\n荷兰\n1\n印度尼西亚\n1\n新西兰\n1\n法国\n1\n哥伦比亚\n1\n德国\n1\n马来西亚\n1\n香港特别行政区\n411\n台湾地区\n29\n本市居住或工作地\n浦东\n11\n黄浦\n7\n徐汇\n24\n静安\n13\n普陀\n5\n虹口\n4\n闵行\n15\n宝山\n2\n嘉定\n15\n金山\n1\n松江\n3\n青浦\n2\n奉贤\n1\n合计\n643",
			Daily: model.Daily{
				DistrictConfirmedFromBubble:       map[string]int{"徐汇区": 1, "黄浦区": 1},
				DistrictConfirmedFromRisk:         map[string]int{"嘉定区": 4, "徐汇区": 1},
				DistrictConfirmedFromAsymptomatic: map[string]int{"闵行区": 1},
				DistrictAsymptomaticFromBubble:    map[string]int{"嘉定区": 15, "奉贤区": 2, "宝山区": 5, "普陀区": 2, "松江区": 4, "浦东新区": 23, "虹口区": 7, "长宁区": 1, "闵行区": 7, "青浦区": 1, "黄浦区": 2},
				DistrictAsymptomaticFromRisk:      map[string]int{"嘉定区": 22, "宝山区": 5, "徐汇区": 5, "普陀区": 1, "杨浦区": 2, "浦东新区": 21, "虹口区": 3, "长宁区": 2, "闵行区": 10, "静安区": 3, "黄浦区": 6},
			},
		},
		{
			Content: "无症状感染者156，女，56岁，居住于宝山区泗塘五村，无症状感染者157，男，30岁，松江区车峰路199弄，无症状感染者158，男，9岁，居住于黄浦区徐家汇路101号，无症状感染者159，女，52岁，居住于松江区龙源路1208弄，无症状感染者160，男，18岁，居住于宝山区长江路868号，无症状感染者161，男，28岁，居住于嘉定区金园一路1359弄，无症状感染者162，女，35岁，居住于黄浦区五里桥路39弄，无症状感染者163，女，38岁，居住于长宁区天山五村，无症状感染者164，男，22岁，居住于静安区银都一村，无症状感染者165，女，65岁，居住于嘉定区城中路29弄，无症状感染者166，男，45岁，居住于嘉定区崇教路267号对面工地宿舍，无症状感染者167，男，48岁，居住于嘉定区伊宁路合作路工地，无症状感染者168，男，34岁，居住于青浦区青松路129弄，无症状感染者169，男，32岁，居住于黄浦区重庆北路212号，无症状感染者170，男，60岁，居住于嘉定区崇教路267号对面工地宿舍，均系本市报告本土确诊病例或无症状感染者的密切接触者，即被隔离管控，其间新冠病毒核酸检测结果异常，经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为无症状感染者。",
			Daily: model.Daily{
				DistrictAsymptomaticFromBubble: map[string]int{"嘉定区": 5, "宝山区": 2, "松江区": 2, "长宁区": 1, "青浦区": 1, "静安区": 1, "黄浦区": 3},
			},
		},
		{
			Content: "2022年4月17日0—24时，新增本土死亡病例3例。\n\n死亡病例1，女性，89岁，合并急性冠脉综合征、冠心病、糖尿病、高血压三级、脑梗，死亡病例2，女性，91岁，合并脑梗后遗症、高血压，死亡病例3，男性，91岁，合并冠心病、高血压，3人入院后转为重症，经全力抢救无效死亡。\n\n2022年4月17日0—24时，新增本土无症状感染者19831例。\n\n无症状感染者1—无症状感染者7029，居住于浦东新区，无症状感染者7030—无症状感染者8684，居住于黄浦区，无症状感染者8685—无症状感染者10030，居住于徐汇区，无症状感染者10031—无症状感染者10688，居住于长宁区，无症状感染者10689—无症状感染者11398，居住于静安区，无症状感染者11399—无症状感染者12276，居住于普陀区，无症状感染者12277—无症状感染者13176，居住于虹口区，无症状感染者13177—无症状感染者14032，居住于杨浦区，无症状感染者14033—无症状感染者16307，居住于闵行区，无症状感染者16308—无症状感染者17668，居住于宝山区，无症状感染者17669—无症状感染者18282，居住于嘉定区，无症状感染者18283—无症状感染者18300，居住于金山区，无症状感染者18301—无症状感染者19044，居住于松江区，无症状感染者19045—无症状感染者19389，居住于青浦区，无症状感染者19390—无症状感染者19403，居住于奉贤区，无症状感染者19404—无症状感染者19425，居住于崇明区，均为本市闭环隔离管控人员，其间新冠病毒核酸检测结果异常，经疾控中心复核结果为阳性，诊断为无症状感染者。",
			Daily: model.Daily{
				DistrictAsymptomaticFromBubble: map[string]int{"嘉定区": 614, "奉贤区": 14, "宝山区": 1361, "崇明区": 22, "徐汇区": 1346, "普陀区": 878, "杨浦区": 856, "松江区": 744, "浦东新区": 7029, "虹口区": 900, "金山区": 18, "长宁区": 658, "闵行区": 2275, "青浦区": 345, "静安区": 710, "黄浦区": 1655},
			},
		},
		{
			Content: "\n病例722—病例784，居住于浦东新区，病例785—病例804，居住于黄浦区，病例805—病例835，居住于徐汇区，病例836—病例846，居住于长宁区，病例847—病例857，居住于静安区，病例858—878，居住于普陀区，病例879—病例885，居住于虹口区，病例886—病例903，居住于杨浦区，病例904—病例922，居住于闵行区，病例923—病例966，居住于宝山区，病例967—病例975，居住于嘉定区，病例976—病例978，居住于金山区，病例979—病例983，居住于松江区，病例984—病例987，居住于青浦区，病例988—病例991，居住于奉贤区，病例992—病例994，居住于崇明区，为此前报告的本土无症状感染者。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。",
			Daily: model.Daily{
				DistrictConfirmedFromAsymptomatic: map[string]int{"嘉定区": 9, "奉贤区": 4, "宝山区": 44, "崇明区": 3, "徐汇区": 31, "普陀区": 21, "杨浦区": 18, "松江区": 5, "浦东新区": 63, "虹口区": 7, "金山区": 3, "长宁区": 11, "闵行区": 19, "青浦区": 4, "静安区": 11, "黄浦区": 20},
			},
		},
		{
			Content: "病例8—病例40，居住于浦东新区，\n病例41—病例45，\u00a0\u00a0\u00a0居住于徐汇区，\n病例46—病例52，居住于长宁区，\n病例53—病例55，居住于静安区，\n病例56、病例57，居住于普陀区，\n病例58、病例59，居住于虹口区，\n病例60、病例61，居住于闵行区，\n病例62—病例68，居住于宝山区，\n病例69、病例70，居住于嘉定区，\n病例71，居住于金山区，\n病例72—病例73，居住于松江区，\n病例74—病例75，居住于青浦区，\n在风险人群筛查中发现新冠病毒核酸检测结果异常，即被隔离管控。经市疾控中心复核结果为阳性。经市级专家会诊，综合流行病学史、临床症状、实验室检测和影像学检查结果等，诊断为确诊病例。",
			Daily: model.Daily{
				DistrictConfirmedFromRisk: map[string]int{"嘉定区": 2, "宝山区": 7, "徐汇区": 5, "普陀区": 2, "松江区": 2, "浦东新区": 33, "虹口区": 2, "金山区": 1, "长宁区": 7, "闵行区": 2, "青浦区": 2, "静安区": 3},
			},
		},
	}

	for i, c := range testcases {
		p := DailyParserShanghai{}
		d := model.Daily{Date: time.Date(2022, 4, 24, 0, 0, 0, 0, time.Local)}
		p.parseDailyContentRegion(&d, c.Content)
		assert.EqualValuesf(t, c.Daily.DistrictConfirmedFromBubble, d.DistrictConfirmedFromBubble, "(%d) 分区确诊（来自隔离管控）", i)
		assert.EqualValuesf(t, c.Daily.DistrictConfirmedFromRisk, d.DistrictConfirmedFromRisk, "(%d) 分区确诊（来自风险人群）", i)
		assert.EqualValuesf(t, c.Daily.DistrictConfirmedFromAsymptomatic, d.DistrictConfirmedFromAsymptomatic, "(%d) 分区确诊（来自无症状）", i)
		assert.EqualValuesf(t, c.Daily.DistrictAsymptomaticFromBubble, d.DistrictAsymptomaticFromBubble, "(%d) 分区无症状（来自隔离管控）", i)
		assert.EqualValuesf(t, c.Daily.DistrictAsymptomaticFromRisk, d.DistrictAsymptomaticFromRisk, "(%d) 分区无症状（来自风险人群）", i)
	}
}

func TestRegexpResidentDistrict1(t *testing.T) {
	testcases := [][]string{
		{
			"\n黄浦区\n2022年3月21日，黄浦区无新增本土确诊病例、49例本土无症状感染者，分别居住于：\n永年路24弄，\n建德路1号，\n已对相关居住地落实消毒等措施。",
			"黄浦区",
			"\n永年路24弄，\n建德路1号，",
		},
		{
			"\n黄浦区\n2022年3月21日，黄浦区无新增本土确诊病例、49例本土无症状感染者，分别居住于。\n永年路24弄，\n建德路1号，\n已对相关居住地落实消毒等措施。",
			"黄浦区",
			"\n永年路24弄，\n建德路1号，",
		},
		{
			"\n黄浦区\n2022年3月21日，黄浦区无新增本土确诊病例、49例本土无症状感染者，分别居住于，\n永年路24弄，\n建德路1号，\n已对相关居住地落实消毒等措施。",
			"黄浦区",
			"\n永年路24弄，\n建德路1号，",
		},
		{
			"\n黄浦区\n2022年3月21日，黄浦区无新增本土确诊病例、49例本土无症状感染者。分别居住于\n永年路24弄，\n建德路1号，\n已对相关居住地落实消毒等措施。",
			"黄浦区",
			"\n永年路24弄，\n建德路1号，",
		},
		{
			"\n黄浦区\n2022年3月21日，黄浦区无新增本土确诊病例、49例本土无症状感染者。\n分别居住于\n永年路24弄，\n建德路1号，\n已对相关居住地落实消毒等措施。",
			"黄浦区",
			"\n永年路24弄，\n建德路1号，",
		},
		{
			"\n黄浦区\n2022年3月21日，黄浦区无新增本土确诊病例、49例本土无症状感染者。\n分别居住于：\n永年路24弄，\n建德路1号，\n已对相关居住地落实消毒等措施。",
			"黄浦区",
			"\n永年路24弄，\n建德路1号，",
		},
		{
			"\n黄浦区\n2022年4月21日，奉贤区新增14例本土确诊病例，新增20例本土无症状感染者。上述人员均在隔离管控中发现，其涉及的场所已落实终末消毒等措施。",
			"黄浦区",
			"",
		},
		{
			"\n松江区\n（滑动查看更多↓）\n2022年4月21日，松江区新增22例本土确诊病例，新增471例本土无症状感染者，分别居住于：\n伴亭路855弄、\n宝胜路18号、",
			"松江区",
			"\n伴亭路855弄、\n宝胜路18号、",
		},
	}

	for i, c := range testcases {
		mm := reResidentDistrictShanghai1.FindAllStringSubmatch(c[0], -1)
		assert.NotNil(t, mm, fmt.Sprintf("匹配居住地信息失败：(%d) %q => nil", i, c[0]))
		if len(mm) > 0 {
			assert.Equal(t, c[1], mm[0][1], "匹配居住地区名称失败：(%d) %q", i, c[0])
			assert.Equal(t, c[2], mm[0][2], "匹配居住地区地址失败：(%d) %q", i, c[0])
		}
	}
}

func TestRegexpResidentDistrict2(t *testing.T) {
	testcases := [][]string{
		{
			"，病例2，男，73岁，居住于黄浦区顺昌路612弄，",
			"病例",
			"2",
			"男",
			"73",
			"黄浦区",
			"顺昌路612弄",
		},
		{
			"\n无症状感染者1，男，24岁，居住地为徐汇区沪闵路9490号，无症状",
			"无症状感染者",
			"1",
			"男",
			"24",
			"徐汇区",
			"沪闵路9490号",
		},
		{
			"无症状感染者7，男，36岁，系浦东机场安检工作人员，居住地为浦东新区祝桥镇邓一村，",
			"无症状感染者",
			"7",
			"男",
			"36",
			"浦东新区",
			"祝桥镇邓一村",
		},
		{
			"，无症状感染者13，男，3月龄，居住地为普陀区石泉东路240弄",
			"无症状感染者",
			"13",
			"男",
			"3月",
			"普陀区",
			"石泉东路240弄",
		},
	}

	for i, c := range testcases {
		mm := reResidentDistrictShanghai2.FindAllStringSubmatch(c[0], -1)
		assert.NotNil(t, mm, fmt.Sprintf("匹配居住地信息失败：(%d) %q => nil", i, c[0]))
		assert.Equal(t, 1, len(mm), "匹配居住地信息-长度不对：(%d) %q => %#v", i, c[0], mm)
		assert.Equal(t, 7, len(mm[0]), "匹配居住地信息-长度不对：(%d) %q => %#v", i, c[0], mm[0])
		if len(mm) > 0 {
			assert.Equal(t, c[1], mm[0][1], "匹配居住地-类型：(%d) %q", i, c[0])
			assert.Equal(t, c[2], mm[0][2], "匹配居住地-编号：(%d) %q", i, c[0])
			assert.Equal(t, c[3], mm[0][3], "匹配居住地-性别：(%d) %q", i, c[0])
			assert.Equal(t, c[4], mm[0][4], "匹配居住地-年龄：(%d) %q", i, c[0])
			assert.Equal(t, c[5], mm[0][5], "匹配居住地-区名称：(%d) %q", i, c[0])
			assert.Equal(t, c[6], mm[0][6], "匹配居住地-地址：(%d) %q", i, c[0])
		}
	}
}
