#!/bin/bash
set -x

k8sconf="conf/aws.kubeconfig.yaml"
#k8sconf="configs/en119.kubeconfig.yaml"

memlimit=105
cpulimit=115
nameSpace="default"
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
options="$options --memLimit $memlimit "
options="$options --cpuLimit $cpulimit "

./podmove $options

