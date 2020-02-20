# go-sandbox-demo

A simple demo site for the [go-sandbox](https://github.com/criyle/go-sandbox), deployed on [heroku](https://go-judger.herokuapp.com).
Under development...

+ Frontend: Vue.js
+ Backend: GO
+ Backend Refresher: Air
+ Judger Client: GO

Benchmark (docker desktop on MacOS): 4 concurrent thread -> 172 op/s

## API

### Database Model

Language:

``` json
{
  "name": "c++",
  "sourceFileName": "a.cc",
  "compileCmd": "g++ -o a a.cc",
  "executables": [ "a" ],
  "runCmd": "a",
}
```

Model:

``` json
{
  "_id": "primary key",
  "language": "<language>",
  "source": "<source code>",
  "date": "<submit date>",
  "status": "<current status (AC / TLE / MLE / OLE / JGF)>",
  "totalTime": "total time",
  "maxMemory": "max memory",
  "results": [
    {
      "time": "<user time (ms)>",
      "memory": "<memory (kb)>",
      "stdin": "<stdin>",
      "stdout": "<stdout>",
      "stderr": "<stderr>",
      "log": "<judger log>",
    },
  ],
}
```

### Rest API

#### POST /api/submit

Request:

```json
{
  "language": "<language>",
  "source": "<source code>",
}
```

Response:

```json
{
  "_id": "<_id>"
}
```

### Client WS

S -> C:

``` json
{
  "id": "<id>",
  "status": "<status>",
  "date": "<date>",
  "language": "language name",
}
```

### Judger WS

Include `Authorization: Token token` in the HTTP Header when call for upgrade.

J -> S:

Progress:

``` json
{
  "id": "<id>",
  "type": "progress",
  "status": "<status>",
  "date": "<date>",
  "language": "language name",
}
```

Finish:

``` json
{
  "id": "<id>",
  "type": "finish",
  "status": "<status>",
  "date": "<date>",
  "language": "language name",
  "results": [ "result" ],
}
```

S -> J:

``` json
{
  "id": "<id>",
  "language": "<language>",
  "source": "source",
}
```
