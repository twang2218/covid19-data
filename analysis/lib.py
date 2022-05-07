import pandas as pd
from sympy import false
print('pandas:', pd.__version__)
import numpy as np
print('numpy:', np.__version__)
import matplotlib
print('matplotlib:', matplotlib.__version__)
import matplotlib.pyplot as plt


def matplotlib_cjk_font_setup():
    import subprocess

    mat_fonts = set(f.name for f in matplotlib.font_manager.FontManager().ttflist)

    output = subprocess.check_output('fc-list :lang=zh -f "%{family}\n"', shell=True)
    output = output.decode('utf-8')
    # print '*' * 10, '系统可用的中文字体', '*' * 10
    # print output
    zh_fonts = set(f.split(',', 1)[0] for f in output.split('\n'))
    available = mat_fonts & zh_fonts

    print('*' * 10, '可用的字体', '*' * 10)
    for f in available:
        print(f)

    # 将中文字体放在备选字体首位

    # ********** 可用的字体 **********
    # AR PL UMing CN
    # Noto Sans CJK JP
    # AR PL UKai CN
    # Noto Serif CJK JP
    # Droid Sans Fallback

    fonts = matplotlib.rcParamsDefault['font.sans-serif'].copy()
    fonts.insert(0, 'Droid Sans Fallback')
    fonts.insert(0, 'Noto Sans CJK JP')
    matplotlib.rcParams['font.sans-serif'] = fonts
    print('<font.sans-serif>: ', matplotlib.rcParams['font.sans-serif'])

    fonts = matplotlib.rcParamsDefault['font.serif'].copy()
    fonts.insert(0, 'Droid Sans Fallback')
    fonts.insert(0, 'Noto Serif CJK JP')
    matplotlib.rcParams['font.serif'] = fonts
    print('<font.serif>: ', matplotlib.rcParams['font.serif'])

    fonts = matplotlib.rcParamsDefault['font.monospace'].copy()
    fonts.insert(0, 'Droid Sans Fallback')
    matplotlib.rcParams['font.monospace'] = fonts
    print('<font.monospace>: ', matplotlib.rcParams['font.monospace'])


def load_data(meta):
    df_daily = pd.read_csv(meta['file_daily'], dtype=int, converters={'来源': str},
                           parse_dates=['日期'], infer_datetime_format=True)
    df_residents = pd.read_csv(meta['file_residents'],
                               parse_dates=['日期'], infer_datetime_format=True)
    # 限定日期范围
    date_from, date_to = meta['date_range']['from'], meta['date_range']['to']
    df_daily = df_daily[df_daily['日期'].between(date_from, date_to)]
    df_residents = df_residents[df_residents['日期'].between(date_from, date_to)]
    # 确保顺序
    df_daily.sort_values(by='日期', ascending=True, inplace=True)
    df_residents.sort_values(by=['日期', '分型', '区', '居住地', '年龄'],
                             ascending=True, inplace=True)
    # 合成信息
    df_residents['地址'] = df_residents['市'] + \
        df_residents['区'] + \
        df_residents['居住地']
    df_residents['标签'] = np.where(df_residents['分型'].notna(),
                                  df_residents['地址'] + '\n' + df_residents['分型'] + '，' +
                                  df_residents['性别'] + '，' +
                                  df_residents['年龄'].astype(str) + '岁',
                                  df_residents['地址'])

    return df_daily, df_residents

from datetime import datetime, timedelta
one_day = timedelta(days=1)

# 绘图函数

def plot_stacked_bar(indexes, data, labels, colors, paddings=[], ax=None):
    # 画 堆叠bar 图
    base = 0
    for i, d in enumerate(data):
        bar = ax.bar(x=indexes, height=d, label=labels[i],
            clip_on=True,
            width=1,
            color=colors[i],
            edgecolor='w',
            linewidth=2,
            alpha=1,
            bottom=base,
            )
        if len(paddings) > 0:
            padding = paddings[i]
        else:
            padding = 0
        # ax.bar_label(bar, d.astype('int32'), rotation=90, padding=-max(d)*0.002, color='w', fontsize=15, weight='bold')
        ax.bar_label(bar, d.astype('int32'), rotation=90, padding=padding, color='w', fontsize=15, weight='bold')
        base = d + base

