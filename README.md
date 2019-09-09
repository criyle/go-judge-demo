# go-sandbox-demo

A simple demo site for the [go-sandbox](https://github.com/criyle/go-sandbox), deployed on [heroku](https://go-judger.herokuapp.com).
Under development...

+   Frontend: Vue.js
+   Backend: GO
+   Judger Client: GO

## API

### Database Model

``` json
{
  "_id": "mongodb primary key",
  "language": "c++",
  "code": "<code>",
  "date": "<submit date @ unit epoch date (ms)",
  "updates": [
    {
      "status": "<current status (AC / TLE / MLE / OLE / JGF)>",
      "time": "<user time (ms)>",
      "memory": "<memory (kb)>",
      "date": "<update date @ unix epoch date (ms)>",
      "stdin": "<stdin>",
      "stdout": "<stdout>",
      "stderr": "<stderr>",
      "log": "<judger log>",
    },
  ],
}
```

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
  "language": "c++",
  "code": "<code>",
  "update": {
    "status": "<status>",
    "time": "<time>",
    "memory": "<memory>",
    "date": "<date>",
    "stdin": "<stdin>",
    "stdout": "<stdout>",
    "stderr": "<stderr>",
    "log": "<judger log>",
  },
}
```

Id is madatory. Others are optional.

### Judger WS

Send:

``` json
{
  "id": "<id>",
  "update": {
    "status": "<status>",
    "time": "<time>",
    "memory": "<memory>",
    "date": "<date>",
    "stdin": "<stdin>",
    "stdout": "<stdout>",
    "stderr": "<stderr>",
    "log": "<judger log>"
  },
}
```

Id is madatory. Others are optional.

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
