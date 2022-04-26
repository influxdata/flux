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
message = (url, title, text, summary="",mention="",button="") => {
    headers = {
        "Content-Type": "application/json",
    }
    // see https://docs.microsoft.com/en-us/outlook/actionable-messages/message-card-reference#card-fields
    // using string body, object cannot be used because '@' is an illegal character in the object property key
    summary2 = if summary == "" 
        then text 
        else summary
        
    shortSummary = if strings.strlen(v: summary2) > summaryCutoff 
        then "${strings.substring(v: summary2, start: 0, end: summaryCutoff)}..."
        else summary2
    body = "{   \"type\": \"message\",
                \"attachments\": [
                    {
                        \"contentType\": \"application/vnd.microsoft.card.adaptive\",
                        \"content\": {
                            \"$schema\": \"http://adaptivecards.io/schemas/adaptive-card.json\",
                            \"version\": \"1.3\",
                            \"type\": \"AdaptiveCard\",
                            \"body\": [
                                    {
                                    \"type\": \"TextBlock\",
                                    \"text\": ${string(v: json.encode(v:title))},
                                    \"weight\": \"bolder\",
                                    \"isSubtle\": false,
                                    \"size\"   : \"Large\"
                                    },
                                    {
                                    \"type\": \"TextBlock\",
                                    \"wrap\": true,
                                    \"text\": ${string(v: json.encode(v:text))}
                                    }              
                            ],
                            \"actions\": [
                                ${button}
                            ],
                            \"msteams\": {
                                \"entities\": [
                                   ${mention}
                                ]
                            }              
                        }
                    }   
                ]
            }"
    return http.post(headers: headers, url: url, data: bytes(v: body))
}

addMention = (name, id) => {
    mention = "{
                    \"type\": \"mention\",
                    \"text\": \"<at>${name}</at>\",
                    \"mentioned\": {
                        \"id\": ${string(v: json.encode(v: id))},
                        \"name\": \"${name}\"
                    }
                },"
    return mention
}

addButton = (type, title, url="") => {
    mention = "{
                \"type\": \"${type}\",
                \"title\": \"${title}\",
                \"url\": \"${url}\"
                },"
    return mention
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
                    summary: if exists obj.summary then obj.summary else "",
                    mention : if exists obj.mention then obj.mention else "", 
                    button: if exists obj.button then obj.button else ""
                ) / 100)}
            })
