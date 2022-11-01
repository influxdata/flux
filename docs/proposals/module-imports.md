# Module Imports

## Summary

Modules are contained within registries.
Registries are configured through package attributes.
Multiple registries can be configured on a single package.
Registries only need to be configured on one file in the package and will apply to the entire package.

The attribute modifies how imports are resolved.
When the registry path is matched by an import, the package will be retrieved from the appropriate module in the registry.

The third parameter may be used to configure additional attributes for accessing the registry.
This is future work, but the additional attributes would be a record that can be used to configure things like authorization using values from the secret service.

## Syntax

The syntax would follow something like this:

```
@registry("fluxlang.dev", "https://fluxlang.dev/modules")
@registry("modules.local", "https://env.aws.cloud2.influxdata.com/modules/myorgid")
package main

import "date"
import "modules.local/my/pkg"
import "fluxlang.dev/leftpad"
```

The first import would reference the standard library as it matches no registry.
The second import would retrieve the package from the module registry of the organization in influx cloud.
The third import would retrieve the package from the global module registry theoretically located at `fluxlang.dev`.

There will be no global registry in the present, but this is just to show what it might look like if we did have one.

## Default registries

Default registries will be configured on a per-host environment basis.
The default flux library will not have any default registries.

Default registries will be limited to DNS-style names that resemble a DNS address with a period.
We do this to prevent conflicts with the standard library and to reduce the chance of conflicts with an existing DNS address.

For non-default registries, there's no restriction on the name.
This is because it's not a default so the writer can see the registry names.
We would recommend using the same standard of DNS-style names to represent a registry.

For influxdb/idpe, the proposed name is `modules.local` although a more suitable DNS name would also be appropriate.
The purpose of this registry is to import modules from your organization-level registry.

For the CLI, we can define a name such as `filesystem.local`, `includes.local`, or something else that makes sense given the context.

The main point is that default registries are determined by the host environment and not determined by flux itself.

## Unused ideas

### Import attributes

The ability to specify a registry by placing an attribute on an import was first considered.
The syntax would have looked like this:

```
package main

@registry("https://fluxlang.dev/modules")
import "leftpad"
```

The idea of this is that you specify the registry an import will come from.
In addition, to make this more intuitive, import groups were suggested.

```
package main

@registry("https://fluxlang.dev/modules")
import (
    "leftpad"
    "pagerduty_utils"
)
```

This recommended syntax would have allowed both imports to come from the same registry.

I believe we should abandon this idea.
It requires us to create a new place where attributes can be attached and the resulting structure is a bit hard to follow.
The information about the import is left to looking at the attribute.

If we were using this method exclusively, it would potentially be an option.
But, we need to include a way to import modules that doesn't require an annotation with defaults that work for user-friendliness.
Since we need a syntax like `modules.local/leftpad`, I don't believe there's a compelling reason to support both.

Removing this also reduces the scope of this change as it won't involve any additional changes to the AST structure or parsing logic.

### Mount name

It can be reminiscent of filesystems and HTTP frameworks to see how the `@registry` annotation acts.
It defines a module registry with a name and creates a mount point for any paths that reference that path.

If we are planning on having import attributes, it's inadvisable to have one import attribute that happens locally and one package attribute that happens globally.
Since the name of the import attribute was already `@registry` and the package level one acted similar to a mount point, we named the package level one `@mount`.

I think this would no longer be needed if only the package attribute exists and I think the name `@mount` would create additional confusion over the name `@registry`.
The name `@registry` is obvious that it relates to the module registry.
The name `@mount` is more about the action rather than what that action represents so I think it could be a confusing name.

For that reason, I changed the package attribute name to be `@registry` in this proposal.
