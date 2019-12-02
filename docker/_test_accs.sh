#!/usr/bin/env bash
cd $(dirname $0)
. ./_params.sh

set -e

if [ -f test_accs.json ]; then
    TEST_ACCS_COUNT=$(wc -l test_accs.json)
    echo -e "\nFound ${TEST_ACCS_COUNT} test accounts.\n"
else
    echo -e "\nMake ${TEST_ACCS_COUNT} test accounts :\n"
    go run ../cmd/acc-gen -from=${TEST_ACCS_START} -count=${TEST_ACCS_COUNT} > test_accs.json
fi
