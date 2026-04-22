#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import os
import math
from typing import List

import humanize as humanize
import matplotlib.pyplot as plt
import matplotlib.ticker
import numpy as np

from report import Report


@matplotlib.ticker.FuncFormatter
def bytes_major_formatter(x, pos):
    return humanize.naturalsize(int(x), binary=True).replace(".0", "")


def _make_axes_grid(nplots: int, ncols: int = 2):
    nrows = max(1, math.ceil(nplots / ncols))
    fig, axes = plt.subplots(ncols=ncols, nrows=nrows, figsize=(12, 5 * nrows))

    # Normalize axes shape for both 1xN and Nx1 cases.
    flat_axes = np.atleast_1d(axes).reshape(-1)

    for ax in flat_axes[nplots:]:
        ax.set_visible(False)

    return fig, flat_axes


def _report_title(report: Report) -> str:
    params = report.session.params
    if params.unlimited:
        return "MemLimiter disabled"

    go_limit = params.go_memory_limit_str
    return f"Cp={params.coefficient_str}, MinGOGC={params.min_gogc}, GoLimit={go_limit}"


def control_params_subplots(reports: List[Report], path: os.PathLike):
    fig, axes = _make_axes_grid(len(reports), ncols=2)

    ls, labels = None, None

    for ix, report in enumerate(reports):
        df = report.df
        ax = axes[ix]

        ax.set_xlim(0, 60)

        ax.set_xlabel('Time, seconds')

        # RSS consumption plot.
        color = 'tab:red'
        l0 = ax.plot(df['elapsed_time'], df['rss'], color=color, label='RSS')
        ax.set_ylabel('RSS, bytes')
        ax.set_ylim(0, 1024 * 1024 * 1024)
        ax.set_yticks([ml * 1024 * 1024 for ml in (256, 512, 512 + 256, 1024)])
        ax.yaxis.set_major_formatter(bytes_major_formatter)

        # GOGC consumption plot.
        color = 'tab:blue'
        twin1 = ax.twinx()
        l1 = twin1.plot(df['elapsed_time'], df['gogc'], color=color, label='GOGC')
        twin1.set_ylabel('GOGC')
        twin1.set_ylim(-5, 105)

        # Throttling plot.
        color = 'tab:green'
        twin2 = ax.twinx()
        twin2.spines.right.set_position(("axes", 1.2))
        l2 = twin2.plot(df['elapsed_time'], df['throttling'], color=color, label='Throttling')
        twin2.set_ylabel('Throttling')
        twin2.set_ylim(-5, 105)

        # Legend.
        if not ls or not labels:
            ls = l0 + l1 + l2
            labels = [l.get_label() for l in ls]

        ax.title.set_text(_report_title(report))

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
            label = (
                f'$C_{{p}}={report.session.params.coefficient_str}$, '
                f'GML={report.session.params.go_memory_limit_str}, '
                f'MinGOGC={report.session.params.min_gogc}'
            )

        ax.plot(report.df['elapsed_time'], report.df['rss'], color=colors[i], label=label)

    ax.legend()
    ax.title.set_text('RSS consumption dependence on $C_{{p}}$')
    fig.tight_layout()
    fig.savefig(path, transparent=False)


def gogc_floor_hits(reports: List[Report], path: os.PathLike):
    active_reports = [report for report in reports if not report.session.params.unlimited]
    if not active_reports:
        raise Exception("no memlimiter-enabled reports")

    labels = []
    ratios = []
    for report in active_reports:
        params = report.session.params
        gogc_series = report.df['gogc']
        floor_hits = (gogc_series <= params.min_gogc).sum()
        ratio = 100.0 * floor_hits / len(gogc_series)

        labels.append(
            f"Cp={params.coefficient_str}\n"
            f"Min={params.min_gogc}\n"
            f"GML={params.go_memory_limit_str}"
        )
        ratios.append(ratio)

    fig, ax = plt.subplots(figsize=(12, 6))
    bars = ax.bar(labels, ratios, color='tab:blue')
    ax.set_ylim(0, 100)
    ax.set_ylabel('Share of samples at GOGC floor, %')
    ax.set_title('How often MinGOGC clamp is hit')
    ax.grid(axis='y', alpha=0.3)

    for bar, value in zip(bars, ratios):
        ax.text(
            bar.get_x() + bar.get_width() / 2,
            value + 1,
            f"{value:.1f}%",
            ha='center',
            va='bottom',
        )

    fig.tight_layout()
    fig.savefig(path, transparent=False)


def memory_limits_overlay(reports: List[Report], path: os.PathLike):
    fig, axes = _make_axes_grid(len(reports), ncols=2)

    for ix, report in enumerate(reports):
        params = report.session.params
        df = report.df
        ax = axes[ix]

        ax.set_xlim(0, 60)
        ax.set_xlabel('Time, seconds')
        ax.set_ylabel('Memory, bytes')
        ax.yaxis.set_major_formatter(bytes_major_formatter)

        ax.plot(df['elapsed_time'], df['rss'], color='tab:red', label='RSS')
        ax.plot(df['elapsed_time'], df['go_runtime_bytes'], color='tab:purple', label='Go runtime memory')
        ax.axhline(params.rss_limit, color='black', linestyle='--', linewidth=1.2, label='RSS limit')

        if params.go_memory_limit > 0:
            ax.axhline(
                params.go_memory_limit,
                color='tab:orange',
                linestyle='--',
                linewidth=1.2,
                label='Go memory limit',
            )

        y_candidates = [params.rss_limit, df['rss'].max(), df['go_runtime_bytes'].max()]
        if params.go_memory_limit > 0:
            y_candidates.append(params.go_memory_limit)
        ymax = max(y_candidates) * 1.1
        ax.set_ylim(0, ymax)

        ax.title.set_text(_report_title(report))
        ax.legend(loc='upper left')

    fig.tight_layout()
    fig.savefig(path, transparent=False)
