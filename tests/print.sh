#!/bin/bash

f=`basename "$0"`

cat > tests/"$f".sb <<EOL
v1="abc"
print(v1)
print("aaa", v1, 123)
EOL

out=$(./shiba tests/"$f".sb)

echo "$out"

rm -f tests/"$f".sb
