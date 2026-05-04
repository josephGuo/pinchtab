# Group 19: iFrame

### 19.1 Read iframe content
Navigate to `http://fixtures/iframe.html` and extract content from inside the embedded same-origin iframe.

**Verify**: The iframe's inner content includes `IFRAME_INNER_CONTENT_LOADED`.

### 19.2 Interact with iframe form
Scope into the embedded iframe, fill the input with "Hello World" and submit. Note that `text` does not pierce iframes — use a scoped observation instead. Reset frame scope when done.

**Verify**: Scoped snapshot contains `IFRAME_INPUT_RECEIVED_HELLO_WORLD`.

---