def plot_adjacent_bar(indexes, data, labels, colors, paddings=[], ax=None):
    # 画 相邻bar 图
    bar_width = 0.9/len(data)
    for i, d in enumerate(data):
        bar_shift = (i*bar_width) * one_day
        bar = ax.bar(x=indexes+bar_shift, height=d, label=labels[i],
                     clip_on=True,
                     width=bar_width,
                     color=colors[i],
                     edgecolor='w',
                     linewidth=0.2,
                     alpha=0.8,
                     )
        if len(paddings) > 0:
            padding = paddings[i]
        else:
            padding = 0
        ax.bar_label(bar, d.astype('int32'), rotation=90,
                     padding=padding, color='w', fontsize=15, weight='bold')


def plot_aux_line(events=[], ax=None):
    # 计算最高位
    position_top = ax.get_ylim()[1] * 0.99

    # 显示事件标注的辅助线
    for e in events:
        if 'to' in e and e['to']:
            # range
            ax.axvspan(
                datetime.fromisoformat(e['from']) - one_day/2,
                datetime.fromisoformat(e['to']) - one_day/2,
                color='grey', alpha=0.4,
                )

            ax.text(
                datetime.fromisoformat(e['from']) + one_day*4,
                position_top,
                e['title'],
                rotation=0,
                horizontalalignment='right',
                verticalalignment='top',
                multialignment='center',
                size=17,
                color='grey',
                alpha=0.6
                )
        else:
            # line
            ax.text(
                datetime.fromisoformat(e['from']) - one_day/1.8,
                position_top,
                e['title'],
                rotation=90,
                horizontalalignment='right',
                verticalalignment='top',
                multialignment='center',
                size=17,
                # color='darkblue',
                color='grey',
                alpha=0.6
                )
            ax.axvline(
                datetime.fromisoformat(e['from']) - one_day/2,
                0,position_top,
                # linestyle='-', linewidth=3, color='red', alpha=0.3,
                linestyle='-', linewidth=3, color='grey', alpha=0.3,
                )


def draw_positive(title, date, confirmed, asymptomatic, events=[], colors=None, paddings=None, show_legend=False, show_minor_label=False, ax=None):
    ax.grid(True, which='major', alpha=0.3)

    if colors is None:
        colors = [
            matplotlib.cm.Reds(matplotlib.colors.LogNorm()(confirmed)),
            matplotlib.cm.Wistia(matplotlib.colors.Normalize()(asymptomatic)),
        ]
    
    if paddings is None:
        paddings = [8, -40]

    # 画 bar 图
    plot_stacked_bar(
        indexes=date,
        data=[confirmed, asymptomatic],
        labels=['确诊病例', '无症状感染者'],
        colors=colors,
        paddings=paddings,
        ax=ax,
    )

    # 显示图例
    if show_legend:
        ax.legend(ncol = 2, loc = 'lower left', borderaxespad=4, fontsize=15)
        # # debug legend
        # leg = ax.get_legend()
        # print('legend color: {}: {}, {}: {}.'.format(
        #     0, leg.legendHandles[0].get_label(),
        #     1, leg.legendHandles[1],
        # ))

    # 显示事件标注的辅助线
    plot_aux_line(events=events, ax=ax)

    # 计算最高位
    position_top = ax.get_ylim()[1]

    # 标题
    t =  "更新至 {} ({})\n(新增 {:.0f} 确诊病例，新增 {:.0f} 无症状感染者)".format(date.max().strftime('%Y年%m月%d日'), title, confirmed.iloc[-1], asymptomatic.iloc[-1])
    ax.text(
        x=date.min() + one_day, y=position_top * 0.97,
        s=t,
        rotation=0,
        horizontalalignment='left',
        verticalalignment='top',
        multialignment='center',
        size='20',
        fontweight='bold',
        color='black',
        alpha=1
        )
    
    # 🎉 庆祝清零（渲染背景色稍微不同）
    if confirmed.iloc[-1] + asymptomatic.iloc[-1] == 0:
        ax.patch.set_facecolor('lavender')
        ax.patch.set_alpha(0.5)


    # X 轴日期格式调整
    ax.tick_params('x', which='both', labelrotation=90)
    # major
    ax.xaxis.set_major_formatter(matplotlib.dates.DateFormatter('%m月%d日'))
    if show_minor_label:
        # minor
        ax.xaxis.set_minor_locator(matplotlib.ticker.MultipleLocator(1))
        ax.xaxis.set_minor_formatter(matplotlib.dates.DateFormatter('%m月%d日'))

    ax.set_xlim([date.min(), date.max() + one_day*4/5])


