export LD_LIBRARY_PATH="${APPDIR}/usr/lib:${LD_LIBRARY_PATH}"
export ZAP="TRUE"

{{ python-executable }} -s ${APPDIR}/opt/python{{ python-version }}/bin/zap "$@"
