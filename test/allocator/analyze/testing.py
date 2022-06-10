#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import dataclasses
import os
from pathlib import Path
from typing import Iterable


@dataclasses.dataclass
class Params:
    unlimited: bool
    rss_limit: str = '1G'
    coefficient: int = 20

    def __str__(self) -> str:
        return f"unlimited_{self.unlimited}_rss_limit_{self.rss_limit}_coefficient_{self.coefficient}"


class Session:
    params: Params
    dir_path: os.PathLike

    def __init__(self, case: Params, root_dir: os.PathLike):
        self.params = case
        self.dir_path = Path(root_dir, str(case))
        os.makedirs(self.dir_path, 0o777)


def make_sessions(root_dir: os.PathLike) -> Iterable[Session]:
    cases = (
        Params(unlimited=True, rss_limit='1G'),
        Params(unlimited=False, rss_limit='1G', coefficient=20),
    )

    return (Session(case=tc, root_dir=root_dir) for tc in cases)
