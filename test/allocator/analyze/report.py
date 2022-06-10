#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import dataclasses
import os
from pathlib import Path

import pandas as pd

from testing import Session


@dataclasses.dataclass
class Report:
    df: pd.DataFrame
    session: Session

    @classmethod
    def from_file(cls, path: os.PathLike, session: Session):
        return Report(
            df=Report.__parse_tracker_stats(path),
            session=session,
        )

    @staticmethod
    def __parse_tracker_stats(path: os.PathLike) -> pd.DataFrame:
        df = pd.read_csv(path)
        df['timestamp'] = pd.to_datetime(df['timestamp'])
        df['utilization'] *= 100
        df['elapsed_time'] = (df['timestamp'] - df['timestamp'].min()).apply(
            lambda x: x.seconds + x.microseconds / 1000000)
        return df

    @property
    def plot_file_path(self) -> os.PathLike:
        return Path(self.session.dir_path, "report.png")