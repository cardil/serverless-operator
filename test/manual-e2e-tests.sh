#!/usr/bin/env bash

# shellcheck disable=SC1091,SC1090
source "$(dirname "${BASH_SOURCE[0]}")/lib.bash"

set -Eeuo pipefail

register_teardown || exit $?
scale_up_workers || exit $?
create_namespaces || exit $?
create_htpasswd_users && add_roles || exit $?

install_catalogsource || exit $?
logger.success 'Cluster prepared for manual testing.'
logger.info 'Press any key to continue...'
read -r -n 1
