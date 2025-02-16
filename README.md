# event-driven

- Wrapeper asynq
- Producer(ctx, payload, consumer, opts)
- Producer(ctx, payload, ...consumer, ...opts)
- Consumer(ctx, payload, opts)

### MONITORAMENTO

- SERVER HTTP
    - SALVA TRANSACAO NO BANCO
    - DASHBOARD COM GRAFANA
    - ESPURGOS MENSAIS NO BANCO COM JOB DIÁRIO

### SAGA PATTERN - ORQUESTRADOR

- SERVER HTTP
    - OPTS
        - RETRY
        - TIMEOUT
        - SEQUENCE PAYLOADS (bool)
        - CONSUMERS

```go
var SagaInput = struct{
Consumers []struct{
Consumer func (ctx, payload, opts)
Retry int
Timeout int
}{
}          
}

```

- Producer(ctx, payload, r, opts)

p1 -> ok (SALVA TX) -> BANCDO AVISA APP PARA FAZER UP
p2 -> ok (SALVA TX) -> BANCDO AVISA APP PARA FAZER UP
p3 -> !ok (SALVA TX)  -> BANCDO AVISA APP PARA FAZER DOWN
p2_down -> ok (SALVA TX) -> BANCDO AVISA APP PARA FAZER DOWN
p1_down -> false (RETRY 5) (SALVA TX) -> ARCHIVED (SALVA TX)

como vou saber se o consumer deu certo?

- Se o consumer não retornar error é porque falhou e preciso tentar x vezes e depois caso não consiga fazer o down

http - GET auth -> RESPONSE ADDR REDIS
CONNECT CONSUMER