#!/bin/bash

make

./shiba > /dev/null 2>&1
ret=$?

if [ $ret -eq 0 ]; then
	echo "filename isn't passed but did not fail"
	exit 1
fi

echo "# test" > test.sb
./shiba test.sb
ret=$?
if [ $ret -ne 0 ]; then
	exit 1
fi

echo "all test passed!"
make clean
exit 0
