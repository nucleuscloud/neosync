VALID_CONTEXTS=("kind-nuc-dev")

assert_context()
{
    if ! kubectl config current-context > /dev/null 2>&1; then
        echo "no cluster context found"
        exit 1
    fi
    context=$(kubectl config current-context)
    is_valid_context=false
    for valid_context in "${VALID_CONTEXTS[@]}"; do
        if [ "$valid_context" == "$context" ]; then
            is_valid_context=true
            break
        fi
    done
    if [ $is_valid_context = false ]; then
        echo "no valid context found: got ${context}, wanted one of ${VALID_CONTEXTS[*]}"
        exit 1
    fi
}
