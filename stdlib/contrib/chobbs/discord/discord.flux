package discord

import "http"
import "json"

option discordURL = "https://discordapp.com/api/webhooks/"
 
// `webhookToken` - string - the secure token of the webhook.
// `webhookID` - string - the ID of the webhook.
// `username` - string - username posting the message.
// `content` - string - the text to display in discord.
// `avatar_url` -  override the default avatar of the webhook.
send = (webhookToken, webhookID, username, content, avatar_url="") => {
  data = {
      username: username,
      content: content,
      avatar_url: avatar_url
    }

  headers = {
      "Content-Type": "application/json"
    }
  encode = json.encode(v:data)

  return http.post(headers: headers, url: discordURL + webhookID + "/" + webhookToken, data: encode)
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send messages to discord for each table row.
// `webhookToken` - string - the secure token of the webhook.
// `webhookID` - string - the ID of the webhook.
// `username` - string - username posting the message.
// `avatar_url` -  override the default avatar of the webhook.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `content`, as defined in the `send` function arguments.
endpoint = (webhookToken, webhookID, username, avatar_url="") =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == send(
                    webhookToken: webhookToken,
                    webhookID: webhookID,
                    username: username,
                    avatar_url: avatar_url,
                    content: obj.content,
                ) / 100)}
            })
