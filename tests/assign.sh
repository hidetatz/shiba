#!/bin/bash

f=`basename "$0"`

cat > tests/"$f".sb <<EOL
v1="abc"
   v2    =        "def"
v1 = "ghi"
v1 = 123
v2 = 123.456
EOL

out=$(./shiba tests/"$f".sb)

echo "$out"

rm -f tests/"$f".sb
