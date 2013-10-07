#!/bin/bash
# Simple speed test for SRE2.

RE=".*(a|(b))+(#*).+"
STR="aba#hello"
RUNS=100000

CMD="go run *.go -runs=$RUNS -re=$RE -s=$STR"

if [ $? != 0 ]; then
  exit $?
fi
echo
echo "RUNS=$RUNS RE=$RE STR=$STR"
echo "CMD=$CMD"
echo

# July 2011 test machine is a late-2010, 13" MacBook Air.
# October 2013 test machine is a late-2009, 27" iMac.

# July 2011: on 2.13ghz Core 2 Duo: ~1.30
# October 2013: on 2.66ghz Intel Core i5: ~1.10
echo "==sre2 simple (fast, no submatches and uses bitset for states)"
time -p $CMD #2>/dev/null >/dev/null
echo

# July 2011: on 2.13ghz Core 2 Duo: ~2.00
# October 2013: on 2.66ghz Intel Core i5: ~1.92
echo "==sre2 submatch (slow, abuse of gc for states)"
time -p $CMD -sub #2>/dev/null >/dev/null
echo

# July 2011: on 2.13ghz Core 2 Duo: ~2.25
# October 2013: on 2.66ghz Intel Core i5: ~1.06
echo "==go regexp (probably very fast)"
time -p $CMD -sub -m #>/dev/null
echo

