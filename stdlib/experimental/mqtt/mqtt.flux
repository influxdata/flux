package mqtt

builtin to : ( <-tables: [A], broker: string, ?topic: string, ?message: string, ?qos: int, ?clientid: string, ?username: string, ?password: string, ?name: string, ?timeout: duration, ?timeColumn: string, ?tagColumns: [string], ?valueColumns: [string]) => [B] where A: Record, B: Record
