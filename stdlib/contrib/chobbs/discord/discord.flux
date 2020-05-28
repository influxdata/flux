package discord

import "http"
import "json"

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

  discordURL = "https://discordapp.com/api/webhooks/"
  return http.post(headers: headers, url: discordURL + webhookID + "/" + webhookToken, data: encode)
}
