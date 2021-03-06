#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import os
from typing import List

import humanize as humanize
import matplotlib.pyplot as plt
import matplotlib.ticker
import numpy as np

from report import Report


@matplotlib.ticker.FuncFormatter
def bytes_major_formatter(x, pos):
    return humanize.naturalsize(int(x), binary=True).replace(".0", "")


def control_params_subplots(reports: List[Report], path: os.PathLike):
    ncols = 2
    nrows = 3
    if len(reports) != ncols * nrows:
        raise Exception("columns and rows mismatch")

    fig, axes = plt.subplots(ncols=2, nrows=3, figsize=(12, 15))

    ls, labels = None, None

    for i in range(nrows):
        for j in range(ncols):
            ix = i * ncols + j

            report = reports[ix]
            df = report.df
            ax = axes[i][j]

            ax.set_xlim(0, 60)

            ax.set_xlabel('Time, seconds')

            # RSS plot
            color = 'tab:red'
            l0 = ax.plot(df['elapsed_time'], df['rss'], color=color, label='RSS')
            ax.set_ylabel('RSS, bytes')
            ax.set_ylim(0, 1024 * 1024 * 1024)
            ax.set_yticks([ml * 1024 * 1024 for ml in (256, 512, 512 + 256, 1024)])
            ax.yaxis.set_major_formatter(bytes_major_formatter)

            # GOGC plot
            color = 'tab:blue'
            twin1 = ax.twinx()
            l1 = twin1.plot(df['elapsed_time'], df['gogc'], color=color, label='GOGC')
            twin1.set_ylabel('GOGC')
            twin1.set_ylim(-5, 105)

            # Throttling plot
            color = 'tab:green'
            twin2 = ax.twinx()
            twin2.spines.right.set_position(("axes", 1.2))
            l2 = twin2.plot(df['elapsed_time'], df['throttling'], color=color, label='Throttling')
            twin2.set_ylabel('Throttling')
            twin2.set_ylim(-5, 105)

            # legend
            if not ls or not labels:
                ls = l0 + l1 + l2
                labels = [l.get_label() for l in ls]

            # title
            if report.session.params.unlimited:
                title = "MemLimiter disabled"
            else:
                coefficient = report.session.params.coefficient_str
                title = f'MemLimiter enabled, $C_{{p}} = {coefficient}$'
            ax.title.set_text(title)

    fig.legend(ls, labels)
    fig.tight_layout()
    fig.savefig(path, transparent=False)


def rss_pivot(reports: List[Report], path: os.PathLike):
    fig, ax = plt.subplots(figsize=(8, 6))
    ax.set_xlim(0, 60)
    ax.set_xlabel('Time, seconds')
    ax.set_ylabel('RSS, bytes')
    ax.set_ylim(0, 1024 * 1024 * 1024)
    ax.set_yticks([ml * 1024 * 1024 for ml in (256, 512, 512 + 256, 1024)])
    ax.yaxis.set_major_formatter(bytes_major_formatter)

    n = len(reports)

    colors = plt.cm.turbo(np.linspace(0, 1, n))
    for i in range(n):
        report = reports[n - 1 - i]
        if report.session.params.unlimited:
            label = 'No limits'
        else:
            label = f'$C_{{p}} = {report.session.params.coefficient_str}$'

        ax.plot(report.df['elapsed_time'], report.df['rss'], color=colors[i], label=label)

    ax.legend()
    ax.title.set_text('RSS consumption dependence on $C_{{p}}$')
    fig.tight_layout()
    fig.savefig(path, transparent=False)
