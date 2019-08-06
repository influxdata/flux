Notification endpoints will require credentials in order to communicate with the various external services.

We do not want users to include those credentials in plain text into Flux scripts. We need to provide a mechanism for users load secrets from a safe storage mechanism and provide those secrets to the endpoints.

**Backing Store:** : There are many backing store possibilities, which need not be discussed here.  Any store that provides secure access for key/value lookup may be used by providing an implementation of the KeyLookup  interface in the flux repository.  

**Fine Grained Authorization**:  keys will generally not be accessible in the flux language.  A secret-enabled function will accept meta-data about the secret (a key) and use it to look up the stored value in memory. The lookup will error if the key is not found or if the user is not authorized to use that key.  

# **Design**

## Internal Functions
A function that requires secret lookup must: 

- Be built-in to the language
- Be secrets-aware, meaning that it can acquire the injected secret service, scan its input for secret references, and look up the values for the given secret keys.  
- Not return secrets under any circumstances (e.g. in error messages) 

The external interface for secret lookup is: 
```Go
type SecretService interface {
// LoadSecret retrieves the secret value v found at key k 
// ctx is the current query context
    LoadSecret(ctx context.Context,k string) (string, error)
}
```

the internal function must be "secret aware" which means that it knows how to search its own input parameters for references to secrets, so that it can make the proper key-for-value substitution at the appropriate time (as close to when the key is used as possible)

for example, a http.post function may reference a token in the header object by having a sub-object with the form  `{secretKey: <string>}`
http.post(header: {"Content Type": "application-json", "ApplicationSecret": {secretKey: "FLUX_HTTP_POST_KEY"}}, url: "http://some/url/", data: json_enc)

## Flux Package
We will provide a `secrets` package that will help the user create an object with the required properties for secret lookup.  
A helper function is useful because we may wish to change the properties of a secret request over time, and we don't want users to have to update their code unless strictly needed.  

The vault package contains: 

- `secrets.get(key: "FLUX_SLACK_ID")`: returns an object with a single property: `{secretKey: "my-secret"}`

The function will provide all that's needed to do the lookup.  It may also be implemented to pre-authorize that the user can access the given key, so that errors occur earlier.  

Finally, a nice-to-have, but not required behavior would be to enforce that the token-lookup object be created by the vault.Get call.  That is, the user should get an error if they construct a request object by hand.  

A final implementation may look like: 
Go code: 
```Go
// calling flux:  
//     import "secret"
//     import "http"
//
//     tokAK = secret.get(key: "FLUX_SLACK_ID")
//     tokURL = secret.get(key: "FLUX_SLACK_URL")
//     // ASSUMPTION:  the docs for http.post say you can provide a token for the URL, 
//     // and also as a sub-object in the header
//     http.post(url: tok2, headers: {AuthenticationKey: tokAK}, data: {message: "hello"})
// in function call: 
var urlstr string
url, ok, err := args.GetObject("url")
if ok {
  urlstr = deps.SecretService.LoadSecret(ctx, url.Get("secretKey"))  
}
token, err := message.Get("token")

header, err := args.GetRequiredObject("header")

// hypothetical helper function
header = ReplaceSecrets(ctx, deps.SecretService, header)
```

