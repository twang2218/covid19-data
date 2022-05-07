import pandas as pd
print('pandas:', pd.__version__)
import numpy as np
print('numpy:', np.__version__)
import matplotlib
print('matplotlib:', matplotlib.__version__)
import matplotlib.pyplot as plt
import seaborn as sns
print('seaborn:', sns.__version__)

plt.rcParams['figure.facecolor'] = 'white'


from lib import *

matplotlib_cjk_font_setup()

import os
file_dir = os.path.dirname(os.path.realpath(__file__))
data_dir = os.path.realpath(os.path.join(file_dir, "../data"))
print("<data_dir>: ", data_dir)
meta = {
    "shanghai": {
        "name": "上海市",
        "name_pinyin": "shanghai",
        "file_daily": os.path.join(data_dir, "shanghai-daily.csv"),
        "file_residents": os.path.join(data_dir, "shanghai-daily-residents.csv"),
        "date_range": {"from": "2022-03-06", "to": "2022-06-01"},
        "districts": [
            "浦东新区", "徐汇区", "闵行区", "黄浦区", "嘉定区", "松江区", "虹口区", "长宁区",
            "青浦区", "静安区", "宝山区", "杨浦区", "普陀区", "崇明区", "金山区", "奉贤区"
        ],
        "events": [
            {"from": "2022-03-18", "title": "非重点区分批核算检测"},
            {"from": "2022-03-31", "title": "上海全域静态管理"},
            # {"from": "2022-04-07", "title": "指数上升趋势放缓"},
            # {"from": "2022-04-10", "to": "2022-04-17", "title": "平台期"},
            {"from": "2022-04-04", "title": "全员核酸检测"},
            {"from": "2022-04-09", "title": "全市抗原检测"},
            {"from": "2022-04-22", "title": "全员核酸检测"},
            {"from": "2022-04-26", "title": "全员核酸检测"},
        ]
    },
    "beijing": {
        "name": "北京市",
        "name_pinyin": "beijing",
        "file_daily": os.path.join(data_dir, "beijing-daily.csv"),
        "file_residents": os.path.join(data_dir, "beijing-daily-residents.csv"),
        "date_range": {"from": "2022-04-15", "to": "2022-06-01"},
        "districts": [
            "朝阳区", "东城区", "西城区", "海淀区", "房山区", "丰台区", "石景山区", "门头沟区",
           	"大兴区", "通州区", "顺义区", "昌平区", "怀柔区", "平谷区", "密云区", "延庆区",
        ],
        "events": [
            {"from": "2022-04-26", "title": "全员核酸检测"},
            {"from": "2022-05-01", "title": "全员核酸检测"},
        ]
    }
}



def draw_city(meta_city, df_daily):
    import os
    import pathlib

    file_dir = os.path.dirname(os.path.realpath(__file__))
    base_dir = os.path.join(file_dir, "figures", meta_city['name_pinyin'])

    pathlib.Path(base_dir).mkdir(parents=True, exist_ok=True)

    # 画全局图
    if max(df_daily['从风险人群中发现的本土病例']) > 0 and max(df_daily['重型']) > 0 and max(df_daily['本土确诊病例']) > 0:
        fig, axes = plt.subplots(2, 2, figsize = (40,25), sharex=True)

        plt.suptitle('{}疫情每日新增数据变化'.format(meta_city['name']), y=0.90, fontsize=30, fontweight='bold', va='center')

        draw_positive('{}全市'.format(meta_city['name']), df_daily['日期'], df_daily['本土确诊病例'], df_daily['本土无症状感染者'], events=meta_city['events'], show_legend=True, show_minor_label=True, ax=axes[0,0])
        draw_positive('{}社会面'.format(meta_city['name']), df_daily['日期'], df_daily['从风险人群中发现的本土病例'], df_daily['从风险人群中发现的无症状感染者'], show_minor_label=True, ax=axes[1,0])

        draw_hospital('住院病例分析', df_daily['日期'], df_daily['本土确诊病例']+df_daily['本土无症状感染者'], df_daily['本土在院治疗'], df_daily['本土病例出院'], df_daily['解除医学观察'], ax=axes[0,1])
        draw_critical('疫情重症分析', df_daily['日期'], df_daily['重型'], df_daily['危重型'], df_daily['本土死亡病例'], ax=axes[1,1])

        plt.subplots_adjust(wspace=0.03, hspace=0.03)

        plt.savefig(os.path.join(base_dir, "daily_overall_analysis.png"), bbox_inches='tight')

        plt.show(block=False)
    
    # 画全局临床分型图
    if max(df_daily['轻型']) > 0 and max(df_daily['本土确诊病例']) > 0:
        fig, ax = plt.subplots(1, 1, figsize=(15, 10), sharex=True)
        draw_type('疫情分型分析', df_daily['日期'],
                df_daily['本土无症状感染者'],
                df_daily['轻型'],
                df_daily['普通型'],
                df_daily['重型'],
                df_daily['危重型'],
                df_daily['本土死亡病例'],
                events=meta_city['events'],
                ax=ax)
        plt.subplots_adjust(wspace=0.02, hspace=0.03)
        plt.savefig(os.path.join(base_dir, "daily_type_analysis.png"), bbox_inches='tight')
        plt.show(block=False)

    
    # 画各区数据图
    import math

    if len(meta_city['districts']) > 0:
        nrows = math.ceil(len(meta_city['districts']) / 4)
        fig, axes = plt.subplots(nrows, 4, figsize = (30,20), sharex=True, sharey=True)
        plt.suptitle('{}分区每日新增数据变化'.format(meta_city['name']), y=0.90, fontsize=30, fontweight='bold', va='center')
        for i, d in enumerate(meta_city['districts']):
            draw_positive(d, df_daily['日期'], df_daily['{}_确诊'.format(d)], df_daily['{}_无症状'.format(d)],
                # events=meta_city['events'],
                paddings=[-20, 0],
                show_minor_label=False, ax=axes[int(i/4), i%4])

        plt.subplots_adjust(wspace=0.02, hspace=0.03)
        plt.savefig(os.path.join(base_dir, "daily_district_positive.png"), bbox_inches='tight')
        plt.show(block=False)

    # 画各区社会面的图
    import math

    if len(meta_city['districts']) > 0 and max(np.max(df_daily[['{}_确诊_来自风险人群'.format(d) for d in meta_city['districts']]])) > 0:
        nrows = math.ceil(len(meta_city['districts']) / 4)
        fig, axes = plt.subplots(nrows, 4, figsize = (30,20), sharex=True, sharey=True)
        plt.suptitle('{}分区(社会面)每日新增数据变化'.format(meta_city['name']), y=0.90, fontsize=30, fontweight='bold', va='center')
        for i, d in enumerate(meta_city['districts']):
            draw_positive(d, df_daily['日期'], df_daily['{}_确诊_来自风险人群'.format(d)], df_daily['{}_无症状_来自风险人群'.format(d)], show_minor_label=False, ax=axes[int(i/4), i%4])

        plt.subplots_adjust(wspace=0.02, hspace=0.03)
        plt.savefig(os.path.join(base_dir, "daily_district_community.png"), bbox_inches='tight')

        plt.show(block=False)


# Entrypoint
if __name__ == "__main__":
    for city in ['shanghai', 'beijing']:
        meta_city = meta[city]
        df_daily, df_residents = load_data(meta=meta_city)
        draw_city(meta_city, df_daily)
