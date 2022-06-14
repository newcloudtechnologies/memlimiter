#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import dataclasses
import os
from pathlib import Path
from typing import Iterable, Final

GIGABYTE: Final = 1024 * 1024 * 1024


@dataclasses.dataclass
class Params:
    unlimited: bool
    rss_limit: int = GIGABYTE
    coefficient: int = 20
    load_duration: str = '60s'

    def __str__(self) -> str:
        return f"unlimited_{self.unlimited}_rss_limit_{self.rss_limit}_coefficient_{self.coefficient}"

    @property
    def rss_limit_str(self):
        return f'{self.rss_limit}b'


class Session:
    params: Params
    dir_path: os.PathLike

    def __init__(self, case: Params, root_dir: os.PathLike):
        self.params = case
        self.dir_path = Path(root_dir, str(case))
        os.makedirs(self.dir_path, 0o777)


def make_sessions(root_dir: os.PathLike) -> Iterable[Session]:
    cases = (
        Params(unlimited=True, load_duration="60s", rss_limit=GIGABYTE),
        Params(unlimited=False, load_duration="60s", rss_limit=GIGABYTE, coefficient=1),
        Params(unlimited=False, load_duration="60s", rss_limit=GIGABYTE, coefficient=5),
        Params(unlimited=False, load_duration="60s", rss_limit=GIGABYTE, coefficient=10),
        Params(unlimited=False, load_duration="60s", rss_limit=GIGABYTE, coefficient=50),
        Params(unlimited=False, load_duration="60s", rss_limit=GIGABYTE, coefficient=100),
    )

    # FIXME: remove after debug
    # cases = (
    #     Params(unlimited=True, load_duration="20s", rss_limit=GIGABYTE),
    #     Params(unlimited=False, load_duration="20s", rss_limit=GIGABYTE, coefficient=10),
    # )

    return (Session(case=tc, root_dir=root_dir) for tc in cases)
