# Feature Library

The feature library is a Go library to provide feature flagging capabilities to Go programs.
It is common in Go programs to introduce new functionality or performance improvements.
But, it is also a common problem that changes to code can potentially have unintended consequences or bugs that will sometimes reach customers.
In order to control the functionality of the code at runtime, some form of feature flag is used.
In its most basic form, a feature flag might just be some configuration option in the server configuration or an environment variable that is set on the deployment.

In some circumstances, more fine-controlled or responsive feature flagging is required.
This library is intended to provide this more fine-controlled and responsive feature flagging.

## Flags

A `Flag` is a definition of a feature flag.
It is represented by a human-readable name, a description of the flag, the programmatic key, and a default value.
It also contains metadata for whether the flag is intended to be exposed to an end-user, the expected lifetime of the flag (temporary or permanent), and a contact for the person, team, or other entity that added the feature flag.

These flags can be represented in YAML form as the following:

```yaml
- name: My Feature
  description: My feature is awesome
  key: myFeature
  default: false
  expose: true
  contact: My Name
  lifetime: temporary
```

The flag type is inferred from the default value.
In this case, the default value is a boolean so this is a boolean flag.
The `feature` command can be used to process the above YAML and to generate the code for a list of feature flags.

    $ mkdir feature/
    $ feature -in flags.yml -out feature/feature.go

From our backend code, we can check the feature flag by importing the package for the generated code and using the following code:

```go
if feature.MyFeature.Enabled(ctx) {
  // new code...
} else {
  // new code...
}
```

To toggle the flag from the default value, the `Flagger` is used.

## Flagger

The `Flagger` is the component that checks the value for a feature flag.
This interface will be implemented by the specific application depending on how it wants to expose feature flags.
A `Flagger` backend can be static or dynamic depending on the underlying implementation.
