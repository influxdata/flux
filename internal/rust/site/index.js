const js = import("@influxdata/parser");
js.then(js => {
  js.js_parse(`package foo
import "bar"

a = x`);
});
