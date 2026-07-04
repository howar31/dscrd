package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Discord JSON error codes that get special treatment.
const (
	codeUnknownGuild   = 10004
	codeUnknownChannel = 10003
	codeUnknownMessage = 10008
	codeUnknownRole    = 10011
	codeUnknownUser    = 10013
	codeMissingAccess  = 50001
	codeCannotDMUser   = 50007
	codeMissingPerms   = 50013
	codeIndexNotReady  = 110000
)

// hints maps Discord error codes to actionable guidance appended to the error.
var hints = map[int]string{
	codeMissingAccess: "Missing Access: the bot is not in this server or cannot see this channel; check the invite (auth invite-url) and channel permissions, and confirm required intents are enabled in the Developer Portal (Bot -> Privileged Gateway Intents)",
	codeMissingPerms:  "Missing Permissions: grant the bot the required permission in Server Settings -> Roles",
	codeCannotDMUser:  "Cannot DM this user: bot DMs require a mutual server and the recipient's privacy settings must allow DMs from server members",
	codeIndexNotReady: "search index is still warming up; retry in a few seconds",
}

// APIError is a structured failure from a Discord REST call.
type APIError struct {
	Method  string // HTTP verb, e.g. "GET"
	Path    string // endpoint path, e.g. "/channels/123/messages"
	Status  int    // HTTP status code
	Code    int    // Discord JSON error code (0 if absent)
	Message string // Discord "message" field
	Hint    string // actionable guidance, when known
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	s := fmt.Sprintf("%s %s: %d %s", e.Method, e.Path, e.Status, msg)
	if e.Code != 0 {
		s += fmt.Sprintf(" (code %d)", e.Code)
	}
	if e.Hint != "" {
		s += "\nhint: " + e.Hint
	}
	return s
}

// ExitCode returns the process exit code for this error.
func (e *APIError) ExitCode() int {
	switch {
	case e.Status == http.StatusUnauthorized:
		return 3
	case e.Status == http.StatusNotFound,
		e.Code == codeUnknownGuild, e.Code == codeUnknownChannel,
		e.Code == codeUnknownMessage, e.Code == codeUnknownRole,
		e.Code == codeUnknownUser:
		return 4
	case e.Status == http.StatusTooManyRequests:
		return 5
	default:
		return 1
	}
}

// newAPIError builds an APIError from a non-2xx Discord response body.
func newAPIError(method, path string, status int, raw []byte) *APIError {
	e := &APIError{Method: method, Path: path, Status: status}
	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if json.Unmarshal(raw, &body) == nil {
		e.Code = body.Code
		e.Message = body.Message
	}
	e.Hint = hints[e.Code]
	return e
}
