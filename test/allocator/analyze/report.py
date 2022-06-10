#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import dataclasses
import os

import pandas as pd

from test_case import TestCase


@dataclasses.dataclass
class Report:
    df: pd.DataFrame
    test_case: TestCase

    @classmethod
    def from_file(cls, path: os.PathLike, test_case: TestCase):
        return Report(
            df=Report.__parse_tracker_stats(path),
            test_case=test_case,
        )

    @staticmethod
    def __parse_tracker_stats(path: os.PathLike) -> pd.DataFrame:
        df = pd.read_csv(path)
        df['timestamp'] = pd.to_datetime(df['timestamp'])
        df['utilization'] *= 100
        df['elapsed_time'] = (df['timestamp'] - df['timestamp'].min()).apply(
            lambda x: x.seconds + x.microseconds / 1000000)
        return df
