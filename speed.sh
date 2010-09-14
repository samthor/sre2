#!/bin/bash
# Simple speed test for SRE2.

make
if [ $? != 0 ]; then
  exit $?
fi
echo

echo "==sre2 simple (fast, no submatches and uses bitset for states)"
time -p ./main 2>/dev/null >/dev/null
echo

echo "==sre2 submatch (slow, abuse of gc for states)"
time -p ./main -sub 2>/dev/null >/dev/null
echo

# NB: Go's regexp module has the same underlying code whether we care about
# submatches or not. This might change in the future.
echo "==go regexp (medium speed, always cares about submatches with good gc)"
time -p ./main -m >/dev/null
echo

