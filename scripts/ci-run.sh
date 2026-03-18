#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:?Usage: $0 <ut|it|at|all|ci|lint|cover> [module-filter]}"
FILTER="${2:-}"

# 命令 → Makefile target 映射
case "$COMMAND" in
    ut)    TARGETS=("test") ;;
    it)    TARGETS=("test-integration") ;;
    at)    TARGETS=("test-at") ;;
    all)   TARGETS=("test" "test-integration") ;;
    ci)    TARGETS=("ci") ;;
    lint)  TARGETS=("lint") ;;
    cover) TARGETS=("test-cover") ;;
    *)     echo "Unknown command: $COMMAND"; exit 1 ;;
esac

# Go 模块
# 注：vega/mdl-data-model, vega/mdl-uniquery, vega/mdl-data-model-job 计划退场，不纳入
GO_MODULES=(
    context-loader/agent-retrieval
    ontology/ontology-manager
    ontology/ontology-query
    vega/vega-backend
    vega/vega-gateway-pro
    execution-factory/operator-integration
    dataflow/flow-automation
)

# Python AT 套件
PYTHON_MODULES=(
    execution-factory/tests
    dataflow/tests
)

PASSED=()
FAILED=()
SKIPPED=()

run_module() {
    local mod="$1" target="$2"

    if [[ -n "$FILTER" && "$mod" != "$FILTER"* ]]; then
        return
    fi

    if [[ ! -f "$mod/Makefile" ]]; then
        SKIPPED+=("$mod:Makefile missing (not yet onboarded to testing standard)")
        return
    fi

    # 跳过模块没有实现的可选 target
    if ! grep -q "^${target}:" "$mod/Makefile" 2>/dev/null; then
        SKIPPED+=("$mod:$target")
        return
    fi

    echo ""
    echo "━━━ $target: $mod ━━━"
    if make -C "$mod" "$target"; then
        PASSED+=("$mod:$target")
    else
        FAILED+=("$mod:$target")
    fi
}

ALL_MODULES=("${GO_MODULES[@]}" "${PYTHON_MODULES[@]}")

for target in "${TARGETS[@]}"; do
    for mod in "${ALL_MODULES[@]}"; do
        run_module "$mod" "$target"
    done
done

# 汇总报告
echo ""
echo "════════════════════════════════"
echo "  PASSED:  ${#PASSED[@]}"
echo "  FAILED:  ${#FAILED[@]}"
echo "  SKIPPED: ${#SKIPPED[@]}"
if [[ ${#FAILED[@]} -gt 0 ]]; then
    for m in "${FAILED[@]}"; do echo "    FAIL $m"; done
    echo "════════════════════════════════"
    exit 1
fi
echo "════════════════════════════════"
