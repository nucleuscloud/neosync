// This module is copied from https://github.com/connectrpc/validate-go
// It was copied due to it lacking updating and we started running into issues with the protovalidate version not being updateable
// This was fine, but broke the camels back when starting to use protoc-gen-connect-openapi
// That module is not available on buf.build yet so must be added as a go dependency
// Maybe we can remove this once validate-go received newer updates
package validate
