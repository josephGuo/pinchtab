# Group 32: Dynamic iframe (inserted after load)

### 32.1 Wait for a late iframe, then interact with it
Navigate to `http://fixtures/iframe-dynamic.html`. The iframe is inserted dynamically after page load — wait for it to appear. Then scope into it, fill the input with "Late World", submit, and verify the inner result marker.

**Verify**: Scoped snapshot contains `IFRAME_INPUT_RECEIVED_LATE_WORLD`.

---

