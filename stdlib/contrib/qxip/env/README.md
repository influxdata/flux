# Env Package

The env package provides methods to access system ENV variables.

To prevent leaking undesired keys, the package requires using the `FLUX_` prefix.

## env.get
Retrive a key from system ENV variables.

Example:

```no_run
import "contrib/qxip/env"
env.get(key: "FLUX_KEY_NAME")
```

### Populate sensitive credentials with ENV variables
```no_run
import "sql"
import "contrib/qxip/env"

username = env.get(key: "FLUX_USERNAME")
password = env.get(key: "FLUX_PASSWORD")

sql.from(
    driverName: "postgres",
    dataSourceName: "postgresql://${username}:${password}@localhost",
    query: "SELECT * FROM example-table",
)
```


## Contact
- Author: Lorenzo Mangani
- Email: lorenzo.mangani@gmail.com
- Github: [@lmangani](https://github.com/lmangani)
- Website: [@qryn](https://qryn.dev)
