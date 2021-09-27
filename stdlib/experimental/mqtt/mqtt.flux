package mqtt


builtin to : (
    <-tables: [A],
    broker: string,
    ?topic: string,
    ?qos: int,
    ?retain: bool,
    ?clientid: string,
    ?username: string,
    ?password: string,
    ?name: string,
    ?timeout: duration,
    ?timeColumn: string,
    ?tagColumns: [string],
    ?valueColumns: [string],
) => [B] where
    A: Record,
    B: Record

builtin publish : (
    broker: string,
    topic: string,
    message: string,
    ?qos: int,
    ?retain: bool,
    ?clientid: string,
    ?username: string,
    ?password: string,
    ?timeout: duration,
) => bool
