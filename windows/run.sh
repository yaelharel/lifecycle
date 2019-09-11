#/bin/bash
set -o errexit -o pipefail -o nounset

cd $(dirname "$BASH_SOURCE")

docker volume rm --force windows-layers
# relies on docker login already on the RDP
docker run --rm -v c:/Users/buildsvc/.docker:C:/Users/ContainerAdministrator/.docker -v windows-layers:c:/layers lifecycle-windows 'c:\cnb\lifecycle\detector.exe'
docker run --rm -v c:/Users/buildsvc/.docker:C:/Users/ContainerAdministrator/.docker -v windows-layers:c:/layers lifecycle-windows 'c:\cnb\lifecycle\analyzer.exe' ekcasey/app-windows
docker run --rm -v c:/Users/buildsvc/.docker:C:/Users/ContainerAdministrator/.docker -v windows-layers:c:/layers lifecycle-windows 'c:\cnb\lifecycle\builder.exe'
docker run --rm -v c:/Users/buildsvc/.docker:C:/Users/ContainerAdministrator/.docker -v windows-layers:c:/layers lifecycle-windows 'c:\cnb\lifecycle\exporter.exe' ekcasey/app-windows
