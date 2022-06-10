#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import humanize as humanize
import matplotlib.pyplot as plt
import matplotlib.ticker

from report import Report


@matplotlib.ticker.FuncFormatter
def bytes_major_formatter(x, pos):
    return humanize.naturalsize(int(x), binary=True).replace(".0", "")


def single_report(report: Report):
    df = report.df

    fig, ax = plt.subplots()
    ax.set_xlim(0, 30)

    twin = ax.twinx()

    ax.set_xlabel('Time, seconds')

    # RSS plot
    color = 'tab:red'
    l1 = ax.plot(df['elapsed_time'], df['rss'], color=color, label='RSS, bytes')
    ax.set_ylabel('RSS, bytes')
    ax.set_ylim(0, 1024 * 1024 * 1024)
    ax.set_yticks([ml * 1024 * 1024 for ml in (256, 512, 512 + 256, 1024)])
    ax.yaxis.set_major_formatter(bytes_major_formatter)

    # GOGC plot
    color = 'tab:blue'
    l2 = twin.plot(df['elapsed_time'], df['gogc'], color=color, label='GOGC')
    twin.set_ylabel('GOGC')
    twin.set_ylim(-5, 105)

    # legend
    ls = l1 + l2
    labels = [l.get_label() for l in ls]
    ax.legend(ls, labels, loc=0)

    # title
    if report.session.params.unlimited:
        title = "MemLimiter disabled"
    else:
        coefficient = report.session.params.coefficient
        rss_limit = report.session.params.rss_limit
        title = f'MemLimiter enabled, $RSS_{{limit}}$ = {rss_limit}, $K_{{p}} = {coefficient}$'
    ax.title.set_text(title)

    fig.tight_layout()
    fig.savefig(report.plot_file_path, transparent=False)
