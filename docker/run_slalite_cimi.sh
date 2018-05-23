#!/usr/bin/env sh
if [ -n "$CIMI_URL" ]; then
    export SLA_CIMIURL=${CIMI_URL}
fi

echo "run_slalite.sh: SLA_CIMIURL=${SLA_CIMIURL:-}"

./SLALite "$@"
