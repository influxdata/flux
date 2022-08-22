use lsp_types::notification::{
    DidChangeTextDocument, DidOpenTextDocument, Initialized, Notification,
};
use lsp_types::{
    ClientCapabilities, CompletionClientCapabilities, CompletionItemCapability, CompletionItemKind,
    CompletionItemKindCapability, CompletionParams, DidChangeConfigurationClientCapabilities,
    DidChangeTextDocumentParams, DidChangeWatchedFilesClientCapabilities,
    DidOpenTextDocumentParams, InitializeParams, InitializedParams, Position,
    TextDocumentClientCapabilities, TextDocumentContentChangeEvent, TextDocumentIdentifier,
    TextDocumentItem, TextDocumentPositionParams, TextDocumentSyncClientCapabilities, Url,
    VersionedTextDocumentIdentifier, WorkspaceClientCapabilities, WorkspaceEditClientCapabilities,
};
use std::fmt::Debug;
use std::process::{Child, Command, Stdio};
use tower_lsp::jsonrpc;

use crate::lsp_suggestion_helper::get_last_ident;
use crate::lsp_suggestion_helper::ExpType::Normal;
use lsp_types::request::{Completion, Initialize, Request};
use lsp_types::MarkupKind::PlainText;
use regex::Regex;
use std::sync::atomic::{AtomicUsize, Ordering};
use tower_lsp::jsonrpc::RequestBuilder;

pub fn add_headers(a: String) -> String {
    format!("Content-Length: {}\r\n\r\n{}", a.len(), a)
}
#[derive(Debug, PartialEq)]
pub enum LSPRequestType {
    DidOpen,
    DidChange,
    Initialize,
    Initialized,
    Completion,
}
#[allow(dead_code)]
#[derive(Debug)]
pub enum LSPError {
    InitError,
    InvalidRequestType,
    InvalidFormatting,
    InternalError,
}

