#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import humanize as humanize
import matplotlib.pyplot as plt
import matplotlib.ticker
import matplotlib.ticker as mtick

from report import Report


@matplotlib.ticker.FuncFormatter
def bytes_major_formatter(x, pos):
    return humanize.naturalsize(x, binary=True)


def single_report(report: Report):
    df = report.df

    fig, ax = plt.subplots()
    fig.subplots_adjust(right=0.75)

    twin1 = ax.twinx()
    twin2 = ax.twinx()

    twin2.spines.right.set_position(("axes", 1.2))

    ax.set_xlabel('Time, seconds')

    color = 'tab:green'
    p1, = ax.plot(df['elapsed_time'], df['rss'], color=color)
    ax.set_ylabel('RSS, bytes', color=color)
    ax.set_ylim(0, 1024 * 1024 * 1024)
    ax.yaxis.set_major_formatter(bytes_major_formatter)

    color = 'tab:red'
    p2, = twin1.plot(df['elapsed_time'], df['utilization'], color=color)
    twin1.set_ylabel('Memory budget utilization, %', color=color)
    twin1.set_ylim(-5, 105)
    twin1.yaxis.set_tick_params(labelcolor=color)
    twin1.yaxis.set_major_formatter(mtick.PercentFormatter())

    color = 'tab:blue'
    p3, = twin2.plot(df['elapsed_time'], df['gogc'], color=color)
    twin2.yaxis.set_tick_params(labelcolor=color)
    twin2.set_ylabel('GOGC', color=color)
    twin2.set_ylim(-5, 105)

    # ax.legend(handles=[p1, p2, p3])

    fig.tight_layout()
    fig.savefig(report.plot_file_path, transparent=False)
