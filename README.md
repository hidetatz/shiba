shiba is a programming language. Aims to be plain like Python, modern like Go/Rust.

Below code works.

```
import os

fd, errno := os.open("/home/hidetatz/shiba/main.go")
if errno != 0 {
    print("failed to open")
    return
}

result, errno := os.read(fd, 100)
if errno != 0 {
    print("failed to read")
    return
}

print(result)
```

For more code example, see [tests/](./tests/).

Installation:

1. Download the latest shiba release from [GitHub Release](https://github.com/hidetatz/shiba/releases/latest)
2. `tar zxf ./shibaX.X.X.linux_amd64.tar.gz`
3. `./shiba main.sb`
4. `./shiba` for REPL
5. `./shiba -h` for help

Author: [@hidetatz](https://github.com/hidetatz)

TODO (@hidetatz):
- struct and method
- error handling
- easy and simple concurrency like Go
- formatter
- package manager
