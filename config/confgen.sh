#!/bin/bash

declare -A confs
confs=(
    [%%AGENT_HTTP%%]=0.0.0.0:1988
    [%%AGGREGATOR_HTTP%%]=0.0.0.0:6055
    [%%GRAPH_HTTP%%]=0.0.0.0:6071
    [%%GRAPH_RPC%%]=0.0.0.0:6070
    [%%HBS_HTTP%%]=0.0.0.0:6031
    [%%HBS_RPC%%]=0.0.0.0:6030
    [%%JUDGE_HTTP%%]=0.0.0.0:6081
    [%%JUDGE_RPC%%]=0.0.0.0:6080
    [%%NODATA_HTTP%%]=0.0.0.0:6090
    [%%QUERY_HTTP%%]=0.0.0.0:9966
    [%%SENDER_HTTP%%]=0.0.0.0:6066
    [%%TASK_HTTP%%]=0.0.0.0:8002
    [%%TRANSFER_HTTP%%]=0.0.0.0:6060
    [%%TRANSFER_RPC%%]=0.0.0.0:8433
    [%%REDIS%%]=127.0.0.1:6379
    [%%MYSQL%%]="root:password@tcp(127.0.0.1:3306)"
)

configurer() {
    for i in "${!confs[@]}"
    do
        search=$i
        replace=${confs[$i]}
        # Note the "" after -i, needed in OS X
        find ./out/*/config/*.json -type f -exec sed -i "s/${search}/${replace}/g" {} \;
    done
}
configurer
