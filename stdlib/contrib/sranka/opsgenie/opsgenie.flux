package opsgenie

import "http"
import "json"
import "strings"

// respondersToJSON converts an array of responder strings to JSON array that can be embedded into an alert message
builtin respondersToJSON : (v: [string]) => string

// `sendAlert` sends a message that creates an alert in Opsgenie. See https://docs.opsgenie.com/docs/alert-api#create-alert for details.
// `url`         - string - Opsgenie API URL. Defaults to "https://api.opsgenie.com/v2/alerts". 
// `apiKey`      - string - API Authorization key. 
// `message`     - string - Alert message text, at most 130 characters. 
// `alias`       - string - Opsgenie alias, at most 250 characters that are used to de-deduplicate alerts. Defaults to message. 
// `description` - string - Description field of an alert, at most 15000 characters. Optional. 
// `priority`    - string - "P1", "P2", "P3", "P4" or "P5". Defaults to "P3". 
// `responders`  - array  - Array of strings to identify responder teams or teams, a 'user:' prefix is used for users, 'teams:' prefix for teams. 
// `tags`        - array  - Array of string tags. Optional. 
// `entity`      - string - Entity of the alert, used to specify domain of the alert. Optional. 
// `actions`     - array  - Array of strings that specifies actions that will be available for the alert. 
// `details`     - string - Additional details of an alert, it must be a JSON-encoded map of key-value string pairs. 
// `visibleTo`   - array  - Arrays of teams and users that the alert will become visible to without sending any notification. Optional. 
sendAlert = (url="https://api.opsgenie.com/v2/alerts", apiKey, message, alias="", description="", priority="P3", responders=[], tags=[], entity="", actions=[], visibleTo=[], details="{}") => {
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": "GenieKey " + apiKey
    }
    cutEncode = (v, max, defV = "") => {
        v2 = if strings.strlen(v: v) != 0 then v else defV
        return if strings.strlen(v: v2) > max 
                then string(v: json.encode(v: "${strings.substring(v: v2, start: 0, end: max)}"))
                else string(v: json.encode(v:v2))
    }
    body = "{
\"message\": ${cutEncode(v:message,max:130)},
\"alias\": ${cutEncode(v:alias,max:512,defV: message)},
\"description\": ${cutEncode(v:description,max:15000)},
\"responders\": ${respondersToJSON(v:responders)},
\"visibleTo\": ${respondersToJSON(v:visibleTo)},
\"actions\": ${string(v: json.encode(v:actions))},
\"tags\": ${string(v: json.encode(v:tags))},
\"details\": ${details},
\"entity\": ${cutEncode(v:entity, max: 512)},
\"priority\": ${cutEncode(v:priority, max: 2)}
}"
    return http.post(headers: headers, url: url, data: bytes(v: body))
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send alerts to opsgenie for each table row.
// `url`         - string - Opsgenie API URL. Defaults to "https://api.opsgenie.com/v2/alerts". 
// `apiKey`      - string - API Authorization key. 
// `entity`      - string - Entity of the alert, used to specify domain of the alert. Optional. 
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with all properties defined in the `sendAlert` function arguments (except url, apiKey and entity).
endpoint = (url="https://api.opsgenie.com/v2/alerts", apiKey, entity = "") =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == sendAlert(
                    url: url,
                    apiKey: apiKey,
                    entity: entity,
                    message: obj.message,
                    alias: obj.alias,
                    description: obj.description,
                    priority: obj.priority,
                    responders: obj.responders,
                    tags: obj.tags,
                    actions: obj.actions,
                    visibleTo: obj.visibleTo,
                    details: obj.details,
                ) / 100)}
            })

