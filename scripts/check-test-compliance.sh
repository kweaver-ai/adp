#!/usr/bin/env bash
set -euo pipefail

ERRORS=0

check() {
    local desc="$1" cmd="$2"
    if eval "$cmd" > /dev/null 2>&1; then
        echo "  PASS  $desc"
    else
        echo "  FAIL  $desc"
        ERRORS=$((ERRORS + 1))
    fi
}

# 注：vega/mdl-data-model, vega/mdl-uniquery, vega/mdl-data-model-job 计划退场，不纳入
GO_MODULES=(
    context-loader/agent-retrieval
    ontology/ontology-manager
    ontology/ontology-query
    vega/vega-backend
    vega/vega-gateway-pro
    execution-factory/operator-integration
    dataflow/flow-automation
    bkn/bkn-backend
    bkn/ontology-query
)

PYTHON_MODULES=(
    execution-factory/tests
    dataflow/tests
)

ALL_MODULES=("${GO_MODULES[@]}" "${PYTHON_MODULES[@]}")

for mod in "${ALL_MODULES[@]}"; do
    echo ""
    echo "=== $mod ==="
    if [[ ! -f "$mod/Makefile" ]]; then
        echo "  INFO  module is not yet onboarded to testing standard"
    fi
    check "Makefile exists" "test -f $mod/Makefile"
    check "make test target" "grep -q '^test:' $mod/Makefile 2>/dev/null"
    check "make lint target" "grep -q '^lint:' $mod/Makefile 2>/dev/null"
    check "make ci target" "grep -q '^ci:' $mod/Makefile 2>/dev/null"
    check "test-result/ in .gitignore" "grep -rq 'test-result' .gitignore $mod/.gitignore 2>/dev/null"

    if [[ -f "$mod/go.mod" ]] || [[ -f "$mod/server/go.mod" ]]; then
        check "make test-cover target (Go)" "grep -q '^test-cover:' $mod/Makefile 2>/dev/null"
        check "no golang/mock (deprecated)" \
            "! grep -rq --include='*.go' --include='go.mod' 'github.com/golang/mock' $mod/"
    fi
done

echo ""
echo "════════════════════════════════"
if [[ $ERRORS -gt 0 ]]; then
    echo "  $ERRORS issues found"
    exit 1
else
    echo "  All checks passed"
fi
echo "════════════════════════════════"
