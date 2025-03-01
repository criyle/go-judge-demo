# go-judge-demo

A simple demo site for the [go-judge](https://github.com/criyle/go-judge), deployed on M1 Mac Mini docker desktop [site](https://goj.ac).
Under development...

Components:

- Frontend: Vue 3, Naive UI, monaco editor
- APIGateway: GO
- Backend: GO
- Judger Client: GO
- Dev Server Compiler: Air, Overmind

Tools:

- <https://github.com/air-verse/air>
- <https://github.com/DarthSim/overmind>

```bash
go install github.com/air-verse/air@latest
go install github.com/DarthSim/overmind@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## API Gateway

### Interface

- GET /api/submission?id=_id: Query history submissions
- POST /api/submit: Submit judge request
- WS /api/ws/judge: Broadcast judge updates
- WS /api/ws/shell: Interactive shell
- GET /: SPA HTML & JS -> /dist

## Backend

Token-based gRPC

- submission(id)
- submit(request)
- updates(): stream judge updates
- judge(): stream for judge client
- shell(): stream for interactive shell

default ports:

- gRPC: `:5081`
- metrics: `:5082`

## Judge Client

Connect to backend with judge()

-metrics: `:2112`

## Development

```bash
# front end
npm run dev
# apigateway 
# demoserver
# judger
overmind start -f Procfile.dev

# mongoDB
docker run -p 27017:27017 mongo
# exec server
air
```

## Docker build

```bash
docker build -t apigateway -f Dockerfile.apigateway .
docker build -t demoserver -f Dockerfile.demoserver .

docker build -t judger -f Dockerfile.judger .
docker build -t judger_exec -f Dockerfile.exec .
```

## Docker run

```bash
docker run --name mongo -d -p 27017:27017 mongo

docker run --name demo --link mongo -d -e MONGODB_URI=mongodb://mongo:27017/test -p 5081:5081 -p 5082:5082 demoserver

docker run --name apigateway --link demo -d -e DEMO_SERVER=demo:5081 -p 5000:5000 apigateway

docker run --name exec -d --privileged -e ES_ENABLE_GRPC=1 -e ES_ENABLE_METRICS=1 -e ES_ENABLE_DEBUG=1 -p 5052:5052 -p 5051:5051 -p 5050:5050 judger_exec

docker run --name judger --link exec --link demo -d -e DEMO_SERVER=demo:5081 -e EXEC_SERVER=exec:5051 -p 2112:2112 judger
```

## Data Model

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

Result:

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

### POST /api/submit

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
  "results": "results[]"
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
