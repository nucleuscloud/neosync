package transformers

//go:generate go run neosync_transformer_generator.go $GOPACKAGE
//go:generate go run neosync_transformer_list_generator.go $GOPACKAGE
//go:generate go run neosync_js_transformer_docs_generator.go ../../../../docs/docs/transformers/gen-javascript-transformer.md
//go:generate go run neosync_transformer_typescript_declaration_generator.go ../../../../frontend/apps/web/@types/neosync-transformers.d.ts
