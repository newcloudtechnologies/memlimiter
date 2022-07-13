#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import dataclasses
import os

import pandas as pd

from testing import Session


@dataclasses.dataclass
class Report:
    df: pd.DataFrame
    session: Session

    @classmethod
    def from_file(cls, path: os.PathLike, session: Session):
        out = Report(
            df=Report.__parse_tracker_stats(path),
            session=session,
        )

        return out

    @staticmethod
    def __parse_tracker_stats(path: os.PathLike) -> pd.DataFrame:
        df = pd.read_csv(path)
        df['timestamp'] = pd.to_datetime(df['timestamp'])
        df['utilization'] *= 100
        return df

    def __post_init__(self):
        # Emulate OOM event for unconstrained process
        if self.session.params.unlimited:
            last_ts, last_but_one_ts = self.df['timestamp'].iloc[-1], self.df['timestamp'].iloc[-2]
            delta = last_ts - last_but_one_ts
            self.df.loc[len(self.df)] = [
                last_ts + delta,
                self.session.params.rss_limit,
                0, 0, 0,
            ]

        # compute elapsed time
        self.df['elapsed_time'] = (self.df['timestamp'] - self.df['timestamp'].min()).apply(
            lambda x: x.seconds + x.microseconds / 1000000)
