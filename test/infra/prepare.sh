#!/usr/bin/bash
set -e
set -x

PROJECT_DIR="${GOPATH}/src/gitlab.stageoffice.ru/UCS-COMMON/memlimiter"

# 1. подготовка контейнера с сервисом аллокатор
ALLOCATOR_SRC_DIR="${PROJECT_DIR}/test/allocator"
ALLOCATOR_DST_DIR="${PROJECT_DIR}/test/infra/allocator"
pushd "${ALLOCATOR_SRC_DIR}"
go build -o ucs-allocator
cp ucs-allocator "${ALLOCATOR_DST_DIR}"
cp ./server/config.json "${ALLOCATOR_DST_DIR}"
popd

# 2. подготовка инфраструктуры Grafana

# не у каждого человека есть репозиторий с Grafana
GRAFANA_SRC_DIR="${GOPATH}/src/gitlab.stageoffice.ru/mdc-devops/ansible_collections/nct.monitoring"

if [ -d "${GRAFANA_SRC_DIR}" ]
then
  pushd "${GRAFANA_SRC_DIR}"
#  git checkout master
#  git pull origin master
else
  mkdir -p "${GRAFANA_SRC_DIR}"
  pushd "${GRAFANA_SRC_DIR}"
  git init
  git remote add origin git@gitlab.stageoffice.ru:mdc-devops/ansible_collections/nct.monitoring.git
  git fetch
  git checkout origin/master -ft
fi

popd