#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import os

import matplotlib.pyplot as plt
import matplotlib.ticker as mtick

from report import Report


def single_report(report: Report):
    df = report.df

    fig, ax1 = plt.subplots()

    color = 'tab:red'
    ax1.plot(df['elapsed_time'], df['utilization'], color=color)
    ax1.set_ylabel('Memory budget utilization, %', color=color)
    ax1.yaxis.set_tick_params(labelcolor=color)
    ax1.yaxis.set_major_formatter(mtick.PercentFormatter())
    ax1.set_xlabel('Time, seconds')

    ax2 = ax1.twinx()

    color = 'tab:blue'
    ax2.plot(df['elapsed_time'], df['gogc'], color=color)
    ax2.yaxis.set_tick_params(labelcolor=color)
    ax2.set_ylabel('GOGC', color=color)

    # fig.tight_layout()
    fig.savefig(report.plot_file_path, transparent=False)

