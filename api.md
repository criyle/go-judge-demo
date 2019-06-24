## API

### Model

``` json
{
  "status": "AC / TLE / MLE / OLE / JGF",
  "time": 1234567,
  "memory": 123456,
  "date": 123456,
  "language": "c++",
  "code": "...",
  "stdin": "",
  "stdout": "",
  "stderr": "",
}
```

+ time in ms
+ memory in kb
+ date in unix epoch

### Client WS

Send:
``` json
{
  "language": "c++",
  "code" : "#include <iostream>\n ....",
}
```

Receive:
``` json
{
  "id": "<id>",
  "status": "<status>",
  "language": "<language>",
  "code": "<code>",
  "time": "<time>",
  "memory": "<memory>",
  "date": "<date>",
  "stdin": "<stdin>",
  "stdout": "<stdout>",
  "stderr": "<stderr>",
}
```

Id, status is madatory. Others are optional.

### Judger WS

Send:
``` json
{
  "id": "<id>",
  "status": "<status>",
  "time": "<time>",
  "memory": "<memory>",
  "date": "<date>",
  "stdin": "<stdin>",
  "stdout": "<stdout>",
  "stderr": "<stderr>",
}
```

Id, status is madatory. Others are optional.

Receive:
``` json
{
  "id": "<id>",
  "language": "c++",
  "code": "...",
}
```

Connect:
Include `Authorization: Token token` in the header when call for upgrade.
