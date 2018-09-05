# Atticus

## Overview

Atticus is a simple http dependency mocking server. The server can be setup with initial configuration and have responses altered at runtime via its control plane.

Atticus uses canned responses, these contain match criteria and templated response body and headers. Templates use the golang text templating syntax

## Matching

- URL: the url, using gorilla mux syntax
- Method: one method per canned response
- Header: headers required to match (planned feature)

## Responses

- Header:
- Body: only strings can be templated. (planning to support, template result type converter)

## Example

```json
{
    "match":{
        "method":"GET",
        "URL":"/example/{id}"
    },
    "template":{
        "body":{
            "some-key":"literal-value",
            "another": 12,
            "templated-value":"{{.Vars.id}} {{.Method}} {{.Header.user_agent}} {{.Header.customer_header}}"
        },
    }
}
```

Then make this request

```bash
curl http://localhost:10001/some/path/with/some-value -H "Custom-Header: one two"
```