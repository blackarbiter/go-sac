#!/bin/zsh

scan_types=("DAST" "SCA" "SAST")
for _ in {1..100}; do
    json_payload=$(jq -n \
        --arg st "${scan_types[$((RANDOM%3))]}" \
        '{
            asset_id: "asset123",
            asset_type: "Domain",
            scan_type: $st,
            priority: 0,
            options: { depth: 3, timeout: 300 }
        }')

    echo "Sending: $json_payload"

    curl -X POST http://localhost:8088/api/v1/tasks/scan \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer 6DNZdF40SaBOOme7V3Ks9cZoj42" \
        -d "$json_payload"
done