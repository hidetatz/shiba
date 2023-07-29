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

Author: [@hidetatz](https://github.com/hidetatz)

TODO (@hidetatz):
- struct and method
- error handling
- easy and simple concurrency like Go
- formatter
- installer
- package manager
