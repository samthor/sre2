#!/bin/bash
# Simple speed test for SRE2.

RE="(a|(b))+"
STR="aba"
RUNS=100000

CMD="./main -runs=$RUNS -re=$RE -s=$STR"

make
if [ $? != 0 ]; then
  exit $?
fi
echo
echo "RUNS=$RUNS RE=$RE STR=$STR"
echo

# On 2.53ghz Core 2 Duo: ~0.50
echo "==sre2 simple (fast, no submatches and uses bitset for states)"
time -p $CMD 2>/dev/null >/dev/null
echo

# On 2.53ghz Core 2 Duo: ~1.32
echo "==sre2 submatch (slow, abuse of gc for states)"
time -p $CMD -sub 2>/dev/null >/dev/null
echo

# NB: Go's regexp module has the same underlying code whether we care about
# submatches or not. This might change in the future.
# On 2.53ghz Core 2 Duo: ~0.87
echo "==go regexp (medium speed, always cares about submatches with good gc)"
time -p $CMD -m >/dev/null
echo

