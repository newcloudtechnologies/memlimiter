#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import dataclasses
import os
from pathlib import Path
from typing import Iterable, Final

GIGABYTE: Final = 1024 * 1024 * 1024
MEBIBYTE: Final = 1024 * 1024
DEFAULT_GO_MEMORY_LIMIT: Final = 800 * MEBIBYTE
DEFAULT_MIN_GOGC: Final = 10


@dataclasses.dataclass
class Params:
    unlimited: bool
    rss_limit: int = GIGABYTE
    go_memory_limit: int = DEFAULT_GO_MEMORY_LIMIT
    min_gogc: int = DEFAULT_MIN_GOGC
    coefficient: float = 20.0
    load_duration: str = '60s'

    def __str__(self) -> str:
        return (
            f"unlimited_{self.unlimited}"
            f"_rss_limit_{self.rss_limit}"
            f"_go_memory_limit_{self.go_memory_limit}"
            f"_min_gogc_{self.min_gogc}"
            f"_coefficient_{self.coefficient_str}"
        )

    @property
    def rss_limit_str(self):
        return f'{self.rss_limit}b'

    @property
    def go_memory_limit_str(self):
        if self.go_memory_limit <= 0:
            return "0"
        return f'{self.go_memory_limit}b'

    @property
    def coefficient_str(self):
        if type(self.coefficient) == float and self.coefficient.is_integer():
            return str(int(self.coefficient))
        else:
            return str(self.coefficient)


class Session:
    params: Params
    dir_path: os.PathLike

    def __init__(self, case: Params, root_dir: os.PathLike):
        self.params = case
        self.dir_path = Path(root_dir, str(case))
        os.makedirs(self.dir_path, 0o777)


def make_sessions(root_dir: os.PathLike) -> Iterable[Session]:
    duration = "60s"
    cases = (
        Params(unlimited=True, load_duration=duration, rss_limit=GIGABYTE, go_memory_limit=0),
        Params(
            unlimited=False,
            load_duration=duration,
            rss_limit=GIGABYTE,
            go_memory_limit=0,
            min_gogc=DEFAULT_MIN_GOGC,
            coefficient=0.5,
        ),
        Params(
            unlimited=False,
            load_duration=duration,
            rss_limit=GIGABYTE,
            go_memory_limit=DEFAULT_GO_MEMORY_LIMIT,
            min_gogc=DEFAULT_MIN_GOGC,
            coefficient=0.5,
        ),
        Params(
            unlimited=False,
            load_duration=duration,
            rss_limit=GIGABYTE,
            go_memory_limit=DEFAULT_GO_MEMORY_LIMIT,
            min_gogc=DEFAULT_MIN_GOGC,
            coefficient=5,
        ),
        Params(
            unlimited=False,
            load_duration=duration,
            rss_limit=GIGABYTE,
            go_memory_limit=DEFAULT_GO_MEMORY_LIMIT,
            min_gogc=DEFAULT_MIN_GOGC,
            coefficient=10,
        ),
        Params(
            unlimited=False,
            load_duration=duration,
            rss_limit=GIGABYTE,
            go_memory_limit=DEFAULT_GO_MEMORY_LIMIT,
            min_gogc=30,
            coefficient=50,
        ),
    )

    # FIXME: Remove after debug.
    # cases = (
    #     Params(unlimited=True, load_duration="10s", rss_limit=GIGABYTE),
    # )

    return (Session(case=tc, root_dir=root_dir) for tc in cases)