def draw_type(title, date, asymptomatic, mild, normal, severe, critical, death, events=[], colors=None, paddings=None, show_minor_label=True, ax=None):
    ax.grid(True, which='major', alpha=0.3)

    if colors is None:
        colors = ['grey', 'brown', 'darkorange',
                  'orangered', 'orange', 'gold']
    if paddings is None:
        paddings = [-10, -10, -10, -20, -20, -15]

    plot_stacked_bar(
        indexes=date,
        data=[death, critical, severe, normal, mild, asymptomatic],
        labels=['死亡', '危重型', '重型', '普通型', '轻型', '无症状感染者'],
        colors=colors,
        paddings=paddings,
        ax=ax,
    )

    # 显示事件标注的辅助线
    plot_aux_line(events=events, ax=ax)

    # 调整X坐标标签
    ax.tick_params('x', which='both', labelrotation=90)
    # major
    ax.xaxis.set_major_formatter(matplotlib.dates.DateFormatter('%m月%d日'))
    if show_minor_label:
        # minor
        ax.xaxis.set_minor_locator(matplotlib.ticker.MultipleLocator(1))
        ax.xaxis.set_minor_formatter(matplotlib.dates.DateFormatter('%m月%d日'))

    # 计算最高位
    position_top = ax.get_ylim()[1]

    t = "更新至 {} ({})\n(今日无症状感染者 {:.0f} 例，轻型 {:.0f} 例，普通型 {:.0f} 例，\n重症 {:.0f} 例，危重症 {:.0f} 例，死亡 {:.0f} 例)".format(
        date.max().strftime('%Y年%m月%d日'), title,
        asymptomatic.iloc[-1], mild.iloc[-1], normal.iloc[-1],
        severe.iloc[-1], critical.iloc[-1], death.iloc[-1])
    ax.text(
        x=date.min() + one_day, y=position_top * 0.98,
        s=t,
        rotation=0,
        horizontalalignment='left',
        verticalalignment='top',
        multialignment='center',
        size='20',
        fontweight='bold',
        color='black',
        alpha=1
    )

    ax.legend(ncol=2, loc='lower left', borderaxespad=4, fontsize=15)
    ax.set_ylim(bottom=0)

