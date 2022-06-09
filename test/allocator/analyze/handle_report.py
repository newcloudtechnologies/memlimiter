#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE

import matplotlib.pyplot as plt
import matplotlib.ticker as mtick
import pandas as pd


def prepare() -> pd.DataFrame:
    df = pd.read_csv('/tmp/tracker.csv')
    df['timestamp'] = pd.to_datetime(df['timestamp'])
    df['utilization'] *= 100
    df['elapsed_time'] = (df['timestamp'] - df['timestamp'].min()).apply(lambda x: x.seconds + x.microseconds / 1000000)
    return df


def render(df: pd.DataFrame):
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

    fig.tight_layout()
    fig.savefig('/tmp/report.png', transparent=False)


def main():
    df = prepare()
    render(df)


if __name__ == '__main__':
    main()
