#!/bin/bash
set -e

OUT=pub.rclone.org:integration-tests/
TMP=/tmp/integration-test-listing
REMOVE=/tmp/integration-test-listing-to-remove
INCLUDE=/tmp/integration-test-listing-include
KEEP=30

echo $(date -Is) Tidy the integration test folder ${OUT}

rclone lsf --dirs-only --dir-slash=false ${OUT} | grep -P "^\d\d\d\d-\d\d-\d\d-\d\d\d\d\d\d$" | sort > ${TMP}

COUNT=$(wc -l ${TMP} | cut -d' ' -f1)
if [[ ${COUNT} > ${KEEP} ]]; then
    echo $(date -Is) Have ${COUNT} directories but only want ${KEEP}
    head -n -${KEEP} ${TMP} > ${REMOVE}
    while IFS= read -r line; do
	echo "/${line}/**"
    done < ${REMOVE} > ${INCLUDE}
    echo $(date -Is) Removing $(wc -l ${INCLUDE}| cut -d' ' -f1) directories
    rclone delete --stats-log-level NOTICE --stats 60s --checkers 16 --transfers 16 --include-from ${INCLUDE} ${OUT}
fi

echo $(date -Is) Done
