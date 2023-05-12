#!/bin/bash

f=`basename "$0"`

cat > tests/"$f".sb <<EOL
v1="abc"
print(v1)
   v2    =        "def"
print(v2)
v1 = "ghi"
print(v1)
v1 = 123
print(v1)
v2 = 123.456
print(v1, v2)
EOL

out=$(./shiba tests/"$f".sb)

expected=`cat << EOS
abc
def
ghi
123
123 123.456
EOS
`

assert $out $expected

rm -f tests/"$f".sb
