#!/bin/sh

function report_sigpipe() {
    echo "got SIGPIPE, exiting..." > /var/trap.log;
    exit 1;
}

trap report_sigpipe SIGPIPE;

while sleep 1
do
    echo "time is $(date)";
done