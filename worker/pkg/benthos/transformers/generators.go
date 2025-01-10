package transformers

//go:generate go run ../../../../tools/generators/neosync_transformer_generator/main.go $GOPACKAGE
//go:generate go run ../../../../tools/generators/neosync_transformer_list_generator/main.go $GOPACKAGE
//go:generate go run ../../../../tools/generators/neosync_js_transformer_docs_generator/main.go ../../../../docs/docs/transformers/gen-javascript-transformer.md
//go:generate go run ../../../../tools/generators/neosync_transformer_typescript_declaration_generator/main.go ../../../../frontend/apps/web/@types/neosync-transformers.d.ts
