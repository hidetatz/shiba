#!/bin/bash

f=`basename "$0"`

cat > tests/"$f".sb <<EOL
# abc
    # aaa

  # def
EOL

out=$(./shiba tests/"$f".sb)

echo $out

rm -f tests/"$f".sb
