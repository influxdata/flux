package teams

import "http"
import "json"
import "strings"

// `summaryCutoff` is used 
option summaryCutoff = 70

// `message` sends a single message to Microsoft Teams via incoming web hook.
// `url` - string - incoming web hook URL
// `title` - string - Message card title.
// `text` - string - Message card text.
// `summary` - string - Message card summary, it can be an empty string to generate summary from text.
message = (url, title, text, summary="") => {
    headers = {
        "Content-Type": "application/json; charset=utf-8",
    }
    // see https://docs.microsoft.com/en-us/outlook/actionable-messages/message-card-reference#card-fields
    // using string body, object cannot be used because '@' is an illegal character in the object property key
    summary2 = if summary == "" 
        then text 
        else summary
    shortSummary = if strings.strlen(v: summary2) > summaryCutoff 
        then "${strings.substring(v: summary2, start: 0, end: summaryCutoff)}..."
        else summary2
    body = "{
\"@type\": \"MessageCard\",
\"@context\": \"http://schema.org/extensions\",
\"title\": ${string(v: json.encode(v:title))},
\"text\": ${string(v: json.encode(v:text))},
\"summary\": ${string(v:json.encode(v:shortSummary))}
}"
    return http.post(headers: headers, url: url, data: bytes(v: body))
}

// `endpoint` creates the endpoint for the Microsoft Teams external service.
// `url` - string - URL of the incoming web hook.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `title`, `text`, and `summary`, as defined in the `message` function arguments.
endpoint = (url) =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == message(
                    url: url,
                    title: obj.title,
                    text: obj.text,
                    summary: if exists obj.summary then obj.summary else ""
                ) / 100)}
            })
