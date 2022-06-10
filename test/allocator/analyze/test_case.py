#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import dataclasses
import os


@dataclasses.dataclass
class TestCase:
    session_dir_path: os.PathLike
    unlimited: bool
    rss_limit: str
    coefficient: int
