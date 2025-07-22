#!/bin/bash

# Parse test results and create a summary for GitHub Actions

set -euo pipefail

JSON_FILE="${1:-test-results.json}"
OUTPUT_FILE="${2:-$GITHUB_STEP_SUMMARY}"

# Initialize counters
total_tests=0
passed_tests=0
failed_tests=0
skipped_tests=0

# Arrays to store failures
declare -a failures

# Parse JSON test results
while IFS= read -r line; do
    # Extract test information
    action=$(echo "$line" | jq -r '.Action // empty')
    package=$(echo "$line" | jq -r '.Package // empty')
    test=$(echo "$line" | jq -r '.Test // empty')
    output=$(echo "$line" | jq -r '.Output // empty')
    
    case "$action" in
        "pass")
            ((passed_tests++))
            ((total_tests++))
            ;;
        "fail")
            if [[ -n "$test" ]]; then
                ((failed_tests++))
                ((total_tests++))
                failures+=("$package - $test")
            fi
            ;;
        "skip")
            ((skipped_tests++))
            ((total_tests++))
            ;;
        "output")
            # Capture panic or error output
            if [[ "$output" =~ "panic:" ]] || [[ "$output" =~ "Error:" ]]; then
                failures+=("  └─ $output")
            fi
            ;;
    esac
done < <(jq -c '.' "$JSON_FILE" 2>/dev/null || echo '{}')

# Generate summary
{
    echo "## 📊 Test Results Summary"
    echo ""
    echo "| Metric | Count |"
    echo "|--------|-------|"
    echo "| Total Tests | $total_tests |"
    echo "| ✅ Passed | $passed_tests |"
    echo "| ❌ Failed | $failed_tests |"
    echo "| ⏭️ Skipped | $skipped_tests |"
    echo ""
    
    if [[ $failed_tests -gt 0 ]]; then
        echo "### ❌ Failed Tests"
        echo ""
        echo "<details>"
        echo "<summary>Click to expand failed test details</summary>"
        echo ""
        echo '```'
        for failure in "${failures[@]}"; do
            echo "$failure"
        done
        echo '```'
        echo "</details>"
        echo ""
    fi
    
    # Calculate pass rate
    if [[ $total_tests -gt 0 ]]; then
        pass_rate=$(( (passed_tests * 100) / total_tests ))
        echo "### 📈 Pass Rate: ${pass_rate}%"
        echo ""
        
        # Progress bar
        echo "<div align=\"center\">"
        echo ""
        echo '```'
        printf "["
        filled=$(( pass_rate / 2 ))
        for ((i=0; i<50; i++)); do
            if [[ $i -lt $filled ]]; then
                printf "█"
            else
                printf "░"
            fi
        done
        printf "] %d%%\n" "$pass_rate"
        echo '```'
        echo ""
        echo "</div>"
    fi
} >> "$OUTPUT_FILE"

# Exit with error if tests failed
if [[ $failed_tests -gt 0 ]]; then
    exit 1
fi