def draw_critical(title, date, severe, critical, death, show_minor_label=True, ax=None):
    ax.grid(True, which='major', alpha=0.3)
    
    plot_stacked_bar(
        indexes=date,
        data=[death, critical, severe],
        labels=['死亡','危重症','重症'],
        colors=[
            # matplotlib.cm.Greys(matplotlib.colors.LogNorm()(death)),
            # matplotlib.cm.Reds(matplotlib.colors.LogNorm()(critical)),
            # matplotlib.cm.Wistia(matplotlib.colors.Normalize()(severe)),
            'grey',
            'brown',
            'darkorange',
        ],
        paddings=[8, 8, -30],
        ax=ax,
    )

    # 计算最高位
    position_top = ax.get_ylim()[1] * 0.97

    # 调整X坐标标签
    ax.tick_params('x', which='both', labelrotation=90)
    # major
    ax.xaxis.set_major_formatter(matplotlib.dates.DateFormatter('%m月%d日'))
    if show_minor_label:
        # minor
        ax.xaxis.set_minor_locator(matplotlib.ticker.MultipleLocator(1))
        ax.xaxis.set_minor_formatter(matplotlib.dates.DateFormatter('%m月%d日'))
    
    t =  "更新至 {} ({})\n(今日 重症 {:.0f} 例，危重症 {:.0f} 例，死亡 {:.0f} 例)".format(date.max().strftime('%Y年%m月%d日'), title, severe.iloc[-1], critical.iloc[-1], death.iloc[-1])
    ax.text(
        x=date.min() + one_day, y=position_top,
        s=t,
        rotation=0,
        horizontalalignment='left',
        verticalalignment='top',
        multialignment='center',
        size='20',
        fontweight='bold',
        color='black',
        alpha=1
        )

    ax.legend(ncol = 3, loc = 'lower left', borderaxespad=4, fontsize=15)
    ax.set_ylim(bottom=0)

def draw_hospital(title, date, positive, in_hospital, discharged_from_hospital, discharged_from_observation, show_minor_label=True, ax=None):
    ax.grid(True, which='major', alpha=0.3)

    # plot_adjacent_bar(
    #     indexes=date,
    #     data=[positive, in_hospital, discharged_from_hospital, discharged_from_observation],
    #     labels=['阳性病例', '在院治疗', '治愈出院', '解除医学观察'],
    #     colors=[
    #         # matplotlib.cm.Greys(matplotlib.colors.LogNorm()(death)),
    #         # matplotlib.cm.Reds(matplotlib.colors.LogNorm()(critical)),
    #         # matplotlib.cm.Wistia(matplotlib.colors.Normalize()(severe)),
    #         'darkorange',
    #         'brown',
    #         'darkgreen',
    #         'green',
    #     ],
    #     paddings=[8, 8, 8, -30],
    #     ax=ax,
    # )
    plot_adjacent_bar(
        indexes=date,
        data=[positive, in_hospital, discharged_from_hospital + discharged_from_observation],
        labels=['阳性病例', '在院治疗', '治愈出院+解除医学观察'],
        colors=[
            # matplotlib.cm.Greys(matplotlib.colors.LogNorm()(death)),
            # matplotlib.cm.Reds(matplotlib.colors.LogNorm()(critical)),
            # matplotlib.cm.Wistia(matplotlib.colors.Normalize()(severe)),
            'darkorange',
            'brown',
            'darkgreen',
        ],
        paddings=[8, 8, 8],
        ax=ax,
    )

    # 计算最高位
    position_top = ax.get_ylim()[1] * 0.97

    # 调整X坐标标签
    ax.tick_params('x', which='both', labelrotation=90)
    # major
    ax.xaxis.set_major_formatter(matplotlib.dates.DateFormatter('%m月%d日'))
    if show_minor_label:
        # minor
        ax.xaxis.set_minor_locator(matplotlib.ticker.MultipleLocator(1))
        ax.xaxis.set_minor_formatter(matplotlib.dates.DateFormatter('%m月%d日'))

    # 标题
    t = "更新至 {} ({})\n(今日阳性病例{:.0f}例，在院治疗{:.0f}例，治愈出院{:.0f}例，解除医学观察{:.0f}例)".format(date.max().strftime(
        '%Y年%m月%d日'), title, positive.iloc[-1], in_hospital.iloc[-1], discharged_from_hospital.iloc[-1], discharged_from_observation.iloc[-1])
    ax.text(
        x=date.min() + one_day, y=position_top,
        s=t,
        rotation=0,
        horizontalalignment='left',
        verticalalignment='top',
        multialignment='center',
        size='20',
        fontweight='bold',
        color='black',
        alpha=1
    )

    # 图例
    ax.legend(ncol=4, loc='center left', borderaxespad=4, fontsize=15)
    ax.set_ylim(bottom=0)
