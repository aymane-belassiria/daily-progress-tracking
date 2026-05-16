import test from "node:test";
import assert from "node:assert/strict";

import { resolveApiBaseUrl, toApiUrl } from "./api.js";

test("resolveApiBaseUrl trims a trailing slash", () => {
  assert.equal(resolveApiBaseUrl("https://api.example.com/"), "https://api.example.com");
});

test("resolveApiBaseUrl falls back to localhost", () => {
  assert.equal(resolveApiBaseUrl(""), "http://localhost:4000");
});

test("toApiUrl joins the normalized base URL with API paths", () => {
  assert.equal(toApiUrl("https://api.example.com/", "/api/health"), "https://api.example.com/api/health");
});