impl From<serde_json::Error> for LSPError {
    fn from(_: serde_json::Error) -> Self {
        Self::InvalidFormatting
    }
}
static NEXT_REQ_ID: AtomicUsize = AtomicUsize::new(1);
pub fn formulate_request(
    request_type: &LSPRequestType,
    text: &str,
    _pos: usize,
) -> Result<String, LSPError> {
    let req_id = NEXT_REQ_ID.fetch_add(1, Ordering::SeqCst) as i64;
    let version = NEXT_REQ_ID.fetch_add(1, Ordering::SeqCst) as i64;

    //TODO: SPECIFY THINGS YOU ARE INTERESTED IN RECEIVING
    //leading question marks are optional
    //use insertion text
    //only want plaintext over snippets
    //completion item kinds
    //kind 5 may be
    //change request_type to an enum
    match request_type {
        LSPRequestType::Initialize => {
            let req: RequestBuilder = jsonrpc::Request::build(Initialize::METHOD)
                .params(
                    serde_json::to_value(InitializeParams {
                        process_id: None,
                        root_path: None,
                        root_uri: Option::from(Url::parse("file:///foo.flux").unwrap()),
                        initialization_options: None,
                        capabilities: ClientCapabilities {
                            workspace: Some(WorkspaceClientCapabilities {
                                apply_edit: Some(true),
                                workspace_edit: Some(WorkspaceEditClientCapabilities {
                                    document_changes: Some(true),
                                    resource_operations: None,
                                    failure_handling: None,
                                    normalizes_line_endings: None,
                                    change_annotation_support: None,
                                }),
                                did_change_configuration: Some(
                                    DidChangeConfigurationClientCapabilities {
                                        dynamic_registration: Some(true),
                                    },
                                ),
                                did_change_watched_files: Some(
                                    DidChangeWatchedFilesClientCapabilities {
                                        dynamic_registration: Some(true),
                                    },
                                ),
                                symbol: None,
                                execute_command: None,
                                workspace_folders: Some(false),
                                configuration: Some(true),
                                semantic_tokens: None,
                                code_lens: None,
                                file_operations: None,
                            }),
                            text_document: Some(TextDocumentClientCapabilities {
                                synchronization: Some(TextDocumentSyncClientCapabilities {
                                    dynamic_registration: Some(true),
                                    will_save: Some(true),
                                    will_save_wait_until: Some(true),
                                    did_save: Some(true),
                                }),
                                completion: Some(CompletionClientCapabilities {
                                    dynamic_registration: None,
                                    completion_item: Some(CompletionItemCapability {
                                        snippet_support: Some(false),
                                        commit_characters_support: None,
                                        documentation_format: Some(vec![PlainText]),
                                        deprecated_support: None,
                                        preselect_support: None,
                                        tag_support: None,
                                        insert_replace_support: None,
                                        resolve_support: None,
                                        insert_text_mode_support: None,
                                    }),
                                    completion_item_kind: Some(CompletionItemKindCapability {
                                        value_set: Some(vec![CompletionItemKind::TEXT]),
                                    }),
                                    context_support: None,
                                }),
                                hover: None,
                                signature_help: None,
                                references: None,
                                document_highlight: None,
                                document_symbol: None,
                                formatting: None,
                                range_formatting: None,
                                on_type_formatting: None,
                                declaration: None,
                                definition: None,
                                type_definition: None,
                                implementation: None,
                                code_action: None,
                                code_lens: None,
                                document_link: None,
                                color_provider: None,
                                rename: None,
                                publish_diagnostics: None,
                                folding_range: None,
                                selection_range: None,
                                linked_editing_range: None,
                                call_hierarchy: None,
                                semantic_tokens: None,
                                moniker: None,
                            }),
                            window: None,
                            general: None,
                            experimental: None,
                        },
                        trace: None,
                        workspace_folders: None,
                        client_info: None,
                        locale: None,
                    })
                    .unwrap(),
                )
                .id(req_id);
            Ok(add_headers(serde_json::to_string(
                &req.id(req_id).finish(),
            )?))
        }
        LSPRequestType::Initialized => {
            let req: RequestBuilder = jsonrpc::Request::build(Initialized::METHOD)
                .params(serde_json::to_value(InitializedParams {}).unwrap());
            Ok(add_headers(req.id(req_id).finish().to_string()))
        }

        LSPRequestType::DidOpen => {
            let req: RequestBuilder = jsonrpc::Request::build(DidOpenTextDocument::METHOD).params(
                serde_json::to_value(DidOpenTextDocumentParams {
                    text_document: TextDocumentItem {
                        uri: Url::parse("file:///foo.flux").unwrap(),
                        language_id: "flux".to_string(),
                        version: version as i32,
                        text: "".to_string(),
                    },
                })?,
            );
            let a = serde_json::to_value(req.id(req_id).finish())?;
            let headed = add_headers(serde_json::to_string(&a)?);
            Ok(headed)
        }
        LSPRequestType::DidChange => {
            let mut text_with_nl = String::from(text);
            text_with_nl.push('\n');
            let basic_change = vec![TextDocumentContentChangeEvent {
                range: None,
                range_length: None,
                text: text_with_nl,
            }];
            let req: RequestBuilder = jsonrpc::Request::build(DidChangeTextDocument::METHOD)
                .params(serde_json::to_value(DidChangeTextDocumentParams {
                    text_document: VersionedTextDocumentIdentifier {
                        uri: (Url::parse("file:///foo.flux").unwrap()),
                        version: version as i32,
                    },
                    content_changes: basic_change,
                })?);
            let a = serde_json::to_value(req.id(req_id).finish())?;
            let headed = add_headers(serde_json::to_string(&a)?);
            Ok(headed)
        }
        LSPRequestType::Completion => {
            let line_num = text.matches("\n").count();

            //FIXME: Cases that don't work x = da will not complete to date and will not autocomplete functions that are args

            let character = match super::super::lsp_suggestion_helper::add_one(text) {
                true => text.len() as u32 + 1,
                false => text.len() as u32,
            };

            // println!("here is the character {}", character);

            let req: RequestBuilder = jsonrpc::Request::build(Completion::METHOD).params(
                serde_json::to_value(CompletionParams {
                    text_document_position: TextDocumentPositionParams {
                        text_document: TextDocumentIdentifier {
                            uri: (Url::parse("file:///foo.flux").unwrap()),
                        },

                        position: Position {
                            line: line_num as u32,
                            character: character as u32,
                        },
                    },
                    work_done_progress_params: Default::default(),
                    partial_result_params: Default::default(),
                    context: Default::default(),
                })?,
            );
            let a = serde_json::to_value(req.id(req_id).finish())?;
            let headed = add_headers(serde_json::to_string(&a)?);
            Ok(headed)
        }
    }
}

pub fn start_lsp() -> Child {
    //step one: start the process
    let child = Command::new("flux-lsp")
        // .arg("-l")
        // .arg("./real_lsp_log/log.txt")
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .expect("failure to execute");
    child
}

pub fn characters_after_last(input: &str, charac: &str) -> Option<u32> {
    if input == "" || charac == "" {
        return None;
    }

    let last_occur = input.chars().position(|c| c == '\n');
    if let Some(occur) = last_occur {
        return Some((occur.abs_diff(input.len()) - 1) as u32);
    }
    None
}
