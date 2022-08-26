### Requirements to build the project
- The Flux LSP has to be in your path
- You have run the make file and have generated a main executable of the headless repl and compiled the rust code
- You are in the headless_repl directory


### Building the Project 
First run a `make` to compile and then ensure that you have `flux-lsp` in your path. Now you are free to `cargo run`.

### TODO
- The hint function currently sends an updated line of what has currently been inputted, but the request to the lsp and the updating of hints only occurs at the end of the function. This means that hints will be delayed by a single character.
- Adding multiline support back into Rustyline, it is possible and it does work.
- Allow flag passing to the Coordinator.
- Add a way to maintain variable state that can be shared with the lsp.
- Formatting for the flux output so the prompt ">>" is not displayed before the new line 
- Add a `make check` command in the makefile that will run `which "flux-lsp"` to see if the flux-lsp is in their path
- Using a json-rpc parsing library rather the `read_json_rpc` function that I have made.
- Add load flux file function back into the repl, easy fix
- Do not allow for argument suggestions to be made when deleting characters ex(date.t if you finished date.truncate( and deleted to there it will try to suggest t:)
