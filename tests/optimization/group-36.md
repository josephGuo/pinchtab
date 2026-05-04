# Group 36: Search results page (SERP)

### 36.1 Find a specific result
Navigate to `http://fixtures/serp.html`. The page has 6 search result cards. Extract just the third result's content using a scoped observation.

**Verify**: The scoped output contains `RESULT_3_TITLE` and `RESULT_3_SNIPPET_MARKER`.

### 36.2 Count all result cards
Extract the full page content to verify all six results are present in one pass. Note that Readability may trim repetitive card layouts.

**Verify**: Output contains all of `RESULT_1_TITLE` through `RESULT_6_TITLE` and the summary `SERP_RESULT_COUNT_6`.

---

