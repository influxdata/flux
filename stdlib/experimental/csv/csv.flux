package csv

import c "csv"
import "experimental/http"

from = (url) => c.from(csv: string(v: http.get(url: url).body))
