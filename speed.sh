#!/bin/bash
# Simple speed test for SRE2.

RE="y"
STR="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxy"
RUNS=100000

CMD="./main -runs=$RUNS -re=$RE -s=$STR"

make
if [ $? != 0 ]; then
  exit $?
fi
echo
echo "RUNS=$RUNS RE=$RE STR=$STR"
echo "CMD=$CMD"
echo

# Test machine is a late-2010, 13" MacBook Air.
# On 2.13ghz Core 2 Duo: ~1.30
echo "==sre2 simple (fast, no submatches and uses bitset for states)"
time -p $CMD #2>/dev/null >/dev/null
echo

# On 2.13ghz Core 2 Duo: ~2.00
echo "==sre2 submatch (slow, abuse of gc for states)"
time -p $CMD -sub #2>/dev/null >/dev/null
echo

# NB: Go's regexp module has the same underlying code whether we care about
# submatches or not. This might change in the future.
# On 2.13ghz Core 2 Duo: ~2.25
echo "==go regexp (medium speed, always cares about submatches with good gc)"
time -p $CMD -sub -m #>/dev/null
echo

