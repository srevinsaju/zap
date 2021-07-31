#!/bin/bash

set -euxo pipefail

"$ZAP_BIN" --help
"$ZAP_BIN" install --help
"$ZAP_BIN" update --help
"$ZAP_BIN" search --help
"$ZAP_BIN" list --help
"$ZAP_BIN" init --help


"$ZAP_BIN" i --github --from=TheAssassin/pyuploadtool --silent pyuploadtool
"$ZAP_BIN" i --github --from=TheAssassin/pyuploadtool --update --silent pyuploadtool
"$ZAP_BIN" remove pyuploadtool
"$ZAP_BIN" list
