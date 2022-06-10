#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import os
from datetime import datetime
from pathlib import Path
from typing import Final

import docker
import jinja2

from report import Report
from test_case import TestCase

image_tag: Final = 'allocator'
dockerfile_path: Final = 'test/allocator'


class PerfConfigRenderer:
    __t: Final = '''
{
  "endpoint": "localhost:1988",
  "rps": 100,
  "load_duration": "30s",
  "allocation_size": "1M",
  "pause_duration": "5s",
  "request_timeout": "1m"
}
    '''
    __template: jinja2.Template

    def __init__(self):
        self.__template = jinja2.Template(self.__t)

    def render(self,
               path: os.PathLike):
        out = self.__template.render()

        with open(path, "w") as f:
            f.write(out)


class ServerConfigRenderer:
    __t: Final = '''
{
    {%  if not unlimited %}
  "memlimiter": {
    "controller_nextgc": {
      "rss_limit": "{{ rss_limit }}",
      "danger_zone_gogc": 50,
      "danger_zone_throttling": 90,
      "period": "100ms",
      "component_proportional": {
        "coefficient": {{ coefficient }},
        "window_size": 20
      }
    }
  },
    {%  endif %}
  "listen_endpoint": "0.0.0.0:1988",
  "tracker": {
    "path": "/etc/allocator/tracker.csv",
    "period": "100ms"
  }
}
    '''
    __template: jinja2.Template

    def __init__(self):
        self.__template = jinja2.Template(self.__t)

    def render(self,
               path: os.PathLike,
               test_case: TestCase,
               ):
        out = self.__template.render(
            unlimited=test_case.unlimited,
            rss_limit=test_case.rss_limit,
            coefficient=test_case.coefficient,
        )

        with open(path, "w") as f:
            f.write(out)


class DockerClient:
    client: docker.client.DockerClient

    def __init__(self):
        self.client = docker.client.from_env()
        self.__build_image()

    def __build_image(self):
        image, logs = self.client.images.build(path=dockerfile_path, tag=image_tag)
        for log in logs:
            print(log)

    def execute(self, mem_limit: str, session_dir_path: os.PathLike):
        container = self.client.containers.run(
            name='allocator',
            image=image_tag,
            mem_limit=mem_limit,
            volumes={
                str(session_dir_path): {
                    'bind': '/etc/allocator',
                    'mode': 'rw',
                }
            },
            detach=True,
        )

        _, logs = container.exec_run(
            cmd='/usr/local/bin/allocator perf -c /etc/allocator/perf_config.json',
            stream=True,
        )

        for log in logs:
            print(log)


def run_session(
        docker_client: DockerClient,
        server_config_renderer: ServerConfigRenderer,
        perf_config_renderer: PerfConfigRenderer,
        root_dir: os.PathLike,
        test_case: TestCase,
) -> Report:
    # make session directory
    now = datetime.now()
    session_dir_path = Path(root_dir, f'allocator_{now.hour}{now.minute}{now.second}')
    os.makedirs(session_dir_path, mode=0o777)

    server_config_path = Path(session_dir_path, "server_config.json")
    server_config_renderer.render(path=server_config_path,
                                  test_case=test_case)

    perf_config_path = Path(session_dir_path, "perf_config.json")
    perf_config_renderer.render(path=perf_config_path)

    # run test session within Docker container
    docker_client.execute(
        mem_limit=test_case.rss_limit,
        session_dir_path=session_dir_path,
    )

    # parse output
    tracker_path = Path(session_dir_path, 'tracker.csv')
    return Report.from_file(test_case=test_case, path=tracker_path)


def main():
    docker_client = DockerClient()
    server_config_renderer = ServerConfigRenderer()
    perf_config_renderer = PerfConfigRenderer()

    test_case = TestCase(unlimited=False, rss_limit='1G', coefficient=20)

    report = run_session(
        docker_client=docker_client,
        server_config_renderer=server_config_renderer,
        perf_config_renderer=perf_config_renderer,
        root_dir=Path('/tmp/allocator'),
        test_case=test_case,
    )

    print(report)


if __name__ == '__main__':
    main()
