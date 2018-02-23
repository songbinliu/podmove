#!/bin/bash
set -x

k8sconf="conf/aws.kubeconfig.yaml"
#k8sconf="configs/en119.kubeconfig.yaml"

nameSpace="default"
#podName="cpu-5-1,limit-mem-256-cpu-20,limit.mem-256-cpu-20,memory-256-2,memory-333,saturated.memory-256-2,under-memory-1024"
#podName="mem-600-group-b3kxg,mem-600-group-csvcs,mem-600-group-wqlm9,turbo-cpu-40"
podName="httpbin-7965799df7-l6nht"
master="ip-172-23-1-107.us-west-2.compute.internal"
slave1="ip-172-23-1-223.us-west-2.compute.internal"
slave2="ip-172-23-1-96.us-west-2.compute.internal"
nodeName=$slave2

options="$options --kubeConfig $k8sconf "
options="$options --v 3 "
options="$options --nameSpace $nameSpace"
options="$options --podName $podName "
options="$options --nodeName $nodeName "

./podmove $options

