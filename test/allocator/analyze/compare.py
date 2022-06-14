#  Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
#  Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
#  License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
import os
from datetime import datetime
from pathlib import Path
from typing import Final

import docker
import jinja2

import render
from report import Report
from testing import Session, make_sessions, Params

image_tag: Final = 'allocator'
dockerfile_path: Final = 'test/allocator'
container_name: Final = 'allocator'


class PerfConfigRenderer:
    __t: Final = '''
{
  "endpoint": "localhost:1988",
  "rps": 100,
  "load_duration": "{{ load_duration }}",
  "allocation_size": "1M",
  "pause_duration": "5s",
  "request_timeout": "1m"
}
    '''
    __template: jinja2.Template

    def __init__(self):
        self.__template = jinja2.Template(self.__t)

    def render(self,
               path: os.PathLike,
               params: Params,
               ):
        out = self.__template.render(load_duration=params.load_duration)

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
    "period": "10ms"
  }
}
    '''
    __template: jinja2.Template

    def __init__(self):
        self.__template = jinja2.Template(self.__t)

    def render(self,
               path: os.PathLike,
               params: Params,
               ):
        out = self.__template.render(
            unlimited=params.unlimited,
            rss_limit=params.rss_limit_str,
            coefficient=params.coefficient,
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
        try:
            # drop container if exists
            container = self.client.containers.get(container_name)
            container.remove(force=True)
        except docker.errors.NotFound:
            pass

        container = self.client.containers.run(
            name=container_name,
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

        container.stop()


def run_session(
        docker_client: DockerClient,
        server_config_renderer: ServerConfigRenderer,
        perf_config_renderer: PerfConfigRenderer,
        session: Session,
) -> Report:
    print(f">>> Start case: {session.params}")

    server_config_path = Path(session.dir_path, "server_config.json")
    server_config_renderer.render(path=server_config_path,
                                  params=session.params)

    perf_config_path = Path(session.dir_path, "perf_config.json")
    perf_config_renderer.render(path=perf_config_path,
                                params=session.params)

    # run test session within Docker container
    docker_client.execute(
        mem_limit=session.params.rss_limit_str,
        session_dir_path=session.dir_path,
    )

    # parse output
    tracker_path = Path(session.dir_path, 'tracker.csv')
    return Report.from_file(session=session, path=tracker_path)


def main():
    docker_client = DockerClient()
    server_config_renderer = ServerConfigRenderer()
    perf_config_renderer = PerfConfigRenderer()

    now = datetime.now()
    root_dir = Path('/tmp/allocator', f'allocator_{now.hour}{now.minute}{now.second}')

    sessions = make_sessions(root_dir)
    reports = [
        run_session(
            docker_client=docker_client,
            server_config_renderer=server_config_renderer,
            perf_config_renderer=perf_config_renderer,
            session=ss)
        for ss in sessions
    ]

    for report in reports:
        render.single_report(report)

    render.multiple_reports(reports, Path(root_dir, 'reports.png'))


if __name__ == '__main__':
    main()
